package server

import (
	"errors"
	"ghenga/db"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/gorilla/mux"
)

// LoginResponseJSON is the structure returned by a login request.
type LoginResponseJSON struct {
	User     string `json:"user"`
	Token    string `json:"token"`
	ValidFor uint   `json:"valid_for"`
	Admin    bool   `json:"admin"`
}

// Login allows users to log in and returns a token.
func Login(ctx context.Context, env *Env, res http.ResponseWriter, req *http.Request) error {
	username, password, ok := req.BasicAuth()
	if !ok {
		return StatusError{
			Code: http.StatusUnauthorized,
			Err:  errors.New("no login data present"),
		}
	}

	env.Debugf("login attempt for user %v", username)

	u, err := env.DB.FindUserName(username)
	if err != nil {
		env.Debugf("error finding user %q in database: %v", username, err)
	}

	if err != nil || !u.CheckPassword(password) {
		return StatusError{
			Code: http.StatusUnauthorized,
			Err:  errors.New("invalid username or password"),
		}
	}

	session, err := env.DB.SaveNewSession(username, env.Cfg.SessionDuration)
	if err != nil {
		return err
	}

	return httpWriteJSON(res, http.StatusOK, LoginResponseJSON{
		User:     u.Login,
		Token:    session.Token,
		ValidFor: uint(env.Cfg.SessionDuration / time.Second),
		Admin:    u.Admin,
	})
}

const authHeaderName = "X-Auth-Token"

// findSession returns a session for the request or an error if none is found.
func findSession(env *Env, req *http.Request) (*db.Session, error) {
	token := req.Header.Get(authHeaderName)
	if token == "" {
		return nil, StatusError{
			Code: http.StatusUnauthorized,
			Err:  errors.New("invalid session token"),
		}
	}

	session, err := env.DB.FindSession(token)
	if err != nil {
		env.Logf("error finding session with token %q in database: %v", token, err)
	}

	if err != nil || session.ValidUntil.Before(time.Now()) {
		return nil, StatusError{
			Code: http.StatusUnauthorized,
			Err:  errors.New("invalid session token"),
		}
	}

	return session, nil
}

// Info allows users to check whether a token is still valid and find the
// current username.
func Info(ctx context.Context, env *Env, res http.ResponseWriter, req *http.Request) error {
	session, err := findSession(env, req)
	if err != nil {
		return err
	}

	u, err := env.DB.FindUserName(session.User)
	if err != nil {
		return err
	}

	return httpWriteJSON(res, http.StatusOK, LoginResponseJSON{
		User:     session.User,
		Token:    session.Token,
		ValidFor: uint(session.ValidUntil.Sub(time.Now()) / time.Second),
		Admin:    u.Admin,
	})
}

// Invalidate deletes a valid session token.
func Invalidate(ctx context.Context, env *Env, res http.ResponseWriter, req *http.Request) error {
	session, err := findSession(env, req)
	if err != nil {
		return err
	}

	return env.DB.Invalidate(session)
}

// LoginHandler adds routes to the for ghenga API in the given environment to r.
func LoginHandler(ctx context.Context, env *Env, r *mux.Router) {
	r.Handle("/api/login/token", Handle(ctx, env, Login)).Methods("GET")
	r.Handle("/api/login/info", Handle(ctx, env, Info)).Methods("GET")
	r.Handle("/api/login/invalidate", Handle(ctx, env, Invalidate)).Methods("GET")
}
