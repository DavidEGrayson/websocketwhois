package domainlist

import (
  "testing"
  "math/rand"
)

func TestBacktrack(t *testing.T) {
  testBacktrack(t, 0, 6)
  testBacktrack(t, 6, 7)
  testBacktrack(t, 1489, 11)
  testBacktrack(t, 1500, 15)
  testBacktrack(t, 1515, 8)
  testBacktrack(t, 1523, 7)
}

func testBacktrack(t *testing.T, lineStart int64, lineLength int) {
  list, err := Open("test_list.txt")
  if err != nil { t.Fatal(err) }

  for i := 0; i < lineLength; i++ {
    initialOffset, err := list.osFile.Seek(lineStart + int64(i), 0)
    if err != nil { t.Fatal(err) }

    offset, err := list.goBackToStartOfLine(initialOffset)
    if err != nil { t.Fatal(err) }

    if offset != lineStart {
      t.Fatalf("Expected goBackToStartOfLine(%d) to return offset %d, got %d\n", initialOffset, lineStart, offset)    
    }

    realOffset, err := list.osFile.Seek(0, 1)
    if offset != realOffset {
      t.Fatalf("goBackToStartOfLine lied about the new address.  real = %d, lie = %d\n", realOffset, offset)
    }
  }
}


func TestFind(t *testing.T) {
  testFind(t, "AARON", false, 0)
  testFind(t, "ALMUD", true, 0)
  testFind(t, "ALMUDE", true, 6)
  testFind(t, "ALMUDES", true, 13)
  testFind(t, "ALPHABETS", true, 533)
  testFind(t, "BLATANT", true, 1515)
  testFind(t, "BROADCAST", false, 1523)
  testFind(t, "KATANA", true, 1523)
  testFind(t, "ZARKANA", false, 1530)
}

func testFind(t *testing.T, entry string, expectToFind bool, expectedOffset int64) {

  list, err := Open("test_list.txt")  
  if err != nil { t.Fatal(err) }
  found, offset, err := list.Find(entry)
  if err != nil {
    t.Fatal(err)
  }
  if found != expectToFind {
    if expectToFind {
      t.Fatalf("Expected Find(\"%s\") to find the entry, but it did not.", entry)
    } else {
      t.Fatalf("Expected Find(\"%s\") not to find the entry, but it did.", entry)
    } 
  }
  if offset != expectedOffset {
    t.Fatalf("Expected Find(\"%s\") to return offset %d, got %d\n", entry, expectedOffset, offset)
  }
}

var bytelist = []byte {
  'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
  'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '-',
}

func randomEntry() string {
  length := 3 + rand.Intn(25)
  str := ""
  for i := 0; i < length; i++ {
    str += string( bytelist[ rand.Intn(len(bytelist)) ] )
  }
  return str
}

func BenchmarkFindRandom(b *testing.B) {
  list, err := Open("../data/org.domains")  
  if err != nil { b.Fatal(err) }

  entries := make([]string, b.N)
  for i, _ := range entries {
    entries[i] = randomEntry()
  }

  b.ResetTimer()

  for _, entry := range entries {
    _, _, err := list.Find(entry)
    if err != nil { b.Fatal(err) }
  }
}

func BenchmarkFindMid(b *testing.B) {
  benchFind(b, "graysonfamily")
}

func BenchmarkFindMid2(b *testing.B) {
  benchFind(b, "rubyonrails")
}

func BenchmarkFindStart(b *testing.B) {
  benchFind(b, "0")
}

func BenchmarkFindEnd(b *testing.B) {
  benchFind(b, "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
}

func benchFind(b *testing.B, entry string) {
  list, err := Open("../data/org.domains")  
  if err != nil { b.Fatal(err) }
  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    _, _, err := list.Find(entry)
    if err != nil { b.Fatal(err) }
  }
}