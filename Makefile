PACKAGE:=github.com/squaremo/flumux
SUBCOMMANDS:=tag list

.PHONY: all clean

all: flumux

clean:
	rm -rf build flumux

flumux: *.go
	mkdir -p build/src/$(PACKAGE)
	cp -R *.go vendor build/src/$(PACKAGE)/
	GOPATH=`pwd`/build go build -o $@ build/src/$(PACKAGE)/*.go
