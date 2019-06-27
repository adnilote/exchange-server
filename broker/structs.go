package main

type Stat struct {
	ID       int     `json:"-"`
	Time     int
	Interval int     `json:"-"`
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   int
	Ticker   string
}

type Request struct {
	ID     int
	UserID int
	Vol    int
	Price  int
	IsBuy int
	Ticker string
}

type Clients struct {
	ID int
	LoginId int
	Balance int

}

type Positions struct {
	ID int
	UserId int
	Ticker string
	Vol int
}

type OrdersHistory struct {
	ID int
	Time int
	Ticker string
	Vol int
	Price float32
	IsBuy int




}