package main 
import (
	"github.com/levigross/grequests"
	"fmt"
	"strings"
	tb "github.com/nsf/termbox-go"
	"encoding/json"
	"strconv"
	"io/ioutil"
	"reflect"
	"sort"
	"time"
)

//define the 30 companies from the DJIA
var DOW_SYMB = []string{"MMM", "AXP", "AAPL", "BA", "CAT", "CVX", "CSCO", "KO", "DIS", "DWDP", "XOM", "GE", "GS", "HD", "IBM", "INTC", "JNJ", "JPM", "MCD", "MRK", "MSFT", "NKE", "PFE", "PG", "TRV", "UTX", "UNH", "VZ", "V", "WMT"}

type stockFigures struct {
	symbol string
	price, open, close, high, low float64
	volume int64
	colour int
}

func check(e error){
	if e != nil{
		panic(e)
	}
}

//store the last call
var oldPV = []stockFigures{}

func apiLookupRealtime(key string) (string){


	symbols_formatted := strings.Join(DOW_SYMB, ",")

	resp, err := grequests.Get(fmt.Sprintf("https://www.alphavantage.co/query?function=BATCH_STOCK_QUOTES&symbols=%s&apikey=%s", symbols_formatted, key), nil)
	check(err)
	ioutil.WriteFile("log.txt", []byte(fmt.Sprintf("https://www.alphavantage.co/query?function=BATCH_STOCK_QUOTES&symbols=%s&apikey=%s	%s", symbols_formatted, key, resp.String())), 0777)
	return resp.String()
}

