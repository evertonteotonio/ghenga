package db

import "testing"

func TestNewFakePerson(t *testing.T) {
	p, err := NewFakePerson("de")
	if err != nil {
		t.Fatalf("NewFakePerson(): %v", err)
	}

	if err = p.Validate(); err != nil {
		t.Fatalf("NewFakePerson() not valid: %v", err)
	}
}

func TestNewFakeUser(t *testing.T) {
	u, err := NewFakeUser("de")
	if err != nil {
		t.Fatalf("new user creation failed: %v", err)
	}

	if !u.CheckPassword("geheim") {
		t.Fatalf("fake user does not have password 'geheim'")
	}
}
