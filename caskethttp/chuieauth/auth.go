package chuieauth

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/gob"
	"errors"
	"net/http"
	"strings"
	"time"

	rdb "github.com/dancannon/gorethink"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type databaseUser struct {
	Username  string    `gorethink:"username"`
	Hash      []byte    `gorethink:"password,omitempty"`
	Permitted []string  `gorethink:"permitted"`
	Expiry    time.Time `gorethink:"expiry"`
}

type sessionData struct {
	Username    string
	Password    string // Actual password
	Renewal     int64  // Cookie must be renewed
	Expiry      int64  // Expires (logs out) no matter what
	Permitted   []string
	IsTemporary bool
	RememberMe  bool
}

const permanentSessionDuration = 2147483646

var dbSession *rdb.Session
var authHostBase string
var sessionRenewalDuration = time.Hour * 12
var sessionStore *sessions.CookieStore
var janitorRunning = false

type verifyError struct {
	message string
}

func (e *verifyError) Error() string {
	return e.message
}

type serverError struct {
	message string
}

func (e *serverError) Error() string {
	return e.message
}

func authenticateSession(session *sessions.Session) (string, []string,
	error) {
	sessionUser, ok := session.Values["SessionUser"].(*sessionData)
	var permitted []string
	var username string

	if !ok {
		return "", []string{}, &verifyError{}
	}

	if !time.Unix(sessionUser.Expiry, 0).Before(time.Unix(0, 0)) &&
		time.Now().After(time.Unix(sessionUser.Expiry, 0)) {
		return "", []string{}, &verifyError{expiredMessage}
	}

	if time.Now().After(time.Unix(sessionUser.Renewal, 0)) {
		if !sessionUser.RememberMe {
			return "", []string{}, &verifyError{expiredMessage}
		}

		user, isTemp, err := queryAuthentication(sessionUser.Username,
			sessionUser.Password)
		if err == errAuthFail {
			return "", []string{}, &verifyError{revokedMessage}
		} else if err != nil {
			return "", []string{}, &serverError{err.Error()}
		}

		sessionUser.Username = user.Username
		sessionUser.Renewal = time.Now().Add(sessionRenewalDuration).Unix()
		sessionUser.Expiry = user.Expiry.Unix()
		sessionUser.Permitted = user.Permitted
		sessionUser.IsTemporary = isTemp

		session.Values["SessionUser"] = sessionUser
		if user.Expiry.Before(time.Unix(0, 0)) {
			session.Options.MaxAge = permanentSessionDuration
		} else {
			session.Options.MaxAge = int(user.Expiry.
				Sub(time.Now()).Seconds() + 1)
		}

		if err != nil {
			return "", []string{}, &serverError{err.Error()}
		}

		permitted = user.Permitted
		username = user.Username
	} else {
		permitted = sessionUser.Permitted
		username = sessionUser.Username
	}

	return username, permitted, nil
}

func (c *ChuieAuth) authenticateSessionAndReturn(w http.ResponseWriter,
	r *http.Request, domainAuthId string) (int, error) {
	session, _ := sessionStore.Get(r, "chuieauth")

	username, permittedDomains, err := authenticateSession(session)

	if err != nil {
		if c.authOptional {
			code, err := c.Next.ServeHTTP(w, r)
			if code < 400 {
				session.Save(r, w)
			}
			return code, err
		}

		switch err := err.(type) {
		case *verifyError:
			return serveVerifyError(w, r, session, err.Error())
		default:
			session.Save(r, w)
			return serveError(w, r, err)
		}
	}

	r.Header.Add("X-Auth-Username", username)
	r.Header.Add("X-Auth-Permitted", strings.Join(permittedDomains, ","))

	authorized := false
	if c.authOptional {
		authorized = true
	}

	for _, permitted := range permittedDomains {
		if permitted == domainAuthId || permitted == "*" {
			authorized = true
			break
		}
	}

	if !authorized {
		session.Save(r, w)
		return serveDenied(w, r)
	}

	code, err := c.Next.ServeHTTP(w, r)
	if code < 400 {
		session.Save(r, w)
	}
	return code, err
}

