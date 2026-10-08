package northapi

import (
	"net/http"
	"strings"
)

func NewOfficialHTTPClient() *http.Client {
	return &http.Client{Transport: officialTransport{base: http.DefaultTransport}}
}

type officialTransport struct {
	base http.RoundTripper
}

func (t officialTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	// North returns a weak ETag for compressed post responses, but If-Match
	// requires a strong validator.
	if postDetailRequest(request) {
		request = request.Clone(request.Context())
		request.Header = request.Header.Clone()
		if request.Header == nil {
			request.Header = make(http.Header)
		}
		request.Header.Set("Accept-Encoding", "identity")
	}

	return t.base.RoundTrip(request)
}

func postDetailRequest(request *http.Request) bool {
	if request.Method != http.MethodGet {
		return false
	}
	parts := strings.Split(strings.Trim(request.URL.EscapedPath(), "/"), "/")

	return len(parts) == 3 && parts[0] == "2" && parts[1] == "tweets" && parts[2] != ""
}