func apiLookup(key string) ([]string) {

	list := []string{}

	for k := range DOW_SYMB {
		resp, err := grequests.Get(fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_INTRADAY&outputsize=compact&symbol=%s&apikey=%s&interval=60min", DOW_SYMB[k], key), nil)
		check(err)
		list = append(list, resp.String())
		time.Sleep(50 * time.Millisecond)
	}
	return list
}

func parseBatchLongJson(json_requests []string, list []stockFigures) ([]stockFigures){
}

	



func parseBatchJson(json_request string) ([]stockFigures){
	var result map[string]interface{}

	list := []stockFigures{}

	json.Unmarshal([]byte(json_request), &result)



	for k := range result["Stock Quotes"].([]interface{}) {
		var err error
		newPair := stockFigures{}

		newPair.symbol = result["Stock Quotes"].([]interface{})[k].(map[string]interface{})["1. symbol"].(string)
		newPair.price, err = strconv.ParseFloat(result["Stock Quotes"].([]interface{})[k].(map[string]interface{})["2. price"].(string),64)

		//new error handling
		if err != nil {
			newPair.price = 0
		}

		newPair.volume, err = strconv.ParseInt(result["Stock Quotes"].([]interface{})[k].(map[string]interface{})["3. volume"].(string), 0, 0)

		//new error handling
		if err != nil {
			newPair.volume = 0
		}

		if k < len(oldPV){
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

		list = append(list, newPair)
	}

	return list
}

func drawHelp(helpMsg string) {
	_, line := tb.Size()

	for i := 0; i < len(helpMsg); i++ {
		tb.SetCell(i, line - 1, rune(helpMsg[i]), tb.ColorWhite, tb.ColorBlue)
	}

}

func drawTicker(fields []string, colours []int, selected int){
	width, _ := tb.Size()

	//set the header
	title := "Tick"
	char := 0
	for i := 0; i < width; i++{
		//check if we should be printing the title letters
		if i > (width - len(title)) / 2 && char < len(title) {
			tb.SetCell(i, 0, rune(title[char]), tb.ColorWhite, 0)
			char++
		} else {
			tb.SetCell(i, 0, rune(0), 0, 0)
		}
	}
	for i := 0; i < width; i++{
		if i < len(fields[0]) {
			tb.SetCell(i, 1, rune(fields[0][i]), tb.ColorBlue, tb.ColorWhite)
		} else {
			tb.SetCell(i, 1, rune(0), 0, tb.ColorWhite)
		}
	}
	//remove the header
	fields = fields[1:]

	//do the rest of the fields
	for row := range fields{
		if row == selected {
			for i := range fields[row]{
				switch colours[row]{
					case 1:
						tb.SetCell(i, row + 2, rune(fields[row][i]), tb.ColorGreen, tb.ColorWhite)
					case 0:
						tb.SetCell(i, row + 2, rune(fields[row][i]), tb.ColorBlack, tb.ColorWhite)
					case -1:
						tb.SetCell(i, row + 2, rune(fields[row][i]), tb.ColorRed, tb.ColorWhite)
				}
			}
		} else{
			for i := range fields[row]{
				switch colours[row]{
					case 1:
						tb.SetCell(i, row + 2, rune(fields[row][i]), tb.ColorGreen, 0)
					case 0:
						tb.SetCell(i, row + 2, rune(fields[row][i]), tb.ColorWhite, 0)
					case -1:
						tb.SetCell(i, row + 2, rune(fields[row][i]), tb.ColorRed, 0)
				}
			} 
		}
	}

	helpMsg := "Close [Ctrl-Q]	Change Selection [Up/Down]	Refresh (temporary) [Space]"

	drawHelp(helpMsg)

	tb.Flush()


	return
}


func setTicker(reqReal string, reqData []string, selected int){

	list := parseBatchJson(reqReal)
	list = parseBatchLongJson(reqData, list)


	if !(reflect.DeepEqual(oldPV, list)) {
		oldPV = list
	}

	fields := []string{}
	colours := []int{}

	fields = append(fields, "SYMBOL      PRICE       VOLUME(m)")

	for s := range list{
		space := strings.Repeat(" ", 12)
		width,_ := tb.Size()

		//calculate the space sizes
		lineSpaces := width - (len(space)*2 + len(fmt.Sprintf("%d", list[s].volume))) - 3
		space1 := space[:len(space) - len(list[s].symbol)]
		space2 := space[:len(space) - len(fmt.Sprintf("%.2f", list[s].price))]
		spaceEnd := strings.Repeat(" ", lineSpaces)

		//format the field
		field := fmt.Sprintf("%s%s%.2f%s%d%s", list[s].symbol, space1, list[s].price, space2, list[s].volume, spaceEnd)

		//add it to the list
		fields = append(fields, field)
		colours = append(colours, list[s].colour)

	}

	drawTicker(fields, colours, selected)
}

func displayAddMenu(){
	width, height := tb.Size()
	tb.Clear(0,0)

	menuText := "Enter the name (symbol) of the stock to add:"

	for i := 0; i < len(menuText); i++ {
		tb.SetCell((width - len(menuText)) / 2 + i, height/2 - 1, rune(menuText[i]), tb.ColorBlack, tb.ColorWhite)
	}

	drawHelp("Cancel [ESC]")

	tb.Flush()
}

func takeInput() (string, bool){
	symbol := ""
	for {
		switch ev := tb.PollEvent(); ev.Type {
		case tb.EventKey:
			if ev.Key == tb.KeyEsc {
				return symbol, false
			} 
			if ev.Key == tb.KeyEnter {
				return symbol, true
			} 
			if ev.Key == tb.KeyBackspace2 && len(symbol) > 0 {
				width, height := tb.Size()
				symbol = symbol[:len(symbol) - 1]

				tb.SetCell(width / 2 - 23 + len(symbol) + 1, height / 2, 0, 0, 0)
				tb.Flush()
			} else {
				width, height := tb.Size()

				symbol += string(ev.Ch)
				tb.SetCell(width / 2 - 23 + len(symbol), height / 2, ev.Ch, 0, 0)
				tb.Flush()
			}
		case tb.EventResize:
			continue

		default:
			return symbol, false
		}
	}
}


func addStock(){
	displayAddMenu()

	symb, complete := takeInput()

	if !complete{
		return
	}

	DOW_SYMB = append(DOW_SYMB, symb)
	sort.Strings(DOW_SYMB)
}


func tickerHandler(API_KEY string){
	//setup the ticker (Initial API call)
	selected := 0
	req := apiLookupRealtime(API_KEY)
	reqLong := apiLookup(API_KEY)
	setTicker(req, reqLong, selected)


//logic that handles the keyboard interaction
HANDLE:
	for {
		switch ev := tb.PollEvent(); ev.Type {
			case tb.EventKey:
				if ev.Key == tb.KeyCtrlQ {
					break HANDLE
				}
				if ev.Key == tb.KeyArrowUp {
					if selected > 0 {
						selected--
						setTicker(req, reqLong, selected)
					}
				}
				if ev.Key == tb.KeyArrowDown {
					if selected < 30 {
						selected++ 
						setTicker(req, reqLong, selected)
					}
				}
				if ev.Key == tb.KeySpace {
					req = apiLookupRealtime(API_KEY)
					setTicker(req, reqLong, selected)
				}
				if ev.Key == tb.KeyCtrlA {
					addStock()
					req = apiLookupRealtime(API_KEY)
					setTicker(req, reqLong, selected)
				}

			case tb.EventResize:
				setTicker(req, reqLong, selected)

		}
	}
}




