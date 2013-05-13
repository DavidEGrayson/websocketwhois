package main

import (
  "fmt"
  "../domainlist"
  "log"
)


func main() {
  fmt.Println("benchmarking...")
  file, err := domainlist.Open("data/org.domains")
  if err != nil {
    log.Fatal(err);
  }
  _, err = file.Find("graysonfamily.org")
  if err != nil {
    log.Fatal(err)
  }
}