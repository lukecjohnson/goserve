# `goserve`

A quick, simple CLI for serving static sites and single page applications.

## Usage
By default, `goserve` will serve the current directory at `localhost:8080`
```
$ goserve
```

A directory and additional options can be specified with `goserve [directory] [options]`:
```
$ goserve public --port 5000
```

To see a list of all available options:
```
$ goserve --help
```