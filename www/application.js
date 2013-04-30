$(function() {
  var conn;
  var userinput = $("#userinput");
  
  if (!window["WebSocket"]) {
    alert("Sorry, your browser does not support WebSockets.  Try Google Chrome!");
    return;
  }

  conn = new WebSocket("ws://" + window.location.hostname + ":" + window.location.port + "/ws");
  conn.onclose = function(evt) {
    appendLog($("<div><b>Connection closed.</b></div>"))
  }
  conn.onmessage = function(evt) {
    appendLog($("<div/>").text(evt.data))
  }

});

