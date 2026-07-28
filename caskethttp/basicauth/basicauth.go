// Copyright 2015 Light Code Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package basicauth implements HTTP Basic Authentication for Casket.
//
// This is useful for simple protections on a website, like requiring
// a password to access an admin interface. This package assumes a
// fairly small threat model.
package basicauth

import (
	"bufio"
	"context"
	"crypto/sha1"
	"crypto/subtle"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jimstudt/http-authentication/basic"
	"github.com/tmpim/casket/caskethttp/httpserver"
)

// BasicAuth is middleware to protect resources with a username and password.
// Note that HTTP Basic Authentication is not secure by itself and should
// not be used to protect important assets without HTTPS. Even then, the
// security of HTTP Basic Auth is disputed. Use discretion when deciding
// what to protect with BasicAuth.
type BasicAuth struct {
	Next     httpserver.Handler
	SiteRoot string
	Rules    []Rule
	Cookie   CookieConfig
}

type CookieConfig struct {
	Enabled   bool
	TTL       time.Duration
	Refresh   time.Duration
	NoRefresh bool
	Name      string
}

const defaultCookieTTL = time.Hour * 24
const defaultCookieName = "casket_basicauth"
const tsCookieSuffix = "_ts"

// ServeHTTP implements the httpserver.Handler interface.
func (a BasicAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	var protected, isAuthenticated bool
	var realm string
	var username string
	var password string
	var ok bool

	// do not check for basic auth on OPTIONS call
	if r.Method == http.MethodOptions {
		// Pass-through when no paths match
		return a.Next.ServeHTTP(w, r)
	}

	// if cookies are enabled for this site, and there's no auth header, pass through the cookie. note this would lead
	// to a possible desync if the user somehow changes the auth header while they have a cookie stored
	var cookie *http.Cookie
	cookieEnabled := a.Cookie.Enabled
	if cookieEnabled {
		if c, err := r.Cookie(a.Cookie.Name); err == nil && c.Value != "" {
			cookie = c

			if r.Header.Get("Authorization") == "" {
				r.Header.Set("Authorization", c.Value)
			}
		}
	}

ruleLoop:
	for _, rule := range a.Rules {
		for _, res := range rule.Resources {
			if !httpserver.Path(r.URL.Path).Matches(res) {
				continue
			}

			for _, exclude := range rule.Exclude {
				if httpserver.Path(r.URL.Path).Matches(exclude) {
					continue ruleLoop
				}
			}

			// path matches; this endpoint is protected
			protected = true
			realm = rule.Realm

			// parse auth header
			username, password, ok = r.BasicAuth()

			// check credentials
			if !ok ||
				username != rule.Username ||
				!rule.Password(password) {
				continue
			}

			// by this point, authentication was successful
			isAuthenticated = true

			// let upstream middleware (e.g. fastcgi and cgi) know about authenticated
			// user; this replaces the request with a wrapped instance
			r = r.WithContext(context.WithValue(r.Context(),
				httpserver.RemoteUserCtxKey, username))

			// Provide username to be used in log by replacer
			repl := httpserver.NewReplacer(r, nil, "-")
			repl.Set("user", username)

			// store the auth header in a cookie if we need to
			if cookieEnabled {
				if a.shouldUpdateCookie(r, cookie) {
					a.setCookies(w, r, r.Header.Get("Authorization"))
				}
			}
		}
	}

	if protected && !isAuthenticated {
		if cookieEnabled {
			a.clearCookies(w, r)
		}

		// browsers show a message that says something like:
		// "The website says: <realm>"
		// which is kinda dumb, but whatever.
		if realm == "" {
			realm = "Restricted"
		}
		w.Header().Set("WWW-Authenticate", "Basic realm=\""+realm+"\"")

		// Get a replacer so we can provide basic info for the authentication error.
		repl := httpserver.NewReplacer(r, nil, "-")
		repl.Set("user", username)
		errstr := repl.Replace("BasicAuth: user \"{user}\" was not found or password was incorrect. {remote} {host} {uri} {proto}")
		err := fmt.Errorf("%s", errstr)
		return http.StatusUnauthorized, err
	}

	// Pass-through when no paths match
	return a.Next.ServeHTTP(w, r)
}

// Rule represents a BasicAuth rule. A username and password
// combination protect the associated resources, which are
// file or directory paths.
type Rule struct {
	Username  string
	Password  func(string) bool
	Resources []string
	Exclude   []string
	Realm     string // See RFC 1945 and RFC 2617, default: "Restricted"
}

