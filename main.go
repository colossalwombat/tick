package main

import (
	tb "github.com/nsf/termbox-go"
	"os"
)

func main() {
	//wipe the log file
	file, err := os.Create("log.txt")
	check(err)
	file.Close()

	//Initialize the UI
	tb.Init()
	defer tb.Close()

	tickerHandler()

}
