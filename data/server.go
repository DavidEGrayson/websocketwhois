package data

import (
  "encoding/json"
  "os"
  "log"
  "../errorwrap"
  "io/ioutil"
  "errors"
  "fmt"
)

type Server struct {
  Name string

  RateLimit bool
  DoNotUse bool
  Protocol string
  Suffixes []string
  NotExistRegexp, ExistRegexp *JsonRegexp
}

type OutputFile struct {
  Servers []*Server
}

func ServerHintsRead() (map[string]*Server, error) {
  servers := make(map[string]*Server)
  
  hintBytes, err := ioutil.ReadFile(Directory + "/server-hints.json")
  if err != nil { return nil, err }

  err = json.Unmarshal(hintBytes, &servers)
  if err != nil {
    if serr, ok := err.(*json.SyntaxError); ok {
      //contextLength := 10
      //if int(serr.Offset) + contextLength >= len(hintBytes) {
      //  contextLength = len(hintBytes) - int(serr.Offset) - 1
      //}
      //nearBytes := hintBytes[serr.Offset:contextLength]
      msg := fmt.Sprintf("Syntax error in server-hints.json at offset %d: %s",
        serr.Offset, serr.Error())
      return nil, errors.New(msg)
    }
    return nil, errorwrap.New("Error decoding server-hints.json", err)
  }

  return servers, nil
}

func ServersWrite(servers []*Server) {
  output := OutputFile { }

  output.Servers = servers

  file, err := os.Create(Directory + "/servers.json")
  if err != nil {
	  log.Fatal("Error opening output file.", err)
  }
  defer file.Close()

  //data, err := json.Marshal(output)  // For production.
  data, err := json.MarshalIndent(output, "", "  ")  // For development.
  if err != nil {
    log.Fatal("Error marshalling json data.", err)
  }

  n, err := file.Write(data)
  if err != nil || n != len(data) {
    log.Fatal("Error writing to output file.", err)
  }
}