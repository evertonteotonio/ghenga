package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/modl"
	"github.com/jmoiron/sqlx"
)

// PeopleDatabase allows handling people.
type PeopleDatabase interface {
	FindPerson(int64) (*Person, error)

	InsertPerson(*Person) error
	ListPeople() ([]*Person, error)
	UpdatePerson(*Person) error
	DeletePerson(int64) error

	FuzzyFindPersons(query string) ([]*Person, error)
}

// Person is a person in the database.
type Person struct {
	ID           int64
	Name         string
	Title        string
	Department   string
	EmailAddress string
	PhoneNumbers PhoneNumbers `db:"-"`

	// Address
	Street     string
	PostalCode string
	State      string
	City       string
	Country    string

	Comment string

	ChangedAt time.Time
	CreatedAt time.Time
	Version   int64
}

// PersonJSON is the JSON representation of a Person as returned or consumed by
// the API.
type PersonJSON struct {
	ID           int64             `json:"id,omitempty"`
	Name         string            `json:"name,omitempty"`
	Title        string            `json:"title,omitempty"`
	Department   string            `json:"department,omitempty"`
	EmailAddress string            `json:"email_address,omitempty"`
	PhoneNumbers []PhoneNumberJSON `json:"phone_numbers"`

	Address AddressJSON `json:"address,omitempty"`

	Comment string `json:"comment,omitempty"`

	ChangedAt string `json:"changed_at,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`

	Version int64 `json:"version"`
}

