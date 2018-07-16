package main

import (
	"encoding/json"
	"fmt"
	"github.com/levigross/grequests"
	tb "github.com/nsf/termbox-go"
	"os"
	"time"
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
	MARKET_OPEN  bool
	STOCKS       []string
	API_RETURN   []string
	FIGURES_LIST []stockFigures
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func logString(str string) {
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY, 0644)
	defer file.Close()
	check(err)

	str += string('\n')

	_, err = file.WriteString(str)

}

func isNYSEOpen() bool {

	loc, err := time.LoadLocation("America/New_York")
	check(err)

	//first check if it's closed
	if time.Now().After(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 15, 59, 0, 0, loc)) || time.Now().Before(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 9, 30, 0, 0, loc)) {
		return false
	} else {
		return true
	}
}

func (tick *Ticker) apiLookup() {

	tick.API_RETURN = make([]string, 0)

	for k := range tick.STOCKS {
		request, err := grequests.Get(fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/quote", tick.STOCKS[k]), nil)
		check(err)
		tick.API_RETURN = append(tick.API_RETURN, request.String())
	}
}

//parse the long form data for every stock
func (tick *Ticker) parseBatch() bool {

	tick.FIGURES_LIST = make([]stockFigures, 0)

	for s := range tick.API_RETURN {

		var result map[string]interface{}
		json.Unmarshal([]byte(tick.API_RETURN[s]), &result)

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

		//this only works when the markets open
		if tick.MARKET_OPEN {
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

		tick.FIGURES_LIST = append(tick.FIGURES_LIST, figure)
	}

	return true

}

func tickerHandler() {

	tick := &Ticker{
		STOCKS:      []string{"MMM", "AXP", "AAPL", "BA", "CAT", "CVX", "CSCO", "KO", "DIS", "DWDP", "XOM", "GE", "GS", "HD", "IBM", "INTC", "JNJ", "JPM", "MCD", "MRK", "MSFT", "NKE", "PFE", "PG", "TRV", "UTX", "UNH", "VZ", "V", "WMT"},
		MARKET_OPEN: isNYSEOpen(),
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

	refresh := make(chan bool)
	go func() {
		for {
			logString("Refreshing the data")
			tick.apiLookup()
			refresh <- tick.parseBatch()
			time.Sleep(3 * time.Second)
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
					if s.Selected > 0 {
						s.Selected--
						s.setTicker(tick)
					}
				}
				if ev.Key == tb.KeyArrowDown {
					if s.Selected < len(tick.STOCKS)-1 {
						s.Selected++
						s.setTicker(tick)
					}
				}
				if ev.Key == tb.KeySpace {
					tick.apiLookup()
					tick.parseBatch()
					s.setTicker(tick)
				}
				if ev.Key == tb.KeyCtrlA {
					tick.addStock(s)
					tick.apiLookup()
					tick.parseBatch()
					s.setTicker(tick)
				}

			case tb.EventResize:
				s.setTicker(tick)

			}
		case <-refresh:
			s.setTicker(tick)
		}
	}
}
