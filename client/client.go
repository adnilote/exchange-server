package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	Tmpl *template.Template
}

type PriceStat struct {
	Time   int
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int
	Ticker string
}

type Transmit struct {
	UserID int `json:"user_id"`
	Vol    int
	Price  int
	IsBuy  int `json:"is_buy"`
	Ticker string
}

func (h *Handler) Transcation(w http.ResponseWriter, r *http.Request) {
	/*	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8082/tickers", nil)
		if err != nil {
			fmt.Println(err)
		}
		client := &http.Client{Timeout: time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		var instruments map[string][]string
		err = json.Unmarshal(body, &instruments)
		if err != nil {
			fmt.Println(err)
		}
		res := instruments["result"]*/
	err := h.Tmpl.ExecuteTemplate(w, "index.html", struct{}{})
	/*		AvailableInstrs []string
	}{
		AvailableInstrs: res,
	})*/
	if err != nil {
		fmt.Println(err)
	}
}

func (h *Handler) Position(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8082/status", nil)
	if err != nil {
		fmt.Println(err)
	}
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(body)
	var statistic map[string][]*Transmit
	err = json.Unmarshal(body, &statistic)
	if err != nil {
		fmt.Println(err)
	}
	res := statistic["result"]
	err = h.Tmpl.ExecuteTemplate(w, "position.html", struct {
		Stats []*Transmit
	}{
		Stats: res,
	})
	if err != nil {
		fmt.Println(err)
	}
}

func (h *Handler) Price(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8082/stat", nil)
	if err != nil {
		fmt.Println(err)
	}
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var statistic map[string][]*PriceStat
	err = json.Unmarshal(body, &statistic)
	if err != nil {
		fmt.Println(err)
	}
	res := statistic["result"]
	err = h.Tmpl.ExecuteTemplate(w, "price.html", struct {
		Stats []*PriceStat
	}{
		Stats: res,
	})
	if err != nil {
		fmt.Println(err)
	}
}

func (h *Handler) Action(w http.ResponseWriter, r *http.Request) {
	sell := r.FormValue("sell")
	buy := 1
	if sell == "Продать" {
		buy = 0
	}
	volume, err := strconv.Atoi(r.FormValue("vol"))
	if err != nil {
		fmt.Println(err)
	}
	price, err := strconv.Atoi(r.FormValue("price"))
	if err != nil {
		fmt.Println(err)
	}
	info := &Transmit{
		UserID: 100500,
		Vol:    volume,
		Price:  price,
		IsBuy:  buy,
		Ticker: r.FormValue("ticker"),
	}
	res, err := json.Marshal(info)
	if err != nil {
		fmt.Println(err)
	}
	reqBody := bytes.NewReader(res)
	req, err := http.NewRequest(http.MethodPut, "http://127.0.0.1:8082/deal", reqBody)

	client := &http.Client{Timeout: time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Будет реализовано в следующей версии"))
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.Transcation(w, r)
	case "/position":
		h.Position(w, r)
	case "/price":
		h.Price(w, r)
	case "/form":
		h.Action(w, r)
	case "/cancel":
		h.Cancel(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	handler := &Handler{
		Tmpl: template.Must(template.ParseGlob("templates/*")),
	}
	fmt.Println("starting server at :8080")
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		fmt.Println(err)
	}
}
