package cors

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func getReq(method string) (*httptest.ResponseRecorder, *http.Request) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(method, "http://foo.com", nil)
	r.Header.Set(requestMethodKey, "POST")
	return w, r
}

func TestDefault_NoOrigin(t *testing.T) {
	w, r := getReq("OPTIONS")
	Default().HandleRequest(w, r)
	if w.Header().Get(allowOriginKey) != "" {
		t.Fatal("Should not have origin header when no origin passed in")
	}
}

func TestDefault_Origin_Supplied(t *testing.T) {
	w, r := getReq("OPTIONS")
	r.Header.Set(originKey, "http://bar.com")
	Default().HandleRequest(w, r)
	if w.Header().Get(allowOriginKey) != "*" {
		t.Fatalf("Expect origin of \"*\". Got \"%s\".", w.Header().Get(allowOriginKey))
	}
	if w.Header().Get(varyKey) != "Origin" {
		t.Fatal("Must include Vary:Origin if allow origin header set to specific domain.")
	}
}

func TestDefault_Dissallowed_Origin(t *testing.T) {
	w, r := getReq("OPTIONS")
	r.Header.Set(originKey, "http://bar.com")
	c := Default()
	c.AllowedOrigins = []string{"http://blog.bar.com"}
	c.HandleRequest(w, r)
	if w.Header().Get(allowOriginKey) != "" {
		t.Fatal("Should not have origin header when no origin not explicitely allowed")
	}
}

func TestDefault_Allowed_Origin_By_Regexp(t *testing.T) {
	c := Default()
	c.AllowedOrigins = []string{}
	c.OriginRegexps = []*regexp.Regexp{
		regexp.MustCompile(".+\\.example\\.com$"),
		regexp.MustCompile("^http\\:\\/\\/(.+)\\.other\\.org$"),
	}

	allowedOrigins := []string{"http://foo.example.com", "http://bar.other.org"}

	for _, origin := range allowedOrigins {
		w, r := getReq("OPTIONS")
		r.Header.Set(originKey, origin)
		c.HandleRequest(w, r)
		if w.Header().Get(allowOriginKey) != origin {
			t.Fatal("Should have origin header with the request origin when the regexp matches")
		}
	}

}

func TestDefault_Methods(t *testing.T) {
	w, r := getReq("OPTIONS")
	r.Header.Set(originKey, "http://bar.com")
	c := Default()
	c.HandleRequest(w, r)
	if w.Header().Get(allowMethodsKey) != "POST, GET, OPTIONS, PUT, DELETE" {
		t.Fatal("Allow methods should be set")
	}
}

func TestDefault_AllowAllHeaders(t *testing.T) {
	w, r := getReq("OPTIONS")
	c := Default()
	c.AllowedHeaders = "*"
	reqHeaders := "Bar, Foo, X-Yz"
	r.Header.Set(originKey, "http://bar.com")
	r.Header.Set(requestHeadersKey, reqHeaders)
	c.HandleRequest(w, r)
	if w.Header().Get(allowHeadersKey) != reqHeaders {
		t.Fatal("If AllowedHeaders is *, it should copy the value of requestHeadersKey to allowHeadersKey")
	}
}
