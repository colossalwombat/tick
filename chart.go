package main

import (
	"fmt"
	"encoding/json"
	"github.com/levigross/grequests"
	tb "github.com/nsf/termbox-go"
	"math"
)

type Chart struct{
	Title string
	Datapoints []float64
	Dates []string
	Symbol string
	Interval int
}

func getMin(array []float64)(float64){
	min := math.MaxFloat64
	for k := range array{
		if array[k] < min{
			min = array[k]
		}
	}
	return min
}

func getMax(array []float64)(float64){
	max := 0.0
	for k := range array{
		if array[k] > max{
			max = array[k]
		}
	}
	return max
}

func (c *Chart)getData(){
	var err error
	var request *grequests.Response
	
	switch c.Interval{
		//cases are in descending order from the other menu
		//1. One Day 2. One Month 3. Three Months 4. Six Months etc
		case 0:
			request, err = grequests.Get(fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/chart/1d", c.Symbol), nil)
		case 1: 
			request, err = grequests.Get(fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/chart/1m", c.Symbol), nil)
		case 2: 
			request, err = grequests.Get(fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/chart/3m", c.Symbol), nil)
		case 3: 
			request, err = grequests.Get(fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/chart/6m", c.Symbol), nil)
		case 4:
			request, err = grequests.Get(fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/chart/ytd", c.Symbol), nil)
		case 5:
			request, err = grequests.Get(fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/chart/1y", c.Symbol), nil)
		case 6:
			request, err = grequests.Get(fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/chart/2y", c.Symbol), nil)
		case 7:
			request, err = grequests.Get(fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/chart/5y", c.Symbol), nil)
	}

	check(err)
	var api_return []map[string]interface{}
	json.Unmarshal([]byte(request.String()), &api_return)

	for k := range api_return{
		c.Datapoints = append(c.Datapoints, api_return[k]["close"].(float64))
	}
}

func (c *Chart) drawChart(s *Screen){
	tb.Clear(0,0)
	s.refreshSize()

	//reference the numbers to the intervals
	posIntervals := []string{"One Day", "One Month", "Three Months", "Six Months", "Year-To-Date", "One Year", "Two Years", "Five Years"}
	//draw the title block 
	titleText := c.Symbol + " - " + posIntervals[c.Interval]
	for i := range titleText{
		tb.SetCell((s.Width - len(titleText)) / 2 + i, 0,  rune(titleText[i]), tb.ColorWhite, 0)
	}

	//chart logic
	//Basic idea: take the points, average the ones that occupy the width of a single character
	bullet := rune('\u2022')
	graphHeight := s.Height - 3
	maxDataPoint, minDataPoint := getMax(c.Datapoints), getMin(c.Datapoints)

	for k := 0; k < len(c.Datapoints); k++{
		for j := 0; j < graphHeight; j++ {
			if j <= int((c.Datapoints[k] - minDataPoint) / (maxDataPoint - minDataPoint) * float64(graphHeight)) {
				tb.SetCell(k, s.Height - 3 - j, bullet, tb.ColorCyan,0)
			}
		}

	}

	tb.Flush()

}


func (s *Screen) chartHandler(symbol string, interval int){
	c := &Chart{
		Symbol: symbol,
		Interval: interval,
	}

	c.getData()
	c.drawChart(s)

	//input loop
	for{
		switch ev := tb.PollEvent(); ev.Type{
		case tb.EventKey:
			if ev.Key == tb.KeyEsc{
				return
			}
		}
	}
}