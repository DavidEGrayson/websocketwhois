package data

import (
  "os"
  "bufio"
  "strings"
)

// This represents a line from tld_serv_list, which came from the
// standard unix whois utility. 
type DebianSuffixInfo struct {
  Name, Server, Note string
}

func DebianSuffixInfosRead() ([]DebianSuffixInfo, error) {
  suffixInfos := make([]DebianSuffixInfo, 0)

  file, err := os.Open(Directory + "/tld_serv_list")
  if err != nil { return nil, err }
  defer file.Close()

  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    line := scanner.Text()
    line = strings.Split(line, "#")[0]     // Remove comments
    fields := strings.Fields(line)         // Split by whitespace.
    
    var suffix DebianSuffixInfo

    if len(fields) == 0 {
      continue   // Empty line.
    }

    suffix.Name = fields[0]
    attrs := fields[1:]
    for _, attr := range attrs {
      if attr[0] >= 'A' && attr[0] <= 'Z' {
        suffix.Note = attr
      } else {
        suffix.Server = attr
      }
    }

    suffixInfos = append(suffixInfos, suffix)
  }
  if (scanner.Err() != nil) {
    return nil, scanner.Err()
  }

  return suffixInfos, nil
}