func serveAuth(w http.ResponseWriter, r *http.Request, session *sessions.Session,
	username string, password string, rememberMe bool) error {
	user, isTemp, err := queryAuthentication(username, password)
	if err == errAuthFail {
		return &verifyError{usernamePasswordMessage}
	} else if err != nil {
		return &serverError{err.Error()}
	}

	sessionUser := &sessionData{
		Username:    user.Username,
		Password:    password,
		Renewal:     time.Now().Add(sessionRenewalDuration).Unix(),
		Expiry:      user.Expiry.Unix(),
		Permitted:   user.Permitted,
		IsTemporary: isTemp,
		RememberMe:  rememberMe,
	}

	session.Values["SessionUser"] = sessionUser

	if !sessionUser.RememberMe {
		session.Options.MaxAge = 0
	} else if time.Unix(sessionUser.Expiry, 0).Before(time.Unix(0, 0)) {
		session.Options.MaxAge = permanentSessionDuration
	} else {
		session.Options.MaxAge = int(user.Expiry.
			Sub(time.Now()).Seconds() + 1)
	}

	return nil
}

func setupSessionStore(encryptionKey string) {
	key := sha512.Sum512([]byte(encryptionKey))
	blockKey := sha256.Sum256([]byte(encryptionKey))
	gob.Register(&sessionData{})
	sessionStore = sessions.NewCookieStore(key[:], blockKey[:])
	sessionStore.Options.Domain = ".chuie.io"
}

func queryAuthentication(username string,
	password string) (databaseUser, bool, error) {
	if len(username) >= 2 && username[:2] == "ac" {
		user, err := queryAccessCode(username)
		if !user.Expiry.Before(time.Unix(0, 0)) &&
			time.Now().After(user.Expiry) {
			return databaseUser{}, true, errAuthFail
		}
		return user, true, err
	}

	user, err := queryUsernamePassword(username, password)
	if !user.Expiry.Before(time.Unix(0, 0)) &&
		time.Now().After(user.Expiry) {
		return databaseUser{}, true, errAuthFail
	}
	return user, false, err
}

func queryUsernamePassword(username string,
	password string) (databaseUser, error) {
	if dbSession == nil {
		return databaseUser{},
			errors.New("chuieauth: no available database sessions")
	}

	c, err := rdb.Table("users").Get(username).Run(dbSession)
	if err != nil {
		return databaseUser{}, errAuthFail
	}

	var result databaseUser
	err = c.One(&result)
	if err != nil {
		return databaseUser{}, errAuthFail
	}

	err = bcrypt.CompareHashAndPassword(result.Hash, []byte(password))
	if err != nil {
		return databaseUser{}, errAuthFail
	}

	return result, nil
}

func queryAccessCode(accessCode string) (databaseUser, error) {
	if dbSession == nil {
		return databaseUser{},
			errors.New("chuieauth: no available database sessions")
	}

	c, err := rdb.Table("access_codes").Get(accessCode).Run(dbSession)
	if err != nil {
		return databaseUser{}, errAuthFail
	}

	var result databaseUser
	err = c.One(&result)
	if err != nil {
		return databaseUser{}, errAuthFail
	}

	result.Username = accessCode
	return result, nil
}

func janitor() {
	if janitorRunning {
		return
	}

	janitorRunning = true

	for {
		time.Sleep(time.Minute)
		if dbSession == nil {
			continue
		}

		// Trash everything that's expired
		rdb.Table("access_codes").Between(rdb.MinVal, time.Now(),
			rdb.BetweenOpts{Index: "expiry"}).Delete().RunWrite(dbSession)
	}
}

func connectToDatabase(host string, authKey string) error {
	var err error
	dbSession, err = rdb.Connect(rdb.ConnectOpts{
		Address:  host,
		Password: authKey,
		Database: "chuieauth",
		MaxIdle:  5,
		MaxOpen:  5,
	})
	go janitor()
	return err
}
