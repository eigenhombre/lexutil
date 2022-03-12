.PHONY: clean deps lint all compile

all:  deps compile lint

deps:
	go get .

compile:
	go build .

test:
	go test

lint:
	golint -set_exit_status .
	staticcheck .

clean:
	rm ${PROG}
