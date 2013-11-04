brog
====

Static blog app.

Usage
-----

Install `brog`.
```bash
go get github.com/aybabtme/brog
```

Use `brog`.
```bash
cd your/blog/path
# Creates brog files, required to run the brog
brog init
# Creates a new brog post named my_post.md
brog create my post
# Starts serving the brog at current location.
brog server
```

Config
------

Look at the `brog_config.json` file, it should be pretty clear.

Development
-----------

You need to [have `GOPATH` properly setup](http://golang.org/doc/code.html#GOPATH).
Additionally, you need to have `$GOPATH/bin` in your `PATH`, because some tools
in the build scripts really on the use of go tools.

To install `brog` on your system:

```
make install
```

To build `brog` for the first time (not required if you `make install`):

```bash
make configure
make
```

To do a normal build:

```
make
```

To remove artifacts resulting in a build/run at the root folder of `brog`:

```
make clean
```
