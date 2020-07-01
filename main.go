package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	flag "github.com/spf13/pflag"
)

// Version : Dynamically set at build time to most recent git tag
var Version = "DEV"

type dir struct {
	http.Dir
}

func (d dir) Open(name string) (http.File, error) {
	f, err := d.Dir.Open(name)
	if os.IsNotExist(err) {
		if f, err := d.Dir.Open(name + ".html"); err == nil {
			return f, nil
		}
	}
	return f, err
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	if runtime.GOOS == "darwin" {
		cmd = "open"
	} else if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "start"}
	} else if runtime.GOOS == "linux" {
		cmd = "xdg-open"
	} else {
		return errors.New("Error: Unsupported platform")
	}

	args = append(args, url)

	return exec.Command(cmd, args...).Start()
}

func main() {
	port := flag.StringP("port", "p", "8080", "Port to serve on")
	cert := flag.StringP("cert", "c", "", "Path to SSL certificate")
	key := flag.StringP("key", "k", "", "Path to the SSL certificate's private key")
	open := flag.BoolP("open", "o", false, "Open browser window")
	version := flag.BoolP("version", "v", false, "Prints the current version of goserve")

	flag.Parse()

	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}

	arguments := flag.Args()

	directory := "."
	if len(arguments) != 0 {
		directory = arguments[0]
	}

	fs := http.FileServer(dir{http.Dir(directory)})
	http.Handle("/", fs)

	secure := *cert != "" && *key != ""

	url := "http://localhost:" + *port
	if secure {
		url = "https://localhost:" + *port
	}

	if *open {
		if err := openBrowser(url); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Printf("Serving %s at %s \n", directory, url)

	if secure {
		log.Fatal(http.ListenAndServeTLS("localhost:"+*port, *cert, *key, nil))
	} else {
		log.Fatal(http.ListenAndServe("localhost:"+*port, nil))
	}
}
