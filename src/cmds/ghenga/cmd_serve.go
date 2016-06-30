package main

import (
	"fmt"
	"ghenga/server"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/gorilla/handlers"
)

type cmdServe struct {
	Port   uint   `short:"p" long:"port"   default:"8080"   description:"set the port for the HTTP server"`
	Addr   string `short:"b" long:"bind"   default:""       description:"bind to this address"`
	Public string `          long:"public" default:"public" description:"directory for serving static files"`
}

func init() {
	_, err := parser.AddCommand("serve",
		"start server",
		"The server command starts the HTTP server",
		&cmdServe{})
	if err != nil {
		panic(err)
	}
}

const sessionDuration = 12 * time.Hour

func expireSessions(ctx context.Context, env *server.Env, d time.Duration) {
	t := time.NewTicker(d)
	defer t.Stop()

	log.Printf("expiring sessions every %v", d)

	for {
		select {
		case <-t.C:
			n, err := env.DB.ExpireSessions(time.Now())
			if err != nil {
				log.Printf("ExpireSessions returned error %v", err)
				continue
			}
			if n > 0 {
				log.Printf("expired %v sessions", n)
			}
		case <-ctx.Done():
			return
		}
	}
}

const sessionExpireInterval = 5 * time.Minute

func (opts *cmdServe) Execute(args []string) (err error) {
	lgr := log.New(os.Stderr, "", log.LstdFlags)

	dbmap, e := OpenDB()
	if e != nil {
		return e
	}
	defer CleanupErr(&err, dbmap.Close)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lgr.Printf("starting server at %v:%d", opts.Addr, opts.Port)

	env := &server.Env{
		DB: dbmap,
		Cfg: server.Config{
			Debug:           globalOpts.Debug,
			SessionDuration: sessionDuration,
		},
	}

	env.Logger.Error = lgr
	if globalOpts.Debug {
		env.Logger.Debug = log.New(os.Stderr, "", log.LstdFlags)
	}

	go expireSessions(ctx, env, sessionExpireInterval)

	router := server.NewRouter(ctx, env)

	// server static files on the root path
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(opts.Public)))

	// activate logging to stdout
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, router))

	return http.ListenAndServe(fmt.Sprintf("%s:%d", opts.Addr, opts.Port), nil)
}
