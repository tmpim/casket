package chuieauth

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

const revokedMessage = "Sorry, your account access has been revoked."
const expiredMessage = "Sorry, your account access has expired."
const verifyFailMessage = "Sorry, your account could not be verified. " +
	"Please login again."
const usernamePasswordMessage = "Sorry, your credentials aren't valid. " +
	"Check they're up to date and that you didn't mistype them."
const serverErrorMessage = "Sorry, a server error occured while logging " +
	"you in. Please try again later."

var router *mux.Router

func (c *ChuieAuth) authMux(w http.ResponseWriter,
	r *http.Request) (int, error) {
	router.ServeHTTP(w, r)
	internalErr := r.Header.Get("X-Auth-Internal-Error")
	r.Header.Del("X-Auth-Internal-Error")
	if internalErr == "" {
		return 0, nil
	} else {
		return 0, errors.New(internalErr)
	}
}

func serveLogin(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "chuieauth")

	_, _, err := authenticateSession(session)
	if err == nil {
		session.Save(r, w)
		http.Redirect(w, r, authHostBase+"/loggedin", http.StatusFound)
		return
	}

	if _, matchesType := err.(*verifyError); matchesType &&
		err.Error() != "" {
		delete(session.Values, "SessionUser")
		session.AddFlash(err.Error(), "error")
	}

	var data struct {
		Error string
	}

	errorFlashes := session.Flashes("error")

	if len(errorFlashes) > 0 {
		data.Error, _ = errorFlashes[0].(string)
	}

	session.Save(r, w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if data.Error != "" {
		loginTemplate.Execute(w, data)
		return
	}

	loginTemplate.Execute(w, nil)
}

func serveLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "chuieauth")
	delete(session.Values, "SessionUser")
	session.Save(r, w)

	http.Redirect(w, r, authHostBase, http.StatusFound)
}

func serveError(w http.ResponseWriter, r *http.Request,
	err error) (int, error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)

	if err != nil {
		serverErrorTemplate.Execute(w, struct {
			Error    string
			Authbase string
		}{
			Error:    err.Error(),
			Authbase: "https://chuie.io",
		})
	} else {
		serverErrorTemplate.Execute(w, nil)
	}

	return 0, err
}

func serveDenied(w http.ResponseWriter,
	r *http.Request) (int, error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	// TODO: Change when HTTPS is available.
	deniedTemplate.Execute(w, struct{ Authbase string }{"https://chuie.io"})
	return 0, nil
}

func serveVerifyError(w http.ResponseWriter, r *http.Request,
	session *sessions.Session, message string) (int, error) {
	delete(session.Values, "SessionUser")
	session.AddFlash(message, "error")
	session.Flashes("redirect") // Clear out existing redirects
	session.AddFlash("//"+r.Host+r.URL.Path, "redirect")
	session.Save(r, w)

	http.Redirect(w, r, authHostBase, http.StatusFound)
	return 0, nil
}

func servePostAuth(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "chuieauth")

	err := r.ParseForm()
	if err != nil {
		session.AddFlash(usernamePasswordMessage, "error")
		session.Save(r, w)
		http.Redirect(w, r, authHostBase, http.StatusFound)
		return
	}

	username := r.Form.Get("username")
	password := r.Form.Get("password")
	rememberMe := r.Form.Get("remember")

	err = serveAuth(w, r, session, username, password, rememberMe == "on")
	if err != nil {
		var errorMessage string
		if _, matchesType := err.(*verifyError); matchesType {
			errorMessage = err.Error()
		} else {
			r.Header.Add("X-Internal-Error", err.Error())
			errorMessage = serverErrorMessage
		}

		session.AddFlash(errorMessage, "error")
		session.Save(r, w)
		http.Redirect(w, r, authHostBase, http.StatusFound)
		return
	}

	sessionUser, ok := session.Values["SessionUser"].(*sessionData)
	if !ok {
		session.Save(r, w)
		serveError(w, r, errors.New("Failed to read SessionUser"))
		return
	}

	if sessionUser.IsTemporary {
		session.Save(r, w)
		http.Redirect(w, r, authHostBase+"/continue", http.StatusFound)
		return
	}

	redirect := session.Flashes("redirect")

	session.Save(r, w)

	if len(redirect) > 0 {
		http.Redirect(w, r, redirect[0].(string), http.StatusFound)
		return
	}

	http.Redirect(w, r, authHostBase+"/loggedin", http.StatusFound)
	return
}

func serveBasicAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Add("WWW-Authenticate", "Basic realm=\"Enter your "+
			"chuie.io authentication credentials to login\"")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Valid chuie.io authentication credentials are " +
			"required to login"))
		return
	}

	session, _ := sessionStore.Get(r, "chuieauth")

	err := serveAuth(w, r, session, username, password, false)
	session.Save(r, w)

	if err != nil {
		if _, matchesType := err.(*verifyError); matchesType {
			w.Header().Add("WWW-Authenticate", "Basic realm=\"Enter your "+
				"chuie.io authentication credentials to login\"")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		} else {
			r.Header.Add("X-Internal-Error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(serverErrorMessage))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully authenticated"))
	return
}

func serveBasicAuthLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "chuieauth")
	delete(session.Values, "SessionUser")
	session.Save(r, w)

	http.Redirect(w, r, "//logout@"+authHostBase[2:]+"/basic", http.StatusFound)
}

func serveLoggedIn(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "chuieauth")

	username, _, err := authenticateSession(session)
	session.Save(r, w)

	if err != nil {
		http.Redirect(w, r, authHostBase, http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	loggedInTemplate.Execute(w, struct{ Username string }{username})
}

func serveTemporaryAccess(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "chuieauth")

	_, _, err := authenticateSession(session)
	if err != nil {
		session.Save(r, w)
		http.Redirect(w, r, authHostBase, http.StatusFound)
		return
	}

	sessionUser, ok := session.Values["SessionUser"].(*sessionData)
	if !ok {
		session.Save(r, w)
		serveError(w, r, errors.New("Failed to read SessionUser"))
		return
	}

	if !sessionUser.IsTemporary {
		session.Save(r, w)
		http.Redirect(w, r, authHostBase, http.StatusFound)
		return
	}

	var duration string
	remaining := time.Unix(sessionUser.Expiry, 0).Sub(time.Now())
	if remaining.Hours() >= 23.5 {
		duration = strconv.FormatFloat(
			math.Floor((remaining.Hours()/24.0)+0.5), 'f', -1, 64) + " days"
	} else if remaining.Minutes() > 59 {
		duration = strconv.FormatFloat(
			math.Floor(remaining.Hours()+0.5), 'f', -1, 64) + " hours"
	} else {
		duration = strconv.FormatFloat(
			math.Floor(remaining.Minutes()+0.5), 'f', -1, 64) + " minutes"
	}

	if duration[:2] == "1 " {
		duration = duration[:len(duration)-1]
	}

	redirectLocation := authHostBase
	redirect := session.Flashes("redirect")
	if len(redirect) > 0 {
		redirectLocation, ok = redirect[0].(string)
		if !ok {
			redirectLocation = authHostBase
		}
	}

	session.Save(r, w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	temporaryAccessTemplate.Execute(w, struct {
		Duration string
		Redirect string
	}{
		Duration: duration,
		Redirect: redirectLocation,
	})
}

func serveStyle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(styleSource)
}

func init() {
	router = mux.NewRouter()
	router.HandleFunc("/", serveLogin).Methods("GET")
	router.HandleFunc("/", servePostAuth).Methods("POST")
	router.HandleFunc("/basic", serveBasicAuth).Methods("GET")
	router.HandleFunc("/basic/logout", serveBasicAuthLogout).Methods("GET")
	router.HandleFunc("/continue", serveTemporaryAccess).Methods("GET")
	router.HandleFunc("/loggedin", serveLoggedIn).Methods("GET")
	router.HandleFunc("/logout", serveLogout).Methods("GET")
	router.HandleFunc("/style.css", serveStyle).Methods("GET")
}
