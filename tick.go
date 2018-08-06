package main

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/levigross/grequests"
	tb "github.com/nsf/termbox-go"
	"math"
	"os"
	"time"
	"reflect"
)

type stockFigures struct {
	Symbol                         string
	Price                          float64
	Volume                         int
	Open, Close, High, Low, Change float64
	MarketCap                      int
	High52, Low52, YTDChange       float64
	Colour                         tb.Attribute
}

type Ticker struct {
	Market_Open  bool
	Stocks       []string
	Api_Return   []string
	Figures_List []stockFigures
	Selected     int
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func logString(args ...interface{}) {
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY, 0644)
	defer file.Close()
	check(err)

	str := spew.Sdump(args)
	str += string('\n')

	_, err = file.WriteString(str)

}

func isNYSEOpen() bool {

	request, err := grequests.Get("https://api.iextrading.com/1.0/market", nil)
	check(err)
	logString(reflect.TypeOf(request))

	var result []map[string]interface{}
	json.Unmarshal([]byte(request.String()), &result)

	lastTime := int64(result[12]["lastUpdated"].(float64) / 1000)
	currentTime := time.Now().Unix()

	if math.Abs(float64(currentTime-lastTime)) > 10.0 {
		return false
	} else {
		return true
	}
}

func duplicateTicker(tick Ticker) *Ticker {
	newTick := tick
	copy(newTick.Stocks, tick.Stocks)
	copy(newTick.Api_Return, tick.Api_Return)
	copy(newTick.Figures_List, tick.Figures_List)

	return &newTick
}

func (tick *Ticker) verifySecurityExists(symbol string) bool {

	request, err := grequests.Get(fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/quote", symbol), nil)
	check(err)

	if request.String() == "Unknown symbol" {
		return false
	}
	return true
}

func (tick *Ticker) apiLookup() {

	tick.Api_Return = make([]string, 0)

	for k := range tick.Stocks {
		request, err := grequests.Get(fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/quote", tick.Stocks[k]), nil)
		check(err)

		//catch API errors
		//TODO make this more efficient, stop crashing when the API lookup fails
		if request.String() == "Unknown symbol" {
			continue
		}

		tick.Api_Return = append(tick.Api_Return, request.String())
	}

}

//parse the long form data for every stock
func (tick *Ticker) parseBatch() bool {

	tick.Figures_List = make([]stockFigures, 0)

	for s := range tick.Api_Return {

		var result map[string]interface{}
		json.Unmarshal([]byte(tick.Api_Return[s]), &result)

		figure := stockFigures{}

		//set the fields by unmapping the json values
		figure.Symbol = result["symbol"].(string)
		figure.Open = result["open"].(float64)
		figure.Close = result["close"].(float64)
		figure.High = result["high"].(float64)
		figure.Low = result["low"].(float64)
		figure.Change = result["changePercent"].(float64) * 100
		figure.Price = result["latestPrice"].(float64)
		figure.MarketCap = int(result["marketCap"].(float64) / 1000000000)
		figure.High52 = result["week52High"].(float64)
		figure.Low52 = result["week52Low"].(float64)
		figure.YTDChange = result["ytdChange"].(float64) * 100

		//this only works when the market's open
		if tick.Market_Open {
			figure.Volume = int(result["iexVolume"].(float64))
		} else {
			figure.Volume = 0
		}

		//set the colour depending on the days change %
		if figure.Change > 0.0 {
			figure.Colour = tb.ColorGreen
		} else if figure.Change == 0.0 {
			figure.Colour = tb.ColorWhite
		} else {
			figure.Colour = tb.ColorRed
		}

		tick.Figures_List = append(tick.Figures_List, figure)
	}

	return true

}

func tickerHandler() {

	tick := &Ticker{
		//define a list of (mostly) S&P stocks
		Stocks:      []string{"ABT", "ABBV", "ACN", "ADBE", "ADT", "AAP", "AES", "AET", "AFL", "AMG", "A", "APD", "AKAM", "AA", "AGN", "ALXN", "ALLE", "ADS", "ALL", "ALTR", "MO", "AMZN", "AEE", "AAL", "AEP", "AXP", "AIG", "AMT", "AMP", "ABC", "AME", "AMGN", "APH", "APC", "ADI", "AON", "APA", "AIV", "AMAT", "ADM", "AIZ", "T", "ADSK", "ADP", "AN", "AZO", "AVGO", "AVB", "AVY", "BHI", "BLL", "BAC", "BK", "BCR", "BXLT", "BAX", "BBT", "BDX", "BBBY", "BRK-B", "BBY", "BLX", "HRB", "BA", "BWA", "BXP", "BSK", "BMY", "BRCM", "BF-B", "CHRW", "CA", "CVC", "COG", "CAM", "CPB", "COF", "CAH", "HSIC", "KMX", "CCL", "CAT", "CBG", "CBS", "CELG", "CNP", "CTL", "CERN", "CF", "SCHW", "CHK", "CVX", "CMG", "CB"},
		Market_Open: isNYSEOpen(),
	}
	//setup the ticker (Initial API call)
	tick.apiLookup()
	tick.parseBatch()

	//Init the screen
	s := &Screen{}
	s.setTicker(tick)

	//established the thread for the listener
	event_queue := make(chan tb.Event)
	go func() {
		for {
			event_queue <- tb.PollEvent()
		}
	}()

	refresh := make(chan *Ticker)

	//screen/data refresh loop, mostly under control now
	go func() {
		for {
			time.Sleep(3 * time.Second)
			newTick := duplicateTicker(*tick)
			newTick.apiLookup()
			newTick.parseBatch()

			s.showRefresh()
			refresh <- newTick
		}
	}()

	//logic that handles the keyboard interaction
HANDLE:
	for {
		select {
		case ev := <-event_queue:
			switch ev.Type {
			case tb.EventKey:
				if ev.Key == tb.KeyCtrlQ {
					break HANDLE
				}
				if ev.Key == tb.KeyArrowUp {
					if tick.Selected > 0 {
						tick.Selected--
						s.setTicker(tick)
					}
				}
				if ev.Key == tb.KeyArrowDown {
					if tick.Selected < len(tick.Stocks)-2 {
						tick.Selected++
						s.setTicker(tick)
					}
				}
				if ev.Key == tb.KeySpace {
					tick.apiLookup()
					tick.parseBatch()
					s.setTicker(tick)
				}
				if ev.Key == tb.KeyCtrlA {

					//checks if adding is successful, calls setTicker to show the main view again on failure
					if !tick.addStock(s) {
						s.setTicker(tick)
						continue
					}
					//refresh the lookup now that there's a new stock in the list
					tick.apiLookup()
					tick.parseBatch()
					s.setTicker(tick)
				}
				if ev.Key == tb.KeyEnter {
					//in the future, this will show a chart
					s.chartMenuHandler(tick.Stocks[tick.Selected])
				}

			case tb.EventResize:
				s.setTicker(tick)

			}
		case r := <-refresh:
			//refresh the screen if there is no user interaction
			r.Selected = tick.Selected
			s.setTicker(r)
		}
	}
}
