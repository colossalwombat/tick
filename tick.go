package main 
import (
	"github.com/levigross/grequests"
	"fmt"
	//"strings"
	"io/ioutil"
	"github.com/gizak/termui"
	"os"
	"encoding/json"
	"strconv"
	//"time"
	"sort"
)

type UnitFigures struct{
	//data structure for each unit of time returned by the lovely alphavantage api's inccorigible json
	Open, High, Low, Close float64
	Volume int64
}

func getFigures(result map[string]interface{})(UnitFigures){
	figures := UnitFigures{}

	var err error
	
	figures.Open, err = strconv.ParseFloat(result["1. open"].(string), 64) 
	figures.High, err = strconv.ParseFloat(result["2. high"].(string), 64)
	figures.Low, err = strconv.ParseFloat(result["3. low"].(string), 64)
	figures.Close, err =  strconv.ParseFloat(result["4. close"].(string), 64) 
	figures.Volume, err = strconv.ParseInt(result["5. volume"].(string), 0, 0)

	check(err)

	return figures
}

func getHigh(values []float64) (float64){
	high := 0.0

	for _, v := range values {
		if v > high{
			high = v
		}
	}
	return high
}

func intCast(array []float64) ([]int) {

	intArray := make([]int, 0)
	for i := range array{
		intArray = append(intArray, int(array[i]))
	}

	//does this operate in place?
	return intArray
}

func displayGraph(values []float64, labels []string, symbol, mode string) {

	linechart := termui.NewLineChart()
	linechart.BorderLabel = fmt.Sprintf("%s Daily %s", symbol, mode)

	//if mode == "Volume" || mode == "volume" {
	//	linechart.Data = intCast(values[len(labels) - termui.TermWidth():])
	//} else {
	linechart.Data = values[len(labels) - termui.TermWidth():]
	//}	

	linechart.Mode = "dot"
	linechart.Width = termui.TermWidth()
	linechart.Height = termui.TermHeight()
	linechart.DataLabels = labels[len(labels) - termui.TermWidth():]
	linechart.X = 0
	linechart.Y = 0
	linechart.AxesColor = termui.ColorGreen
	linechart.LineColor = termui.ColorCyan | termui.AttrBold


	termui.Render(linechart)
	termui.Handle("/sys/kbd/q", func(termui.Event){
		termui.StopLoop()
		})
	//termui.Handle("
	termui.Loop()

}

func apiLookup(symbol string) (string){

	key, err := ioutil.ReadFile("key")
	check(err)

	//if you really want to steal my API key, go for it
	API_KEY := string(key)

	resp, err := grequests.Get(fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&outputsize=full&symbol=%s&apikey=%s", symbol, API_KEY), nil)

	check(err)

	return resp.String()

}

func processJson (json_request, mode string)([]float64, []string) {
	var result map[string]interface{}

	json.Unmarshal([]byte(json_request), &result)

	figures := []UnitFigures{}
	labels := []string{}

	keys := make([]string, 0)

	//sort the keys so we can iterate over the map in order
	for k, _ := range result[fmt.Sprintf("Time Series (%s)", "Daily")].(map[string]interface{}) {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {

		figures = append(figures, getFigures(result[fmt.Sprintf("Time Series (%s)", "Daily")].(map[string]interface{})[k].(map[string]interface{})))
		labels = append(labels, k)
	}

	figures_narrow := []float64{}

	//grab the values indicated by the mode string

	switch mode {
		case "High", "high":
			for _, figure := range figures{
				figures_narrow = append(figures_narrow, float64(figure.High))
				}

		case "Low", "low":
			for _, figure := range figures{
				figures_narrow = append(figures_narrow, float64(figure.Low))
			}

		case "Open", "open":
			for _, figure := range figures{
				figures_narrow = append(figures_narrow, float64(figure.Open))
			}
		
		case "Close", "close":
			for _, figure := range figures{
				figures_narrow = append(figures_narrow, float64(figure.Close))
			}
		
		case "Volume", "volume":
			for _, figure := range figures{
				figures_narrow = append(figures_narrow, float64(figure.Volume))
			}
		
	}

	return figures_narrow, labels
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


func main () {

	//set the symbol from the command line (temporary)
	SYMBOL, MODE := os.Args[1] ,os.Args[2]

	json_request := apiLookup(SYMBOL)

	//Initialize the UI
	initTermui()


	figures_processed, labels := processJson(json_request, MODE)
	

	displayGraph(figures_processed, labels, SYMBOL, MODE)

	termui.Close()

}