// AddressJSON is the JSON representation of an address.
type AddressJSON struct {
	Street     string `json:"street,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	State      string `json:"state,omitempty"`
	City       string `json:"city,omitempty"`
	Country    string `json:"country,omitempty"`
}

// NewPerson returns a new person record.
func NewPerson(name string) *Person {
	ts := time.Now()
	return &Person{
		Name:      name,
		CreatedAt: ts,
		ChangedAt: ts,
	}
}

// timeLayout is the format used for the timestamps in the JSON representation
// of a Person.
const timeLayout = "2006-01-02T15:04:05-07:00"

// MarshalJSON returns the JSON representation of p.
func (p Person) MarshalJSON() ([]byte, error) {
	jp := PersonJSON{
		ID:   p.ID,
		Name: p.Name,

		ChangedAt: p.ChangedAt.Format(timeLayout),
		CreatedAt: p.CreatedAt.Format(timeLayout),
		Version:   p.Version,
	}

	jp.Title = p.Title
	jp.Department = p.Department
	jp.EmailAddress = p.EmailAddress

	jp.PhoneNumbers = []PhoneNumberJSON{}
	for _, pn := range p.PhoneNumbers {
		jp.PhoneNumbers = append(jp.PhoneNumbers, PhoneNumberJSON{
			Type:   pn.Type,
			Number: pn.Number,
		})
	}

	jp.Address = AddressJSON{
		Street:     p.Street,
		PostalCode: p.PostalCode,
		State:      p.State,
		City:       p.City,
		Country:    p.Country,
	}

	jp.Comment = p.Comment
	return json.Marshal(jp)
}

// UnmarshalJSON returns a person from JSON.
func (p *Person) UnmarshalJSON(data []byte) error {
	var jp PersonJSON

	err := json.Unmarshal(data, &jp)
	if err != nil {
		return err
	}

	createdAt, err := time.Parse(timeLayout, jp.CreatedAt)
	if err != nil {
		return err
	}

	changedAt, err := time.Parse(timeLayout, jp.ChangedAt)
	if err != nil {
		return err
	}

	*p = Person{
		ID:           jp.ID,
		Name:         jp.Name,
		Title:        jp.Title,
		Department:   jp.Department,
		EmailAddress: jp.EmailAddress,

		Street:     jp.Address.Street,
		PostalCode: jp.Address.PostalCode,
		State:      jp.Address.State,
		City:       jp.Address.City,
		Country:    jp.Address.Country,

		Comment: jp.Comment,

		CreatedAt: createdAt,
		ChangedAt: changedAt,
		Version:   jp.Version,
	}

	for _, num := range jp.PhoneNumbers {
		n := PhoneNumber{
			Number: num.Number,
			Type:   num.Type,
		}
		p.PhoneNumbers = append(p.PhoneNumbers, n)
	}

	return nil
}

// Validate checks if p is valid and returns an error if not.
func (p *Person) Validate() error {
	if p.Name == "" {
		return errors.New("name is empty")
	}

	if p.CreatedAt.IsZero() || p.ChangedAt.IsZero() {
		return errors.New("invalid timestamps")
	}

	return nil
}

// PostInsert is run after a person is saved into the database. It is
// used to handle phone numbers associated with a person.
func (p *Person) PostInsert(db modl.SqlExecutor) error {
	if len(p.PhoneNumbers) == 0 {
		return nil
	}

	for _, num := range p.PhoneNumbers {
		num.PersonID = p.ID
		err := db.Insert(&num)
		if err != nil {
			return err
		}
	}

	return nil
}

// PostGet loads the phone numbers associated with the person.
func (p *Person) PostGet(db modl.SqlExecutor) error {
	return db.Select(&p.PhoneNumbers, "SELECT * FROM phone_numbers WHERE person_id = $1", p.ID)
}

// in is a small wrapper around the sqlx.In() function which handles rebinding
// to a different bindtype.
func in(query string, args ...interface{}) (string, []interface{}, error) {
	query, args, err := sqlx.In(query, args...)
	if err != nil {
		return query, args, err
	}

	bindType := sqlx.BindType(dialect)
	query = sqlx.Rebind(bindType, query)

	return query, args, err
}

// PostUpdate is run after a person has been updated. It handles updating the
// phone numbers for a person.
func (p *Person) PostUpdate(db modl.SqlExecutor) error {
	var ids []int64
	for _, num := range p.PhoneNumbers {
		num.PersonID = p.ID
		var err error
		if num.ID != 0 {
			_, err = db.Update(&num)
		} else {
			err = db.Insert(&num)
		}

		if err != nil {
			return err
		}

		ids = append(ids, num.ID)
	}

	if len(ids) > 0 {
		// remove excess phone numbers
		query, args, err := in("DELETE FROM phone_numbers WHERE person_id = ? AND id NOT IN (?)", p.ID, ids)
		if err != nil {
			return err
		}

		_, err = db.Exec(query, args...)
		return err
	}

	// else remove all phone numbers
	_, err := db.Exec("DELETE FROM phone_numbers WHERE person_id = $1", p.ID)
	return err
}

// Update updates p with the fields from other.
func (p *Person) Update(other PersonJSON) {
	p.Name = other.Name
	p.Title = other.Title
	p.Department = other.Department
	p.EmailAddress = other.EmailAddress

	p.PhoneNumbers = nil

	for _, num := range other.PhoneNumbers {
		p.PhoneNumbers = append(p.PhoneNumbers, PhoneNumber{
			Type:   num.Type,
			Number: num.Number,
		})
	}

	p.Street = other.Address.Street
	p.PostalCode = other.Address.PostalCode
	p.State = other.Address.State
	p.City = other.Address.City
	p.Country = other.Address.Country

	p.Comment = other.Comment

	p.Version = other.Version
}

func (p Person) String() string {
	numbers := ""
	if len(p.PhoneNumbers) > 0 {
		for _, num := range p.PhoneNumbers {
			numbers += fmt.Sprintf(", %v [%v]", num.Number, num.Type)
		}
	}

	return fmt.Sprintf("<Person[%v] (%v)%s>", p.ID, p.Name, numbers)
}

// FindPerson returns the person struct with the given id.
func (db *Database) FindPerson(id int64) (*Person, error) {
	var p Person

	err := db.dbmap.SelectOne(&p, "SELECT * FROM people WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// UpdatePerson modifies an existing person.
func (db *Database) UpdatePerson(p *Person) error {
	_, err := db.dbmap.Update(p)
	return err
}

// InsertPerson creates a new person.
func (db *Database) InsertPerson(p *Person) error {
	return db.dbmap.Insert(p)
}

// ListPeople returns the list of people.
func (db *Database) ListPeople() ([]*Person, error) {
	var people []*Person
	err := db.dbmap.Select(&people, "select * from people")
	return people, err
}

// DeletePerson removes a person.
func (db *Database) DeletePerson(id int64) error {
	res := db.dbmap.Dbx.MustExec("delete from people where id = $1", id)
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return errors.New("person not found")
	}

	return nil
}
