package main

import (
  "regexp"
  "errors"
  "fmt"
)

var notExistPatterns, existPatterns patternSet

func responseAnalysisInit() {
  notExistPatterns = newPatternSet([]string {
    `(?i)^no entries found.?$`,
    `(?i)^no matching record$`,
    `(?i)^domain (.+) not registe?red.$`,
    `(?i)^% no entries found for the selected source\(s\).$`,
    `(?i)^% object (\S+) not found.$`,
    `(?i)^object does not exist$`,
    `(?i)^no entries found in the \.\S+ database`,
  })

  existPatterns = newPatternSet([]string {
    `(?i)^domain +name: *(\S+)\s*$`,
    `(?i)^domain\s*:\s*(\S+)\s*$`,
    `(?i)^ *complete domain name\.+: *(\S+)\s*$`,
  })
}

func analyzeNotExistResponse(r queryResult) (*regexp.Regexp, error) {
  notExistScore := notExistPatterns.score(r)
  existScore := existPatterns.score(r)

  if (notExistScore.MatchCount == 1 && existScore.MatchCount == 0) {
    // Totally unambiguous success.  This is a not-exist reponse.
    return notExistScore.FirstMatchedPattern, nil
  }

  msg := fmt.Sprintf("Expected response to indicate domain non-existence, but it did not (%d,%d): %s", notExistScore.MatchCount, existScore.MatchCount, r.String())

  return nil, errors.New(msg)
}

func analyzeExistResponse(r queryResult, domain string) (*regexp.Regexp, error) {
  notExistScore := notExistPatterns.score(r)
  existScore := existPatterns.score(r)

  if (notExistScore.MatchCount == 0 && existScore.MatchCount == 1) {
    // Totally unambiguous success.  This is a not-exist reponse.
    return existScore.FirstMatchedPattern, nil
  }

  // TODO: make sure that the domain actually appears on the matching
  // line of the response!  Need some more functionality in patternSet

  msg := fmt.Sprintf("Expected response to indicate domain existence, but it did not (%d,%d): %s", notExistScore.MatchCount, existScore.MatchCount, r.String())

  return nil, errors.New(msg)
}