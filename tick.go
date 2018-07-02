package main 
import (
	"github.com/levigross/grequests"
	"fmt"
	"strings"
	"github.com/gizak/termui"
	"encoding/json"
	"strconv"
	"io/ioutil"
	"time"

)

//define the 30 companies from the DJIA
var DOW_SYMB = []string{"MMM", "AXP", "AAPL", "BA", "CAT", "CVX", "CSCO", "KO", "DIS", "DWDP", "XOM", "GE", "GS", "HD", "IBM", "INTC", "JNJ", "JPM", "MCD", "MRK", "MSFT", "NKE", "PFE", "PG", "TRV", "UTX", "UNH", "VZ", "V", "WMT"}

type PVpair struct {
	symbol string
	price float64
	volume int64
	colour int
}

//store the last call
var oldPV = []PVpair{}


func apiLookupRealtime(key string) (string){


	symbols_formatted := strings.Join(DOW_SYMB, ",")

	resp, err := grequests.Get(fmt.Sprintf("https://www.alphavantage.co/query?function=BATCH_STOCK_QUOTES&symbols=%s&apikey=%s", symbols_formatted, key), nil)
	check(err)
	ioutil.WriteFile("log.txt", []byte(fmt.Sprintf("https://www.alphavantage.co/query?function=BATCH_STOCK_QUOTES&symbols=%s&apikey=%s	%s", symbols_formatted, key, resp.String())), 0777)
	return resp.String()
}

func parseBatchJson(json_request string) ([]PVpair){
	var result map[string]interface{}

	list := []PVpair{}

	json.Unmarshal([]byte(json_request), &result)


	for k := range result["Stock Quotes"].([]interface{}) {
		var err error
		newPair := PVpair{}

		newPair.symbol = result["Stock Quotes"].([]interface{})[k].(map[string]interface{})["1. symbol"].(string)
		newPair.price, err = strconv.ParseFloat(result["Stock Quotes"].([]interface{})[k].(map[string]interface{})["2. price"].(string),64)
		newPair.volume, err = strconv.ParseInt(result["Stock Quotes"].([]interface{})[k].(map[string]interface{})["3. volume"].(string), 0, 0)

		if len(oldPV) > 1{
			if newPair.price > oldPV[k].price {
				newPair.colour = 1
			} else if newPair.price == oldPV[k].price {
				newPair.colour = 0
			} else {
				newPair.colour = -1
			}
		} else {
			newPair.colour = 0
		}

		check(err)
		list = append(list, newPair)
	}

	return list

	
}

func initTermui(){
	err := termui.Init()
	check(err)	
}



func check(e error){
	if e != nil{
		panic(e)
	}
}

func showTicker(req string, selected int){



	list := parseBatchJson(req)

	oldPV = list

	fields := []string{}

	fields = append(fields, "SYMBOL      PRICE       VOLUME")

	for s := range list{
		var str string
		space1 := "            "
		space2 := "            "

		//more gross one liners...
		lineSpaces := termui.TermWidth() - (len(space1) + len(space2) + len(fmt.Sprintf("%d", list[s].volume))) - 3

		//calculate the formatting for the spaces
		space1 = space1[:len(space1) - len(list[s].symbol)]
		space2 = space2[:len(space2) - len(fmt.Sprintf("%.2f", list[s].price))]


		spaceEnd := strings.Repeat(" ", lineSpaces)

		//color the proper text
		var cur_symbol string

		cur_symbol = list[s].symbol

		


		if s == selected {
			str = fmt.Sprintf("[%s%s%.2f%s%d%s](fg-black,bg-white)", list[s].symbol, space1, list[s].price, space2, list[s].volume, spaceEnd)
		} else {
			switch list[s].colour {
						case 1:
							str = fmt.Sprintf("[%s%s%.2f%s%d%s](fg-green)", cur_symbol, space1, list[s].price, space2, list[s].volume, spaceEnd)
						case 0:
							str = fmt.Sprintf("[%s%s%.2f%s%d%s](fg-blue)", cur_symbol, space1, list[s].price, space2, list[s].volume, spaceEnd)
						case -1:
							str = fmt.Sprintf("[%s%s%.2f%s%d%s](fg-red)", cur_symbol, space1, list[s].price, space2, list[s].volume, spaceEnd)	
					}
			
		}
		fields = append(fields, str)

	}


	ls := termui.NewList()
	ls.Items = fields
	ls.BorderLabel = "DJIA"
	ls.Height = termui.TermHeight()
	ls.Width = termui.TermWidth()


	termui.Render(ls)



}

func tickerHandler(API_KEY string){
	//setup the ticker (Initial API call)
	selected := 1
	req := apiLookupRealtime(API_KEY)
	showTicker(req, selected)

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})
	termui.Handle("/sys/kbd/<up>", func(termui.Event) {
		if selected > 0 {
			selected--
			showTicker(req, selected)
		}
	})
	termui.Handle("/sys/kbd/<down>", func(termui.Event) {
		if selected < 30 {
			selected++
			showTicker(req, selected)
		}
	})
	termui.Handle("/sys/kbd/<enter>", func(termui.Event) {
		graphHandler(API_KEY, DOW_SYMB[selected])
	})

	//init the new timer
	termui.Merge("/timer/4s", termui.NewTimerCh(time.Second * 4))
	
	termui.Handle("/timer/4s", func(termui.Event) {
		req = apiLookupRealtime(API_KEY)
		showTicker(req, selected)
	})

	termui.Loop()

}




