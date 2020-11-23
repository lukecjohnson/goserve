package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"
)

var currentVersion = "DEV"

func main() {
	port := flag.StringP("port", "p", "8080", "Port to bind (Default: 8080)")
	host := flag.StringP("host", "h", "localhost", "Hostname to bind (Default: localhost)")
	cert := flag.StringP("cert", "c", "", "Path to SSL certificate")
	key := flag.StringP("key", "k", "", "Path to the SSL certificate's private key")
	single := flag.BoolP("single", "s", false, "Serve as single page application")
	version := flag.BoolP("version", "v", false, "Display the current version of serve")

	flag.Usage = func() {
		flag.CommandLine.SortFlags = false

		fmt.Printf("\n%s\n\n", "Usage:")
		fmt.Println("  serve [directory] [options]")

		fmt.Printf("\n\n%s\n\n", "Options:")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Printf("  -%s, --%s\t\t%s\n\n", f.Shorthand, f.Name, f.Usage)
		})

		os.Exit(0)
	}

	flag.Parse()

	if *version {
		fmt.Println(currentVersion)
		os.Exit(0)
	}

	args := flag.Args()

	if len(args) > 1 {
		printError("please provide just one directory path")
		os.Exit(1)
	}

	dir := "."
	if len(args) != 0 {
		dir = args[0]

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			printError("the provided directory could not be found")
			os.Exit(1)
		}
	}

	secure := *cert != "" && *key != ""

	addr := *host + ":" + *port

	url := "http://" + addr
	if secure {
		url = "https://" + addr
	}

	var root http.FileSystem
	if *single {
		root = spaRoot{http.Dir(dir)}
	} else {
		root = htmlRoot{http.Dir(dir)}
	}

	fs := notFoundHanlder(http.FileServer(root), dir)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")

		fs.ServeHTTP(w, r)
	})

	displayDir := dir
	if displayDir == "." {
		wd, _ := os.Getwd()
		displayDir = filepath.Base(wd)
	}

	fmt.Printf("\n\033[1;32m•\033[0m Serving \033[4m%s\033[0m at \033[4m%s\033[0m\n\n", displayDir, url)

	if secure {
		log.Fatal(http.ListenAndServeTLS(addr, *cert, *key, nil))
	} else {
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}

type htmlRoot struct {
	http.Dir
}

func (d htmlRoot) Open(name string) (http.File, error) {
	f, err := d.Dir.Open(name)

	if os.IsNotExist(err) && filepath.Ext(name) == "" {
		if f, err := d.Dir.Open(name + ".html"); err == nil {
			return f, nil
		}
	}

	return f, err
}

type spaRoot struct {
	http.Dir
}

func (d spaRoot) Open(name string) (http.File, error) {
	if filepath.Ext(name) == "" && name != "/" {
		return d.Dir.Open("/index.html")
	}

	return d.Dir.Open(name)
}

// http.ResponseWriter wrapper
type notFoundResponseWriter struct {
	http.ResponseWriter
	status int
}

// Wrapper for the WriteHeader method of notFoundResponseWriter's embedded ResponseWriter
func (w *notFoundResponseWriter) WriteHeader(status int) {
	// Save status for later use
	w.status = status

	// Write header with status code unless status code is 404 - handle404 will write the header
	if status != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(status)
	}
}

// Wrapper for the Write method of notFoundResponseWriter's embedded ResponseWriter
func (w *notFoundResponseWriter) Write(p []byte) (int, error) {
	// Write response unless status code is 404 - handle404 will write a custom response
	if w.status != http.StatusNotFound {
		return w.ResponseWriter.Write(p)
	}

	// Fake that a response was written
	return len(p), nil
}

// Returns handler with middleware for intercepting 404 errors
func notFoundHanlder(h http.Handler, root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create an instance of notFoundResponseWriter
		writer := &notFoundResponseWriter{ResponseWriter: w}

		// Pass control to next handler
		h.ServeHTTP(writer, r)

		// Check if status code is 404
		if writer.status == http.StatusNotFound {
			// Call handle404 to write custom response
			handle404(w, root)
		}
	}
}

// Writes custom 404 response
func handle404(w http.ResponseWriter, root string) {
	// Set default response content
	content := []byte(resourceNotFoundTemplate)

	// Check if custom 404.html file exists
	custom404Page := filepath.Join(root, "404.html")
	if _, err := os.Stat(custom404Page); !os.IsNotExist(err) {
		// Read 404.html file
		content, err = ioutil.ReadFile(custom404Page)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Set "Content-Type" header for HTML response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Write header with 404 status code
	w.WriteHeader(http.StatusNotFound)

	// Write HTML response
	w.Write(content)
}

// Prints formatted error message ("• Error: {message}")
func printError(msg string) {
	fmt.Printf("\n\033[1;31m•\033[0m Error: %s\n\n", msg)
}
