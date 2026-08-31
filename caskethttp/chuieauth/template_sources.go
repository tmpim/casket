package chuieauth

import (
	"html/template"
)

var styleSource []byte
var loginTemplate *template.Template
var deniedTemplate *template.Template
var serverErrorTemplate *template.Template
var temporaryAccessTemplate *template.Template
var loggedInTemplate *template.Template

func init() {
	var loginSource = `<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no">
<link href="https://fonts.googleapis.com/css?family=Lato:400,300,100,700" rel="stylesheet" type="text/css">
<link href="/style.css" rel="stylesheet" type="text/css">
<title>chuie.io authentication</title>
</head>
<body>
	<div class="header-bar">
		<div class="container">
			<span>chuie.io authentication</span>
		</div>
	</div>
	<div class="container">
		<div class="half">
			<h1 class="welcome">Hey there!</h1>
			{{- if .Error}}
			<p class="error">
				{{.Error}}
				<br>
				<span class="help">
					Contact Jason (<a href="mailto:me@chuie.io">me@chuie.io</a>) for more help.
				</span>
			</p>
			{{- else}}
			<p><strong>Welcome to chuie.io!</strong> The page you are trying to access requires appropriate authorization.</p>
			{{- end}}
			<form action="/" method="post">
				<input name="username" required type="text" placeholder="Username or access code">
				<input name="password" type="password" placeholder="Password">
				<label class="checkbox-label">
				<input name="remember" checked type="checkbox">
				<span class="custom-checkbox"></span>
				<span class="remember-me">Remember me</span></label>
				<input type="submit" value="Login">
			</form>
		</div>
	</div>
	<div class="footer">
		<div class="footer-text container">
			You're on chuie.io. Made by Jason Chu (1lann) -
			<a href="mailto:me@chuie.io">me@chuie.io</a>.
		</div>
	</div>
	<script>
	function getInputByName(name) {
		var allElements = document.getElementsByTagName("input");
		for (var i = 0, n = allElements.length; i < n; i++) {
			if (allElements[i].getAttribute("name") === name) {
				return allElements[i];
			}
		}
	}

	function getSubmitButton() {
		var allElements = document.getElementsByTagName("input");
		for (var i = 0, n = allElements.length; i < n; i++) {
			if (allElements[i].getAttribute("type") === "submit") {
				return allElements[i];
			}
		}
	}

	var usernameInput = getInputByName("username");
	var submitButton = getSubmitButton();

	usernameInput.addEventListener("input", function() {
		if (this.value.slice(0, 2) == "ac") {
			submitButton.value = "Use Access Code";
			usernameInput.className = "access-code";
		} else {
			submitButton.value = "Login";
			usernameInput.className = "";
		}
	});
	</script>
</body>
</html>`

	var deniedSource = `<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no">
<link href="https://fonts.googleapis.com/css?family=Lato:400,300,100,700" rel="stylesheet" type="text/css">
<link href="{{.Authbase}}/style.css" rel="stylesheet" type="text/css">
<title>Access denied</title>
<style>
body {
	background: linear-gradient(135deg, rgba(244,67,54,1) 0%, rgba(255, 193, 7,1) 100%);
}
</style>
</head>
<body>
	<div class="header-bar">
		<div class="container">
			<span>chuie.io authentication</span>
		</div>
	</div>
	<div class="container">
		<div class="half">
			<h1 class="welcome">Access denied</h1>
			<p>
				You don't seem to be permitted to access this. If you think
				this is in error, contact Jason (1lann) for help.</p>
			<p>
				If you would like to log into a different account, click the button
				below and you will be logged out.
			</p>
			<a class="button" href="{{.Authbase}}/logout">Switch Accounts</a>
		</div>
	</div>
	<div class="footer">
		<div class="footer-text container">
			You're on chuie.io. Made by Jason Chu (1lann) -
			<a href="mailto:me@chuie.io">me@chuie.io</a>.
		</div>
	</div>
</body>
</html>`

	var serverErrorSource = `<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no">
<link href="https://fonts.googleapis.com/css?family=Lato:400,300,100,700" rel="stylesheet" type="text/css">
<link href="{{.Authbase}}/style.css" rel="stylesheet" type="text/css">
<title>Server error</title>
<style>
body {
	background: linear-gradient(135deg, rgba(244,67,54,1) 0%, rgba(255, 193, 7,1) 100%);
}
</style>
</head>
<body>
	<div class="header-bar">
		<div class="container">
			<span>chuie.io authentication</span>
		</div>
	</div>
	<div class="container">
		<div class="half">
			<h1 class="welcome">Server error</h1>
			<p>
				Sorry, authentication is not available right now due to a
				server error. The resource you're trying to access will not be
				accessible until the error is resolved. You can try again later
				or contact Jason (1lann) for help.
			</p>
			{{- if .Error}}
			<p class="error-code">
				Error: {{.Error}}
			</p>
			{{- end}}
			<p><strong>Sorry again for the inconvenience!</strong></p>
		</div>
	</div>
	<div class="footer">
		<div class="footer-text container">
			You're on chuie.io. Made by Jason Chu (1lann) -
			<a href="mailto:me@chuie.io">me@chuie.io</a>.
		</div>
	</div>
</body>
</html>`

	var temporaryAccessSource = `<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no">
<link href="https://fonts.googleapis.com/css?family=Lato:400,300,100,700" rel="stylesheet" type="text/css">
<link href="/style.css" rel="stylesheet" type="text/css">
<title>Temporary access</title>
</head>
<body>
	<div class="header-bar">
		<div class="container">
			<span>chuie.io authentication</span>
		</div>
	</div>
	<div class="container">
		<div class="half">
			<h1 class="welcome">Temporary access</h1>
			<p>You're using a temporary access code which lasts for
			<strong>{{.Duration}}</strong>. After this time period, you'll need to request
			for another access code if you need one.</p>
			<a class="button" href="{{.Redirect}}">I understand</a>
		</div>
	</div>
	<div class="footer">
		<div class="footer-text container">
			You're on chuie.io. Made by Jason Chu (1lann) -
			<a href="mailto:me@chuie.io">me@chuie.io</a>.
		</div>
	</div>
</body>
</html>`

	var loggedInSource = `<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no">
<link href="https://fonts.googleapis.com/css?family=Lato:400,300,100,700" rel="stylesheet" type="text/css">
<link href="/style.css" rel="stylesheet" type="text/css">
<title>Logged in</title>
</head>
<body>
	<div class="header-bar">
		<div class="container">
			<span>chuie.io authentication</span>
		</div>
	</div>
	<div class="container">
		<div class="half">
			<h1 class="welcome">You're logged in</h1>
			<p>
				as <strong>{{.Username}}</strong> to chuie.io services.
				If you would like to log out, you can do so with the button below.
			</p>
			<a class="button" href="/logout">Log out</a>
		</div>
	</div>
	<div class="footer">
		<div class="footer-text container">
			You're on chuie.io. Made by Jason Chu (1lann) -
			<a href="mailto:me@chuie.io">me@chuie.io</a>.
		</div>
	</div>
</body>
</html>`

	styleSource = []byte(`/*! normalize.css v3.0.3 | MIT License | github.com/necolas/normalize.css */img,legend{border:0}legend,td,th{padding:0}html{font-family:sans-serif;-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%}body{margin:0}article,aside,details,figcaption,figure,footer,header,hgroup,main,menu,nav,section,summary{display:block}audio,canvas,progress,video{display:inline-block;vertical-align:baseline}audio:not([controls]){display:none;height:0}[hidden],template{display:none}a{background-color:transparent}a:active,a:hover{outline:0}abbr[title]{border-bottom:1px dotted}b,optgroup,strong{font-weight:700}dfn{font-style:italic}h1{font-size:2em;margin:.67em 0}mark{background:#ff0;color:#000}small{font-size:80%}sub,sup{font-size:75%;line-height:0;position:relative;vertical-align:baseline}sup{top:-.5em}sub{bottom:-.25em}svg:not(:root){overflow:hidden}figure{margin:1em 40px}hr{box-sizing:content-box;height:0}pre,textarea{overflow:auto}code,kbd,pre,samp{font-family:monospace,monospace;font-size:1em}button,input,optgroup,select,textarea{color:inherit;font:inherit;margin:0}button{overflow:visible}button,select{text-transform:none}button,html input[type=button],input[type=reset],input[type=submit]{-webkit-appearance:button;cursor:pointer}button[disabled],html input[disabled]{cursor:default}button::-moz-focus-inner,input::-moz-focus-inner{border:0;padding:0}input{line-height:normal}input[type=checkbox],input[type=radio]{box-sizing:border-box;padding:0}input[type=number]::-webkit-inner-spin-button,input[type=number]::-webkit-outer-spin-button{height:auto}input[type=search]{-webkit-appearance:textfield;box-sizing:content-box}input[type=search]::-webkit-search-cancel-button,input[type=search]::-webkit-search-decoration{-webkit-appearance:none}fieldset{border:1px solid silver;margin:0 2px;padding:.35em .625em .75em}table{border-collapse:collapse;border-spacing:0}

html {
	position: relative;
	min-height: 100%;
	margin: 0px;
	padding: 0px;
}

body {
	font-family: "Lato", "Helvetica Neue", "Helvetica", "Arial", sans-serif;
	background: rgb(0, 176, 255);
	background: linear-gradient(135deg, rgba(0, 176, 255,1) 0%, rgba(0,230,118,1) 100%);
	min-height: 100%;
	color: #ffffff;
	margin: 0px;
	padding: 0px;
	font-size: 24px;
	margin-bottom: 5rem;
}

.container {
	width: calc(100% - 40px);
	max-width: 1000px;
	margin-left: auto;
	margin-right: auto;
	padding: 0px 20px 0px 20px;
}

.half {
	max-width: 500px;
}

.welcome {
	font-weight: 300;
	font-size: 3em;
	margin-top: 0.4em;
	margin-bottom: 0.5em;
}

.header-bar {
	height: 60px;
	background: rgba(0, 0, 0, 0.15);
	display: flex;
	align-items: center;
	font-size: 1.2rem;
}

.error {
	color: #ffffff;
	background-color: #FF5722;
	padding: 20px;
	position: relative;
	border-radius: 4px;
}

.error-code {
	opacity: 0.6;
}

.help {
	opacity: 0.8;
	font-size: 0.8em;
}

a {
	text-decoration: none;
	color: #ffffff;
	opacity: 0.8;
	transition: opacity 0.2s;
}

a:hover {
	text-decoration: underline;
	opacity: 1;
}

form {
	width: 100%;
}

input[type=text], input[type=password] {
	background: rgba(0, 0, 0, 0.1);
	border: 1px solid rgba(255, 255, 255, 0.6);
	padding: 15px;
	width: calc(100% - 30px);
	transition: background 0.2s, border-color 0.2s;
}

input::-webkit-input-placeholder {
	color: rgba(255, 255, 255, 0.6);
}

input[type=text] {
	border-bottom-width: 0px;
	border-top-right-radius: 4px;
	border-top-left-radius: 4px;
}

input[type=text].access-code {
	border-bottom-right-radius: 4px;
	border-bottom-left-radius: 4px;
}

input[type=password] {
	border-bottom-right-radius: 4px;
	border-bottom-left-radius: 4px;
}

input[type=text]:focus, input[type=password]:focus {
	outline: 0px;
	background: rgba(0, 0, 0, 0.2);
	border-color: rgba(255, 255, 255, 1);
}

input[type=text]:focus + input[type=password] {
	border-top-color: rgba(255, 255, 255, 1);
}

input[type=text].access-code {
	border-bottom-width: 1px;
}

input[type=text].access-code + input[type=password] {
	display: none;
}

input[type=checkbox] {
	float: left;
	opacity: 0;
	width: 0px;
}

.custom-checkbox {
	float: left;
	margin: 0px;
	height: 1.2em;
	width: 1.2em;
	border: 1px solid rgba(255, 255, 255, 0.6);
	border-radius: 4px;
	transition: border-color 0.2s;
}

.remember-me {
	position: relative;
	left: 10px;
}

.checkbox-label {
	display: inline-block;
	margin-top: 1em;
	cursor: pointer;
}

input[type=checkbox]:focus + .custom-checkbox {
	border-color: rgba(255, 255, 255, 1);
}

input[type=checkbox]:hover + .custom-checkbox {
	border-color: rgba(255, 255, 255, 1);
}

.custom-checkbox:before {
	position: relative;
	content: 'X';
	top: -0.11em;
	left: 0.18em;
	font-size: 1.2em;
	opacity: 0;
	transition: opacity 0.2s;
}

input[type=checkbox]:checked + .custom-checkbox:before {
	opacity: 1;
}

input[type=submit], .button {
	display: block;
	margin-top: 1em;
	border: 1px solid rgba(255, 255, 255, 0.6);
	background: rgba(0, 0, 0, 0.1);
	border-radius: 4px;
	padding: 15px 20px 15px 20px;
	min-width: 200px;
	text-decoration: none;
	color: #ffffff;
	transition: background 0.2s, border-color 0.2s;
}

.button {
	display: inline-block;
	margin-top: 0.5em;
	opacity: 1;
	text-align: center;
	text-decoration: none !important;
}

input[type=submit]:hover, .button:hover {
	background: rgba(0, 0, 0, 0.05);
	border-color: rgba(255, 255, 255, 1);
}

input[type=submit]:active, .button:active {
	background: rgba(0, 0, 0, 0.2);
	border-color: rgba(255, 255, 255, 1);
	outline: 0px;
}

input[type=submit]:focus, .button:focus {
	border-color: rgba(255, 255, 255, 1);
	outline: 0px;
}

.footer {
	position: absolute;
	bottom: 0;
	left: 0;
	width: 100%;
	height: 3.5rem;
	font-size: 0.8em;
}

@media (max-width: 500px) {
	body {
		font-size: 18px;
	}

	.footer {
		font-size: 1em;
	}
}`)

	loginTemplate = template.Must(template.New("login").Parse(loginSource))
	deniedTemplate = template.Must(template.New("denied").Parse(deniedSource))
	serverErrorTemplate = template.Must(template.New("error").
		Parse(serverErrorSource))
	temporaryAccessTemplate = template.Must(template.New("access").
		Parse(temporaryAccessSource))
	loggedInTemplate = template.Must(template.New("loggedin").
		Parse(loggedInSource))
}
