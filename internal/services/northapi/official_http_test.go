package northapi

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestOfficialHTTPClientRequestsIdentityEncodingForPostDetails(t *testing.T) {
	t.Parallel()

	var encodings []string
	client := NewOfficialHTTPClient()
	client.Transport = officialTransport{base: roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		encodings = append(encodings, request.Header.Get("Accept-Encoding"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("{}")),
			Request:    request,
		}, nil
	})}

	for _, rawURL := range []string{
		"https://api.north.rip/2/tweets/123",
		"https://api.north.rip/2/tweets?ids=123",
		"https://api.north.rip/2/tweets/123/conversation",
	} {
		request, err := http.NewRequest(http.MethodGet, rawURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		response, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		if err := response.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}

	if len(encodings) != 3 || encodings[0] != "identity" || encodings[1] != "" || encodings[2] != "" {
		t.Fatalf("Accept-Encoding headers = %#v", encodings)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
