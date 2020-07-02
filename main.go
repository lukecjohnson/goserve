package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"

	flag "github.com/spf13/pflag"
)

// Version : Dynamically set at build time to most recent git tag
var Version = "DEV"

func main() {
	port := flag.StringP("port", "p", "8080", "Port to serve on")
	host := flag.StringP("host", "h", "localhost", "Hostname to serve on")
	cert := flag.StringP("cert", "c", "", "Path to SSL certificate")
	key := flag.StringP("key", "k", "", "Path to the SSL certificate's private key")
	single := flag.BoolP("single", "s", false, "Serve as single page application")
	open := flag.BoolP("open", "o", false, "Open browser window")
	version := flag.BoolP("version", "v", false, "Prints the current version of goserve")

	flag.Parse()

	arguments := flag.Args()

	directory := "."
	if len(arguments) != 0 {
		directory = arguments[0]
	}

	secure := *cert != "" && *key != ""

	address := *host + ":" + *port

	url := "http://" + address
	if secure {
		url = "https://" + address
	}

	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}

	if *open {
		if err := openBrowser(url); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fs := http.FileServer(http.Dir(directory))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if path.Ext(r.URL.Path) == "" {
			if *single {
				r.URL.Path = "/"
			} else {
				if _, err := os.Stat(directory + r.URL.Path); os.IsNotExist(err) {
					r.URL.Path += ".html"
				}
			}
		}

		fs.ServeHTTP(w, r)
	})

	fmt.Printf("Serving %s at %s \n", directory, url)

	if secure {
		log.Fatal(http.ListenAndServeTLS(address, *cert, *key, nil))
	} else {
		log.Fatal(http.ListenAndServe(address, nil))
	}
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
