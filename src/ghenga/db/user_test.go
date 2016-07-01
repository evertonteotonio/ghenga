package db

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func testUserAdd(t *testing.T, db DB) {
	u, err := NewUser("foo", "bar")
	if err != nil {
		t.Fatal(err)
	}

	err = db.InsertUser(u)
	if err != nil {
		t.Fatal(err)
	}

	u2, err := db.FindUserName("foo")
	if err != nil {
		t.Fatal(err)
	}

	if !u2.CheckPassword("bar") {
		t.Fatalf("password for test user does not match hash")
	}

	if u2.CheckPassword("xxx") {
		t.Fatalf("wrong password for test user was accepted")
	}
}

func TestDBUserAdd(t *testing.T) {
	testUserAdd(t, testDB)
}

func TestMockDBUserAdd(t *testing.T) {
	testUserAdd(t, NewMockDB(20, 5))
}

var testUsers = []struct {
	name string
	u    User
}{
	{
		name: "testuser1",
		u: User{
			Login:        "foobar",
			Admin:        false,
			PasswordHash: "foobarbaz",
			ChangedAt:    parseTime("2016-04-24T10:30:07+02:00"),
			CreatedAt:    parseTime("2016-04-24T10:30:07+02:00"),
			Version:      23,
		},
	},
	{
		name: "testuser2",
		u: User{
			Login:        "x",
			Admin:        true,
			PasswordHash: "xxy",
			ChangedAt:    parseTime("2016-03-24T10:30:07+02:00"),
			CreatedAt:    parseTime("2016-01-24T10:30:07+02:00"),
			Version:      5,
		},
	},
}

func testUserVersion(t *testing.T, db DB) {
	u, err := db.FindUserName("admin")
	if err != nil {
		t.Fatal(err)
	}

	u.Version = 25
	err = db.UpdateUser(u)
	if err == nil {
		t.Fatalf("expected error due to outdated version not found")
	}
}

func TestDBUserVersion(t *testing.T) {
	testUserVersion(t, testDB)
}

func TestMockDBUserVersion(t *testing.T) {
	testUserVersion(t, NewMockDB(20, 5))
}

func TestUserMarshal(t *testing.T) {
	for i, test := range testUsers {
		buf := marshal(t, test.u)

		golden := filepath.Join("testdata", "TestUserMarshal_"+test.name+".golden")
		if *update {
			err := ioutil.WriteFile(golden, buf, 0644)
			if err != nil {
				t.Fatalf("test %d: update golden file %v failed: %v", i, golden, err)
			}
		}

		expected, err := ioutil.ReadFile(golden)
		if err != nil {
			t.Errorf("test %d: unable to read golden file %v", i, golden)
			continue
		}
		if !bytes.Equal(buf, expected) {
			t.Errorf("test %d (%v) wrong JSON returned:\nwant:\n%s\ngot:\n%s", i, test.name, expected, buf)
		}
	}
}

func TestUserUnmarshal(t *testing.T) {
	for i, test := range testUsers {
		golden := filepath.Join("testdata", "TestUserMarshal_"+test.name+".golden")
		buf, err := ioutil.ReadFile(golden)
		if err != nil {
			t.Errorf("test %d: unable to read golden file %v", i, golden)
			continue
		}

		var u User
		unmarshal(t, buf, &u)

		buf2 := marshal(t, u)

		if !bytes.Equal(buf, buf2) {
			t.Errorf("test %d (%v) wrong JSON returned:\nwant:\n%s\ngot:\n%s", i, test.name, buf, buf2)
		}
	}
}

var testUserValidate = []struct {
	name  string
	valid bool
	u     User
}{
	{
		name:  "invalid1",
		valid: false,
		u: User{
			Login: "",
		},
	},
}

func TestUserValidate(t *testing.T) {
	for i, test := range testUsers {
		if err := test.u.Validate(); err != nil {
			t.Errorf("test %v (%v) failed: test User is invalid: %v", test.name, i, err)
		}
	}

	for i, test := range testUserValidate {
		err := test.u.Validate()
		if test.valid && err != nil {
			t.Errorf("test %v (%v) failed: test User should be valid but is invalid: %v", test.name, i, err)
		}

		if !test.valid && err == nil {
			t.Errorf("test %v (%v) failed: test User should be invalid but is valid", test.name, i)
		}
	}
}

func testUserUpdate(t *testing.T, db DB) {
	u, err := db.FindUserName("user")
	if err != nil {
		t.Fatalf("unable to load user %q: %v", "user", err)
	}

	u.Login = "foo bar"
	if err = db.UpdateUser(u); err != nil {
		t.Fatalf("unable to update user: %v", err)
	}

	v := u.Version
	u.Admin = !u.Admin
	u.Version = 1
	if err = db.UpdateUser(u); err == nil {
		t.Fatalf("update did not fail despite wrong version field")
	}

	u.Admin = !u.Admin
	u.Login = "user"
	u.Version = v

	if err = db.UpdateUser(u); err != nil {
		t.Fatalf("unable to update user: %v", err)
	}
}

func TestDBUserUpdate(t *testing.T) {
	testUserUpdate(t, testDB)
}

func TestMockDBUserUpdate(t *testing.T) {
	testUserUpdate(t, NewMockDB(20, 5))
}

func testUserUpdatePassword(t *testing.T, db DB) {
	u, err := db.FindUserName("user")
	if err != nil {
		t.Fatalf("unable to load user %q: %v", "user", err)
	}

	if !u.CheckPassword("geheim") {
		t.Fatalf("password for account `user` is not `geheim`")
	}

	u.Password = "foobar2"
	if err = db.UpdateUser(u); err != nil {
		t.Errorf("unable to update user: %v", err)
	}

	if u.CheckPassword("geheim") {
		t.Errorf("password for account `user` is still `geheim`")
	}

	if !u.CheckPassword("foobar2") {
		t.Errorf("changed password for account `user` is not `foobar2`")
	}

	u2, err := db.FindUserName("user")
	if err != nil {
		t.Fatalf("unable to load user %q: %v", "user", err)
	}

	if !u2.CheckPassword("foobar2") {
		t.Errorf("changed password for account `user` in the db is not `foobar2`")
	}
}

func TestDBUserUpdatePassword(t *testing.T) {
	testUserUpdatePassword(t, testDB)
}

func TestMockDBUserUpdatePassword(t *testing.T) {
	testUserUpdatePassword(t, NewMockDB(20, 5))
}
