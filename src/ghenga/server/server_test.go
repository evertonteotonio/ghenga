package server

import "ghenga/db"

func init() {
	// use weaker scrypt hash parameters to decrease test execution time
	db.TestUseWeakPasswordHashParameters()
}
