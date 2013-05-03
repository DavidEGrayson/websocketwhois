// TODO: Favorites list using HTML5 storage

// Useful: http://matthewlein.com/experiments/easing.html

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

  user_input.keyup(keyUpCallback = function() {
    var input_string = user_input.val().toLowerCase();
    if (!input_string.match(/^[a-z-.0-9]*$/)) {
      userInputInvalid = true;
      domainNames = []
      // TODO: indicate that the input was invalid
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
    if (userInputInvalid)
    {
      $("#userInputInvalidMessage").slideDown(500);
    }
    else
    {
      $("#userInputInvalidMessage").slideUp(200);
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
    //console.log(domainResults.join(", "));

    // Naive re-rendering of everything even if it didn't change.
    var resultArea = document.createElement("div");
    resultArea.id = "result-area";
    //resultArea.appendChild(document.createTextNode(domainResults.join(", ")));

    $.each(domainResults, function(index, domainResult) {
      var resultNode = document.createElement("div");
      resultNode.className = "result " + domainResult.state;
      var firstParagraph = document.createElement("p");
      var nameNode = document.createElement("span");
      nameNode.className = "name"
      nameNode.appendChild(document.createTextNode(domainResult.name));
      firstParagraph.appendChild(nameNode);
      resultNode.appendChild(firstParagraph);
      
      switch(domainResult.state)
      {
      case "pending":
        firstParagraph.innerHTML += "<span class='animatedEllipsis'>" +
          "<span>.</span><span>.</span><span>.</span></span>";
        break;
      case "available":
        firstParagraph.appendChild(document.createTextNode(" is available!"));
        break;
      case "taken":
        firstParagraph.appendChild(document.createTextNode(" is taken."));
        var goToWebsite = document.createElement("a");
        goToWebsite.href = "http://" + domainResult.name
        goToWebsite.appendChild(document.createTextNode("Go to website"));
        resultNode.appendChild(goToWebsite);
        break;
      case "error":
        firstParagraph.appendChild(document.createTextNode(" is unknown because there was an error!"));
        break;
      }
      resultArea.appendChild(resultNode);
    });
    // $("<div id='result-area'/>");
    $("#result-area").replaceWith(resultArea);
  });

  // Uncomment the code below to make it easier to debug certain things.
  //window.setTimeout(function() {
  //  $("#user-input").val("davidegrayson");
  //  keyUpCallback();
  //}, 500);

  //window.setTimeout(function() {
  // domainResults = [
  //   {name: "davidegrayson.com", state: "taken"},
  //   {name: "davidegrayson.net", state: "available"},
  //   {name: "davidegrayson.org", state: "pending"},
  //   {name: "davidegrayson.info", state: "error"}
  // ]
  // domainResultsChangeCallbacks.fire();
  //}, 50);

  // Just in case the user clicked back and there is stuff left in the box.
  window.setTimeout(function() {
    keyUpCallback();
  }, 100);
});

