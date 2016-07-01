### A tool for asking about container images. An image asking tool.

## To build

It's all written in Go, so you need that.

You also need libgit2. On `apt`-packaged systems:

    apt-get install libgit2-dev

On MacOSX,

    brew install libgit2

The Go dependencies are submodules; you need a

    git submodule update --init --recursive

to fetch those.

After that, a simple

    make

will "make" the binary, which it puts in the top directory. You don't
need to use Go's funny path scheme, but if you do, it will still work.

