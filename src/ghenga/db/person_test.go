package db

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "update golden files")

func parseTime(s string) time.Time {
	t, err := time.Parse(timeLayout, s)
	if err != nil {
		panic(err)
	}

	return t
}

var testPersons = []struct {
	name string
	p    Person
}{
	{
		name: "testperson1",
		p: Person{
			Name:         "Tamara Skibicki",
			EmailAddress: "pit@ackermannsehls.org",
			PhoneNumbers: []PhoneNumber{
				{Type: "work", Number: "(03867) 3074101"},
				{Type: "mobile", Number: "+49-077-1634655"},
				{Type: "other", Number: "2134"},
			},
			Comment:   "fake profile",
			ChangedAt: parseTime("2016-04-24T10:30:07+00:00"),
			CreatedAt: parseTime("2016-04-24T10:30:07+00:00"),
			Version:   23,
		},
	},
	{
		name: "testperson2",
		p: Person{
			Name:         "Mario Drees",
			EmailAddress: "bela_freigang@herweg.com",
			ChangedAt:    parseTime("2016-04-24T10:30:07+00:00"),
			CreatedAt:    parseTime("2016-04-24T10:30:07+00:00"),
			Version:      1,
		},
	},
	{
		name: "testperson3",
		p: Person{
			Name:         "Mario Drees",
			EmailAddress: "bela_freigang@herweg.com",
			PhoneNumbers: []PhoneNumber{
				{Type: "wörk", Number: "1234123 3074101"},
			},

			Street:     "Lower High St. 23",
			Country:    "GB",
			City:       "London",
			State:      "California",
			PostalCode: "1234",

			ChangedAt: parseTime("2016-04-24T10:30:07+00:00"),
			CreatedAt: parseTime("2016-04-24T10:30:07+00:00"),
			Version:   5,
		},
	},
}

func testPersonInsertSelect(t *testing.T, db DB) {
	var ids []int64
	for _, test := range testPersons {
		err := db.InsertPerson(&test.p)
		if err != nil {
			t.Errorf("saving %v failed: %v", test.name, err)
			continue
		}

		ids = append(ids, test.p.ID)
	}

	for i, test := range testPersons {
		p, err := db.FindPerson(ids[i])
		if err != nil {
			t.Errorf("loading %v failed: %v", test.p.ID, err)
			continue
		}

		if p.ID == 0 {
			t.Errorf("ID of new person is zero")
		}

		if p.Version != test.p.Version+1 {
			t.Errorf("%v: wrong version loaded from db, want %v, got %v",
				test.name, test.p.Version+1, p.Version)
		}

		p.ID = test.p.ID
		p.Version = test.p.Version

		buf1 := marshal(t, test.p)
		buf2 := marshal(t, p)

		if !bytes.Equal(buf1, buf2) {
			t.Errorf("loading %v returned different data:\n  want: %s\n   got: %s",
				test.name, buf1, buf2)
			continue
		}
	}
}

func TestDBPersonInsertSelect(t *testing.T) {
	testPersonInsertSelect(t, testDB)
}

func TestMockDBPersonInsertSelect(t *testing.T) {
	db := NewMockDB(20, 5)
	testPersonInsertSelect(t, db)
}

func testPersonVersion(t *testing.T, db DB) {
	p, err := db.FindPerson(14)
	if err != nil {
		t.Fatal(err)
	}

	p.Version = 25
	err = db.UpdatePerson(p)
	if err == nil {
		t.Fatalf("expected error due to outdated version not found")
	}
}

func TestDBPersonVersion(t *testing.T) {
	testPersonVersion(t, testDB)
}

func TestMockDBPersonVersion(t *testing.T) {
	db := NewMockDB(20, 5)
	testPersonVersion(t, db)
}

func marshal(t *testing.T, item interface{}) []byte {
	buf, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		t.Fatalf("json.Marshal(): %v", err)
	}

	return buf
}

func unmarshal(t *testing.T, buf []byte, item interface{}) {
	err := json.Unmarshal(buf, item)
	if err != nil {
		t.Fatalf("json.Unmarsha(%s): %v", buf, err)
	}
}

