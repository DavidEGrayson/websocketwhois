package data

import (
  "encoding/json"
  "regexp"
)

type JsonRegexp regexp.Regexp

func (r *JsonRegexp) UnmarshalJSON(input []byte) error {
  var str *string
  err := json.Unmarshal(input, str)
  if err != nil { return err }

  regexp, err := regexp.Compile(*str)
  if err != nil { return err }

  *r = JsonRegexp(*regexp)

  return nil
}