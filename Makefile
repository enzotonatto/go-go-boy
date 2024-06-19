.PHONY: all build

all: build

go.mod:
	go mod init game
	go get github.com/nsf/termbox-go

build: go.mod
	cd server && go build -o ../bin/server server.go
	cd client && go build -o ../bin/client client.go
	
clean:
	rm -rf bin

distclean: clean
	rm -f go.mod go.sum
