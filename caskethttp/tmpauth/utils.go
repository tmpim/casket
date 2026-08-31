package tmpauth

import (
	"encoding/base64"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/tidwall/gjson"
)

func (t *Tmpauth) DebugLog(fmtString string, args ...interface{}) {
	if !t.Config.Debug {
		return
	}

	t.Logger.Output(2, fmt.Sprintf(fmtString, args...))
}

func getJSONPath(jsonData, path string) string {
	result := gjson.Get(jsonData, path)
	if !result.Exists() {
		return ""
	}

	return result.String()
}

func getJSONPathMany(jsonData, path string) []string {
	var results []string
	result := gjson.Get(jsonData, path)
	for _, val := range result.Array() {
		results = append(results, val.String())
	}

	return results
}

type HeaderOption struct {
	Format   string
	Optional bool
}

func (h *HeaderOption) Evaluate(jsonData string) (string, error) {
	result := getJSONPath(jsonData, h.Format)
	if result == "" && !h.Optional {
		return "", fmt.Errorf("tmpauth: requested header format yielded no results on claim")
	}

	return result, nil
}

func (t *Tmpauth) CookieName() string {
	return "__Host-tmpauth_" + t.Config.ClientID
}

func (t *Tmpauth) StateIDCookieName(id string) string {
	t.hmacMutex.Lock()
	t.HMAC.Reset()
	t.HMAC.Write([]byte(id))
	name := base64.RawURLEncoding.EncodeToString(t.HMAC.Sum(nil))
	t.hmacMutex.Unlock()

	return "__Host-tmpauth-stateid_" + name
}

func (t *Tmpauth) VerifyWithPublicKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
		return nil, fmt.Errorf("tmpauth: expected ECDSA signing method, got: %v", token.Header["alg"])
	}

	return t.Config.PublicKey, nil
}

func (t *Tmpauth) VerifyWithSecret(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("tmpauth: expected HMAC signing method, got: %v", token.Header["alg"])
	}

	return t.Config.Secret, nil
}
