package main

import (
	"container/heap"
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"hakaton-2018-2-2-msu/exchange-server/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// поток ценовых данных от биржи к брокеру
// мы каждую секнуду будем получать отсюда событие с ценами, которые броке аггрегирует у себя в минуты и показывает клиентам
// устанавливается 1 раз брокером
func (ex *ExchangeServer) Statistic(id *exchange.BrokerID, stream exchange.Exchange_StatisticServer) error {
	ex.brokers.addStatListener(*id, stream)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case <-ex.brokers.b[id.ID].close:
				return
			}
		}
	}(wg)
	wg.Wait()
	return nil
}

// отправка на биржу заявки от брокера
func (ex *ExchangeServer) Create(ctx context.Context, deal *exchange.Deal) (*exchange.DealID, error) {
	tool := deal.Ticker

	id := uuid.New().ID()
	deal.ID = int64(id)
	if deal.BrokerID == 0 {
		deal.BrokerID = int32(uuid.New().ID())
	}
	if deal.Amount < 0 {
		ex.stocks[tool].buyMu.Lock()
		heap.Push(&ex.stocks[tool].buy, deal)
		ex.stocks[tool].buyMu.Unlock()
	} else {
		ex.stocks[tool].sellMu.Lock()
		heap.Push(&ex.stocks[tool].sell, deal)
		ex.stocks[tool].sellMu.Unlock()
	}

	res := exchange.DealID{
		ID:       deal.ID,
		BrokerID: int64(deal.BrokerID),
	}

	return &res, nil
}

// отмена заявки
func (ex *ExchangeServer) Cancel(ctx context.Context, dealID *exchange.DealID) (*exchange.CancelResult, error) {
	for _, stock := range ex.stocks {
		i := stock.buy.Find(*dealID)
		if i != -1 {
			heap.Remove(stock.buy, i)
			return &exchange.CancelResult{Success: true}, nil
		}

		i = stock.sell.Find(*dealID)
		if i != -1 {
			heap.Remove(stock.sell, i)
			return &exchange.CancelResult{Success: true}, nil
		}

	}
	return &exchange.CancelResult{Success: false}, nil
}

// исполнение заявок от биржи к брокеру
// устанавливается 1 раз брокером и при исполнении какой-то заявки
func (ex *ExchangeServer) Results(idBroker *exchange.BrokerID, stream exchange.Exchange_ResultsServer) error {

	ex.brokers.addResListener(idBroker, stream)

	if !ex.brokers.b[idBroker.ID].resStarted {
		ex.brokers.numResListener++

		// get old results
		ex.mu.RLock()
		for i := 0; i < len(ex.results); i++ { //todo add bd
			ex.mu.RLock()
			resID := int64(ex.results[i].BrokerID)
			ex.mu.RUnlock()

			if resID == idBroker.ID {

				err := stream.Send(&ex.results[i])
				if err != nil {
					fmt.Println("error in send")
					ex.mu.RUnlock()
					return err
				}
			}
		}
		ex.mu.RUnlock()
		ex.brokers.b[idBroker.ID].resStarted = true
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case <-ex.brokers.b[idBroker.ID].close:
				return
			}
		}
	}(wg)
	wg.Wait()

	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatalln("cant listen port", err)
	}

	server := grpc.NewServer()

	exchange.RegisterExchangeServer(server, NewExchangeServer())

	fmt.Println("starting server at :8090")
	err = server.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
