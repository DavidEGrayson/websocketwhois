package domainlist

import (
  "testing"
)


func BenchmarkMidLookup(b *testing.B) {
  list, err := Open("../data/org.domains")  
  if err != nil { b.Fail() }
  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    offset, err := list.Find("graysonfamily")
    if err != nil || offset < 0 { b.Fail() }
  }
}