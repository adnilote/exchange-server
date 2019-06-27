package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
)

type SliceToJSON struct {
	Result []Stat `json:"result"`
}

type RequestToJSON struct {
	Result []Request `json:"result"`
}

func CallDB(st []byte, Method string, db *sql.DB) []byte {
	result := make([]Stat, 0)
	switch Method {
	case "stat":
		rows, _ := db.Query("SELECT * FROM stat")
		defer rows.Close()

		for rows.Next() {
			MyStat := &Stat{} //new(Stat)
			//rows.Scan(&Stat)
			err := rows.Scan(&MyStat.ID, &MyStat.Time, &MyStat.Interval, &MyStat.Open, &MyStat.High, &MyStat.Low, &MyStat.Close, &MyStat.Volume, &MyStat.Ticker)
			if err != nil {
				fmt.Printf("%v\n", err)
			}
			//fmt.Printf("%v\n", MyStat)
			result = append(result, *MyStat)
		}
		out := &SliceToJSON{Result: result}
		js, _ := json.Marshal(out)
		return js
	case "deal":
		u := &Request{}
		err := json.Unmarshal(st, u)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		// todo GRPC
		rows, err := db.Exec("insert into request (id ,user_id,vol,Price,is_buy,Ticker) values (?,?,?,?,?,?)", &u.ID, &u.UserID, &u.Vol, &u.Price, &u.IsBuy, &u.Ticker)
		if err != nil {
			fmt.Println("Error with deal!!!", err)
			return nil
		}
		LastId, err := rows.LastInsertId()
		if err != nil {
			fmt.Println("Error LastId")
			return nil
		}

		rows, err = db.Exec("insert into request (id ,user_id,vol,Price,is_buy,Ticker) values (?,?,?,?,?,?)", &u.ID, &u.UserID, &u.Vol, &u.Price, &u.IsBuy, &u.Ticker)
		if err != nil {
			fmt.Println("Error with deal!!!", err)
			return nil
		}
		LastId, err = rows.LastInsertId()
		if err != nil {
			fmt.Println("Error LastId")
			return nil
		}

		return []byte(`{"body": {"id": "` + strconv.Itoa(int(LastId)) + `" }}`)

		////	case "ticker":
		////		rows, err := db.Query("select ticker from stat")
		////		if err != nil {
		////			fmt.Println("Error with ticker!!!", err)
		////			return nil
		////		}
		////		defer rows.Close()
		////		tick []string
		////		for rows.Next() {
		////			var str string
		////			rows.Scan(&str)
		////			tick = append(tick, str)
		////		}
		////		result = tick
		////		return result
	case "status":
		result := make([]Request, 0)
		/*u := &Request{}
		fmt.Println(string(st))
		err := json.Unmarshal(st, u)
		if err != nil {
			fmt.Println(err)
			fmt.Println("absd")
			return nil
		}*/

		rows, err := db.Query("select * from request")
		for rows.Next() {
			MyStat := &Request{}
			err := rows.Scan(&MyStat.ID, &MyStat.UserID, &MyStat.Vol, &MyStat.Price, &MyStat.IsBuy, &MyStat.Ticker)
			if err != nil {
				fmt.Printf("%v\n", err)
			}
			//fmt.Printf("%v\n", MyStat)
			result = append(result, *MyStat)
		}
		defer rows.Close()
		if err != nil {
			fmt.Println("Error with deal!!!", err)
			return nil
		}
		//LastId:=result[len(result)-1].ID
		if err != nil {
			fmt.Println("Error LastId")
			return nil
		}

		out := &RequestToJSON{Result: result}
		js, _ := json.Marshal(out)
		return js
	case "StatInsert":
		MyStat := &Stat{}
		_ = json.Unmarshal(st, MyStat)
		rows, _ := db.Query("insert into stat (id,time,`interval`,open,high,low,close,volume,ticker) values (?,?,?,?,?,?,?,?,?)", &MyStat.ID, &MyStat.Time, &MyStat.Interval, &MyStat.Open, &MyStat.High, &MyStat.Low, &MyStat.Close, &MyStat.Volume, &MyStat.Ticker)
		defer rows.Close()
		out := &SliceToJSON{Result: result}
		js, _ := json.Marshal(out)
		return js
	default:
		return []byte{}
	}

}
