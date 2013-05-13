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
// its line.  Returns -1 if the domain was not found.
func (f *File) Find(domainName string) (offset int64, err error) {
  _, _ = f.osFile.Seek(1053725408*3/4, 0)
  bytes := make([]byte, 300)
  _, _ = f.osFile.Read(bytes)
  fmt.Println(string(bytes))
  return -1
}


