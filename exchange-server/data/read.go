package exchangedata

// todo ??
//exit chan
import (
	"bufio"
	"hakaton-2018-2-2-msu/exchange-server/proto"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type record struct {
	Ticker string //название торгуемого инструмента
	Per    int    // период, у нас тики ( отдельные сделки ), игнорируйте это поле
	Time   int32
	Last   float32 // цена прошедшей сделки
	Vol    int32   // объём продшей сделки
}

// parseRecordList parse record from string
func parseRecordList(inputRecordList []string) record {
	result := record{}
	result.Ticker = inputRecordList[0]
	result.Per, _ = strconv.Atoi(inputRecordList[1])
	last, _ := strconv.ParseFloat(inputRecordList[4], 32)
	result.Last = float32(last)
	vol, _ := strconv.Atoi(inputRecordList[5])
	result.Vol = int32(vol)
	year, _ := strconv.Atoi(inputRecordList[2][:4])
	month, _ := strconv.Atoi(inputRecordList[2][4:6])
	day, _ := strconv.Atoi(inputRecordList[2][6:8])
	hour, _ := strconv.Atoi(inputRecordList[3][:2])
	minute, _ := strconv.Atoi(inputRecordList[3][2:4])
	second, _ := strconv.Atoi(inputRecordList[3][4:6])
	result.Time = int32(time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC).Unix())
	return result
}

// ReadPrices write to chan OHLCV of tool every second
func ReadPrices(tool string, out chan exchange.OHLCV, exit chan struct{}) {
	filePath := tool + "_180518_180518.txt"
	file, err := os.Open("data/" + filePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer close(out)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var i int64
	var beginPrice, minPrice, maxPrice float32
	var sumVol int32
	prevRecord := record{}
	scanner.Scan()
	for scanner.Scan() {
		select {
		case <-exit:
			return
		default:
		}
		inputStr := scanner.Text()
		inputRecordList := strings.Split(inputStr, ";")
		curRecord := parseRecordList(inputRecordList)

		if prevRecord.Ticker == "" {
			beginPrice = curRecord.Last
			minPrice = curRecord.Last
			maxPrice = curRecord.Last
			sumVol = curRecord.Vol
			prevRecord.Time = curRecord.Time
		} else {
			if curRecord.Time == prevRecord.Time {
				sumVol += curRecord.Vol
				if minPrice > curRecord.Last {
					minPrice = curRecord.Last
				}
				if maxPrice < curRecord.Last {
					maxPrice = curRecord.Last
				}
			} else {
				out <- exchange.OHLCV{
					ID:       i,
					Time:     prevRecord.Time,
					Interval: 1,
					Open:     beginPrice,
					High:     maxPrice,
					Low:      minPrice,
					Close:    prevRecord.Last,
					Volume:   sumVol,
					Ticker:   prevRecord.Ticker,
				}
				i++
				beginPrice = curRecord.Last
				minPrice = curRecord.Last
				maxPrice = curRecord.Last
				sumVol = curRecord.Vol
				time.Sleep(time.Second)
			}
		}
		prevRecord = curRecord
	}
}
