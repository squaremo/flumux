PACKAGE:=github.com/squaremo/flumux

.PHONY: all clean

all: flumux

clean:
	rm -rf build flumux

flumux: main.go tag/*.go
	mkdir -p build/src/$(PACKAGE)
	cp -R main.go tag vendor build/src/$(PACKAGE)/
	GOPATH=`pwd`/build go build -o $@ build/src/$(PACKAGE)/main.go
