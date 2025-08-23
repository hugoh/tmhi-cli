package pkg

import "net/http"

// mockRoundTripper implements http.RoundTripper for testing
type mockRoundTripper struct {
	resp *http.Response
	err  error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.resp, m.err
}

func newTestClient(resp *http.Response, err error) *http.Client {
	return &http.Client{
		Transport: &mockRoundTripper{resp: resp, err: err},
	}
}
