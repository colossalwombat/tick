SRCS=tick.go main.go draw.go chart.go 

all: ${SRCS}
	touch tick
	rm tick
	go build -o tick ${SRCS}


