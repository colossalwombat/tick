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
	first := 0
	last := len(s.Fields)
	/*
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
		}*/

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

func (s *Screen) displaySymbolNotFound() {
	s.refreshSize()

	text := "SYMBOL NOT FOUND"

	for i := 0; i < len(text); i++ {
		tb.SetCell((s.Width-len(text))/2+i, s.Height/2+1, rune(text[i]), tb.ColorBlack, tb.ColorWhite)
	}
	tb.Flush()

}

func (s *Screen) takeInput(tick *Ticker) (string, bool) {
	symbol := ""
	for {
		//poll for the key entry, breaking on esc
		switch ev := tb.PollEvent(); ev.Type {
		case tb.EventKey:
			if ev.Key == tb.KeyEsc {
				return symbol, false
			}
			if ev.Key == tb.KeyEnter {
				if tick.verifySecurityExists(symbol) {
					return strings.ToUpper(symbol), true
				} else {
					s.displaySymbolNotFound()
				}

			}
			if ev.Key == tb.KeyBackspace2 && len(symbol) > 0 {
				symbol = symbol[:len(symbol)-1]
				tb.SetCell(s.Width/2-23+len(symbol)+1, s.Height/2, 0, 0, 0)
				tb.Flush()
			} else {
				symbol += string(ev.Ch)
				tb.SetCell(s.Width/2-23+len(symbol), s.Height/2, ev.Ch, 0, 0)
				tb.Flush()
			}
		case tb.EventResize:
			continue

		default:
			return symbol, false
		}
	}
}

func (s *Screen) printTextCentered(text string, height, padding int, fg, bg tb.Attribute) {
	//print the padding
	for i := 0; i < padding; i++ {
			tb.SetCell((s.Width+len(text))/2+i, height, 0, 0, bg)
		}
	//dont even ask
	if len(text) % 2 == 1 {
		padding++
	}
	
	for i := 0; i < padding; i++ {
		tb.SetCell((s.Width-len(text))/2-i, height, 0, 0, bg)
	}
	
	//print the text
	for i := 0; i < len(text); i++ {
		tb.SetCell((s.Width-len(text))/2+i, height, rune(text[i]), fg, bg)
	}
}

func (s *Screen) displayChartMenu(symbol string, selected int) {
	s.refreshSize()
	tb.Clear(0, 0)

	//draw the header
	headerText := fmt.Sprintf("Chart for %s", symbol)
	s.printTextCentered(headerText, s.Height/2-5, (24-len(headerText))/2, tb.ColorBlack, tb.ColorCyan)

	//draw the options
	options := []string{"One Day", "One Month", "Three Months", "Six Months", "Year To Date", "One Year", "Two Years", "Five Years"}

	for option := range options {
		if option == selected {
			s.printTextCentered(options[option], s.Height/2-(4-option), (24-len(options[option]))/2, tb.ColorRed, tb.ColorWhite)

		} else {
			s.printTextCentered(options[option], s.Height/2-(4-option), (24-len(options[option]))/2, tb.ColorWhite, tb.ColorRed)
		}
	}

	tb.Flush()
}

func (s *Screen) chartMenuHandler(symbol string) {
	selected := 0
	s.displayChartMenu(symbol, selected)
	for {
		switch ev := tb.PollEvent(); ev.Type {
		case tb.EventKey:
			if ev.Key == tb.KeyArrowUp && selected > 0 {
				selected--
				s.displayChartMenu(symbol, selected)
			}
			if ev.Key == tb.KeyArrowDown && selected < 7 {
				selected++
				s.displayChartMenu(symbol, selected)
			}
			if ev.Key == tb.KeyEsc {
				return
			}
		case tb.EventResize:
			s.displayChartMenu(symbol, selected)
		}
	}

}

func (tick *Ticker) addStock(s *Screen) bool {
	s.displayAddMenu()

	symb, complete := s.takeInput(tick)

	if !complete {
		return false
	}

	tick.STOCKS = append(tick.STOCKS, symb)
	sort.Strings(tick.STOCKS)

	return true

}
