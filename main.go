package main

import (
  "fmt"
  "log"
  "net/http"
)

func main() {
  fs := http.FileServer(http.Dir("."))
  http.Handle("/", fs)

  fmt.Println("Server running on port 8080")
  log.Fatal(http.ListenAndServe(":8080", nil))
}