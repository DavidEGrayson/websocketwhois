package main

// TODO: configure emacs to not put tab characters in go code

import (
  "code.google.com/p/go.net/websocket"
  //"os/exec"
  "log"
  "fmt"
  "net"
  "io"
  //"bufio"
  "os"
  "net/http"
  "reflect"
  "syscall"
  "strings"
)

// The connection between a websocket and a process.
type connection struct {
  ws *websocket.Conn
  Id int
  log log.Logger

  closeRequest chan bool
  whoisResults chan whoisResult
  Closed bool
}

var idChannel chan int = make(chan int)

// Called once when the program starts.
func connInit() {
  go assignIds()
}

// Adds 0, 1, 2, 3, ... to the idChannel.
func assignIds() {
  i := 0
  for {
    idChannel <- i
    i++
  }
  // TODO: prevent integer overflow at some point?
}

func websocketClosedLocally(err error) bool {
  if operror, ok := err.(*net.OpError); ok {
    if errno, ok := operror.Err.(syscall.Errno); ok {
      if (errno==1236) { // Microsoft Windows ERROR_CONNECTION_ABORTED
        return true
      }
    }
  }
  return false
}

func (c *connection) websocketToProcess() {
  defer c.Close()
  for {
    var message string
    err := websocket.Message.Receive(c.ws, &message)
    if err != nil {
      switch {
        case c.Closed: // Suppress error message
        case websocketClosedLocally(err): c.log.Println("Websocket closed locally.")
        case err == io.EOF: c.log.Println("Websocket closed remotely.")
        default: c.log.Println("Error reading from websocket:", err, reflect.TypeOf(err))
      }
      return
    }

    if strings.HasPrefix(message, "w") {
      domainFragment := strings.ToLower(message[1:])
      c.whoisRequest(domainFragment)
    }

  }
}

func (c *connection) whoisRequest(domainFragment string) {
  c.log.Println("Whois request:", domainFragment)

	domain := domainFragment + ".com" // TODO: this needs a lot of work

  var request whoisRequest
  request.Domain = domain
  request.ResultChannel = c.whoisResults
  whoisRequestChannel <- request
}

func (c *connection) websocketWrite() {
  for result := range c.whoisResults {
    c.log.Println("Result: ", result.WhoisRequest.Domain, result.Exists);

		var existsString string
		if result.Exists {
			existsString = "1"
		} else {
			existsString = "0"
		}

		str := "r" + result.WhoisRequest.Domain + "," + existsString
		//log.Println("Sending to websocket:", str)
		err := websocket.Message.Send(c.ws, str)
		if err != nil {
			if !c.Closed {
				log.Println("Error sending to websocket:", err)
			}
		}
    
  }
}

func firstProtocol(req *http.Request) string {
  protocols := req.Header["Sec-Websocket-Protocol"]
  if len(protocols) >= 1 {
    return protocols[0]
  }
  return ""
}

func (c *connection) run() {
  c.Id = <-idChannel
  c.log = *log.New(os.Stdout, fmt.Sprintf("#%d ", c.Id), log.Flags())
  c.log.Println("New connection from " + c.ws.Request().RemoteAddr + ".")

  // closeRequest must be buffered in case multiple goroutines
  // call Close() at nearly the same time.  Is there a better way?
  c.closeRequest = make(chan bool, 10)

  c.whoisResults = make(chan whoisResult, 100)  

  //go c.processToWebsocket()
  go c.websocketToProcess()
  go c.websocketWrite()
  
  <-c.closeRequest
  c.cleanup()
}

func (c *connection) Close() {
  c.log.Println("Sending request to close.")
  c.closeRequest <- true
}

func (c *connection) cleanup() {
  c.Closed = true
  c.ws.Close()
  c.log.Println("Connection cleaned up.")
}

func wsHandler(ws *websocket.Conn) {
  c := &connection { ws: ws }
  c.run()
}
