package main

import (
  "fmt"
  "log"
  "net/http"

  flag "github.com/spf13/pflag"
)

type Dir struct {
  http.Dir
}

func (d Dir) Open(name string) (http.File, error) {
  f, err := d.Dir.Open(name)
  if os.IsNotExist(err) {
    if f, err := d.Dir.Open(name + ".html"); err == nil {
      return f, nil
    }
  }
  return f, err
}

func main() {
  port := flag.StringP("port", "p", "8080", "Port to serve on")
  flag.Parse()

  arguments := flag.Args()
  directory := arguments[0]
  
  fs := http.FileServer(Dir{http.Dir(directory)})
  http.Handle("/", fs)

  fmt.Printf("Serving %s on port %s \n", directory, *port)
  log.Fatal(http.ListenAndServe("localhost:" + *port, nil))
}