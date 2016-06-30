package db

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

// SessionDatabase allows handling sessions.
type SessionDatabase interface {
	SaveNewSession(string, time.Duration) (*Session, error)
	FindSession(string) (*Session, error)
	Invalidate(*Session) error
	ExpireSessions(time.Time) (int, error)
}

// Session contains the authentication token of a logged-in user.
type Session struct {
	Token      string
	User       string
	ValidUntil time.Time
}

func (s Session) String() string {
	return fmt.Sprintf("<Session %v, user %v (valid %v)>",
		s.Token[:8], s.User, s.ValidUntil.Sub(time.Now()))
}

const tokenLength = 32

// newSession generates a new session for a user.
func newSession(user string, valid time.Duration) (*Session, error) {
	buf := make([]byte, tokenLength)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return nil, err
	}

	s := &Session{
		Token:      hex.EncodeToString(buf),
		User:       user,
		ValidUntil: time.Now().Add(valid),
	}

	return s, nil
}

// SaveNewSession generates a new session for the user and saves it to the db.
func (db *Database) SaveNewSession(user string, valid time.Duration) (*Session, error) {
	s, err := newSession(user, valid)
	if err != nil {
		return nil, err
	}

	err = db.dbmap.Insert(s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// FindSession searches the session with the given token in the database.
func (db *Database) FindSession(token string) (*Session, error) {
	var s Session
	err := db.dbmap.SelectOne(&s, "SELECT * FROM sessions WHERE token = $1", token)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// ExpireSessions removes expired sessions from the db.
func (db *Database) ExpireSessions(until time.Time) (sessionsRemoved int, err error) {
	res := db.dbmap.Dbx.MustExec("DELETE FROM sessions WHERE valid_until < $1", until)
	n, err := res.RowsAffected()
	return int(n), err
}

// Invalidate removes the session from the database.
func (db *Database) Invalidate(s *Session) error {
	_, err := db.dbmap.Delete(s)
	if err != nil {
		return err
	}

	return nil
}
