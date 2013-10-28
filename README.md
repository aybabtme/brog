brog
====

Static blog app.

Usage
-----

```bash
go build

./brog -port=8080         # Listens on port 8080
BROG_PORT=9000 ./brog     # Listens on port 9000
./brog                    # Listens on DefaultPort
```

Config
------

Set the `BROG_PORT` variable to change which port Brog listens on by default.
Set the `port` flag to achieve the same effect while overriding the value of `
BROG_PORT`.
