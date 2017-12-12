BIN:=flumux
LDFLAGS:=

.PHONY: all clean

all: flumux

clean:
	rm -rf ${BIN}

${BIN}: *.go Gopkg.toml Gopkg.lock
	go build -ldflags="${LDFLAGS}" -o $@ .
