package main

import (
	tm "github.com/nsf/termbox-go"
	"os"
)

func main() {
	//wipe the log file
	file, err := os.Create("log.txt")
	check(err)
	file.Close()

	//Initialize the UI
	tm.Init()
	defer tm.Close()

	tickerHandler()

}
