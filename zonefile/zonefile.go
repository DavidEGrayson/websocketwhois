package zonefile

import (
  "fmt"
  "os"
)

func Open(name string) (file *File, err error) {
	fmt.Println("Hello.  I should open " + name)
  osFile, err := os.Open(name)
  if err != nil {
    return nil, err
  }
  return &File { osFile }, nil
}

type File struct {
  osFile *os.File
}

func (f *File) Close() error {
  return f.osFile.Close()
}

// Find the given domain name and return the offset of the first character of
// its line.
func (f *File) Find(domainName string) int {
  return -1
}


