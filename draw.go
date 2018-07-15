package main

import(
	"github.com/levigross/grequests"
	"fmt"
	"strings"
	tb "github.com/nsf/termbox-go"
	"os"
	"reflect"
	"sort"
	"time"
	"encoding/json"
)

func drawHelp(helpMsg string) {
	_, line := tb.Size()

	for i := 0; i < len(helpMsg); i++ {
		tb.SetCell(i, line - 1, rune(helpMsg[i]), tb.ColorWhite, tb.ColorBlue)
	}

}

func drawTicker(fields []string, colours []tb.Attribute, selected int){
	width, height := tb.Size()

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

	//calculate which fields should be displayed
	var first, last int
	if selected < height - 3 {
		first = 0
		last = len(fields)
	} else {
		first = selected - ((height - 3) / 2)

		//check if the end of the array is still off the bottom of the screen
		if selected + ((height - 3) / 2) < len(fields){
			last = selected + ((height - 3) / 2)
		} else {
			last = len(fields)
		}
	}


	//do the rest of the fields
	for row := range fields[first:last]{
		if row == selected {
			for i := range fields[row]{
				if colours[row] != tb.ColorWhite {
					tb.SetCell(i, row + 2, rune(fields[row][i]), colours[row], tb.ColorWhite)
				} else {
					tb.SetCell(i, row + 2, rune(fields[row][i]), tb.ColorBlack, tb.ColorWhite)
				}
			}
		} else{
			for i := range fields[row]{
				tb.SetCell(i, row + 2, rune(fields[row][i]), colours[row], 0)
			} 
		}
	}

	helpMsg := "Close [Ctrl-Q]	Change Selection [Up/Down]	Refresh (temporary) [Space] Add Stock [Crtl-A]"

	drawHelp(helpMsg)

	tb.Flush()


	return
}


func setTicker(list []stockFigures, selected int){

	
	if !(reflect.DeepEqual(oldPV, list)) {
		oldPV = list
	}

	fields := []string{}
	colours := []tb.Attribute{}

	fields = append(fields, "SYMBOL      PRICE       VOLUME(m)   OPEN        CLOSE       HIGH        LOW         CHANGE      MARKETCAP   52WKHIGH    52WKLOW     YTDCHANGE")

	for s := range list{
		/*space := strings.Repeat(" ", 12)
		width,_ := tb.Size()

		//calculate the space sizes
		lineSpaces := width - (len(space)*2 + len(fmt.Sprintf("%d", list[s].volume))) - 3
		space1 := space[:len(space) - len(list[s].symbol)]
		space2 := space[:len(space) - len(fmt.Sprintf("%.2f", list[s].price))]
		space3 := space[:len(space) - len(fmt.Sprintf("%d", list[s].volume))]
		space4 := space[:len(space) - len(fmt.Sprintf("%.2f", list[s].open))]
		space5 := space[:len(space) - len(fmt.Sprintf("%.2f", list[s].close))]
		spaceEnd := strings.Repeat(" ", lineSpaces)

		//format the field
		field := fmt.Sprintf("%s%s%.2f%s%d%s%.2f%s%.2f%s%.2f%s", list[s].symbol, space1, list[s].price, space2, list[s].volume, space3, list[s].open, space4, list[s].close, space5, list[s].change, spaceEnd)*/


		//calculate the spaces and format the field
		v := reflect.ValueOf(&list[s]).Elem()
		var field string

		for i := 0; i < v.NumField(); i++{
			var strToAdd string
			switch v.Field(i).Interface().(type) {
			case float64:
				strToAdd += fmt.Sprintf("%.2f", v.Field(i).Interface())
			case int:
				strToAdd += fmt.Sprintf("%d", v.Field(i).Interface())
			case string:
				strToAdd += v.Field(i).Interface().(string)
			}
			field += strToAdd + strings.Repeat(" ", 12 - len(strToAdd))
		}

		//add it to the list
		fields = append(fields, field)
		colours = append(colours, list[s].Colour)

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

	STOCKS = append(STOCKS, symb)
	sort.Strings(STOCKS)
}
