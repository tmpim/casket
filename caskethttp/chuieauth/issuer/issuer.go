package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	r "github.com/dancannon/gorethink"
	"github.com/mailgun/mailgun-go"
)

type databaseUser struct {
	Username  string    `gorethink:"username"`
	Hash      []byte    `gorethink:"password,omitempty"`
	Permitted []string  `gorethink:"permitted"`
	Expiry    time.Time `gorethink:"expiry"`
}

var issued = make(map[string]time.Time)
var issuedMutex = new(sync.Mutex)
var session *r.Session
var issueDuration = time.Hour * 6
var emailSuffix = "@student.ccgs.wa.edu.au"

func issueNewAccessCode() (string, error) {
	var accessCode string

	for {
		rawAccessCode := make([]byte, 8)
		num, err := rand.Read(rawAccessCode)
		if num != 8 {
			return "", errors.New("issuer: read discrepancy")
		}
		if err != nil {
			return "", err
		}

		accessCode = "ac" + hex.EncodeToString(rawAccessCode)

		resp, err := r.Table("access_codes").Insert(databaseUser{
			Username:  accessCode,
			Permitted: []string{"bok"},
			Expiry:    time.Now().Add(issueDuration),
		}).RunWrite(session)
		if err != nil {
			return "", err
		}

		if resp.Errors > 0 {
			continue
		}

		break
	}

	return accessCode, nil
}

func emailRegistrationCode(email, accessCode string) {
	gun := mailgun.NewMailgun("bot.chuie.io", "[redacted]")

	msg := gun.NewMessage(
		"chuie.io bot <hello@bot.chuie.io>",
		"Your access code: "+accessCode,
		"Hello!\n\nThanks for using chuie.io services, your access "+
			"code is\n\n"+accessCode+"\n\nPlease follow the instructions "+
			"on the next page you were on.\nIf you need additional help, "+
			"you can contact my master at me@chuie.io\n\nRegards,\nchuie.io "+
			"bot\n\nP.S. Yes I am really only a bot, and no I am not "+
			"sentient.",
		email,
	)

	_, _, err := gun.Send(msg)
	if err != nil {
		log.Println("failed to send message:", err)
	}
}

func verifyEmail(email string) {
	log.Println("verifying:", email)

	if !strings.HasSuffix(email, emailSuffix) {
		return
	}
	username := strings.TrimSuffix(email, emailSuffix)
	parsedUsername, err := strconv.Atoi(username)
	if err != nil {
		return
	}

	issuedMutex.Lock()
	_, exists := issued[strconv.Itoa(parsedUsername)]
	issuedMutex.Unlock()

	if exists {
		return
	}

	accessCode, err := issueNewAccessCode()
	if err != nil {
		log.Println("issue access code error:", err)
		return
	}

	log.Println("issued access code:", parsedUsername, ":",
		accessCode)

	issuedMutex.Lock()
	issued[strconv.Itoa(parsedUsername)] =
		time.Now().Add(issueDuration)
	issuedMutex.Unlock()

	emailRegistrationCode(strconv.Itoa(parsedUsername)+
		emailSuffix, accessCode)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Auth-Username") != "" {
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}

	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		email := r.Form.Get("email")
		if email == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		go verifyEmail(email)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(checkSource))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(registrationSource))
}

func styleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(styleSource)
}

func main() {
	for {
		var err error
		session, err = r.Connect(r.ConnectOpts{
			Address:  "127.0.0.1:28015",
			AuthKey:  os.Getenv("DBAUTH"),
			Database: "chuieauth",
			MaxIdle:  5,
			MaxOpen:  5,
		})
		if err != nil {
			log.Println("failed to connect to database:", err)
			time.Sleep(time.Second * 2)
		} else {
			break
		}
	}

	http.HandleFunc("/", registerHandler)
	http.HandleFunc("/style.css", styleHandler)
	go janitor()
	log.Fatal(http.ListenAndServe(":9110", nil))
}

func janitor() {
	for {
		time.Sleep(time.Minute)
		for key, issue := range issued {
			if time.Now().After(issue) {
				issuedMutex.Lock()
				delete(issued, key)
				issuedMutex.Unlock()
			}
		}
	}
}

