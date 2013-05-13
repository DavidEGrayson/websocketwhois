package main

import (
  "fmt"
  "../zonefile"
  "log"
)


func main() {
  fmt.Println("benchmarking...")
  file, err := zonefile.Open("../data/org.zone")
  if err != nil {
    log.Fatal(err);
  }
  file.Find("graysonfamily.org")
}