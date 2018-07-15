package main 
import (
	"github.com/levigross/grequests"
	"fmt"
	"strings"
	tb "github.com/nsf/termbox-go"
	"os"
	"reflect"
	"sort"
	"time"
	"encoding/json"
	"math/rand"
)

type stockFigures struct {
	Symbol string
	Price float64
	Volume int
	Open, Close, High, Low, Change float64
	MarketCap int
	High52, Low52, YTDChange float64
	Colour tb.Attribute
}

type Ticker struct {
	MARKET_OPEN bool
	STOCKS []string
	SELECTED int

	}

func check(e error){
	if e != nil{
		panic(e)
	}
}

func logString(str string){
	file, err := os.OpenFile("log.txt", os.O_APPEND | os.O_WRONLY, 0644)
	defer file.Close()
	check(err)

	str += string('\n')

	_, err = file.WriteString(str)

}


func isNYSEOpen() (bool){

	loc, err := time.LoadLocation("America/New_York")
	check(err)

	//first check if it's closed
	if time.Now().After(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 15, 59, 0, 0, loc)) || time.Now().Before(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 9, 30, 0, 0, loc)){
		return false
	} else {
		return true
	}
}

func apiLookup() ([]string) {

	list := []string{}

	if MARKET_OPEN{

		for k := range STOCKS {
			request, err := grequests.Get(fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/quote", STOCKS[k]), nil)
			check(err)
			list = append(list, request.String())
		}
		
	} else {
		//otherwise pull the old data
		for k := range STOCKS {
			request, err := grequests.Get(fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/chart/1m", STOCKS[k]), nil)
			check(err)
			list = append(list, request.String())
		}
	}

	return list
}


//parse the long form data for every stock
func parseBatch(reqs []string) ([]stockFigures){

	list := []stockFigures{}

	for s := range reqs{
		
		var result map[string]interface{}
		json.Unmarshal([]byte(reqs[s]), &result)

		figure := stockFigures{}

		//set the fields by unmapping the json values
		figure.Symbol = result["symbol"].(string)
		figure.Open = result["open"].(float64)
		figure.Close = result["close"].(float64)
		figure.High = result["high"].(float64)
		figure.Low = result["low"].(float64)
		figure.Volume = int(result["iexVolume"].(float64))
		figure.Change = result["changePercent"].(float64) * 100
		figure.Price = result["latestPrice"].(float64)
		figure.MarketCap = int(result["marketCap"].(float64) / 1000000000)
		figure.High52 = result["week52High"].(float64)
		figure.Low52 = result["week52Low"].(float64)
		figure.YTDChange = result["ytdChange"].(float64) * 100



		//set the colour depending on the days change %
		if figure.Change > 0.0 {
			figure.Colour = tb.ColorGreen
		} else if figure.Change == 0.0 {
			figure.Colour = tb.ColorWhite
		} else {
			figure.Colour = tb.ColorRed
		}


		list = append(list, figure)
	}

	return list
}


func tickerHandler() (*Ticker){

	tick := &Ticker{
		STOCKS: ["MMM", "AXP", "AAPL", "BA", "CAT", "CVX", "CSCO", "KO", "DIS", "DWDP", "XOM", "GE", "GS", "HD", "IBM", "INTC", "JNJ", "JPM", "MCD", "MRK", "MSFT", "NKE", "PFE", "PG", "TRV", "UTX", "UNH", "VZ", "V", "WMT"]
		MARKET_OPEN: isNYSEOpen(),
		SELECTED: 0
	}
	//setup the ticker (Initial API call)
	reqLong := apiLookup()
	list := parseBatch(reqLong)
	setTicker(list, selected)

	//established the thread for the listener
	event_queue := make(chan tb.Event)
	go func() {
		for {
			event_queue <- tb.PollEvent()
		}
	}()

	refresh := make(chan []stockFigures)
	go func(){
		for {
			logString("Refreshing the data")
			req := apiLookup()
			refresh <- parseBatch(req)
			time.Sleep(3 * time.Second)
		}
	}()


//logic that handles the keyboard interaction
HANDLE:
	for {
		select{
		case ev := <-event_queue:
			switch ev.Type {
				case tb.EventKey:
					if ev.Key == tb.KeyCtrlQ {
						break HANDLE
					}
					if ev.Key == tb.KeyArrowUp {
						if selected > 0 {
							selected--
							setTicker(list, selected)
						}
					}
					if ev.Key == tb.KeyArrowDown {
						if selected < len(STOCKS) - 1 {
							selected++ 
							setTicker(list, selected)
						}
					}
					if ev.Key == tb.KeySpace {
						reqLong = apiLookup()
						list = parseBatch(reqLong)
						setTicker(list, selected)
					}
					if ev.Key == tb.KeyCtrlA {
						addStock()
						reqLong = apiLookup()
						list = parseBatch(reqLong)
						setTicker(list, selected)
					}

				case tb.EventResize:
					setTicker(list, selected)

				}
		case r := <-refresh:
			setTicker(r, selected)
		}
	}
}