var registrationSource = []byte(`<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no">
<link href="https://fonts.googleapis.com/css?family=Lato:400,300,100,700" rel="stylesheet" type="text/css">
<link href="/style.css" rel="stylesheet" type="text/css">
<title>Limited access</title>
<style>
body {
	background: linear-gradient(135deg, rgba(170,0,255,1) 0%, rgba(2,136,209,1) 100%);
}
input[name=email] {
	border-bottom-width: 1px;
	border-radius: 4px;
}
</style>
</head>
<body>
	<div class="header-bar">
		<div class="container">
			<span>chuie.io access code issuer</span>
		</div>
	</div>
	<div class="container">
		<div class="half">
			<h1 class="welcome">Limited access</h1>
			<p>
				To limit and prevent unauthorized use of books.chuie.io, a
				barrier to entry has been made. You now require a valid
				access code which lasts for 6 hours in order to access
				books.chuie.io. You may retrieve your access code at any time
				by entering your school email below.
			</p>
			<form action="/" method="post">
				<input name="email" required type="text" placeholder="School email address">
				<input type="submit" value="Submit">
			</form>
		</div>
	</div>
	<div class="footer">
		<div class="footer-text container">
			You're on chuie.io. Made by Jason Chu (1lann) -
			<a href="mailto:me@chuie.io">me@chuie.io</a>.
		</div>
	</div>
</body>
</html>`)

var checkSource = []byte(`<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no">
<link href="https://fonts.googleapis.com/css?family=Lato:400,300,100,700" rel="stylesheet" type="text/css">
<link href="/style.css" rel="stylesheet" type="text/css">
<title>Check your email</title>
<style>
body {
	background: linear-gradient(135deg, rgba(170,0,255,1) 0%, rgba(2,136,209,1) 100%);
}
input[name=email] {
	border-bottom-width: 1px;
	border-radius: 4px;
}
</style>
</head>
<body>
	<div class="header-bar">
		<div class="container">
			<span>chuie.io access code issuer</span>
		</div>
	</div>
	<div class="container">
		<div class="half">
			<h1 class="welcome">Check your inbox</h1>
			<p>
				Check your email inbox for the access code to use to login at
				<a href="/home">books.chuie.io/home</a>. If you
				didn't receive an email, make sure you entered your
				<strong>school email address</strong> correctly and that
				you haven't already requested for an access code in the
				past 6 hours.
			</p>
			<p>
				Contact Jason (<a href="mailto:me@chuie.io">me@chuie.io</a>)
				if you have any issues.
			</p>
		</div>
	</div>
	<div class="footer">
		<div class="footer-text container">
			You're on chuie.io. Made by Jason Chu (1lann) -
			<a href="mailto:me@chuie.io">me@chuie.io</a>.
		</div>
	</div>
</body>
</html>`)

var styleSource = []byte(`/*! normalize.css v3.0.3 | MIT License | github.com/necolas/normalize.css */img,legend{border:0}legend,td,th{padding:0}html{font-family:sans-serif;-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%}body{margin:0}article,aside,details,figcaption,figure,footer,header,hgroup,main,menu,nav,section,summary{display:block}audio,canvas,progress,video{display:inline-block;vertical-align:baseline}audio:not([controls]){display:none;height:0}[hidden],template{display:none}a{background-color:transparent}a:active,a:hover{outline:0}abbr[title]{border-bottom:1px dotted}b,optgroup,strong{font-weight:700}dfn{font-style:italic}h1{font-size:2em;margin:.67em 0}mark{background:#ff0;color:#000}small{font-size:80%}sub,sup{font-size:75%;line-height:0;position:relative;vertical-align:baseline}sup{top:-.5em}sub{bottom:-.25em}svg:not(:root){overflow:hidden}figure{margin:1em 40px}hr{box-sizing:content-box;height:0}pre,textarea{overflow:auto}code,kbd,pre,samp{font-family:monospace,monospace;font-size:1em}button,input,optgroup,select,textarea{color:inherit;font:inherit;margin:0}button{overflow:visible}button,select{text-transform:none}button,html input[type=button],input[type=reset],input[type=submit]{-webkit-appearance:button;cursor:pointer}button[disabled],html input[disabled]{cursor:default}button::-moz-focus-inner,input::-moz-focus-inner{border:0;padding:0}input{line-height:normal}input[type=checkbox],input[type=radio]{box-sizing:border-box;padding:0}input[type=number]::-webkit-inner-spin-button,input[type=number]::-webkit-outer-spin-button{height:auto}input[type=search]{-webkit-appearance:textfield;box-sizing:content-box}input[type=search]::-webkit-search-cancel-button,input[type=search]::-webkit-search-decoration{-webkit-appearance:none}fieldset{border:1px solid silver;margin:0 2px;padding:.35em .625em .75em}table{border-collapse:collapse;border-spacing:0}

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
