package main

import (
	"fmt"
	tb "github.com/nsf/termbox-go"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Screen struct {
	ColourFields  []ColourField
	Width, Height int
}

type ColourField struct {
	Field                  string
	Foreground, Background tb.Attribute
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

func (s *Screen) drawTicker(tick *Ticker) {
	s.refreshSize()

	//set the header
	title := "Tick"
	char := 0
	for i := 0; i < s.Width; i++ {
		//center and print the title letters
		if i > (s.Width-len(title))/2 && char < len(title) {
			tb.SetCell(i, 0, rune(title[char]), tb.ColorWhite, 0)
			char++
		} else {
			tb.SetCell(i, 0, rune(0), 0, 0)
		}
	}
	for i := 0; i < s.Width; i++ {
		if i < len(s.ColourFields[0].Field) {
			tb.SetCell(i, 1, rune(s.ColourFields[0].Field[i]), tb.ColorBlack, tb.ColorCyan)
		} else {
			tb.SetCell(i, 1, rune(0), 0, tb.ColorCyan)
		}
	}
	//remove the header
	s.ColourFields = s.ColourFields[1:]

	//calculate which fields should be displayed
	var first, last int

	if tick.Selected-(s.Height/2)-1 > 0 {
		first = tick.Selected - (s.Height / 2) - 1
	} else {
		first = 0
	}

	if s.Height-3+first < len(s.ColourFields) {
		last = s.Height - 3 + first
	} else {
		last = len(s.ColourFields)
	}

	s.ColourFields = s.ColourFields[first:last]

	for row := range s.ColourFields {
		for i := range s.ColourFields[row].Field {
			tb.SetCell(i, row+2, rune(s.ColourFields[row].Field[i]), s.ColourFields[row].Foreground, s.ColourFields[row].Background)
		}
	}

	//clear the rest of the screen
	for row := len(s.ColourFields) + 1; row < s.Height; row++ {
		for col := 0; col < s.Width; col++ {
			tb.SetCell(col, row, 0, 0, 0)
		}
	}

	helpMsg := "Close [Ctrl-Q]	Change Selection [Up/Down]	Refresh (temporary) [Space] Add Stock [Crtl-A]"

	s.drawHelp(helpMsg)

	//only for debugging
	sel := strconv.Itoa(tick.Selected)
	if tick.Selected > 9 {
		tb.SetCell(101, 0, rune(sel[1]), tb.ColorWhite, 0)
		tb.SetCell(100, 0, rune(sel[0]), tb.ColorWhite, 0)
	} else {
		tb.SetCell(101, 0, rune(sel[0]), tb.ColorWhite, 0)
	}

	tb.Flush()

}

func (s *Screen) setTicker(tick *Ticker) {
	s.ColourFields = make([]ColourField, 0)

	s.ColourFields = append(s.ColourFields, ColourField{"SYMBOL      PRICE       VOLUME(m)   OPEN        CLOSE       HIGH        LOW         CHANGE      MARKETCAP   52WKHIGH    52WKLOW     YTDCHANGE", 0, 0})

	for k := range tick.Figures_List {

		//calculate the spaces and format the field
		v := reflect.ValueOf(&tick.Figures_List[k]).Elem()
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
		if k == tick.Selected {
			if tick.Figures_List[k].Colour != tb.ColorWhite {
				newCF := ColourField{field, tick.Figures_List[k].Colour, tb.ColorWhite}
				s.ColourFields = append(s.ColourFields, newCF)
			} else {
				newCF := ColourField{field, tb.ColorBlack, tb.ColorWhite}
				s.ColourFields = append(s.ColourFields, newCF)
			}
		} else {
			newCF := ColourField{field, tick.Figures_List[k].Colour, 0}
			s.ColourFields = append(s.ColourFields, newCF)
		}
	}
	s.drawTicker(tick)
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

func (s *Screen) displayAddingMessage(text string) {
	s.refreshSize()

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
				s.displayAddingMessage("CHECKING...")
				if tick.verifySecurityExists(symbol) {
					logString("Returning true from takeInput")
					return strings.ToUpper(symbol), true
				} else {
					s.displayAddingMessage("SYMBOL NOT FOUND")
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
	//rough attempt to fix an alignment issue, TODO fix this
	if len(text)%2 == 1 {
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
	selectedInterval := 0
	s.displayChartMenu(symbol, selectedInterval)
	for {
		switch ev := tb.PollEvent(); ev.Type {
		case tb.EventKey:
			if ev.Key == tb.KeyArrowUp && selectedInterval > 0 {
				selectedInterval--
				s.displayChartMenu(symbol, selectedInterval	)
			}
			if ev.Key == tb.KeyArrowDown && selectedInterval < 7 {
				selectedInterval++
				s.displayChartMenu(symbol, selectedInterval)
			}
			if ev.Key == tb.KeyEsc {
				return
			}
			if ev.Key == tb.KeyEnter{
				//show the chart
				s.chartHandler(symbol, selectedInterval)
				return
			}
		case tb.EventResize:
			s.displayChartMenu(symbol, selectedInterval)
		}
	}

}

func (s *Screen) showRefresh() {
	tb.SetCell(s.Width-1, 0, rune('e'), 0, tb.ColorWhite)
	tb.Flush()

	go func() {
		time.Sleep(500 * time.Millisecond)
		tb.SetCell(s.Width-1, 0, rune(0), 0, 0)
		tb.Flush()
	}()
}

func (tick *Ticker) addStock(s *Screen) bool {
	s.displayAddMenu()

	symb, complete := s.takeInput(tick)

	if !complete {
		return false
	}
	logString("Adding stock")

	tick.Stocks = append(tick.Stocks, symb)
	sort.Strings(tick.Stocks)

	return true

}