func TestPersonMarshal(t *testing.T) {
	for i, test := range testPersons {
		buf := marshal(t, test.p)

		golden := filepath.Join("testdata", "TestPersonMarshal_"+test.name+".golden")
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

func TestPersonUnmarshal(t *testing.T) {
	for i, test := range testPersons {
		golden := filepath.Join("testdata", "TestPersonMarshal_"+test.name+".golden")
		buf, err := ioutil.ReadFile(golden)
		if err != nil {
			t.Errorf("test %d: unable to read golden file %v", i, golden)
			continue
		}

		var p Person
		unmarshal(t, buf, &p)

		buf2 := marshal(t, p)

		if !bytes.Equal(buf, buf2) {
			t.Errorf("test %d (%v) wrong JSON returned:\nwant:\n%s\ngot:\n%s", i, test.name, buf, buf2)
		}
	}
}

var testPersonValidate = []struct {
	name  string
	valid bool
	p     Person
}{
	{
		name:  "invalid1",
		valid: false,
		p: Person{
			Name: "",
		},
	},
}

func TestPersonValidate(t *testing.T) {
	for i, test := range testPersons {
		if err := test.p.Validate(); err != nil {
			t.Errorf("test %v (%v) failed: testPerson is invalid: %v", test.name, i, err)
		}
	}

	for i, test := range testPersonValidate {
		err := test.p.Validate()
		if test.valid && err != nil {
			t.Errorf("test %v (%v) failed: testPerson should be valid but is invalid: %v", test.name, i, err)
		}

		if !test.valid && err == nil {
			t.Errorf("test %v (%v) failed: testPerson should be invalid but is valid", test.name, i)
		}
	}
}

func testPersonUpdate(t *testing.T, db DB) {
	p, err := db.FindPerson(12)
	if err != nil {
		t.Fatalf("unable to load person 12: %v", err)
	}

	p.Name = "foo bar"
	if err = db.UpdatePerson(p); err != nil {
		t.Fatalf("unable to update person: %v", err)
	}

	p.Title = "CTO"
	p.Version = 1
	if err = db.UpdatePerson(p); err == nil {
		t.Fatalf("update did not fail despite wrong version field")
	}
}

func TestDBPersonUpdate(t *testing.T) {
	testPersonUpdate(t, testDB)
}

func TestMockDBPersonUpdate(t *testing.T) {
	db := NewMockDB(20, 5)
	testPersonUpdate(t, db)
}

func findPerson(t *testing.T, db DB, id int64) *Person {
	p, err := db.FindPerson(id)
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func updatePerson(t *testing.T, db DB, p *Person) {
	if err := db.UpdatePerson(p); err != nil {
		t.Fatal(err)
	}
}

func testPersonUpdatePhoneNumbers(t *testing.T, db DB) {
	p := findPerson(t, db, 14)
	p.PhoneNumbers = append(p.PhoneNumbers, PhoneNumber{Type: "test", Number: "12345"})

	updatePerson(t, db, p)

	p2 := findPerson(t, db, p.ID)
	if !p.PhoneNumbers.Equals(p2.PhoneNumbers) {
		t.Fatalf("changing phone numbers did not work, want:\n%v\n  got:\n%v", p.PhoneNumbers, p2.PhoneNumbers)
	}
}

func TestDBPersonUpdatePhoneNumbers(t *testing.T) {
	testPersonUpdatePhoneNumbers(t, testDB)
}

func TestMockDBPersonUpdatePhoneNumbers(t *testing.T) {
	db := NewMockDB(20, 5)
	testPersonUpdatePhoneNumbers(t, db)
}

func testPersonDeletePhoneNumber(t *testing.T, db DB) {
	p := findPerson(t, db, 14)
	if len(p.PhoneNumbers) > 0 {
		p.PhoneNumbers = p.PhoneNumbers[1:]
	}

	updatePerson(t, db, p)

	p2 := findPerson(t, db, p.ID)
	if !p.PhoneNumbers.Equals(p2.PhoneNumbers) {
		t.Fatalf("changing phone numbers did not work, want:\n%v\n  got:\n%v", p.PhoneNumbers, p2.PhoneNumbers)
	}
}

func TestDBPersonDeletePhoneNumber(t *testing.T) {
	testPersonDeletePhoneNumber(t, testDB)
}

func TestMockDBPersonDeletePhoneNumber(t *testing.T) {
	db := NewMockDB(20, 5)
	testPersonDeletePhoneNumber(t, db)
}

func testPersonDeleteAllPhoneNumbers(t *testing.T, db DB) {
	p := findPerson(t, db, 14)
	p.PhoneNumbers = PhoneNumbers{}

	updatePerson(t, db, p)

	p2 := findPerson(t, db, p.ID)
	if len(p2.PhoneNumbers) > 0 {
		t.Fatalf("removing phone numbers did not work, got:\n%v", p2.PhoneNumbers)
	}
}

func TestDBPersonDeleteAllPhoneNumbers(t *testing.T) {
	testPersonDeleteAllPhoneNumbers(t, testDB)
}

func TestMockDBPersonDeleteAllPhoneNumbers(t *testing.T) {
	db := NewMockDB(20, 5)
	testPersonDeleteAllPhoneNumbers(t, db)
}

func testPersonReplacePhoneNumbers(t *testing.T, db DB) {
	p := findPerson(t, db, 14)
	p.PhoneNumbers = PhoneNumbers{PhoneNumber{Type: "test", Number: "12345"}}

	updatePerson(t, db, p)

	p2 := findPerson(t, db, p.ID)
	if !p.PhoneNumbers.Equals(p2.PhoneNumbers) {
		t.Fatalf("changing phone numbers did not work, want:\n%v\n  got:\n%v", p.PhoneNumbers, p2.PhoneNumbers)
	}
}

func TestDBPersonReplacePhoneNumbers(t *testing.T) {
	testPersonReplacePhoneNumbers(t, testDB)
}
