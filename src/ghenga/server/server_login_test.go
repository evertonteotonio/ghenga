package server

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func loginRequest(t testing.TB, srv *TestSrv, username, password string) (status int, body []byte) {
	req, err := http.NewRequest("GET", srv.URL+"/api/login/token", nil)
	if err != nil {
		t.Fatalf("unable to create login request: %v", err)
	}

	if username != "" {
		req.SetBasicAuth(username, password)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("login attempt failed: %v", err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatalf("error closing body: %v", err)
		}
	}()

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}

	return res.StatusCode, buf
}

type loginResponse struct {
	Token    string `json:"token"`
	ValidFor uint   `json:"valid_for"`
}

func login(t testing.TB, srv *TestSrv, username, password string) (token string) {
	status, body := loginRequest(t, srv, username, password)

	if status != http.StatusOK {
		t.Fatalf("unexpected status %v, expected 200", status)
	}

	var response loginResponse
	unmarshal(t, body, &response)

	if response.Token == "" || response.ValidFor == 0 {
		t.Fatalf("invalid response from login endpoint:\n%s", body)
	}

	return response.Token
}

var invalidUsernamePasswords = []struct {
	u string
	p string
}{
	{"admin", "geheimX"},
	{"admin", ""},
	{"", "geheimX"},
	{"", ""},
}

func testLogin(t testing.TB) {
	srv, cleanup := TestServer(t)
	defer cleanup()

	token := login(t, srv, "admin", "geheim")
	if token == "" {
		t.Fatalf("invalid response for valid login request: token %v", token)
	}

	for _, test := range invalidUsernamePasswords {
		status, body := loginRequest(t, srv, test.u, test.p)
		if status != http.StatusUnauthorized {
			t.Errorf("invalid response for invalid login request (%v/%v): status %v, body:\n%s",
				test.u, test.p, status, body)
		}
	}
}

func TestLogin(t *testing.T) {
	testLogin(t)
}

func BenchmarkLogin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testLogin(b)
	}
}

func TestInfo(t *testing.T) {
	srv, cleanup := TestServer(t)
	defer cleanup()

	token := login(t, srv, "admin", "geheim")
	if token == "" {
		t.Fatalf("invalid response for valid login request: token %v", token)
	}

	req, err := http.NewRequest("GET", srv.URL+"/api/login/info", nil)
	if err != nil {
		t.Fatalf("NewRequest() %v", err)
	}

	req.Header.Add("X-Auth-Token", token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}

	err = res.Body.Close()
	if err != nil {
		t.Fatalf("Body.Close() %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("http response status %v unexpected, want 200", res.Status)
	}

	var response loginResponse
	unmarshal(t, buf, &response)

	if response.Token != token || response.ValidFor == 0 {
		t.Fatalf("invalid response for check request: %v", response)
	}
}
