package main

import (
  "regexp"
  "errors"
  "fmt"
)

var notExistPatterns, existPatterns patternSet

func responseAnalysisInit() {
  notExistPatterns = newPatternSet([]string {
    `(?i)^no entries found\.?$`,
    `(?i)^no matching record\.?$`,
    `(?i)^domain (\S+) not registe?red\.$`,
    `(?i)^% no entries found for the selected source\(s\)\.$`,
    `(?i)^%? ?object (\S+) not found.$`,
    `(?i)^object does not exist$`,
    `(?i)^no entries found in the \.\S+ database`,
    `(?i)^sorry, but domain: "(\S+)", not found in database`,
    `(?i)^domain not found$`,
    `(?i)^(\S+) is available\.$`,
    `(?i)^no entries found for the selected source\.$`,
    `(?i)^%error: no entries found$`,
    `(?i)^% no such domain$`,
    `(?i)^No information available about domain name (\S+) in the registry nask database.$`,
    `(?i)^(\S+) no match$`,
    `(?i)^(\S+) is free$`,
    `(?i)^% not registered - the domain you have requested (\S+) is not a registered \S+ domain name\.$`,
    `(?i)^key not found$`,
    `(?i)^no match found for (\S+)$`,
    `(?i)^the domain has not been registered\.$`,
    `(?i)^% This query returned 0 objects\.$`,
    `(?i)^domain (\S+) is free\.\s*$`,
    `(?i)^% no entries found for query "(\S+)"\.$`,
    `(?i)^% no data was found to match the request criteria\.$`,
    `(?i)^no such domain (\S+)$`,
    `(?i)^no match!!$`,
    `(?i)^above domain name is not registered to krnic\.$`,
  })

  existPatterns = newPatternSet([]string {
    `(?i)^domain +name\s*:\s*(\S+)\s*$`,
    `(?i)^domain\s*:\s*(\S+)\s*$`,
    `(?i)^ *complete domain name\.+: *(\S+)\s*$`,
    `(?i)^Domain is not available or is reserved by the registry\.$`,
    `(?i)^Nome de dom.nio / Domain Name: (\S+)$`,
    `(?i)^\[domain name\]\s*(\S+)$`,
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