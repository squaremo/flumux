PACKAGE:=github.com/squaremo/flumux
SUBCOMMANDS:=tag list

.PHONY: all clean

all: flumux

clean:
	rm -rf build flumux

flumux: main.go tag/*.go $(foreach cmd,$(SUBCOMMANDS),$(wildcard $(cmd)/*.go))
	mkdir -p build/src/$(PACKAGE)
	cp -R main.go $(SUBCOMMANDS) vendor build/src/$(PACKAGE)/
	GOPATH=`pwd`/build go build -o $@ build/src/$(PACKAGE)/main.go
