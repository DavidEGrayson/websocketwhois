package main

import (
  "regexp"
)

var notExistPatterns []*regexp.Regexp
var existPatterns []*regexp.Regexp

func initData() {
  notExistStrings := []string {
    "(?i)no entries found",
    "(?i)no matching record",
  }
  notExistPatterns = compileAll(notExistStrings)

  existStrings := []string {
    "domain +name: +(.+)/i",
  }
  existPatterns = compileAll(existStrings)
}

var bytelist = []byte {
  'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
  'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
  // Do NOT include '-' in this list because we are tryin to generate a domain
  // name that probably does not exist but would be valid, and a hyphen
  // in certain spots is not allowed.
}

