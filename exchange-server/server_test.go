package main

import (
	"context"
	exchange "hakaton-2018-2-2-msu/exchange-server/proto"
	"io"
	"log"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var lis *bufconn.Listener

const bufSize = 1024 * 1024

var ex *ExchangeServer
var client exchange.ExchangeClient
var ctx context.Context
var wg *sync.WaitGroup
var myBrokerId *exchange.BrokerID

func init() {
	wg = &sync.WaitGroup{}

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	ex = NewExchangeServer()
	exchange.RegisterExchangeServer(s, ex)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
		ex = NewExchangeServer()
	}()

	ctx = context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}

	client = exchange.NewExchangeClient(conn)
	myBrokerId = &exchange.BrokerID{ID: 1}
}

func bufDialer(string, time.Duration) (net.Conn, error) {
	return lis.Dial()
}

// test first statistics
func TestStatistic(t *testing.T) {

	cases := []exchange.OHLCV{
		exchange.OHLCV{
			ID:       0,
			Time:     1526637601,
			Interval: 1,
			Open:     2323.75,
			High:     2323.75,
			Low:      2323.75,
			Close:    2323.75,
			Volume:   2639122,
			Ticker:   "IMOEX",
		},
		exchange.OHLCV{
			ID:       0,
			Time:     1526637600,
			Interval: 1,
			Open:     62.14,
			High:     62.1875,
			Low:      62.14,
			Close:    62.1875,
			Volume:   4000,
			Ticker:   "USD000UTSTOM",
		},
		exchange.OHLCV{
			ID:       0,
			Time:     1526637600,
			Interval: 1,
			Open:     117230,
			High:     117350,
			Low:      117230,
			Close:    117350,
			Volume:   136,
			Ticker:   "SPFB.RTS",
		},
	}
	stream, err := client.Statistic(ctx, myBrokerId)
	if err != nil {
		t.Error("Statisitcs return err: ", err)
	}

	var result *exchange.OHLCV

	for i := 0; i < 6; i++ {
		// receive data from stream
		result, err = stream.Recv()
		if err == io.EOF {
			t.Error("stream closed from server side")
			return
		}
		if err != nil {
			t.Errorf("receive error %v", err)
			continue
		}

		for _, item := range cases {
			if result.GetTicker() == item.Ticker && result.GetID() == 0 {
				if !cmp.Equal(result, &item) {
					t.Error("expected", &item, "have", result)
				}
				break
			}
		}
	}

}

// dealIds returns from create()
var dealIDs []*exchange.DealID

func TestCreate(t *testing.T) {

	dealIDs = []*exchange.DealID{}
	cases := []*exchange.Deal{
		&exchange.Deal{
			Ticker: "SPFB.RTS",
			Amount: -100,
			Price:  11,
			Time:   10000,
		},
		&exchange.Deal{
			Ticker: "SPFB.RTS",
			Amount: 100,
			Price:  11,
			Time:   10000,
		},
	}

	for _, deal := range cases {
		dealID, err := client.Create(ctx, deal)
		if err != nil {
			t.Error("Statisitcs return err: ", err)
		}
		dealIDs = append(dealIDs, dealID)

		foundB := ex.stocks[deal.Ticker].buy.Find(*dealID)
		if foundB == -1 {
			foundS := ex.stocks[deal.Ticker].sell.Find(*dealID)
			if foundS == -1 {
				t.Error("Create fails")
			}
		}

	}
}
func TestCancel(t *testing.T) {

	for _, dealID := range dealIDs {
		res, err := client.Cancel(ctx, dealID)
		if err != nil {
			t.Error(err)
		}
		if !res.GetSuccess() {
			t.Error("Result must be true, got false")
		}
	}
}
func TestFakeCancel(t *testing.T) {

	dealID := &exchange.DealID{
		ID:       456456,
		BrokerID: 123132,
	}
	res, err := client.Cancel(ctx, dealID)
	if err != nil {
		t.Error(err)
	}
	if res.GetSuccess() {
		t.Error("Result must be false, got true")
	}
}

