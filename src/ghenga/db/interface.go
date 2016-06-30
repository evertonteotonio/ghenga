package db

// DB collects all methods for a database.
type DB interface {
	Close() error

	UserDatabase
	PeopleDatabase
	SessionDatabase
}
