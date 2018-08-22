package main

import (
	"encoding/json"
	"fmt"
	"github.com/levigross/grequests"
	tb "github.com/nsf/termbox-go"
	"math"
)

type Chart struct {
	Title      string
	Datapoints []float64
	Dates      []string
	Symbol     string
	Interval   int
}

func getMin(array []float64) float64 {
	min := math.MaxFloat64
	for k := range array {
		if array[k] < min {
			min = array[k]
		}
	}
	return min
}

func getMax(array []float64) float64 {
	max := 0.0
	for k := range array {
		if array[k] > max {
			max = array[k]
		}
	}
	return max
}

func (c *Chart) getData() {
	var err error
	var request *grequests.Response

	switch c.Interval {
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

	for k := range api_return {
		switch api_return[k]["close"].(type) {
		case float64:
			c.Datapoints = append(c.Datapoints, api_return[k]["close"].(float64))
		case nil:
			c.Datapoints = append(c.Datapoints, api_return[k]["average"].(float64))
		}
	}
}

func (c *Chart) drawChart(s *Screen) {
	tb.Clear(0, 0)
	s.refreshSize()

	//reference the numbers to the intervals
	posIntervals := []string{"One Day", "One Month", "Three Months", "Six Months", "Year-To-Date", "One Year", "Two Years", "Five Years"}
	//draw the title block
	titleText := c.Symbol + " - " + posIntervals[c.Interval]
	for i := range titleText {
		tb.SetCell((s.Width-len(titleText))/2+i, 0, rune(titleText[i]), tb.ColorWhite, 0)
	}

	//chart logic
	//Basic idea: take the points, average the ones that occupy the width of a single character
	bullet := rune('\u2022')
	graphHeight := s.Height - 3
	maxDataPoint, minDataPoint := getMax(c.Datapoints), getMin(c.Datapoints)

	//two cases: either there are more datapoints than we can display, and we have to average OR there are less and we need to extrapolate
	if len(c.Datapoints) > s.Width {
		compressedDataPoints := []float64{}

		for i := 0; i < s.Width; i++ {
			var average float64
			for j := 0; j < len(c.Datapoints)/s.Width; j++ {
				average += c.Datapoints[i+j]
			}
			average /= float64(len(c.Datapoints) / s.Width)

			compressedDataPoints = append(compressedDataPoints, average)
		}
		c.Datapoints = compressedDataPoints
	} else {
		//ie there are less points

		extrapolatedDataPoints := []float64{}
		dataToWidth := float64(len(c.Datapoints)) / float64(s.Width)
		widthToData := 1.0 / dataToWidth

		//simple linear extrapolation algorithm
		for i := 0; i < s.Width; i++ {
			j := int(float64(i) * dataToWidth)

			//this shouldn't be necessary
			if j >= len(c.Datapoints)-1 {
				break
			}

			slope := (c.Datapoints[j+1] - c.Datapoints[j]) / widthToData

			extrapolatedDataPoints = append(extrapolatedDataPoints, (float64(i)-float64(j)*widthToData)*slope)
		}
		c.Datapoints = extrapolatedDataPoints
	}
	logString(c.Datapoints)

	//jog the graph at 2/3 of the minimum data point
	jogGraph := 2 * minDataPoint / 3

	for k := 0; k < len(c.Datapoints); k++ {
		for j := graphHeight; j > 0; j-- {
			if j <= int((c.Datapoints[k]-jogGraph)/(maxDataPoint-jogGraph)*float64(graphHeight)) {
				tb.SetCell(k, s.Height-3-j, bullet, tb.ColorCyan, 0)
				break
			}
		}

	}

	//draw the labels TODO

	tb.Flush()
}

func (s *Screen) chartHandler(symbol string, interval int) {
	c := &Chart{
		Symbol:   symbol,
		Interval: interval,
	}

	c.getData()
	c.drawChart(s)

	//input loop
	for {
		switch ev := tb.PollEvent(); ev.Type {
		case tb.EventKey:
			if ev.Key == tb.KeyEsc {
				return
			}
		}
	}
}
