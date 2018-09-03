package main

import "mime"

type asset struct {
	content     []byte
	contentType string
}

var assets map[string]asset

func init() {
	assets = make(map[string]asset)
	assets["magnet.css"] = asset{
		contentType: mime.TypeByExtension(".css"),
		content: []byte(`
html, body {
	font-family:Tahoma, Geneva, sans-serif;
	font-size:13px;
	line-height: 1.3;
	background-color:#FFF;
	color:#000;
	height: 100%;
}

a:link, a:visited, a:active{
	color: #000099;
	text-decoration:none;
}

a:hover{
	text-decoration:underline;
}`),
	}

	assets["magnet.js"] = asset{
		contentType: mime.TypeByExtension(".js"),
		content: []byte(`function createRequestObject() {
	if (window.XMLHttpRequest)
	{
	  // code for IE7+, Firefox, Chrome, Opera, Safari
	  return new XMLHttpRequest();
	}
	if (window.ActiveXObject)
	{
	  // code for IE6, IE5
	  return new ActiveXObject("Microsoft.XMLHTTP");
	} else {
	  // There is an error creating the object,
	  // just as an old browser is being used.
	  alert("Your Browser Does Not Support This Script - Please Upgrade Your Browser ASAP");
	}
  }
  
  // Make the XMLHttpRequest object
  var http = createRequestObject();
  
  function sendRequest(page) {
	// Open PHP script for requests
	http.open('GET', page, true);
	http.send(null);
  }`),
	}
}
