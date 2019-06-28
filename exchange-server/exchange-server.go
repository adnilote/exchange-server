package main

import (
	"container/heap"
	"fmt"
	readOHLCV "hakaton-2018-2-2-msu/exchange-server/data"
	"hakaton-2018-2-2-msu/exchange-server/proto"
	"math"
	"sync"
	"time"
)

var tools []string = []string{"IMOEX", "SPFB.RTS", "USD000UTSTOM"}

type Brokers struct {
	mu              sync.RWMutex
	b               map[int64]*Broker
	numResListener  int
	numStatListener int
}

type Broker struct {
	statMu     sync.RWMutex
	resMu      sync.RWMutex
	statStream exchange.Exchange_StatisticServer
	resStream  exchange.Exchange_ResultsServer
	close      chan struct{}
	resStarted bool
}

func (b *Brokers) addStatListener(id exchange.BrokerID, stream exchange.Exchange_StatisticServer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.b[id.ID]; ok {
		b.b[id.ID].statStream = stream
	} else {
		b.b[id.ID] = &Broker{
			statStream: stream,
			close:      make(chan struct{}),
		}
	}
	b.numStatListener++
}

func (b *Brokers) addResListener(id *exchange.BrokerID, stream exchange.Exchange_ResultsServer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.b[id.ID]; ok {
		b.b[id.ID].resStream = stream
	}
	b.b[id.ID] = &Broker{
		resStream: stream,
		close:     make(chan struct{}),
	}

}

// send sends OHLCV to all brokers
func (b *Brokers) sendOHLCV(ohlcv *exchange.OHLCV) {
	// todo paralel send to brokers
	b.mu.RLock()
	defer b.mu.RUnlock()
	wg := &sync.WaitGroup{}
	for _, broker := range b.b {
		if broker.statStream == nil {
			continue
		}
		wg.Add(1)
		go func(broker *Broker, wg *sync.WaitGroup) {
			defer wg.Done()
			broker.statMu.RLock()
			defer broker.statMu.RUnlock()

			stream := broker.statStream

			ctx := stream.Context()
			select {
			case <-ctx.Done():
				broker.statStream = nil
				broker.close <- struct{}{}
				return
			default:
			}

			err := stream.Send(ohlcv)
			if err != nil {
				fmt.Println("Send error: ", err)
				return
			}
		}(broker, wg)

	}
	wg.Wait()
}

// sendRes send deal related to broker
func (b *Brokers) sendRes(res *exchange.Deal) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	wg := &sync.WaitGroup{}

	for idBroker, broker := range b.b {
		if broker.resStream == nil {
			continue
		}

		wg.Add(1)
		func(wg *sync.WaitGroup) {
			broker.resMu.RLock()
			defer broker.resMu.RUnlock()
			defer wg.Done()
			stream := broker.resStream
			ctx := stream.Context()
			select {
			case <-ctx.Done():
				broker.resStream = nil
				broker.close <- struct{}{}
				return
			default:
			}

			if int64(res.BrokerID) == idBroker {
				err := stream.Send(res)
				if err != nil {
					fmt.Println("Send error: ", err)
				}
			}
		}(wg)

	}
	wg.Wait()
}

type ExchangeServer struct {
	brokers     *Brokers
	dataSources map[string]chan exchange.OHLCV
	mu          sync.RWMutex // for results
	results     []exchange.Deal
	tools       []string
	stocks      map[string]*Stock
	res         chan *exchange.Deal
	statExit    chan struct{} // to stop reading from file
}

// Stock stores buy and sell list of deals
type Stock struct {
	buyMu  sync.RWMutex
	sellMu sync.RWMutex
	buy    BuyQuery
	sell   SellQuery
}

// NewExchangeServer initiates exchange Server.
// Use start() to start it working
func NewExchangeServer() *ExchangeServer {
	data := make(map[string]chan exchange.OHLCV, len(tools))
	stocks := make(map[string]*Stock)

	ex := ExchangeServer{
		dataSources: data,
		brokers:     &Brokers{b: make(map[int64]*Broker)},
		tools:       tools,
		statExit:    make(chan struct{}),
		res:         make(chan *exchange.Deal),
	}

	for _, name := range tools {
		data[name] = make(chan exchange.OHLCV, 0)
		go readOHLCV.ReadPrices(name, data[name], ex.statExit)

		queryBuy := Query{}
		querySell := Query{}
		stock := &Stock{buy: BuyQuery{Query: &queryBuy}, sell: SellQuery{Query: &querySell}}
		stocks[name] = stock
		heap.Init(&stocks[name].buy)
		heap.Init(&stocks[name].sell)
	}
	ex.stocks = stocks

	go ex.startStatistic()
	go ex.startResults()

	fmt.Println("stock starts")
	for _, tool := range ex.tools {
		go ex.workTool(tool)
	}
	return &ex
}

func (ex *ExchangeServer) startStatistic() {
	fmt.Println("Statisitc starts")
	for _, source := range ex.dataSources {
		for ohlcv := range source {
			ex.brokers.sendOHLCV(&ohlcv)
		}
	}
}
func (ex *ExchangeServer) startResults() {
	fmt.Println("Results starts")
	for res := range ex.res {
		ex.brokers.sendRes(res)
	}
}