// PasswordMatcher determines whether a password matches a rule.
type PasswordMatcher func(pw string) bool

var (
	htpasswords   map[string]map[string]PasswordMatcher
	htpasswordsMu sync.Mutex
)

// GetHtpasswdMatcher matches password rules.
func GetHtpasswdMatcher(filename, username, siteRoot string) (PasswordMatcher, error) {
	filename = filepath.Join(siteRoot, filename)
	htpasswordsMu.Lock()
	if htpasswords == nil {
		htpasswords = make(map[string]map[string]PasswordMatcher)
	}
	pm := htpasswords[filename]
	if pm == nil {
		fh, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("open %q: %v", filename, err)
		}
		defer fh.Close()
		pm = make(map[string]PasswordMatcher)
		if err = parseHtpasswd(pm, fh); err != nil {
			return nil, fmt.Errorf("parsing htpasswd %q: %v", fh.Name(), err)
		}
		htpasswords[filename] = pm
	}
	htpasswordsMu.Unlock()
	if pm[username] == nil {
		return nil, fmt.Errorf("username %q not found in %q", username, filename)
	}
	return pm[username], nil
}

func parseHtpasswd(pm map[string]PasswordMatcher, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.IndexByte(line, '#') == 0 {
			continue
		}
		i := strings.IndexByte(line, ':')
		if i <= 0 {
			return fmt.Errorf("malformed line, no color: %q", line)
		}
		user, encoded := line[:i], line[i+1:]
		for _, p := range basic.DefaultSystems {
			matcher, err := p(encoded)
			if err != nil {
				return err
			}
			if matcher != nil {
				pm[user] = matcher.MatchesPassword
				break
			}
		}
	}
	return scanner.Err()
}

// PlainMatcher returns a PasswordMatcher that does a constant-time
// byte comparison against the password passw.
func PlainMatcher(passw string) PasswordMatcher {
	// compare hashes of equal length instead of actual password
	// to avoid leaking password length
	passwHash := sha1.New()
	if _, err := passwHash.Write([]byte(passw)); err != nil {
		log.Printf("[ERROR] unable to write password hash: %v", err)
	}
	passwSum := passwHash.Sum(nil)
	return func(pw string) bool {
		pwHash := sha1.New()
		if _, err := pwHash.Write([]byte(pw)); err != nil {
			log.Printf("[ERROR] unable to write password hash: %v", err)
		}
		pwSum := pwHash.Sum(nil)
		return subtle.ConstantTimeCompare([]byte(pwSum), []byte(passwSum)) == 1
	}
}

func (a BasicAuth) shouldUpdateCookie(r *http.Request, cookie *http.Cookie) bool {
	if cookie == nil {
		return true // we don't have a cookie, set a new one
	}

	if a.Cookie.NoRefresh || a.Cookie.Refresh <= 0 {
		return false // no refresh, leave it as-is to expire according to ttl
	}

	tsCookie, err := r.Cookie(a.Cookie.Name + tsCookieSuffix)
	if err != nil {
		return true // timestamp cookie missing, re-issue anyway i guess?
	}

	issuedAtUnix, err := strconv.ParseInt(tsCookie.Value, 10, 64)
	if err != nil {
		return true // whatever
	}

	issuedAt := time.Unix(issuedAtUnix, 0)

	// refresh if we're past time
	return time.Since(issuedAt) > a.Cookie.Refresh
}

func (a BasicAuth) setCookies(w http.ResponseWriter, r *http.Request, authHeader string) {
	now := time.Now()

	authCookie := &http.Cookie{
		Name:     a.Cookie.Name,
		Value:    authHeader,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}

	tsCookie := &http.Cookie{
		Name:     a.Cookie.Name + tsCookieSuffix,
		Value:    fmt.Sprintf("%d", now.Unix()),
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}

	if a.Cookie.TTL > 0 {
		exp := now.Add(a.Cookie.TTL)
		maxAge := int(a.Cookie.TTL.Seconds())

		authCookie.Expires = exp
		authCookie.MaxAge = maxAge

		tsCookie.Expires = exp
		tsCookie.MaxAge = maxAge
	}

	http.SetCookie(w, authCookie)
	http.SetCookie(w, tsCookie)
}

func (a BasicAuth) clearCookies(w http.ResponseWriter, r *http.Request) {
	deadAuth := &http.Cookie{
		Name:     a.Cookie.Name,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}

	deadTS := &http.Cookie{
		Name:     a.Cookie.Name + tsCookieSuffix,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, deadAuth)
	http.SetCookie(w, deadTS)
}
