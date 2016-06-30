package main

import "github.com/jessevdk/go-flags"

type globalOptions struct {
	DB    string `short:"d" long:"database" default:"" env:"GHENGA_DB"   description:"Connection string for postgresql database" default:"host=/var/run/postgresql"`
	Debug bool   `short:"D" long:"debug"                            env:"GHENGA_DEBUG" description:"Enable debug messages for development"`
}

var globalOpts = globalOptions{}
var parser = flags.NewParser(&globalOpts, flags.HelpFlag|flags.PassDoubleDash)
