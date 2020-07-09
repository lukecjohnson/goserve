package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
	version := flag.BoolP("version", "v", false, "Display the current version of serve")

	// Override default usage function
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

	// Parse flags
	flag.Parse()

	// Parse non-flag arguments
	args := flag.Args()

	// Check that no more than 1 argument was provided
	if len(args) > 1 {
		printError("please provide just one directory path")
		os.Exit(1)
	}

	// Set directory to serve (defaults to current directory if no argument was provided)
	dir := "."
	if len(args) != 0 {
		dir = args[0]

		// Ensure that the provided directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			printError("the provided directory could not be found")
			os.Exit(1)
		}
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

	// Check for --version flag (print version and exit)
	if *version {
		fmt.Println(currentVersion)
		os.Exit(0)
	}

	// Check for --open flag (open browser window)
	if *open {
		if err := openBrowser(url); err != nil {
			printError("unable to open browser: " + err.Error())
			os.Exit(1)
		}
	}

	// Check for --single flag and set the http.FileSystem root with the appropriate wrapper
	var root http.FileSystem
	if *single {
		root = spaRoot{http.Dir(dir)}
	} else {
		root = htmlRoot{http.Dir(dir)}
	}

	// Wrap file server in handler with custom 404 handling
	fs := notFoundHanlder(http.FileServer(root), dir)

	// Register handler for all routes and add middleware for setting response headers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check for --cors flag (set "Access-Control-Allow-Origin" header)
		if *cors {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// Set "Cache-Control" header (defaults to "no-store" if max-age option wasn't provided)
		if *maxAge != "" {
			w.Header().Set("Cache-Control", "max-age="+*maxAge)
		} else {
			w.Header().Set("Cache-Control", "no-store")
		}

		// Pass control to next handler
		fs.ServeHTTP(w, r)
	})

	// If dir is the current working directory, display the directory name instead of "."
	displayDir := dir
	if displayDir == "." {
		wd, _ := os.Getwd()
		displayDir = filepath.Base(wd)
	}

	// Print "Serving {directory} at {url}" message
	fmt.Printf("\n\033[1;32m•\033[0m Serving \033[4m%s\033[0m at \033[4m%s\033[0m\n\n", displayDir, url)

	// Start server
	if secure {
		log.Fatal(http.ListenAndServeTLS(addr, *cert, *key, nil))
	} else {
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}

// http.Dir wrapper for standard static sites
type htmlRoot struct {
	http.Dir
}

// Wrapper for the Open method of htmlRoot's embedded http.Dir to handle clean urls
func (d htmlRoot) Open(name string) (http.File, error) {
	// Try opening requested file
	f, err := d.Dir.Open(name)

	// Check if requested file exists and if an extension was included
	if os.IsNotExist(err) && filepath.Ext(name) == "" {
		// Return file with .html extention if it exists
		if f, err := d.Dir.Open(name + ".html"); err == nil {
			return f, nil
		}
	}

	// Return original requested file or error
	return f, err
}

// http.Dir wrapper for single page applications
type spaRoot struct {
	http.Dir
}

// Wrapper for the Open method of spaRoot's embedded http.Dir to handle rewriting urls to index.html
func (d spaRoot) Open(name string) (http.File, error) {
	// Check that the requested file doesn't have an extension and isn't already index.html
	if filepath.Ext(name) == "" && name != "/" {
		// Return index.html instead of original requested file
		return d.Dir.Open("/index.html")
	}

	// Return original requested file or error
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

// Launches a browser window to a specified url
func openBrowser(url string) error {
	var cmd string
	var args []string

	// Set platform-specific command (w/ args) for opening a browser window
	if runtime.GOOS == "darwin" {
		cmd = "open"
	} else if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "start"}
	} else if runtime.GOOS == "linux" {
		cmd = "xdg-open"
	} else {
		return errors.New("unsupported platform")
	}

	// Add url as the last argument
	args = append(args, url)

	// Execute command
	return exec.Command(cmd, args...).Start()
}

// Prints formatted error message ("• Error: {message}")
func printError(msg string) {
	fmt.Printf("\n\033[1;31m•\033[0m Error: %s\n\n", msg)
}
