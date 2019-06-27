package main

import (
	"hakaton-2018-2-2-msu/exchange-server/proto"
	"math"
)

type Query struct {
	data []*exchange.Deal
}

func (query *Query) Len() int     { return len(query.data) }
func (query Query) Swap(i, j int) { query.data[i], query.data[j] = query.data[j], query.data[i] }

func (query *Query) Push(x interface{}) {
	item := x.(*exchange.Deal)
	query.data = append(query.data, item)
}

// Pop returns best deal removing it from heap
func (query *Query) Pop() interface{} {
	old := query.data
	n := len(old)
	if n <= 0 {
		return nil
	}
	item := old[n-1]
	query.data = old[0 : n-1]
	return item
}

func (query *Query) Find(deal exchange.DealID) int {
	for i, q := range query.data {
		if q.GetID() == deal.ID {
			return i
		}
	}
	return -1
}

// Get returns best deal without removing it from heap
func (query *Query) Get() *exchange.Deal {
	if query.Len() > 0 {
		return query.data[0]
	}
	return nil
}

type SellQuery struct {
	*Query
}

func (query SellQuery) Less(i, j int) bool {
	return query.data[i].Price < query.data[j].Price && query.data[i].GetTime() < query.data[j].GetTime()
}

// GetBestPrice return min price in sellQuery.
// Can be use with RLock()
func (query *SellQuery) GetBestPrice() float32 {
	deal := query.Get()
	if deal != nil {
		return deal.GetPrice()
	}
	return math.MaxFloat32
}

type BuyQuery struct {
	*Query
}

// GetBestPrice return max price in buyQuery.
// Can be use with RLock()
func (query *BuyQuery) GetBestPrice() float32 {
	deal := query.Get()
	if deal != nil {
		return deal.GetPrice()
	}
	return 0
}

func (query BuyQuery) Less(i, j int) bool {
	return query.data[i].Price > query.data[j].Price && query.data[i].GetTime() < query.data[j].GetTime()
}
func (query BuyQuery) Len() int { return len(query.data) }

// ///////////////////////////////
// type SellQuery []*exchange.Deal

// func (query SellQuery) Len() int { return len(query) }
// func (query SellQuery) Less(i, j int) bool {
// 	return query[i].Price < query[j].Price && query[i].GetTime() < query[j].GetTime()
// }
// func (query SellQuery) Swap(i, j int) { query[i], query[j] = query[j], query[i] }

// func (query *SellQuery) Push(x interface{}) {
// 	item := x.(*exchange.Deal)
// 	*query = append(*query, item)
// }

// func (query *SellQuery) Pop() interface{} {
// 	old := *query
// 	n := len(old)
// 	if n <= 0 {
// 		return nil
// 	}
// 	item := old[n-1]
// 	*query = old[0 : n-1]
// 	return item
// }

// func (query *SellQuery) Get() (*exchange.Deal, error) {
// 	if query.Len() > 0 {
// 		return (*query)[0], nil
// 	}
// 	return nil, fmt.Errorf("SellQuery is empty")
// }

// func (query *SellQuery) Find(x interface{}) int64 {
// 	for _, q := range *query {
// 		deal := x.(exchange.DealID)
// 		if q.GetID() == deal.ID {
// 			return deal.ID
// 		}
// 	}
// 	return -1
// }

// func (query *SellQuery) Delete(iDel int64) {
// 	k := 0
// 	old := *query
// 	n := len(old) - 1
// 	new := SellQuery{}

// 	for i := 0; i < n; i++ {
// 		if int64(i) == iDel {
// 			k++
// 		}
// 		new[i] = old[k]
// 		k++
// 	}
// 	*query = new
// }

// type BuyQuery []*exchange.Deal

// func (query BuyQuery) Len() int { return len(query) }
// func (query BuyQuery) Less(i, j int) bool {
// 	return query[i].Price > query[j].Price && query[i].GetTime() < query[j].GetTime()
// }
// func (query BuyQuery) Swap(i, j int) {
// 	query[i], query[j] = query[j], query[i]
// }

// func (query *BuyQuery) Push(x interface{}) {
// 	item := x.(*exchange.Deal)
// 	*query = append(*query, item)
// }

// func (query *BuyQuery) Pop() interface{} {
// 	old := *query
// 	n := len(old)
// 	if n <= 0 {
// 		return nil
// 	}
// 	item := old[n-1]
// 	*query = old[0 : n-1]
// 	return item
// }

// func (query *BuyQuery) Find(x interface{}) int64 {
// 	for _, q := range *query {
// 		deal := x.(exchange.DealID)
// 		if q.GetID() == deal.ID {
// 			return deal.ID
// 		}
// 	}
// 	return -1
// }

// func (query *BuyQuery) Delete(iDel int64) {
// 	k := 0
// 	old := *query
// 	n := len(old) - 1
// 	new := BuyQuery{}

// 	for i := 0; i < n; i++ {
// 		if int64(i) == iDel {
// 			k++
// 		}
// 		new[i] = old[k]
// 		k++
// 	}
// 	*query = new
// }
