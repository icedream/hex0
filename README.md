# hex0 Go implementation

This is my own implementation of the hex0 minimalist assembler tool used for
full-source bootstrapping.

It is neither the most optimal nor the most trustable implementation, however it
is another alternative to the [existing hex0 seed](https://github.com/oriansj/bootstrap-seeds)
that can be ported to multiple platforms.

**Do not just trust anything, do your own verifications.**

## Building

The `Makefile` will download a specific version of [TinyGo](https://tinygo.org),
extract it and use it to compile the binary in a size-optimized way.

To build debug and optimized binaries, run `make`.

To just build the optimized hex0 binary, run `make hex0`.

Of course, the above can also be done by hand, just read into the `Makefile` for
that.

The resulting `hex0` binary will be 66520 bytes in size.

## Usage

`hex0` works in all modes implemented by the C implementation. That means you
can feed it hex0 source code via standard input pipe or via path as the first
argument, as well as specify an output file as second argument or skip it to
output to standard output pipe.
