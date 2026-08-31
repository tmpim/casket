package tmpauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/tmpim/casket/caskethttp/httpserver"
)

type StatusResponse struct {
	Tmpauth        bool            `json:"tmpauth"`
	ClientID       string          `json:"clientID"`
	IsLoggedIn     bool            `json:"isLoggedIn"`
	UserDescriptor json.RawMessage `json:"loggedInUser,omitempty"`
}

func (t *Tmpauth) serveStatus(w http.ResponseWriter, r *http.Request, token *CachedToken) (int, error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := &StatusResponse{
		Tmpauth:    true,
		ClientID:   t.Config.ClientID,
		IsLoggedIn: token != nil,
	}

	if token != nil {
		resp.UserDescriptor = json.RawMessage(token.UserDescriptor)
	}

	json.NewEncoder(w).Encode(resp)

	return 0, nil
}

func (t *Tmpauth) janitor() {
	t.DebugLog("token cache janitor started")

	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-t.done:
			t.DebugLog("stopping token cache janitor")
			ticker.Stop()
			return
		case <-ticker.C:
			t.DebugLog("running token cache janitor")

			t.tokenCacheMutex.Lock()

			now := time.Now()
			for k, v := range t.TokenCache {
				if now.After(v.RevalidateAt) {
					delete(t.TokenCache, k)
				}
			}

			t.tokenCacheMutex.Unlock()
		}
	}
}

