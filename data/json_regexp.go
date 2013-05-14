package data

import (
  "encoding/json"
  "regexp"
  "bytes"
)

type JsonRegexp regexp.Regexp

func (r *JsonRegexp) UnmarshalJSON(input []byte) error {
  decoder := json.NewDecoder(bytes.NewReader(input))

  var str string
  err := decoder.Decode(&str)
  if err != nil { return err }

  regexp, err := regexp.Compile(str)
  if err != nil { return err }

  *r = JsonRegexp(*regexp)

  return nil
}