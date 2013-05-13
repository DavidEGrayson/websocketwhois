package domainlist

import (
  "testing"
)

func BenchmarkLookupMid(b *testing.B) {
  benchLookup(b, "graysonfamily")
}

func BenchmarkLookupMid2(b *testing.B) {
  benchLookup(b, "rubyonrails")
}

func BenchmarkLookupStart(b *testing.B) {
  benchLookup(b, "0")
}

func BenchmarkLookupEnd(b *testing.B) {
  benchLookup(b, "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
}

func benchLookup(b *testing.B, entry string) {
  list, err := Open("../data/org.domains")  
  if err != nil { b.Fatal(err) }
  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    offset, err := list.Find(entry)
    if err != nil || offset < 0 {
      b.Fatal(err)
    }
  }
}