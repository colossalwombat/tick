package main 
import (
	"github.com/levigross/grequests"
	"fmt"
	"strings"
	"time"
	"io/ioutil"
)


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

	for _, line := range strings.Split(strings.TrimSuffix(resp.String(), "\n"), "\n"){
		fmt.Printf("\r                                                                              ")
		fmt.Printf("\r %s",line)
		time.Sleep(100 * time.Millisecond)
	}

	
}
