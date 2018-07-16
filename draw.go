package main

import (
	"fmt"
	tb "github.com/nsf/termbox-go"
	"reflect"
	"sort"
	"strings"
)

type Screen struct {
	Fields                  []string
	Colours                 []tb.Attribute
	Width, Height, Selected int
}

func (s *Screen) refreshSize() {
	s.Width, s.Height = tb.Size()
}

func (s *Screen) drawHelp(helpMsg string) {
	s.refreshSize()

	for i := 0; i < len(helpMsg); i++ {
		tb.SetCell(i, s.Height-1, rune(helpMsg[i]), tb.ColorBlack, tb.ColorCyan)
	}

}

func (s *Screen) drawTicker() {
	s.refreshSize()

	//set the header
	title := "Tick"
	char := 0
	for i := 0; i < s.Width; i++ {
		//check if we should be printing the title letters
		if i > (s.Width-len(title))/2 && char < len(title) {
			tb.SetCell(i, 0, rune(title[char]), tb.ColorWhite, 0)
			char++
		} else {
			tb.SetCell(i, 0, rune(0), 0, 0)
		}
	}
	for i := 0; i < s.Width; i++ {
		if i < len(s.Fields[0]) {
			tb.SetCell(i, 1, rune(s.Fields[0][i]), tb.ColorBlack, tb.ColorCyan)
		} else {
			tb.SetCell(i, 1, rune(0), 0, tb.ColorCyan)
		}
	}
	//remove the header
	s.Fields = s.Fields[1:]

	//calculate which fields should be displayed TODO fix this
	var first, last int
	if s.Selected < s.Height-3 {
		first = 0
		last = len(s.Fields)
	} else {
		first = s.Selected - ((s.Height - 3) / 2)

		//check if the end of the array is still off the bottom of the screen
		if s.Selected+((s.Height-3)/2) < len(s.Fields) {
			last = s.Selected + ((s.Height - 3) / 2)
		} else {
			last = len(s.Fields)
		}
	}

	//do the rest of the fields
	for row := range s.Fields[first:last] {
		if row == s.Selected {
			for i := range s.Fields[row] {
				if s.Colours[row] != tb.ColorWhite {
					tb.SetCell(i, row+2, rune(s.Fields[row][i]), s.Colours[row], tb.ColorWhite)
				} else {
					tb.SetCell(i, row+2, rune(s.Fields[row][i]), tb.ColorBlack, tb.ColorWhite)
				}
			}
		} else {
			for i := range s.Fields[row] {
				tb.SetCell(i, row+2, rune(s.Fields[row][i]), s.Colours[row], 0)
			}
		}
	}

	//clear the rest of the screenhttp://reddit.com/http://reddit.com/http://reddit.com/
	for row := last + 1; row < s.Height; row++ {
		for col := 0; col < s.Width; col++ {
			tb.SetCell(col, row, 0, 0, 0)
		}
	}

	helpMsg := "Close [Ctrl-Q]	Change Selection [Up/Down]	Refresh (temporary) [Space] Add Stock [Crtl-A]"

	s.drawHelp(helpMsg)

	tb.Flush()

}

func (s *Screen) setTicker(tick *Ticker) {
	s.Fields = make([]string, 0)

	s.Fields = append(s.Fields, "SYMBOL      PRICE       VOLUME(m)   OPEN        CLOSE       HIGH        LOW         CHANGE      MARKETCAP   52WKHIGH    52WKLOW     YTDCHANGE")

	for k := range tick.FIGURES_LIST {

		//calculate the spaces and format the field
		v := reflect.ValueOf(&tick.FIGURES_LIST[k]).Elem()
		var field string

		for i := 0; i < v.NumField(); i++ {
			var strToAdd string
			switch v.Field(i).Interface().(type) {
			case float64:
				strToAdd += fmt.Sprintf("%.2f", v.Field(i).Interface())
			case int:
				strToAdd += fmt.Sprintf("%d", v.Field(i).Interface())
			case string:
				strToAdd += v.Field(i).Interface().(string)
			}
			//fix for the negative values being out of line
			field += strToAdd + strings.Repeat(" ", 12-len(strToAdd))
		}

		//add it to the list
		s.Fields = append(s.Fields, field)
		s.Colours = append(s.Colours, tick.FIGURES_LIST[k].Colour) //this is stupid

	}

	s.drawTicker()
}

func (s *Screen) displayAddMenu() {
	s.refreshSize()
	tb.Clear(0, 0)

	menuText := "Enter the name (symbol) of the stock to add:"

	for i := 0; i < len(menuText); i++ {
		tb.SetCell((s.Width-len(menuText))/2+i, s.Height/2-1, rune(menuText[i]), tb.ColorBlack, tb.ColorWhite)
	}

	s.drawHelp("Cancel [ESC]")

	tb.Flush()
}

func takeInput() (string, bool) {
	symbol := ""
	for {
		//poll for the key entry, breaking on esc
		switch ev := tb.PollEvent(); ev.Type {
		case tb.EventKey:
			if ev.Key == tb.KeyEsc {
				return symbol, false
			}
			if ev.Key == tb.KeyEnter {
				return strings.ToUpper(symbol), true
			}
			if ev.Key == tb.KeyBackspace2 && len(symbol) > 0 {
				width, height := tb.Size()
				symbol = symbol[:len(symbol)-1]

				tb.SetCell(width/2-23+len(symbol)+1, height/2, 0, 0, 0)
				tb.Flush()
			} else {
				width, height := tb.Size()

				symbol += string(ev.Ch)
				tb.SetCell(width/2-23+len(symbol), height/2, ev.Ch, 0, 0)
				tb.Flush()
			}
		case tb.EventResize:
			continue

		default:
			return symbol, false
		}
	}
}

func (tick *Ticker) addStock(s *Screen) {
	s.displayAddMenu()

	symb, complete := takeInput()

	if !complete {
		return
	}
	logString("Before")
	logString(fmt.Sprintln(tick.STOCKS))

	tick.STOCKS = append(tick.STOCKS, symb)
	sort.Strings(tick.STOCKS)

	logString("After")
	logString(fmt.Sprintln(tick.STOCKS))
}
