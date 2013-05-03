$(function() {
  var conn;
  var user_input = $("#user-input");
  var tldsOfInterest = ["com", "net"];
  var userInputInvalid = false;
  var domainsOfInterest = [];
  var userInputChangeCallbacks = $.Callbacks();

  if (!window["WebSocket"]) {
    alert("Sorry, your browser does not support websockets.  Try Google Chrome!");
    location.href = "http://www.google.com/chrome/";
    return;
  }

  //whois.init();

  user_input.keyup(function() {
    var input_string = user_input.val().toLowerCase();
    if (!input_string.match(/[a-z.]*/)) {
      userInputInvalid = true;
    }
    else {
      userInputInvalid = false;
      var parts = input_string.split("\.")
      parts = $.grep(parts, function(e){ return e; })
      if (parts.length == 0) {
        domainsOfInterest = []
      }
      else if (parts.length == 1) {
        domainsOfInterest = $.map(tldsOfInterest, function(tld) {
          return parts[0] + "." + tld;
        });
      }
      else {
        domainsOfInterest = [parts[parts.length-2] + "." + parts[parts.length-1]]; 
      }
    }

    userInputChangeCallbacks.fire();
  });

  userInputChangeCallbacks.add(function() {
    if (userInputInvalid) {
      console.log("User input invalid.");
    }
    else {
      console.log("User interested in: " + domainsOfInterest);
    }        
  });

  userInputChangeCallbacks.add(function() {
    
  });

  conn = new WebSocket("ws://" + window.location.hostname + ":" + window.location.port + "/ws");
  conn.onclose = function(evt) {
    console.log("Connection closed.");
  }
  conn.onmessage = function(evt) {
    appendLog("Received message: " + evt.data);
  }

});

