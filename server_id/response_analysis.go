package main

import (
  "regexp"
  "errors"
  "fmt"
)

type patternScheme struct {
  notExistPatterns, existPatterns patternSet
}

var scheme1, scheme2 patternScheme

func responseAnalysisInit() {
  scheme1.notExistPatterns = newPatternSet([]string {
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
    `(?i)^no match$`,
    `(?i)^above domain name is not registered to krnic\.$`,
    `(?i)^domain "(\S+)" - available$`,
    `(?i)^"(\S+)" not found\.$`,
    `(?i)^not found: (\S+)$`,
    `(?i)^not found\.\.\.$`,
    `(?i)^% nothing found$`,
    `(?i)^we do not have an entry in our database matching your query\.$`,
    `(?i)^(\S+): no existe$`,
    `(?i)^no domain records were found to match`,
  })

  scheme1.existPatterns = newPatternSet([]string {
    `(?i)^\s*domain +name\s*:\s*(\S+)\s*$`,
    `(?i)^domain\s*:\s*(\S+)\s*$`,
    `(?i)^ *complete domain name\.+: *(\S+)\s*$`,
    `(?i)^domain is not available or is reserved by the registry\.$`,
    `(?i)^nome de dom.nio / domain name: (\S+)$`,
    `(?i)^\[domain name\]\s*(\S+)$`,
    `(?i)^domain "(\S+)" - not available$`,
  })

  scheme2.notExistPatterns = newPatternSet([]string {
    `(?i)^domain status:\s*available$`,
    `(?i)^status:\s*available$`,
    `(?i)^status\.?:\s*not found$`,
    `(?i)^status:\s*not registered$`,
    `(?i)^status: free$`,
  })

  scheme2.existPatterns = newPatternSet([]string {
    `(?i)^domain status:\s*registered$`,
    `(?i)^domain status:\s*ok$`,
    `(?i)^status\.?:\s*registered$`,
    `(?i)^status:\s*not available$`,
    `(?i)^status:\s*connect$`,
    `(?i)^status:\s*active$`,
  })
}

func analyzeResponsePair(notExistResponse, existResponse queryResult) (notExistRegexp *regexp.Regexp, existRegexp *regexp.Regexp, err error) {

  notExistRegexp, existRegexp, err1 := responsePairMatchesScheme(scheme1, notExistResponse, existResponse) 
  if err1 == nil { return }

  notExistRegexp, existRegexp, err2 := responsePairMatchesScheme(scheme2, notExistResponse, existResponse) 
  if err2 == nil { return }

  return nil, nil, errors.New(
    "Responses did not fit scheme 1: " + err1.Error() + "\n" +
    "Responses did not fit scheme 2: " + err2.Error() + "\n")
}

func responsePairMatchesScheme(scheme patternScheme, notExistResponse queryResult, existResponse queryResult) (notExistRegexp *regexp.Regexp, existRegexp *regexp.Regexp, err error) {

  nnScore := scheme.notExistPatterns.score(notExistResponse)
  neScore := scheme.existPatterns.score(notExistResponse)

  if !(nnScore.MatchCount == 1 && neScore.MatchCount == 0) {
    msg := fmt.Sprintf("Expected response to indicate domain non-existence, but it did not (%d,%d): %s", nnScore.MatchCount, neScore.MatchCount, notExistResponse.String())
    return nil, nil, errors.New(msg)
  }

  enScore := scheme.notExistPatterns.score(existResponse)
  eeScore := scheme.existPatterns.score(existResponse)

  if !(enScore.MatchCount == 0 && eeScore.MatchCount == 1) {
    msg := fmt.Sprintf("Expected response to indicate domain existence, but it did not (%d,%d): %s", enScore.MatchCount, eeScore.MatchCount, existResponse.String())
    return nil, nil, errors.New(msg)
  }

  return nnScore.FirstMatchedPattern, eeScore.FirstMatchedPattern, nil
}
