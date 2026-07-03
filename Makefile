.PHONY: all build build-cli build-gui run clean

all: build

build: build-cli build-gui

build-cli:
	go build -o anjal ./cmd/anjal

build-gui:
	cd gui && wails build

run: build-cli
	./anjal

clean:
	rm -f anjal
	rm -rf gui/build/bin/
