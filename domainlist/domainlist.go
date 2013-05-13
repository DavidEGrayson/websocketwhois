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
  return
}


// TODO: Consider forcing the beginning of the file to have a newline.
// Then we can get rid of all the checks for offset == 0 in this method.
// Only do it if the benchmarks get noticeably faster!
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
func (f *File) Find(domainNameStr string) (offset int64, err error) {
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
      return -1, nil
    }

    bisectPoint := (lowerBound + upperBound) / 2

    _, err := f.osFile.Seek(bisectPoint, 0)
    if err != nil { return -1, err }

    bisectPoint, err = goBackToStartOfLine(f.osFile, bisectPoint)
    if err != nil { return -1, err }

    fragment := make([]byte, 80)
    n, err := f.osFile.Read(fragment)
    if err != nil && err != io.EOF { return -1, err }
    fragment = fragment[0:n]

    newlineIndex := bytes.IndexByte(fragment, '\n')
    if newlineIndex == -1 {
      // We found a line longer than expected.  Should not happen.
      return -1, errors.New("Domain list file has a line longer than 80 bytes.")
    }
    bisectingDomainName := fragment[0:newlineIndex]

    comparison := bytes.Compare(bisectingDomainName, domainName)

    //fmt.Printf("%11d %11d %11d %11d %s %d\n",
    //  lowerBound, bisectPoint, upperBound, upperBound - lowerBound,
    //  bisectingDomainName, comparison)

    switch comparison {
    case  0: return bisectPoint, nil
    case  1: upperBound = bisectPoint
    case -1: lowerBound = bisectPoint + int64(len(bisectingDomainName)) + 1
    }
  }

  return
}