// workTool matches sell and buy stock.
// If sell price is less than buy price,
// trade is done - partialy or fully.
// results are saved in ex.results
func (ex *ExchangeServer) workTool(tool string) {
	fmt.Println("WorkerTool starts " + tool)
	var bestSellPrice, bestBuyPrice float32
	bestBuyPrice, bestSellPrice = 0, math.MaxFloat32

	for {
		// if bestPrices didnt change, do nothing
		// Get is less expensive than Pop, that's why
		// we dont pop deals and push them back in loop,
		// but pop it only when bestPrices change
		ex.stocks[tool].buyMu.RLock()
		curBuyPr := ex.stocks[tool].buy.GetBestPrice()
		ex.stocks[tool].buyMu.RUnlock()

		ex.stocks[tool].sellMu.RLock()
		curSellPr := ex.stocks[tool].sell.GetBestPrice()
		ex.stocks[tool].sellMu.RUnlock()

		if bestBuyPrice >= curBuyPr && bestSellPrice <= curSellPr {
			time.Sleep(time.Millisecond * 10)
			continue
		}

		// Pop deals from heaps
		fmt.Println("WorkerTool continues " + tool)
		fmt.Println(ex.stocks[tool].buy.Query.data)
		fmt.Println(ex.stocks[tool].sell.Query.data)

		sellDeal, err := ex.PopSellDeal(tool)
		if err != nil {
			if bestBuyPrice < curBuyPr {
				bestBuyPrice = curBuyPr
			} else {
				bestSellPrice = curSellPr
			}
			continue
		}
		bestSellPrice = sellDeal.GetPrice()

		buyDeal, err := ex.PopBuyDeal(tool)
		if err != nil { // if no buy deal, push sellDeal back
			ex.mu.Lock()
			heap.Push(&ex.stocks[tool].sell, &sellDeal)
			ex.mu.Unlock()
			if bestBuyPrice < curBuyPr {
				bestBuyPrice = curBuyPr
			} else {
				bestSellPrice = curSellPr
			}
			continue
		}
		bestBuyPrice = buyDeal.GetPrice()

		if buyDeal.GetPrice() >= sellDeal.GetPrice() { // todo write results to db
			// Trade success
			fmt.Println("Trade success!!!!")

			buyDeal.Price = sellDeal.GetPrice()

			difAmount := sellDeal.GetAmount() + buyDeal.GetAmount()
			switch {
			case difAmount == 0: // Full trade
				ex.results = append(ex.results, *sellDeal)
				ex.results = append(ex.results, *buyDeal)
				bestSellPrice = math.MaxFloat32
				bestBuyPrice = 0

			case difAmount < 0: // Partial trade
				// sellDeal done
				ex.results = append(ex.results, *sellDeal)
				bestSellPrice = math.MaxFloat32

				//save partial buy and push back
				saveb := *buyDeal
				buyDeal.Amount = 0 - sellDeal.GetAmount()
				buyDeal.Partial = true
				ex.results = append(ex.results, *buyDeal)

				saveb.Amount += sellDeal.GetAmount()
				ex.mu.Lock()
				heap.Push(&ex.stocks[tool].buy, &saveb)
				ex.mu.Unlock()

			case difAmount > 0: // Partial trade
				// buyDeal done
				ex.results = append(ex.results, *buyDeal)
				bestBuyPrice = 0

				// save partial sellDeal
				saves := *sellDeal
				sellDeal.Amount = 0 - buyDeal.GetAmount()
				sellDeal.Partial = true
				ex.results = append(ex.results, *sellDeal)

				saves.Amount += buyDeal.GetAmount()
				ex.mu.Lock()
				heap.Push(&ex.stocks[tool].sell, &saves)
				ex.mu.Unlock()

			}

			// inform listeners
			ex.brokers.mu.RLock()
			brNum := ex.brokers.numResListener
			ex.brokers.mu.RUnlock()

			if brNum != 0 {
				ex.res <- sellDeal
				ex.res <- buyDeal
			}
		}

	}
}

// PopSellDeal pop Deal with min Price and min Time from heap.
// Return error if heap is empty.
func (ex *ExchangeServer) PopSellDeal(tool string) (*exchange.Deal, error) {
	if ex.stocks[tool].sell.Len() == 0 {
		return &exchange.Deal{}, fmt.Errorf("Sell stock %s is empty", tool)
	}

	ex.stocks[tool].sellMu.Lock()
	sellDeal := heap.Pop(&ex.stocks[tool].sell)
	ex.stocks[tool].sellMu.Unlock()

	return sellDeal.(*exchange.Deal), nil
}

// PopBuyDeal pop Deal with max Price and min Time from heap.
// Return error if heap is empty
func (ex *ExchangeServer) PopBuyDeal(tool string) (*exchange.Deal, error) {
	if ex.stocks[tool].buy.Len() == 0 {
		return &exchange.Deal{}, fmt.Errorf("Buy stock %s is empty", tool)
	}

	ex.stocks[tool].buyMu.Lock()
	buyDeal := heap.Pop(&ex.stocks[tool].buy)
	ex.stocks[tool].buyMu.Unlock()

	return buyDeal.(*exchange.Deal), nil
}
