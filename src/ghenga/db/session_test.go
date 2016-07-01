package db

import (
	"testing"
	"time"
)

func TestSession(t *testing.T) {
	for i := 0; i < 20; i++ {
		s, err := testDB.SaveNewSession("user", 300)
		if err != nil {
			t.Fatalf("unable to generate new token: %v", err)
		}

		token := s.Token

		if token == "" || len(token) != 2*tokenLength || token == "0000000000000000000000000000000000000000000000000000000000000000" {
			t.Fatalf("invalid token %q", token)
		}

		err = testDB.Invalidate(s)
		if err != nil {
			t.Fatalf("invalidate() %v", err)
		}
	}
}

func TestSessionSave(t *testing.T) {
	var tokens []string

	for i := 0; i < 10; i++ {
		session, err := testDB.SaveNewSession("user", time.Duration(i)*time.Second)
		if err != nil {
			t.Fatalf("SaveNewSession() error %v", err)
		}

		s, err := testDB.FindSession(session.Token)
		if err != nil {
			t.Fatalf("unable to find newly generated token in the session database: %v", err)
		}

		if s.Token != session.Token {
			t.Fatalf("FindSession returned a different token")
		}

		tokens = append(tokens, s.Token)
	}

	n, err := testDB.ExpireSessions(time.Now())
	if err != nil {
		t.Fatalf("error expire sessions: %v", err)
	}

	if n != 1 {
		t.Errorf("expected 2 expired sessions, got %v", n)
	}

	if _, err = testDB.FindSession(tokens[0]); err == nil {
		t.Fatalf("expired session token %v still found in database", tokens[0])
	}
}