func (t *Tmpauth) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	t.janitorOnce.Do(func() {
		go t.janitor()
		backgroundWorker.Start()
	})

	if len(t.Config.Headers) > 0 {
		for header := range t.Config.Headers {
			r.Header.Del(header)
		}
	}

	statusRequested := false
	whomstRequested := false

	if httpserver.Path(r.URL.Path).Matches("/.well-known/tmpauth/") {
		switch strings.TrimPrefix(r.URL.Path, "/.well-known/tmpauth/") {
		case "callback":
			return t.authCallback(w, r)
		case "status":
			statusRequested = true
			break
		case "whomst":
			whomstRequested = true
			break
		default:
			return http.StatusBadRequest, fmt.Errorf("tmpauth: no such path")
		}
	}

	// determine if auth is required
	authRequired := true

	// If the URL path is weird, it signals a possible attack attempt.
	// always require authentication in such a condition.
	if path.Clean(r.URL.Path) == r.URL.Path {
		if len(t.Config.Except) > 0 {
			for _, exempt := range t.Config.Except {
				if httpserver.Path(r.URL.Path).Matches(exempt) {
					authRequired = false
					break
				}
			}
		} else if len(t.Config.Include) > 0 {
			found := false
			for _, included := range t.Config.Include {
				if httpserver.Path(r.URL.Path).Matches(included) {
					found = true
					break
				}
			}
			if !found {
				authRequired = false
			}
		}
	} else {
		t.DebugLog("url path is suspicious, authentication being mandated: %v != %v",
			path.Clean(r.URL.Path), r.URL.Path)
	}

	t.DebugLog("auth requirement for %q: %v", r.URL.Path, authRequired)

	cachedToken, err := t.authFromCookie(r)
	if err != nil {
		t.DebugLog("failed to get JWT token: %v", err)

		if _, err := r.Cookie(t.CookieName()); err != http.ErrNoCookie {
			t.DebugLog("cookie exists and deemed to be invalid, requesting client to delete cookie")

			http.SetCookie(w, &http.Cookie{
				Name:     t.CookieName(),
				Value:    "",
				MaxAge:   -1,
				Path:     "/",
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
		}

		// Not authed, return an empty status or whomst response if requested
		if statusRequested {
			return t.serveStatus(w, r, nil)
		} else if whomstRequested {
			return t.serveWhomst(w, nil)
		}

		// Begin auth flow
		if authRequired {
			return t.startAuth(w, r)
		}
	} else if len(t.Config.Headers) > 0 {
		err := t.SetHeaders(cachedToken, r.Header)
		if err != nil {
			t.DebugLog("failed to set headers: %v", err)
			return http.StatusPreconditionRequired, fmt.Errorf("tmpauth: missing required header value")
		}
	}

	// Token is available (authenticated, but not necessarily allowed), serve the status response if requested
	if statusRequested {
		return t.serveStatus(w, r, cachedToken)
	}

	if authRequired {
		userAuthorized := false
		if len(t.Config.AllowedUsers) > 0 {
			t.DebugLog("checking if user is allowed on allowed users list: %v", cachedToken.UserIDs)
			userIDs := make(map[string]bool)
			for _, userID := range cachedToken.UserIDs {
				userIDs[userID] = true
			}

			for _, allowedUser := range t.Config.AllowedUsers {
				if userIDs[allowedUser] {
					userAuthorized = true
					break
				}
			}
		} else {
			userAuthorized = true
		}

		if !userAuthorized {
			t.DebugLog("user not on allowed users list")
			return http.StatusForbidden, fmt.Errorf("tmpauth: user not in allowed list")
		}
	}

	// Now serve the whomst response if requested (authenticated and authorized)
	if whomstRequested {
		return t.serveWhomst(w, cachedToken)
	}

	return t.Next.ServeHTTP(w, r)
}

func (t *Tmpauth) startAuth(w http.ResponseWriter, r *http.Request) (int, error) {
	now := time.Now()
	expiry := time.Now().Add(5 * time.Minute)
	tokenID := generateTokenID()

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &stateClaims{
		CallbackURL: "https://" + t.Config.Host.Host + t.Config.Host.Path + "/.well-known/tmpauth/callback",
		StandardClaims: jwt.StandardClaims{
			Id:        tokenID,
			Issuer:    TmpAuthHost + ":server:" + t.Config.ClientID,
			Audience:  TmpAuthHost + ":central:state",
			IssuedAt:  now.Unix(),
			NotBefore: now.Unix(),
			ExpiresAt: expiry.Unix(),
		},
	}).SignedString(t.Config.Secret)
	if err != nil {
		t.DebugLog("failed to sign state token: %v", err)
		return http.StatusInternalServerError, errors.New("tmpauth: failed to start authentication")
	}

	requestURI := r.URL.RequestURI()
	if requestURI[0] != '/' {
		requestURI = "/" + requestURI
	}

	t.stateIDCache.SetDefault(tokenID, requestURI)

	// store request URIs in cookies sometimes in case this is a distributed
	// casket instance or something and it'll still work in most cases
	if len(requestURI) <= 128 {
		http.SetCookie(w, &http.Cookie{
			Name:     t.StateIDCookieName(tokenID),
			Value:    url.QueryEscape(requestURI),
			Expires:  expiry,
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	} else {
		http.SetCookie(w, &http.Cookie{
			Name:     t.StateIDCookieName(tokenID),
			Value:    "ok",
			Expires:  expiry,
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	queryParams := url.Values{
		"state":     []string{token},
		"client_id": []string{t.Config.ClientID},
		"method":    []string{"tmpauth"},
	}

	http.Redirect(w, r, "https://"+TmpAuthHost+"/auth?"+queryParams.Encode(), http.StatusSeeOther)

	return 0, nil
}

// authFromCookie attempts to get the auth token from the cookie or the X-Tmpauth-Token header, and returns the
// cachedToken (if it was successfully parsed), and any error.
func (t *Tmpauth) authFromCookie(r *http.Request) (*CachedToken, error) {
	token := r.Header.Get("X-Tmpauth-Token")
	if token != "" {
		return t.parseWrappedAuthJWT(token)
	}

	cookie, err := r.Cookie(t.CookieName())
	if err != nil {
		return nil, err
	}

	return t.parseWrappedAuthJWT(cookie.Value)
}

type WhomstResponse struct {
	Whomst json.RawMessage `json:"whomst"`
}

// serveWhomst returns the entire whomst database if the user is logged in.
func (t *Tmpauth) serveWhomst(w http.ResponseWriter, token *CachedToken) (int, error) {
	// If the user is not logged in, return an error
	if token == nil {
		return http.StatusUnauthorized, fmt.Errorf("tmpauth: must be logged in to retrieve whomst database")
	}

	// Fetch the whomst database from tmpauth using the user's token
	resp, err := t.HttpClient.Get("https://" + TmpAuthHost + "/whomst/tmpauth/db?token=" + url.QueryEscape(token.RawToken))
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("tmpauth: failed to create request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return http.StatusInternalServerError, fmt.Errorf("tmpauth: got non OK response when retrieving whomst database: %v", resp.Status)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("tmpauth: failed to write response: %w", err)
	}

	return 0, nil
}
