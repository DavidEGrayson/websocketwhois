package main

import (
  "regexp"
  "errors"
  "fmt"
)

var notExistPatterns, existPatterns patternSet

func initData() {
  notExistPatterns = newPatternSet([]string {
    "(?i)no entries found",
    "(?i)no matching record",
  })

  existPatterns = newPatternSet([]string {
    "domain +name: +(.+)/i",
  })
}

func analyzeNotExistResponse(r queryResult) (*regexp.Regexp, error) {
  notExistScore := notExistPatterns.score(r)
  existScore := existPatterns.score(r)

  if (notExistScore.MatchCount == 1 && existScore.MatchCount == 0) {
    // Totally unambiguous success.  This is a not-exist reponse.
    return notExistScore.FirstMatchedPattern, nil
  }

  msg := fmt.Sprintf("Expected response to indicate domain non-existence, but it did not (%d,%d): %s", notExistScore.MatchCount, existScore.MatchCount, r)

  return nil, errors.New(msg)
}