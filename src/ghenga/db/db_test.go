package db

import (
	"os"
	"testing"

	"github.com/elithrar/simple-scrypt"
)

var testDB *DB

func TestMain(m *testing.M) {
	var cleanup func()

	// use weaker scrypt parameters to increase test speed
	scryptParameters = scrypt.Params{N: 128, R: 8, P: 1, SaltLen: 16, DKLen: 32}

	testDB, cleanup = TestDB(20, 5)
	res := m.Run()
	cleanup()
	os.Exit(res)
}
