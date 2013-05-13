package main

// TODO: assert that the lines that contain domain names are contiguous block
//   If not, throw an error.

// TODO: assert that no domain names are longer than 80 characters

// TODO: assert that each domain name is lexigraphically after the previous

import (
  "os"
  "log"
  "io"
  "bufio"
  "bytes"
)

func main() {
  log.SetOutput(os.Stderr)
  if (len(os.Args) != 2) {
    log.Print("Usage: zcat zonefile.gz | zonefile_process TLD > domainlist")
    os.Exit(2)
  }

  reader := bufio.NewReader(os.Stdin)
  domainNameSuffix := bytes.ToUpper([]byte("." + os.Args[1] + "."))
  var lastDomainName string

  ns := []byte("NS")

  for {
    str, err := reader.ReadString('\n')
    if err == io.EOF {
      break
    }
    if err != nil {
      log.Fatal(err)
    }

    fields := bytes.Fields( []byte(str) )
    if len(fields) == 0 {
      continue
    }

    if bytes.HasSuffix(fields[0], domainNameSuffix) && bytes.Compare(fields[1], "NS") == 0 {

      // Found a domain name.  Process it.
      domainName := bytes.ToLower(fields[0])
      domainName = domainName[0 : len(domainName) - len(domainNameSuffix) + 1]
      domainName[len(domainName)-1] = '\n'
      
      comparison = bytes.Compare(fields[0], lastDomainName)

      if comparison == 0 {
        // We already printed this domain name.
        continue
      }

      _, err := os.Stdout.Write(domainName)
      if err != nil {
        log.Fatal(err);
      }

      lastDomainName = domainName
    }

  }
}