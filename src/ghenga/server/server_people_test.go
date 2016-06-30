package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func marshal(t *testing.T, item interface{}) []byte {
	buf, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}

	return buf
}

func unmarshal(t testing.TB, data []byte, item interface{}) {
	err := json.Unmarshal(data, item)
	if err != nil {
		t.Fatal(err)
	}
}

func readBody(t *testing.T, res *http.Response) (int, []byte) {
	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("cannot read response body: %v", err)
	}

	err = res.Body.Close()
	if err != nil {
		t.Fatalf("close body: %v", err)
	}

	return res.StatusCode, responseBody
}

func request(t *testing.T, token, method, url string, body []byte) (int, []byte) {
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, rd)
	if err != nil {
		t.Fatalf("NewRequest() %v", err)
	}

	if token != "" {
		req.Header.Add(authHeaderName, token)
	}

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%v request to %v failed: %v", method, url, err)
	}

	t.Logf("%v %v -> %v (%v)", method, url, res.StatusCode, res.Status)

	return readBody(t, res)
}

func readFixture(t *testing.T, filename string) []byte {
	p, err := ioutil.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("unable to read test fixture: %v", err)
	}

	return p
}

type Person struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Version int    `json:"version"`
}

func verifyPerson(t *testing.T, name string, data []byte) Person {
	var person Person

	unmarshal(t, data, &person)

	if person.ID == 0 {
		t.Fatalf("person has ID 0")
	}

	if person.Name != name {
		t.Fatalf("name does not match, want %q, got %q", name, person.Name)
	}

	return person
}

func deletePerson(t *testing.T, token, url string, id int) {
	status, body := request(t, token, "DELETE", fmt.Sprintf("%s/api/person/%d", url, id), nil)
	if status != 200 {
		t.Fatalf("reading person again yielded unexpected status %d", status)
	}

	if strings.TrimSpace(string(body)) != "{}" {
		t.Fatalf("expected empty JSON body not found, got:\n%s", body)
	}
}

func TestPersonCRUD(t *testing.T) {
	srv, cleanup := TestServer(t)
	defer cleanup()

	p := readFixture(t, "sample_person.json")

	token := login(t, srv, "admin", "geheim")

	status, body := request(t, token, "POST", srv.URL+"/api/person", p)
	if status != 201 {
		t.Fatalf("invalid status code, want 201, got %v, body:\n  %s", status, string(p))
	}

	person := verifyPerson(t, "Nicolai Person", body)

	status, body = request(t, token, "GET", fmt.Sprintf("%s/api/person/%d", srv.URL, person.ID), nil)
	if status != 200 {
		t.Fatalf("reading person again yielded unexpected status %d: %s", status, body)
	}

	t.Logf("person: %v", person)

	person = verifyPerson(t, person.Name, body)
	person.Name = "Robert Niemand"

	t.Logf("person: %v", person)

	status, body = request(t, token, "PUT", fmt.Sprintf("%s/api/person/%d", srv.URL, person.ID), marshal(t, person))
	if status != 200 {
		t.Fatalf("updating person, invalid status %d", status)
	}

	verifyPerson(t, person.Name, body)

	status, body = request(t, token, "GET", fmt.Sprintf("%s/api/person/%d", srv.URL, person.ID), nil)
	if status != 200 {
		t.Fatalf("reading person again yielded unexpected status %d", status)
	}

	verifyPerson(t, person.Name, body)

	deletePerson(t, token, srv.URL, person.ID)
}

func TestPersonList(t *testing.T) {
	srv, cleanup := TestServer(t)
	defer cleanup()

	token := login(t, srv, "admin", "geheim")

	status, body := request(t, token, "GET", srv.URL+"/api/person", nil)
	if status != 200 {
		t.Fatalf("reading list of persons failed with invalid status: want 200, got %d", status)
	}

	var list []Person
	unmarshal(t, body, &list)
	if len(list) == 0 {
		t.Fatalf("got no persons from test server")
	}

	t.Logf("loaded %d person records", len(list))
}

var invalidPersonTests = []string{
	`{}`,
	`{"id": 23}`,
	`{"email_address": "foo@example.com"}`,
}

func TestInvalidPerson(t *testing.T) {
	srv, cleanup := TestServer(t)
	defer cleanup()

	token := login(t, srv, "admin", "geheim")

	for _, test := range invalidPersonTests {
		status, body := request(t, token, "POST", srv.URL+"/api/person", []byte(test))
		if status != 400 {
			t.Fatalf("status code for invalid person not found, want 400, got %v, body:\n  %s", status, body)
		}
	}
}
