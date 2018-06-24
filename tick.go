package main 
import (
	"github.com/levigross/grequests"
	"fmt"
	//"strings"
	"time"
	"io/ioutil"
	//"io"
	"golang.org/x/crypto/ssh/terminal"
	//"math/rand"
	"os"
	"encoding/json"
	"strconv"
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

func displayGraph(values []float64, wipe bool) {
	width, height, err := terminal.GetSize(int(os.Stdin.Fd()))
	check(err)

	if wipe {
		fmt.Printf("\033[%dA", height -1)
	}

	high := getHigh(values)

	i := height - 1

	for i > 0 {
		//make sure not to overrun the line
		if height - i > width {
			fmt.Println("\n\n\n\n\n\n\ndoes this appear?")
			time.Sleep(2 * time.Second)
			break
		}

		fmt.Printf("\r")
		//loop through the values, print an X if the value is within the range

		for _, value := range values{
			if float64(value) >= (float64(i)/float64(height)) * float64(high) {
				fmt.Printf("X")
			} else {
				fmt.Printf(" ")
			}
		}
		fmt.Printf("\n")
		i--
	}


}


func check(e error){
	if e != nil{
		panic(e)
	}
}


func main () {
	key, err := ioutil.ReadFile("key")
	check(err)

	API_KEY := string(key)

	resp, err := grequests.Get(fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_INTRADAY&symbol=MSFT&interval=60min&apikey=%s", API_KEY), nil)

	check(err)

	json_request := resp.String()

	//OFFLINE VERSION
	//json_request, err := ioutil.ReadFile("/Users/Jack/Desktop/json_request")


	check(err)

	var result map[string]interface{}

	json.Unmarshal([]byte(json_request), &result)

	figures := []UnitFigures{}

	for _, v := range result["Time Series (60min)"].(map[string]interface{}){
		figures = append(figures, getFigures(v.(map[string]interface{})))
	}



	figures_high := []float64{}

	for _, figure := range figures{
		figures_high = append(figures_high, float64(figure.Volume))
	}
	

	for{
		displayGraph(figures_high, true)
		time.Sleep(1 * time.Second)
	}


}

