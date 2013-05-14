package main

import (
  "regexp"
)

type patternSet []*regexp.Regexp

type patternSetScore struct {
  MatchCount int
  FirstMatchedPattern *regexp.Regexp
}

func newPatternSet(s []string) patternSet {
  r := make([]*regexp.Regexp, len(s))
  for i, str := range s {
    r[i] = regexp.MustCompile(str)
  }
  return r
}

func (p *patternSet) score(strs []string) patternSetScore {
  score := patternSetScore{}
  for _, line := range strs {
    for _, pattern := range *p {
      if pattern.MatchString(line) {
        score.MatchCount += 1
        if (score.FirstMatchedPattern == nil) {
          score.FirstMatchedPattern = pattern
        }
      }
    }
  }
  return score
}
