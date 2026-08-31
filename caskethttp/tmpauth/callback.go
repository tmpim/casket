package tmpauth

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var ErrInvalidCallbackToken = &CallbackError{
	errorCode:    "invalid_token",
	humanMessage: "The callback token from tmpauth failed to be validated",
}

type CallbackError struct {
	errorCode    string
	humanMessage string
}

func (c *CallbackError) Error() string {
	return "tmpauth: callback failed with error code: " + c.errorCode
}

func (t *Tmpauth) failRedirect(w http.ResponseWriter, r *http.Request, err *CallbackError) (int, error) {
	params := make(url.Values)
	params.Set("error", err.errorCode)
	if err.humanMessage != "" {
		params.Set("error_description", err.humanMessage)
	}

	if t.Config.Redirect != "" {
		http.Redirect(w, r, t.Config.Redirect+"?"+params.Encode(), http.StatusSeeOther)
		return 0, nil
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("X-Tmpauth-Error", err.errorCode)
	w.WriteHeader(http.StatusBadRequest)
	w.Write(append([]byte(err.humanMessage), '\n'))

	return 0, nil
}

func (t *Tmpauth) authCallback(w http.ResponseWriter, r *http.Request) (int, error) {
	params := r.URL.Query()

	t.DebugLog("executing authCallback flow")

	// We use metatokens in case the primary token is so large that it cannot fit in a URL query parameter.
	// Hence we retrieve the whole token out of band using a token ID.
	tokenStr := params.Get("token")
	stateStr := params.Get("state")

	state, err := jwt.ParseWithClaims(stateStr, &stateClaims{
		clientID: t.Config.ClientID,
	}, t.VerifyWithSecret)
	if err != nil {
		t.DebugLog("failed to verify state token: %v", err)
		return t.failRedirect(w, r, ErrInvalidCallbackToken)
	}

	claims := state.Claims.(*stateClaims)

	redirectURI, err := t.consumeStateID(r, w, claims.Id)
	if err != nil {
		t.DebugLog("failed to verify state ID against session: %v", err)
		return t.failRedirect(w, r, ErrInvalidCallbackToken)
	}

	if params.Get("error") != "" {
		return t.failRedirect(w, r, &CallbackError{
			errorCode:    params.Get("error"),
			humanMessage: params.Get("error_description"),
		})
	}

	token, err := t.parseAuthJWT(tokenStr, backgroundWorker.MinValidationTime())
	if err != nil {
		t.DebugLog("failed to verify callback token: %v", err)
		return t.failRedirect(w, r, ErrInvalidCallbackToken)
	}

	if token.StateID != claims.Id {
		t.DebugLog("failed to verify state ID: token(%v) != state(%v)", token.StateID, claims.Id)
		return t.failRedirect(w, r, ErrInvalidCallbackToken)
	}

	expires := token.Expiry.Unix()
	if expires < 60*60*24*366 {
		expires = 0
	}

	wToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &wrappedToken{
		Token: tokenStr,
		StandardClaims: jwt.StandardClaims{
			Audience:  TmpAuthHost + ":server:user_cookie:" + t.Config.ClientID,
			Issuer:    TmpAuthHost + ":server:" + t.Config.ClientID,
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: expires,
		},
	}).SignedString(t.Config.Secret)
	if err != nil {
		t.DebugLog("failed to sign wrapped token: %v", err)
		return http.StatusInternalServerError, fmt.Errorf("tmpauth: failed to sign wrapped token")
	}

	// token validated, can cache now
	tokenID := sha256.Sum256([]byte(wToken))
	t.tokenCacheMutex.Lock()
	t.TokenCache[tokenID] = token
	t.tokenCacheMutex.Unlock()

	t.DebugLog("auth callback successful, setting cookie")

	http.SetCookie(w, &http.Cookie{
		Name:     t.CookieName(),
		Value:    wToken,
		Expires:  token.Expiry,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	if redirectURI == "" {
		if t.Config.Redirect != "" {
			http.Redirect(w, r, t.Config.Redirect, http.StatusSeeOther)
			return 0, nil
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`you have been successfully authenticated, however we could ` +
			`not tell what the original page was that you were trying to visit.` + "\n" +
			`please try re-visiting the page you were trying to visit again`))
		return 0, nil
	}

	if t.Config.Host.Path != "" {
		redirectURI = path.Join(t.Config.Host.Path, redirectURI)
	}

	if t.Config.Redirect != "" {
		params := make(url.Values)
		params.Set("redirect_uri", redirectURI)

		http.Redirect(w, r, t.Config.Redirect+"?"+params.Encode(), http.StatusSeeOther)
		return 0, nil
	}

	http.Redirect(w, r, redirectURI, http.StatusSeeOther)
	return 0, nil
}

func (t *Tmpauth) consumeStateID(r *http.Request, w http.ResponseWriter, stateID string) (string, error) {
	t.DebugLog("consuming state ID: %v", stateID)

	defer func() {
		for _, cookie := range r.Cookies() {
			if !strings.HasPrefix(cookie.Name, "__Host-tmpauth-stateid_") {
				continue
			}

			cookie.Value = ""
			cookie.Expires = time.Time{}
			cookie.MaxAge = -1
			cookie.HttpOnly = true
			cookie.Secure = true
			cookie.Path = "/"
			cookie.SameSite = http.SameSiteLaxMode

			http.SetCookie(w, cookie)
		}
	}()

	stateCookie, err := r.Cookie(t.StateIDCookieName(stateID))
	if err != nil {
		return "", fmt.Errorf("tmpauth: state ID cookie not present")
	}

	value, err := url.QueryUnescape(stateCookie.Value)
	if err != nil {
		return "", fmt.Errorf("tmpauth: state ID cookie invalid")
	}

	var redirectURI string
	redirectURIRaw, found := t.stateIDCache.Get(stateID)
	if found {
		t.stateIDCache.Delete(stateID)
		redirectURI = redirectURIRaw.(string)
	}

	if value == "ok" {
		if !found {
			return "", nil
		}

		return redirectURI, nil
	} else if value[0] == '/' {
		if !found || redirectURI == value {
			return value, nil
		}

		return "", fmt.Errorf("tmpauth: state ID cookie mis-match")
	}

	return "", fmt.Errorf("tmpauth: state ID cookie invalid")
}
