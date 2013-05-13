package domainlist

// TODO: in the part of the code that uses this, warn users that the list will
// not contain "inactive" domain names because those are not part of the zone file.

import (
  "os"
  "io"
  "bytes"
  "errors"
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
  buffer []byte
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

  f.buffer = make([]byte, 80)
  return
}


// TODO: Consider forcing the beginning of the file to have a newline.
// Then we can get rid of all the checks for offset == 0 in this method.
// Only do it if the benchmarks get noticeably faster!
func (f *File) goBackToStartOfLine(currentOffset int64) (offset int64, err error) {
  offset = currentOffset

  if (offset == 0) { return }

  var buffer []byte
  if (offset <= 80) {

    buffer = f.buffer[0:offset]

    _, err = f.osFile.Seek(0, 0)
    if err != nil { return }

    _, err = f.osFile.Read(buffer)
    if err != nil { return }

    for i := 0; i < len(buffer); i++ {
      c := buffer[len(buffer) - i - 1]
      if c == '\n' {
        if i > 0 {
          return f.osFile.Seek(int64(-i), 1)
        }
        return offset, nil
      }
    }
    return f.osFile.Seek(0, 0)

  } else {

    buffer = f.buffer

    _, err = f.osFile.Seek(-80, 1)
    if err != nil { return }
    
    _, err = f.osFile.Read(buffer)
    if err != nil { return }

    for i := 0; i < len(buffer); i++ {
      c := buffer[len(buffer) - i - 1]
      if c == '\n' {
        if i > 0 {
          return f.osFile.Seek(int64(-i), 1)
        }
        return offset, nil
      }
    }
    return -1, errors.New("Domain list file has line longer than 79 bytes.")
  }

  panic("unreachable")
}

// Tries to find the given domain name and retuns three values:
// found bool:  Whether the domain name was found or not.
// offset int64: If the domain name was found, this is the offset where it exists.
//   If the domain name was not found, this is the offset in the file where
//   we would expect the domain name to be.
// err error: Any operating system errors or unexpected data in the file.
//   If err is non-nil, then found is false and offset is -1.
func (f *File) Find(domainNameStr string) (found bool, offset int64, err error) {
  domainName := []byte(domainNameStr)

  // lowerBound points to the first byte of the first line that might contain the
  //   domain we a looking for
  // upperBound points to the first byte AFTER the last line that might contain the
  //   domain we are looking for.
  var lowerBound, upperBound int64
  lowerBound = 0
  upperBound = f.size

  for {
    if upperBound == lowerBound {
      // Specified domain name does not exist in the file.
      return false, lowerBound, nil
    }

    bisectPoint := (lowerBound + upperBound) / 2

    _, err := f.osFile.Seek(bisectPoint, 0)
    if err != nil { return false, -1, err }

    bisectPoint, err = f.goBackToStartOfLine(bisectPoint)
    if err != nil { return false, -1, err }

    fragment := make([]byte, 80)
    n, err := f.osFile.Read(fragment)
    if err != nil && err != io.EOF { return false, -1, err }
    fragment = fragment[0:n]

    newlineIndex := bytes.IndexByte(fragment, '\n')
    if newlineIndex == -1 {
      // We found a line longer than expected.  Should not happen.
      return false, -1, errors.New("Domain list file has a line longer than 80 bytes.")
    }
    bisectingDomainName := fragment[0:newlineIndex]

    comparison := bytes.Compare(bisectingDomainName, domainName)

    //fmt.Printf("%11d %11d %11d %11d %s %d\n",
    //  lowerBound, bisectPoint, upperBound, upperBound - lowerBound,
    //  bisectingDomainName, comparison)

    switch comparison {
    case  0: return true, bisectPoint, nil
    case  1: upperBound = bisectPoint
    case -1: lowerBound = bisectPoint + int64(len(bisectingDomainName)) + 1
    }
  }

  return
}