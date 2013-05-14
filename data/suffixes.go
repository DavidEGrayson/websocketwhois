package data

import (
  "io/ioutil"
  "encoding/json"
)

/** TODO: move this to its own package (errorwrap) */
type errorWrap struct {
  Message string
  InnerError error
}
func New(message string, err error) error {
  return &errorWrap{message, err}
}
func (e *errorWrap) Error() string {
  return e.Message + "  " + e.InnerError.Error()
}

type Suffix struct {
  ExampleExistingDomain string
}

func SuffixesRead() (map[string]Suffix, error) {
  suffixes := make(map[string]Suffix)
  
  suffixes_bytes, err := ioutil.ReadFile(Directory + "/suffixes.json")
  if err != nil { return nil, err }

  err = json.Unmarshal(suffixes_bytes, &suffixes)
  if err != nil { return nil, New("Error decoding suffixes.json.", err) }

  return suffixes, nil
}