package main

import (
	"ghenga/db"
	"log"
)

type cmdFakedata struct {
	People int `short:"p" long:"people" default:"500" description:"Number of fake people profiles to create"`
	User   int `short:"u" long:"user" default:"5" description:"Number of fake user to create"`
}

func init() {
	_, err := parser.AddCommand("fakedata",
		"fill db with fake dat",
		"This command fills the data with fake information that looks real",
		&cmdFakedata{})
	if err != nil {
		panic(err)
	}
}

func (opts *cmdFakedata) Execute(args []string) (err error) {
	dbm, e := db.Open(globalOpts.DB)
	if e != nil {
		return e
	}
	defer CleanupErr(&err, func() error {
		return dbm.Close()
	})

	log.Printf("inserting fake data into the db...")

	err = db.InsertFakeData(dbm, opts.People, opts.User)
	if err != nil {
		panic(err)
	}
	log.Printf("done")

	return nil
}
