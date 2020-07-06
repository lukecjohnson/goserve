package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"

	flag "github.com/spf13/pflag"
)

// Dynamically set at build time to most recent git tag
var currentVersion = "DEV"

func main() {
	// Command line flags
	port := flag.StringP("port", "p", "8080", "Port to bind (Default: 8080)")
	host := flag.StringP("host", "h", "localhost", "Hostname to bind (Default: localhost)")
	cert := flag.StringP("cert", "c", "", "Path to SSL certificate")
	key := flag.StringP("key", "k", "", "Path to the SSL certificate's private key")
	maxAge := flag.StringP("max-age", "m", "", "Set the max age for resources in seconds (Cache-Control: max-age=<seconds>)")
	cors := flag.BoolP("cors", "C", false, "Enable CORS (Access-Control-Allow-Origin: *)")
	single := flag.BoolP("single", "s", false, "Serve as single page application")
	open := flag.BoolP("open", "o", false, "Open browser window")
	version := flag.BoolP("version", "v", false, "Display the current version of goserve")

	// Override default usage function
	flag.Usage = func() {
		flag.CommandLine.SortFlags = false

		fmt.Printf("\n%s\n\n", "Usage:")
		fmt.Println("  goserve [directory] [options]")

		fmt.Printf("\n\n%s\n\n", "Options:")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Printf("  -%s, --%s\t\t%s\n\n", f.Shorthand, f.Name, f.Usage)
		})

		os.Exit(0)
	}

	// Parse flags
	flag.Parse()

	// Parse non-flag arguments
	args := flag.Args()

	// Set directory to serve (defaults to current directory)
	dir := "."
	if len(args) != 0 {
		dir = args[0]
	}

	// Check if both cert and key options were provided
	secure := *cert != "" && *key != ""

	// Set address (host + port)
	addr := *host + ":" + *port

	// Set url (protocol + host + port)
	url := "http://" + addr
	if secure {
		url = "https://" + addr
	}

	// Check for --version flag
	if *version {
		fmt.Println(currentVersion)
		os.Exit(0)
	}

	// Check for --open flag
	if *open {
		var openBrowserCmd string
		var openBrowserArgs []string

		if runtime.GOOS == "darwin" {
			openBrowserCmd = "open"
		} else if runtime.GOOS == "windows" {
			openBrowserCmd = "cmd"
			openBrowserArgs = []string{"/c", "start"}
		} else if runtime.GOOS == "linux" {
			openBrowserCmd = "xdg-open"
		} else {
			fmt.Println("Error: Unsupported platform")
			os.Exit(1)
		}

		openBrowserArgs = append(openBrowserArgs, url)

		if err := exec.Command(openBrowserCmd, openBrowserArgs...).Start(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// Wrap file server in handler with custom 404 handling
	fs := wrapHandler(http.FileServer(http.Dir(dir)), dir)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if *cors {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		if *maxAge != "" {
			w.Header().Set("Cache-Control", "max-age="+*maxAge)
		} else {
			w.Header().Set("Cache-Control", "no-store")
		}

		if path.Ext(r.URL.Path) == "" {
			if *single {
				r.URL.Path = "/"
			} else {
				if _, err := os.Stat(dir + r.URL.Path); os.IsNotExist(err) {
					r.URL.Path += ".html"
				}
			}
		}

		fs.ServeHTTP(w, r)
	})

	fmt.Printf("Serving %s at %s \n", dir, url)

	// Start server
	if secure {
		log.Fatal(http.ListenAndServeTLS(addr, *cert, *key, nil))
	} else {
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}

type resourceNotFoundResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *resourceNotFoundResponseWriter) WriteHeader(status int) {
	w.status = status
	if status != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *resourceNotFoundResponseWriter) Write(p []byte) (int, error) {
	if w.status != http.StatusNotFound {
		return w.ResponseWriter.Write(p)
	}
	return len(p), nil
}

func wrapHandler(h http.Handler, root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writer := &resourceNotFoundResponseWriter{ResponseWriter: w}

		h.ServeHTTP(writer, r)

		if writer.status == http.StatusNotFound {
			handle404(w, root)
		}
	}
}

func handle404(w http.ResponseWriter, root string) {
	content := []byte(resourceNotFoundTemplate)

	custom404Page := path.Join(root, "404.html")
	if _, err := os.Stat(custom404Page); !os.IsNotExist(err) {
		content, err = ioutil.ReadFile(custom404Page)
		if err != nil {
			log.Fatal(err)
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write(content)
}
