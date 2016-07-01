package db

import (
	"os"
	"testing"
)

var testDB DB

func TestMain(m *testing.M) {
	var cleanup func()

	TestUseWeakPasswordHashParameters()

	testDB, cleanup = TestDB(20, 5)
	res := m.Run()
	cleanup()
	os.Exit(res)
}
