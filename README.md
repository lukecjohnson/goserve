# `serve`

A quick, simple CLI for serving static sites and single page applications.

## Usage
By default, `serve` will serve the current directory at `localhost:8080`:
```
$ serve
```

A directory and additional options can be specified with `serve [directory] [options]`:
```
$ serve public --port 5000 --single
```

To see a list of all available options:
```
$ serve --help
```

## Installation
Currently, `serve` is only officially distributed for macOS via [Homebrew](https://brew.sh/). To install:
```
$ brew install lukecjohnson/packages/serve
```

For other platforms, prebuilt binaries can be downloaded directly from the [releases page](https://github.com/lukecjohnson/serve/releases).