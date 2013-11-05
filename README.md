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

Something you should know
-------------------------

By default, `brog` will regenerate any harm caused to its vital structure.
That is, when `brog server` runs, it watches its templates and reload them
at every change.

* If a vital template has been deleted or renamed, `brog` will regenerate the
template at its original location.
* If a vital template has been modified and `brog` find that it is
corrupted (fails to parse), `brog` will repeal the threat and rewrite the file
with its default version.

You can change this setting in `brog_config.json`.

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

License
-------

All of `brog` proper is MIT licensed.  See [`LICENSE.md`](LICENSE.md).

Some files bundled with `brog` have their own license. Their licenses can also be found in [`LICENSE.md`](LICENSE.md).  If you do not agree to their terms, feel free to remove the related parts.

[`highlight.js`](https://github.com/isagalaev/highlight.js) is copyright (c) 2006, Ivan Sagalaev
