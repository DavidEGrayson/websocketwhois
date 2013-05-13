package domainlist

// TODO: in the part of the code that uses this, warn users that the list will
// not contain "inactive" domain names because those are not part of the zone file.

import (
  "fmt"
  "os"
  "io"
)

func Open(name string) (file *File, err error) {
  osFile, err := os.Open(name)
  if err != nil {
    return
  }

  file = &File {}
  file.osFile = osFile

  err = file.setup()
  if err != nil {
    file.Close()
    return
  }

  return
}

type File struct {
  osFile *os.File
  size int64
}

// Compares two byte slices alphabetically.
// Returns 1 is the first is greater.
// Returns -1 is the second is greater.
// Return 0 is the two are exactly the same.
//
// The byte slices are not necessarily null terminated.
// Slice a is less than slice a + b as long as b is nonempty.
func Compare(x, y []byte) int {
  i := 0
  for {
    if i == len(x) && i == len(y) { return 0; }
    if i == len(x) { return -1 }
    if i == len(y) { return 1 }
    if x[i] < y[i] { return -1 }
    if x[i] > y[i] { return 1 }
    i += 1
  }
  return 0  // TODO: upgrade to go 1.1 and run "go vet" to get rid of things like this
}

func (f *File) Close() error {
  return f.osFile.Close()
}

func (f *File) setup() (err error) {
  info, err := f.osFile.Stat()
  if err != nil {
    return
  }
  f.size = info.Size()
  fmt.Printf("size = %d\n", f.size)
  return
}

func goBackToStartOfLine(osFile * os.File, currentOffset int64) (offset int64, err error) {
  byteslice := make([]byte, 1)

  offset = currentOffset

  if (offset == 0) { return }

  offset, err = osFile.Seek(-1, 1)
  if err != nil { return }

  for offset > 0 {
    _, err = osFile.Read(byteslice)
    offset += 1
    if err != nil { return }
    if byteslice[0] == '\n' { return }
    offset, err = osFile.Seek(-2, 1)
    if err != nil { return }
  }
  return
}

// Finds the given domain name and return the offset of the first character of
// its line.  Returns -1 if the domain was not found.
func (f *File) Find(domainName string) (offset int64, err error) {
	fmt.Println("Hello.  I should find " + domainName)

  offset = -1

  // lowerBound points to the first byte of the first line that might contain the
  //   domain we a looking for
  // upperBound points to the first byte AFTER the last line that might contain the
  //   domain we are looking for.
  var lowerBound, upperBound int64
  lowerBound = 0
  upperBound = f.size

  for {
    bisectPoint := (lowerBound + upperBound) / 2

    fmt.Printf("points: %d %d %d\n", lowerBound, bisectPoint, upperBound)

    _, err = f.osFile.Seek(bisectPoint, 0)
    if err != nil { return }

    bisectPoint, err = goBackToStartOfLine(f.osFile, bisectPoint)
    if err != nil { return }

    fmt.Printf("bisect point now = %d\n", bisectPoint)

    bytes := make([]byte, 80)
    obytes := bytes
    var n int
    n, err = f.osFile.Read(bytes)
    if err != nil && err != io.EOF { return }
    bytes = bytes[0:n]

    foundNewline := false
    for i, byte := range bytes {
      if byte == '\n' {
        foundNewline = true
        bytes = bytes[0:i]
        break
      }
    }

    if !foundNewline {
      // TODO: reutrn an error here
      //eturn -1, err
    }

    fmt.Printf("bytes: %d %s\n", len(bytes), bytes)
    fmt.Printf("bytes: %d %s\n", len(obytes), obytes)
    return -1, nil

  }

  return
}