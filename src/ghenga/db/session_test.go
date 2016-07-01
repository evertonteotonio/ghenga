package db

import (
	"testing"
	"time"
)

func testSession(t *testing.T, db DB) {
	for i := 0; i < 20; i++ {
		s, err := db.SaveNewSession("user", 300)
		if err != nil {
			t.Fatalf("unable to generate new token: %v", err)
		}

		token := s.Token

		if token == "" || len(token) != 2*tokenLength || token == "0000000000000000000000000000000000000000000000000000000000000000" {
			t.Fatalf("invalid token %q", token)
		}

		err = db.Invalidate(s)
		if err != nil {
			t.Fatalf("invalidate() %v", err)
		}
	}
}

func TestDBSession(t *testing.T) {
	testSession(t, testDB)
}

func TestMockDBSession(t *testing.T) {
	db := NewMockDB(20, 5)
	testSession(t, db)
}

func testSessionSave(t *testing.T, db DB) {
	var tokens []string

	for i := 0; i < 10; i++ {
		session, err := db.SaveNewSession("user", time.Duration(i)*time.Second)
		if err != nil {
			t.Fatalf("SaveNewSession() error %v", err)
		}

		s, err := db.FindSession(session.Token)
		if err != nil {
			t.Fatalf("unable to find newly generated token in the session database: %v", err)
		}

		if s.Token != session.Token {
			t.Fatalf("FindSession returned a different token")
		}

		tokens = append(tokens, s.Token)
	}

	n, err := db.ExpireSessions(time.Now())
	if err != nil {
		t.Fatalf("error expire sessions: %v", err)
	}

	if n != 1 {
		t.Errorf("expected 2 expired sessions, got %v", n)
	}

	if _, err = db.FindSession(tokens[0]); err == nil {
		t.Fatalf("expired session token %v still found in database", tokens[0])
	}
}

func TestDBSessionSave(t *testing.T) {
	testSessionSave(t, testDB)
}

func TestMockDBSessionSave(t *testing.T) {
	testSessionSave(t, NewMockDB(20, 5))
}
