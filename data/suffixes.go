package data

import (
  "io/ioutil"
  "encoding/json"
)

type Suffix struct {
  ExampleExistingDomain string
}

func SuffixesRead() (map[string]Suffix, error) {
  suffixes := make(map[string]Suffix)
  
  suffixes_bytes, err := ioutil.ReadFile(Directory + "/suffixes.json")
  if err != nil { return nil, err }

  err = json.Unmarshal(suffixes_bytes, &suffixes)
  if err != nil { return nil, err }

  return suffixes, nil
}