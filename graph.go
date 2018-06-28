package main

import (
	"strconv"
	"github.com/gizak/termui"
	"fmt"
	"github.com/levigross/grequests"
	"encoding/json"
	"sort"
)

type UnitFigures struct{
	//data structure for each unit of time returned by the lovely alphavantage api's inccorigible json
	Open, High, Low, Close float64
	Volume int64
}


//parse the values from the json map
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


func displayGraph(values []float64, labels []string, symbol, mode string, xshift int) {

	linechart := termui.NewLineChart()
	linechart.BorderLabel = fmt.Sprintf("%s Daily %s", symbol, mode)

    //limit the data set to what shold be on screen, specified by the xshift
    //xshift is added, but is a negative value so the net result is to subtract it
    //this is because having a positive value indicate a leftward shift hurts my brain

	linechart.Data = values[len(values) - termui.TermWidth() + xshift:len(values) + xshift]
    linechart.DataLabels = labels[len(labels) - termui.TermWidth() + xshift:len(labels) + xshift]

	linechart.Mode = "dot"
	linechart.Width = termui.TermWidth()
	linechart.Height = termui.TermHeight()
	linechart.X = 0
	linechart.Y = 0
	linechart.AxesColor = termui.ColorGreen
	linechart.LineColor = termui.ColorCyan | termui.AttrBold


	termui.Render(linechart)
}

func apiLookup(symbol, API_KEY string, full bool) (string){

	if full{
		resp, err := grequests.Get(fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&outputsize=full&symbol=%s&apikey=%s", symbol, API_KEY), nil)
		check(err)
		return resp.String()
	} else {
		resp, err := grequests.Get(fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&outputsize=compact&symbol=%s&apikey=%s", symbol, API_KEY), nil)
		check(err)
		return resp.String()
	}
}

func processJsonGraph (json_request, mode string)([]float64, []string) {
	var result map[string]interface{}

	json.Unmarshal([]byte(json_request), &result)

	figures := []UnitFigures{}
	labels := []string{}

	keys := make([]string, 0)

	fmt.Println(json_request)

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

func graphHandler(API_KEY, symbol string){
	xshift := 0
	//init graph
	jsonString := apiLookup(symbol, API_KEY, true)
	values, labels := processJsonGraph(jsonString, "Close")
	displayGraph(values, labels, symbol, "Close", xshift)

	//key triggers, q to quit

	termui.Handle("/sys/kbd/<escape>", func(termui.Event){
		//you would think this would leak memory like crazy, but it doesn't seem to at all
		//maybe go is smarter than I thought
		tickerHandler(API_KEY)
		})
	termui.Handle("/sys/kbd/<left>", func(termui.Event){
		if -(xshift) < len(labels){
			xshift -= 5
			displayGraph(values, labels, symbol, "Close", xshift)
		}
	})
	termui.Handle("/sys/kbd/<right>", func(termui.Event){
		if xshift < 0 {
			xshift += 5
			displayGraph(values, labels, symbol, "Close", xshift)
		}
	})
	termui.Loop()

}