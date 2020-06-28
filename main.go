package main

import (
  "fmt"
  "log"
  "net/http"

  flag "github.com/spf13/pflag"
)

func main() {
  port := flag.StringP("port", "p", "8080", "Port to serve on")
  flag.Parse()

  arguments := flag.Args()
	directory := arguments[0]
  
  fs := http.FileServer(http.Dir(directory))
  http.Handle("/", fs)

  fmt.Printf("Serving %s on port %s \n", directory, *port)
  log.Fatal(http.ListenAndServe("localhost:" + *port, nil))
}