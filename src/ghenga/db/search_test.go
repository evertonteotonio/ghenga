package db

import "testing"

var searchTestPersons = []Person{
	Person{
		Name:         "Tamara Skibicki",
		EmailAddress: "pit@ackermannsehls.org",
		PhoneNumbers: []PhoneNumber{
			{Type: "work", Number: "(03867) 3074101"},
			{Type: "mobile", Number: "+49-077-1634655"},
			{Type: "other", Number: "2134"},
		},
		Comment:   "fake profile",
		ChangedAt: parseTime("2016-04-24T10:30:07+02:00"),
		CreatedAt: parseTime("2016-04-24T10:30:07+02:00"),
		Version:   23,
	},
	Person{
		Name:         "Mario Drees",
		EmailAddress: "bela_freigang@herweg.com",
		ChangedAt:    parseTime("2016-04-24T10:30:07+00:00"),
		CreatedAt:    parseTime("2016-04-24T10:30:07+00:00"),
		Version:      1,
	},
}

//  fuzzyFindPersons makes sure that at least people are contained within the
//  result set.
func fuzzyFindPersons(t *testing.T, db DB, query string, in []Person, out []Person) {
	result, err := db.FuzzyFindPersons(query)
	if err != nil {
		t.Fatalf("FuzzyFindPersons(%q) returned error %v", query, err)
	}

	for _, p := range in {
		found := false

		for _, r := range result {
			if p.Name == r.Name {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("FuzzyFindPerson(%q) person %v not found in result set", query, p)
		}
	}

	for _, p := range out {
		found := false

		for _, r := range result {
			if p.Name == r.Name {
				found = true
				break
			}
		}

		if found {
			t.Errorf("FuzzyFindPerson(%q) person %v was found in result set", query, p)
		}
	}
}

var fuzzyFindTests = []struct {
	query string
	in    []Person
	out   []Person
}{
	{
		query: "tamara",
		in:    []Person{searchTestPersons[0]},
		out:   []Person{searchTestPersons[1]},
	},
	{
		query: "ama",
		in:    []Person{searchTestPersons[0]},
		out:   []Person{searchTestPersons[1]},
	},
	{
		query: "Mar",
		in:    []Person{searchTestPersons[1]},
	},
	{
		query: "a",
		in:    []Person{searchTestPersons[0], searchTestPersons[1]},
	},
	{
		query: "y",
		out:   []Person{searchTestPersons[0], searchTestPersons[1]},
	},
}

func testFuzzyFindPersons(t *testing.T, db DB) {
	for _, p := range searchTestPersons {
		err := db.InsertPerson(&p)
		if err != nil {
			t.Fatalf("insert test persons returned error %v", err)
		}
	}

	for _, test := range fuzzyFindTests {
		fuzzyFindPersons(t, db, test.query, test.in, test.out)
	}
}

func TestDBFuzzyFindPersons(t *testing.T) {
	testFuzzyFindPersons(t, testDB)
}

func TestMockDBFuzzyFindPersons(t *testing.T) {
	db := NewMockDB(20, 5)
	testFuzzyFindPersons(t, db)
}
