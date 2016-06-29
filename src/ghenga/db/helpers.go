package db

import (
	"math/rand"
	"os"

	"github.com/fd0/probe"
	"github.com/jmoiron/modl"
	"github.com/manveru/faker"
)

// NewFakePerson returns a Person struct filled with fake data.
func NewFakePerson(lang string) (*Person, error) {
	f, err := faker.New(lang)
	if err != nil {
		return nil, probe.Trace(err, lang)
	}

	p := NewPerson(f.FirstName() + " " + f.LastName())
	if rand.Float32() <= 0.2 {
		p.Title = "CEO"
	}
	p.Department = "Testers"
	p.EmailAddress = f.Email()

	for _, d := range []struct {
		probability float32
		tpe         string
		gen         func() string
	}{
		{0.5, "mobile", f.CellPhoneNumber},
		{0.9, "work", f.PhoneNumber},
		{0.1, "fax", f.PhoneNumber},
		{0.3, "other", f.PhoneNumber},
	} {
		if rand.Float32() < d.probability {
			p.PhoneNumbers = append(p.PhoneNumbers, PhoneNumber{
				Type:   d.tpe,
				Number: d.gen(),
			})
		}
	}

	p.Comment = "fake profile"
	p.ID = rand.Int63n(20)

	if rand.Float32() <= 0.6 {
		p.Street = f.StreetAddress()
		p.PostalCode = f.PostCode()
		if rand.Float32() < 0.4 {
			p.State = "CA"
		}
		p.City = f.City()
		p.Country = f.Country()
	}

	return p, nil
}

// NewFakeUser returns a User struct filled with fake data. The password is
// always set to "geheim".
func NewFakeUser(lang string) (*User, error) {
	f, err := faker.New(lang)
	if err != nil {
		return nil, probe.Trace(err, lang)
	}

	return NewUser(f.UserName(), "geheim")
}

// InsertFakeData will populate the db with fake (but realistic) data. Among
// others, users named "admin" and "user" with the password "geheim" are
// created.
func InsertFakeData(dbm *modl.DbMap, people, user int) error {
	for i := 0; i < people; i++ {
		p, err := NewFakePerson("de")
		if err != nil {
			return probe.Trace(err, people)
		}

		err = dbm.Insert(p)
		if err != nil {
			return probe.Trace(err, user)
		}
	}

	for _, s := range []struct {
		name  string
		admin bool
	}{{"admin", true}, {"user", false}} {
		u, err := NewUser(s.name, "geheim")
		if err != nil {
			return probe.Trace(err, s.name)
		}

		u.Admin = s.admin
		if err := dbm.Insert(u); err != nil {
			return probe.Trace(err, u)
		}
	}

	for i := 0; i < user; i++ {
		u, err := NewFakeUser("de")
		if err != nil {
			return probe.Trace(err)
		}

		err = dbm.Insert(u)
		if err != nil {
			// ignore errors for fake data
			continue
		}
	}

	return nil
}

const defaultDataSource = "host=/var/run/postgresql"

// TestDataSource returns a datasource used for testing. This defaults to the
// postgresql running on the default unix domain socket. If the environment
// variable GHENGA_TEST_DB is set, this value is returned instead.
func TestDataSource() string {
	dataSource := os.Getenv("GHENGA_TEST_DB")
	if dataSource == "" {
		dataSource = defaultDataSource
	}

	return dataSource
}

// testCleanupDB removes everything in the database. On error it panics.
func testCleanupDB(db *modl.DbMap) {
	var queries []string
	err := db.Select(&queries, `SELECT 'DROP TABLE "' || tablename || '" CASCADE;' FROM pg_tables WHERE schemaname='public'`)
	if err != nil {
		panic(err)
	}

	for _, query := range queries {
		_, err = db.Exec(query)
		if err != nil {
			panic(err)
		}
	}
}

// TestDB returns a database suitable for testing. The database is emptied
// and filled with fake data. It should be called in TestMain. On error, TestDB
// panics.
func TestDB(people, user int) (*modl.DbMap, func()) {
	dbmap, err := Init(TestDataSource())
	if err != nil {
		panic(err)
	}

	testCleanupDB(dbmap)

	err = InsertFakeData(dbmap, people, user)
	if err != nil {
		panic(err)
	}

	return dbmap, func() {
		testCleanupDB(dbmap)

		if err := dbmap.Db.Close(); err != nil {
			panic(err)
		}
	}
}
