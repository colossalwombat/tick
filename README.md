## tick
Terminal based stock watching app, capable of graphing all 30 companies on the DJIA. For know, you'll need to grab an API key from https://alphavantage.com and save it in a file with no extension called `key`. This will probably change in the future, if this project every becomes anything more than a novelty.

# Currently
- Pull data from 30 fixed companies
- Display graph of closing price over 20 years

# TODO
- Scrollable stock list
- Multigraphing
- Custom entries
- Fix flickering(OSX)

# Dependencies
Can be installed with:
`go get -u github.com/gizak/termui github.com/levigross/grequests`

