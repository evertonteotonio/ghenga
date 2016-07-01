package db

import (
	"errors"
	"strings"
	"time"
)

// MockDB implements the DB interface but only stores data in memory.
type MockDB struct {
	users    []User
	people   []Person
	sessions []Session
}

// ensure that *MockDB implements DB
var _ DB = &MockDB{}

// Close does nothing.
func (db *MockDB) Close() error {
	return nil
}

// InsertUser adds a new user to the db.
func (db *MockDB) InsertUser(u *User) error {
	db.users = append(db.users, *u)
	return nil
}

// ListUsers returns a list of all users.
func (db *MockDB) ListUsers() ([]*User, error) {
	list := make([]*User, len(db.users))
	for _, u := range db.users {
		list = append(list, &u)
	}
	return list, nil
}

// UpdateUser modifies an existing user record in the db.
func (db *MockDB) UpdateUser(u *User) error {
	for i, user := range db.users {
		if user.ID == u.ID {
			db.users[i] = *u
			return nil
		}
	}

	return errors.New("user not found")
}

// DeleteUser removes a record from the db.
func (db *MockDB) DeleteUser(id int64) error {
	for i, user := range db.users {
		if user.ID == id {
			db.users = append(db.users[:i], db.users[i+1:]...)
			return nil
		}
	}

	return errors.New("user not found")
}

// FindUser returns the user with the given id.
func (db *MockDB) FindUser(id int64) (*User, error) {
	for _, user := range db.users {
		if user.ID == id {
			return &user, nil
		}
	}

	return nil, errors.New("user not found")
}

// FindUserName searches for a user with the given login name.
func (db *MockDB) FindUserName(name string) (*User, error) {
	for _, user := range db.users {
		if user.Login == name {
			return &user, nil
		}
	}

	return nil, errors.New("user not found")
}

// InsertPerson adds a new person to the db.
func (db *MockDB) InsertPerson(p *Person) error {
	db.people = append(db.people, *p)
	return nil
}

// ListPeople returns a list of all people in the database.
func (db *MockDB) ListPeople() ([]*Person, error) {
	list := make([]*Person, len(db.people))
	for _, u := range db.people {
		list = append(list, &u)
	}
	return list, nil
}

// UpdatePerson modifies a person in the db.
func (db *MockDB) UpdatePerson(p *Person) error {
	for i, person := range db.people {
		if person.ID == p.ID {
			db.people[i] = *p
			return nil
		}
	}

	return errors.New("person not found")
}

// DeletePerson removes a person from the db.
func (db *MockDB) DeletePerson(id int64) error {
	for i, person := range db.people {
		if person.ID == id {
			db.people = append(db.people[:i], db.people[i+1:]...)
			return nil
		}
	}

	return errors.New("person not found")
}

// FindPerson searches for a person.
func (db *MockDB) FindPerson(id int64) (*Person, error) {
	for _, person := range db.people {
		if person.ID == id {
			return &person, nil
		}
	}

	return nil, errors.New("person not found")
}

// FuzzyFindPersons returns all people matching query.
func (db *MockDB) FuzzyFindPersons(query string) ([]*Person, error) {
	query = strings.ToLower(query)
	var list []*Person
	for _, person := range db.people {
		if strings.Contains(strings.ToLower(person.Name), query) {
			p := person
			list = append(list, &p)
		}
	}

	return list, nil
}

// SaveNewSession creates a new session and saves it in the db.
func (db *MockDB) SaveNewSession(login string, until time.Duration) (*Session, error) {
	s, err := newSession(login, until)
	if err != nil {
		return nil, err
	}

	db.sessions = append(db.sessions, *s)
	return s, nil
}

// FindSession returns the session for the given token.
func (db *MockDB) FindSession(token string) (*Session, error) {
	for _, s := range db.sessions {
		if s.Token == token {
			return &s, nil
		}
	}

	return nil, errors.New("session not found")
}

// Invalidate removes the session from the database.
func (db *MockDB) Invalidate(s *Session) error {
	for i, session := range db.sessions {
		if session.Token == s.Token {
			db.sessions = append(db.sessions[:i], db.sessions[i+1:]...)
			return nil
		}
	}

	return errors.New("session not found")
}

// ExpireSessions removes all sessions which have timed out.
func (db *MockDB) ExpireSessions(now time.Time) (n int, err error) {
	var out []Session

	for _, session := range db.sessions {
		if session.ValidUntil.Before(now) {
			n++
			continue
		}

		out = append(out, session)
	}

	db.sessions = out

	return n, err
}
