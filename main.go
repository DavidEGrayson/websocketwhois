package main

// TODO: support secure websockets

import (
  "flag"
  "log"
  "net/http"
  "code.google.com/p/go.net/websocket"
  "strings"
  "fmt"
  "os"
)

// Commandline arguments
const defaultAddr string = "localhost:8080"
var addr * string
var webDirName * string

func homeHandler(conn http.ResponseWriter, request *http.Request) {
  if strings.ToLower(request.Header.Get("Upgrade")) == "websocket" {
    websocket.Handler(wsHandler).ServeHTTP(conn, request)
  } else if (webDirName != nil) {
    if (request.Method != "GET") {
      conn.Header().Set("Allow", "GET")
      http.Error(conn, "This server only accepts GET or websocket requests.", http.StatusMethodNotAllowed)
    } else {
      http.FileServer(http.Dir(*webDirName)).ServeHTTP(conn, request)
    }
  } else {
		http.NotFound(conn, request)
  }
}

func printHelpScreen() {
  fmt.Println("Usage: websocketwhois [OPTIONS]")
  fmt.Println("")
  fmt.Println("Options:")
  fmt.Println("  --addr=ADDR        Address and port to bind to.  Default is " + defaultAddr + ".")
  fmt.Println("  --dir=DIRNAME      Directory of files that can be fetched with plain HTTP.")
}

func parseArgs() bool {
  addr = flag.String("addr", defaultAddr, "")
  webDirName = flag.String("dir", "", "")
  
  // Parse the flags
  flag.Parse()
  
  if *webDirName == "" {
    webDirName = nil
  }
  
  if len(flag.Args()) != 0 {
    printHelpScreen()
    return false
  }
  return true
}

func run() {
  log.Printf("wsc server starting on %s.\n", *addr)
  if err := http.ListenAndServe(*addr, http.HandlerFunc(homeHandler)); err != nil {
    log.Fatal("Error in ListenAndServe:", err)
  }
  log.Printf("Serving ending?");
}

func main() {
  log.SetOutput(os.Stdout)

  if (!parseArgs()) {
    return
  }

  connInit()
  run()
}