package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	exchange "hakaton-2018-2-2-msu/exchange-server/proto"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
)

var (
	// DSN это соединение с базой
	// вы можете изменить этот на тот который вам нужен
	// docker run -p 3306:3306 -v $(PWD):/docker-entrypoint-initdb.d -e MYSQL_ROOT_PASSWORD=1234 -e MYSQL_DATABASE=golang -d mysql
	DSN = "root:golang2018@tcp(localhost:3306)/golang2017?charset=utf8"
	// DSN = "coursera:5QPbAUufx7@tcp(localhost:3306)/coursera?charset=utf8"
)

func main() {
	db, err := sql.Open("mysql", DSN)
	err = db.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		panic(err)
	}
	grpcConn, err := grpc.Dial(
		"127.0.0.1:8090",
		grpc.WithInsecure(),
	)
	if err != nil {
		return
	}

	clientExchange := exchange.NewExchangeClient(grpcConn)

	go func() {
		// Test Statisitc
		id := exchange.BrokerID{ID: 1}
		conn, _ := clientExchange.Statistic(context.Background(), &id)

		for {
			res, _ := conn.Recv()
			jsonRes, _ := json.Marshal(res)
			CallDB(jsonRes, "StatInsert", db)
		}

		// Test Create and Cancle
		deal1, err := clientExchange.Create(context.Background(), &exchange.Deal{
			ID:       1,
			BrokerID: 2,
			ClientID: 3,
			Ticker:   "IMOEX",
			Amount:   5,
			Partial:  true,
			Time:     7,
			Price:    8,
		})
		if err != nil {
			fmt.Printf("Create fail: %s", err)
		} else {
			fmt.Println("Create Done")
		}
		fmt.Println(deal1)
		deal2, err := clientExchange.Create(context.Background(), &exchange.Deal{
			ID:       2,
			BrokerID: 2,
			ClientID: 3,
			Ticker:   "IMOEX",
			Amount:   -5,
			Partial:  true,
			Time:     7,
			Price:    8,
		})
		if err != nil {
			fmt.Printf("Create fail: %s", err)
		} else {
			fmt.Println("Create Done")
		}
		fmt.Println(deal2)
		deal3, err := clientExchange.Cancel(context.Background(), &exchange.DealID{
			ID:       1,
			BrokerID: 2,
		})
		if err != nil {
			fmt.Printf("Create fail: %s", err)
		} else {
			fmt.Println("Create Done")
		}
		fmt.Println(deal3)

	}()

	handler, err := DbExplorer(db, &clientExchange)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("starting server at :8082")
	http.ListenAndServe(":8082", handler)
}
