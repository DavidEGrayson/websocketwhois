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
  var lastDomainName []byte

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

    if bytes.HasSuffix(fields[0], domainNameSuffix) && bytes.Compare(fields[1], ns) == 0 {

      // Found a domain name.  Process it.
      domainName := bytes.ToLower(fields[0])
      domainName = domainName[0 : len(domainName) - len(domainNameSuffix)]
      
      comparison := bytes.Compare(domainName, lastDomainName)

      if comparison == 0 {
        // We already printed this domain name.
        continue
      }

      if comparison == -1 {
        // The file is not sorted the way we expect!
        log.Printf("Zone file is not sorted as we expected.  %s comes after %s\n",
          string(domainName), string(lastDomainName))
      }

      domainNameLine := append(domainName, '\n')

      _, err := os.Stdout.Write(domainNameLine)
      if err != nil { log.Fatal(err) }

      lastDomainName = domainName
    }

  }
}