package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	exchange "hakaton-2018-2-2-msu/exchange-server/proto"
	"io/ioutil"
	"net/http"
	"strings"
)

/*
type InfoFromRequest struct {
	HTTPMethod string
	Method     string
}
*/
type Handler struct {
	DB     *sql.DB
	Client *exchange.ExchangeClient
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db := h.DB

	var method string

	body, _ := ioutil.ReadAll(r.Body)

	if strings.Contains(r.URL.Path, "stat") {
		method = "stat"
	}

	if strings.Contains(r.URL.Path, "tickers") {
		method = "tickers"
	}

	if strings.Contains(r.URL.Path, "cancel") {
		method = "cancel"
	}

	if strings.Contains(r.URL.Path, "deal") {
		method = "deal"
	}

	if strings.Contains(r.URL.Path, "status") {
		method = "status"
	}
	fmt.Println(string(body))
	jsonAns := CallDB(body, method, db)
	//fmt.Printf("%v\n", sliceFromDB)

	resultToClient(jsonAns, w)
}

func resultToClient(jsonAns []byte, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonAns)
}

func requestHandler(r *http.Request, fromBody *interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &*fromBody)
	if err != nil {
		return err
	}
	return nil
}

func DbExplorer(db *sql.DB, cl *exchange.ExchangeClient) (http.Handler, error) {

	handler := &Handler{}
	handler = &Handler{DB: db, Client: cl}

	return handler, nil
}
