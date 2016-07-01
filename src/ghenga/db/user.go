package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/elithrar/simple-scrypt"
	"github.com/jmoiron/modl"
)

// UserDatabase stores user records.
type UserDatabase interface {
	FindUser(int64) (*User, error)
	FindUserName(string) (*User, error)

	InsertUser(*User) error
	ListUsers() ([]*User, error)
	UpdateUser(*User) error
	DeleteUser(int64) error
}

// User is a user of the system in the database.
type User struct {
	ID           int64
	Login        string
	PasswordHash string
	Admin        bool

	Password string `db:"-"`

	ChangedAt time.Time
	CreatedAt time.Time
	Version   int64
}

// UserJSON is the JSON representation of a User.
type UserJSON struct {
	ID       int64  `json:"id,omitempty"`
	Login    string `json:"login,omitempty"`
	Admin    bool   `json:"admin"`
	Password string `json:"password,omitempty"`

	ChangedAt string `json:"changed_at"`
	CreatedAt string `json:"created_at"`
	Version   int64  `json:"version"`
}

// NewUser returns a new User initialized with the given password.
func NewUser(login, password string) (*User, error) {
	u := &User{
		Login:     login,
		CreatedAt: time.Now(),
		ChangedAt: time.Now(),
	}

	if err := u.UpdatePasswordHash(password); err != nil {
		return nil, err
	}

	return u, nil
}

// NewUserWithHash returns a new User initialized with the given password hash.
func NewUserWithHash(login, hash string) (*User, error) {
	u := &User{
		Login:        login,
		CreatedAt:    time.Now(),
		ChangedAt:    time.Now(),
		PasswordHash: hash,
	}

	return u, nil
}

// CheckPassword returns true iff the password matches the user's password hash.
func (u User) CheckPassword(password string) bool {
	err := scrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

var scryptParameters = scrypt.DefaultParams

func init() {
	p, err := scrypt.Calibrate(500*time.Millisecond, 64, scrypt.Params{})
	if err != nil {
		panic(err)
	}

	scryptParameters = p
}

// WeakenPasswordHash configures weaker scrypt password hash parameters. This
// must only be used in tests.
func TestUseWeakPasswordHashParameters() {
	scryptParameters = scrypt.Params{N: 128, R: 8, P: 1, SaltLen: 16, DKLen: 32}
}

// UpdatePasswordHash updates the password hash for u.
func (u *User) UpdatePasswordHash(password string) error {
	hash, err := scrypt.GenerateFromPassword([]byte(password), scryptParameters)
	if err != nil {
		return err
	}

	u.PasswordHash = string(hash)
	return nil
}

func (u User) String() string {
	return fmt.Sprintf("<User %v (%v)>", u.Login, u.ID)
}

// MarshalJSON returns the JSON representation of u.
func (u User) MarshalJSON() ([]byte, error) {
	ju := UserJSON{
		ID:    u.ID,
		Login: u.Login,
		Admin: u.Admin,

		ChangedAt: u.ChangedAt.Format(timeLayout),
		CreatedAt: u.CreatedAt.Format(timeLayout),
		Version:   u.Version,
	}

	if u.Password != "" && u.PasswordHash == "" {
		ju.Password = u.Password
	}

	return json.Marshal(ju)
}

// UnmarshalJSON returns a user from JSON.
func (u *User) UnmarshalJSON(data []byte) error {
	var ju UserJSON

	err := json.Unmarshal(data, &ju)
	if err != nil {
		return err
	}

	hash, err := scrypt.GenerateFromPassword([]byte(ju.Password), scrypt.DefaultParams)
	if err != nil {
		return err
	}

	createdAt, err := time.Parse(timeLayout, ju.CreatedAt)
	if err != nil {
		return err
	}

	changedAt, err := time.Parse(timeLayout, ju.ChangedAt)
	if err != nil {
		return err
	}

	*u = User{
		ID:           ju.ID,
		Login:        ju.Login,
		Admin:        ju.Admin,
		PasswordHash: string(hash),

		CreatedAt: createdAt,
		ChangedAt: changedAt,
		Version:   ju.Version,
	}

	return nil
}

// PreInsert is run before a person is saved into the database. It is used to
// update the password hash when the field `Password` is set.
func (u *User) PreInsert(db modl.SqlExecutor) error {
	if u.Password == "" {
		return nil
	}

	return u.UpdatePasswordHash(u.Password)
}

// PreUpdate is run before a person is saved into the database. It is used to
// update the password hash when the field `Password` is set.
func (u *User) PreUpdate(db modl.SqlExecutor) error {
	if u.Password == "" {
		return nil
	}

	return u.UpdatePasswordHash(u.Password)
}

// Validate checks whether the user record does not contain any errors.
func (u User) Validate() error {
	if u.Login == "" {
		return errors.New("login must not be empty")
	}

	if u.PasswordHash == "" {
		return errors.New("user must have a password hash")
	}

	if u.CreatedAt.IsZero() || u.ChangedAt.IsZero() {
		return errors.New("invalid timestamps")
	}

	return nil
}

// Update updates some fields from other.
func (u *User) Update(other UserJSON) {
	u.Login = other.Login
	u.Admin = other.Admin

	if other.Password != "" {
		u.UpdatePasswordHash(other.Password)
	}
}

// FindUserName searches the database for a user based on their login name.
func (db *Database) FindUserName(login string) (*User, error) {
	var u User
	err := db.dbmap.SelectOne(&u, "SELECT * FROM users WHERE login = $1", login)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// FindUser searches the database for a user based on their id.
func (db *Database) FindUser(id int64) (*User, error) {
	var u User
	err := db.dbmap.SelectOne(&u, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// ListUsers returns the list of User.
func (db *Database) ListUsers() ([]*User, error) {
	var user []*User
	err := db.dbmap.Select(&user, "select * from users")
	return user, err
}

// UpdateUser modifies an existing user.
func (db *Database) UpdateUser(u *User) error {
	_, err := db.dbmap.Update(u)
	return err
}

// InsertUser creates a new user.
func (db *Database) InsertUser(u *User) error {
	return db.dbmap.Insert(u)
}

// DeleteUser removes a user.
func (db *Database) DeleteUser(id int64) error {
	res := db.dbmap.Dbx.MustExec("delete from people where id = $1", id)
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return errors.New("user not found")
	}

	return nil
}
