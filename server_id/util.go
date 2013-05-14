package main

import (
  "regexp"
)

func compileAll(s []string) []*regexp.Regexp {
  r := make([]*regexp.Regexp, len(s))
  for i, str := range s {
    r[i] = regexp.MustCompile(str)
  }
  return r
}

