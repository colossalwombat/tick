# tick
Terminal based stock watching app, capable of tracking a user-modifiable list of securities and charting their performance over a defined interval. Still a work in progress, but mostly usable if you just want to check up on your portfolio throughout the day.

![There should be screenshot here](https://github.com/colossalwombat/tick/blob/master/screenshot.png 'Screenshots are always better than words')

(Volume shows 0 because the market was closed)

## TODO
- Fix the charting for shorter time periods (< one year)
- Improve the handling of API requests
- Handle network errors better (less explosive)
- Make the chart less ugly

## Dependencies
Depends on:
https://github.com/nsf/termbox-go

https://github.com/levigross/grequests

https://github.com/davecgh/go-spew

Use the latest everything and you should be good.

Can be installed with:
`go get -u github.com/nsf/termbox-go github.com/levigross/grequests github.com/davecgh/go-spew`

## License
Licensed under the MIT License