// test results and trade matching full trade and partial trades
func TestResults(t *testing.T) {
	deals := []struct {
		deal    *exchange.Deal
		partial bool
		dealID  *exchange.DealID
	}{
		{
			deal: &exchange.Deal{
				Ticker:   "SPFB.RTS",
				Amount:   -100,
				Price:    11,
				Time:     10000,
				BrokerID: int32(myBrokerId.GetID()),
			},
			partial: false,
		},

		{
			deal: &exchange.Deal{
				Ticker:   "SPFB.RTS",
				Amount:   100,
				Price:    11,
				Time:     10000,
				BrokerID: int32(myBrokerId.GetID()),
			},
			partial: false,
		},
		{
			deal: &exchange.Deal{
				Ticker:   "IMOEX",
				Amount:   -100,
				Price:    11,
				Time:     10000,
				BrokerID: int32(myBrokerId.GetID()),
			},
			partial: true,
		},
		{
			deal: &exchange.Deal{
				Ticker:   "IMOEX",
				Amount:   10,
				Price:    11,
				Time:     10000,
				BrokerID: int32(myBrokerId.GetID()),
			},
			partial: false,
		},
		{
			deal: &exchange.Deal{
				Ticker:   "SPFB.RTS",
				Amount:   -10,
				Price:    11,
				Time:     10000,
				BrokerID: int32(myBrokerId.GetID()),
			},
			partial: false,
		},

		{
			deal: &exchange.Deal{
				Ticker:   "SPFB.RTS",
				Amount:   100,
				Price:    11,
				Time:     10000,
				BrokerID: int32(myBrokerId.GetID()),
			},
			partial: true,
		},
	}

	for i, deal := range deals {
		dealID, err := client.Create(ctx, deal.deal)
		if err != nil {
			t.Error("Create return err: ", err)
		}
		deals[i].dealID = dealID
		time.Sleep(time.Nanosecond * 3)
	}

	stream, err := client.Results(ctx, myBrokerId)
	if err != nil {
		t.Error("Results return err: ", err)
	}

	for i := 0; i < len(deals); i++ {

		result, err := stream.Recv()
		if err == io.EOF {
			t.Error("stream closed from server side")
			return
		}
		if err != nil {
			t.Errorf("receive error %v", err)
		}
		found := false
	dealLoop:
		for i, deal := range deals {
			if deal.dealID.ID == result.ID {
				if deal.partial != result.Partial {
					t.Error("Expect partial, got =")
				}
				found = true
				deals[i] = deals[0]
				deals = deals[1:]
				break dealLoop
			}
		}
		if !found {
			t.Error("No result")
		}

	}

}
func TestClosingStat(t *testing.T) {
	stream, err := client.Statistic(ctx, myBrokerId)
	if err != nil {
		t.Error("Statisitcs return err: ", err)
	}

	for i := 0; i < 2; i++ {
		// receive data from stream
		_, err := stream.Recv()
		if err == io.EOF {
			t.Error("stream closed from server side")
			return
		}
		if err != nil {
			t.Errorf("receive error %v", err)
			continue
		}
	}
	ex.brokers.b[myBrokerId.ID].close <- struct{}{}

	_, err = stream.Recv()
	if err == nil {
		t.Errorf("stream must close from server side: %s", err)
	}

}
func TestClosingResults(t *testing.T) {
	deals := []struct {
		deal    *exchange.Deal
		partial bool
		dealID  *exchange.DealID
	}{
		{
			deal: &exchange.Deal{
				Ticker:   "SPFB.RTS",
				Amount:   -100,
				Price:    11,
				Time:     10000,
				BrokerID: int32(myBrokerId.GetID()),
			},
			partial: false,
		},

		{
			deal: &exchange.Deal{
				Ticker:   "SPFB.RTS",
				Amount:   100,
				Price:    11,
				Time:     10000,
				BrokerID: int32(myBrokerId.GetID()),
			},
			partial: false,
		},
		{
			deal: &exchange.Deal{
				Ticker:   "IMOEX",
				Amount:   -100,
				Price:    11,
				Time:     10000,
				BrokerID: int32(myBrokerId.GetID()),
			},
			partial: true,
		},
		{
			deal: &exchange.Deal{
				Ticker:   "IMOEX",
				Amount:   10,
				Price:    11,
				Time:     10000,
				BrokerID: int32(myBrokerId.GetID()),
			},
			partial: false,
		},
	}

	for i, deal := range deals {
		dealID, err := client.Create(ctx, deal.deal)
		if err != nil {
			t.Error("Create return err: ", err)
		}
		deals[i].dealID = dealID
		time.Sleep(time.Nanosecond * 3)
	}

	stream, err := client.Results(ctx, myBrokerId)
	if err != nil {
		t.Error("Results return err: ", err)
	}

	for i := 0; i < 4; i++ {
		_, err = stream.Recv()
		if err == io.EOF {
			t.Error("stream closed from server side")
			return
		}
		if err != nil {
			t.Errorf("receive error %v", err)
		}
	}

	ex.brokers.b[myBrokerId.ID].close <- struct{}{}

	_, err = stream.Recv()
	if err == nil {
		t.Errorf(" stream must close from server side: %s ", err)
	}

}
