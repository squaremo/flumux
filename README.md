### A tool for asking about container images. An image asking tool.

## To install with Go

You need libgit2. On `apt`-packaged systems:

    apt-get install golang libgit2-dev

On MacOSX,

    brew install go libgit2

Then you can do

    go get github.com/squaremo/flumux

to fetch, build and install the binary. You can also build it by
cloning the repository -- see below.

## To use

Flumux is just a way of associating your container image builds with
git commits. When you build an image, you tag it with the current
state of the repository:

    $ docker build -t myimage:`flumux tag` .

Once these are pushed to an image registry, e.g., quay.io, you can
query for the images and match them to git commits. `flumux lookup`
takes input on stdin, and for each line it takes the first field (the
text before any whitespace), assumes it's a commit ID, and looks for
an image named after it. This is designed to work well with git
output, e.g.,

    $ git log --oneline | flumux lookup myimage

## To build in a cloned repository

You can also build it in the repository. The Go dependencies are
submodules; you need a

    git submodule update --init --recursive

to fetch those.

After that, a simple

    make

will build the binary, and put it in the top directory. You don't need
to use Go's funny path scheme, but if you do, it will still work
(albeit by ignoring your `$GOPATH`).
