$(function() {
  var user_input = $("#user-input");
  var tldsOfInterest = ["com", "net"];
  var userInputInvalid = false;
  var domainNames = [];
  var userInputChangeCallbacks = $.Callbacks();
  var domainResults = [];
  var domainResultsChangeCallbacks = $.Callbacks();

  if (!window["WebSocket"]) {
    alert("Sorry, your browser does not support websockets.  Try Google Chrome!");
    location.href = "http://www.google.com/chrome/";
    return;
  }

  whois.init();

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
        domainNames = []
      }
      else if (parts.length == 1) {
        domainNames = $.map(tldsOfInterest, function(tld) {
          return parts[0] + "." + tld;
        });
      }
      else {
        domainNames = [parts[parts.length-2] + "." + parts[parts.length-1]]; 
      }
    }

    userInputChangeCallbacks.fire();
  });

  userInputChangeCallbacks.add(function() {
    if (userInputInvalid) {
      console.log("User input invalid.");
    }
    else {
      console.log("User interested in: " + domainNames);
    }        
  });

  userInputChangeCallbacks.add(function() {
    domainResults = $.map(domainNames, function(name) {
      return whois.domain(name);
    });

    domainResultsChangeCallbacks.fire();
  });

  whois.resultCallbacks.add(function(domain) {
    domainResultsChangeCallbacks.fire();
  });

  domainResultsChangeCallbacks.add(function() {
    console.log(domainResults.join(", "));
  });

});

