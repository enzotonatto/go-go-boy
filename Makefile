.PHONY: all build

all: build

go.mod:
	go mod init game
	go get github.com/nsf/termbox-go

build: go.mod
	go build game.go
	
clean:
	rm -f game

distclean: clean
	rm -f go.mod go.sum
