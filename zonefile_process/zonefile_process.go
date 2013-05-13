package main

// TODO: change this to use the 'bytes' package, not strings

// TODO: assert that the lines that contain domain names are contiguous block
//   If not, throw an error.

// TODO: assert that no domain names are longer than 80 characters

import (
  "os"
  "log"
  "io"
  "bufio"
  "strings"
)

func main() {
  log.SetOutput(os.Stderr)
  if (len(os.Args) != 2) {
    log.Print("Usage: cat zonefile | zonefile_process tld > domainlist")
    os.Exit(2)
  }

  reader := bufio.NewReader(os.Stdin)
  domainNameSuffix := "." + strings.ToUpper(os.Args[1]) + "."
  var lastDomainName string

  for {
    str, err := reader.ReadString('\n')
    if err == io.EOF {
      break
    }
    if err != nil {
      log.Fatal(err)
    }

    fields := strings.Fields(str)
    if len(fields) == 0 {
      continue
    }

    if strings.HasSuffix(fields[0], domainNameSuffix) &&
      fields[1] == "NS" && fields[0] != lastDomainName {

      // Found a new domain name.  Write it to stdout.
      lastDomainName = fields[0]

      domainName := []byte(strings.ToLower(fields[0]))

      // Replace the TLD with newline.  ".org." -> "\n"
      domainName = domainName[0 : len(domainName) - len(domainNameSuffix) + 1]
      domainName[len(domainName)-1] = '\n'

      _, err := os.Stdout.Write(domainName)
      if err != nil {
        log.Fatal(err);
      }
    }

  }
}