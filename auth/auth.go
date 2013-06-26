/*
Copyright 2011 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var kBasicAuthPattern *regexp.Regexp = regexp.MustCompile(`^Basic ([a-zA-Z0-9\+/=]+)`)

var (
	mode = &UserPass{Username: "robert", Password: "password"}
)

func basicAuth(req *http.Request) (string, string, error) {
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return "", "", fmt.Errorf("Missing \"Authorization\" in header")
	}
	matches := kBasicAuthPattern.FindStringSubmatch(auth)
	if len(matches) != 2 {
		return "", "", fmt.Errorf("Bogus Authorization header")
	}
	encoded := matches[1]
	enc := base64.StdEncoding
	decBuf := make([]byte, enc.DecodedLen(len(encoded)))
	n, err := enc.Decode(decBuf, []byte(encoded))
	if err != nil {
		return "", "", err
	}
	pieces := strings.SplitN(string(decBuf[0:n]), ":", 2)
	if len(pieces) != 2 {
		return "", "", fmt.Errorf("didn't get two pieces")
	}
	return pieces[0], pieces[1], nil
}

// UserPass is used when the auth string provided in the config
// is of the kind "userpass:username:pass"
type UserPass struct {
	Username, Password string
}

func (up *UserPass) IsAuthorized(req *http.Request) bool {
	user, pass, err := basicAuth(req)
	if err != nil {
		return false
	}
	return user == up.Username && pass == up.Password
}

func SendUnauthorized(conn http.ResponseWriter) {
	realm := "rcas"
	conn.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=%q", realm))
	conn.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(conn, "<h1>Unauthorized</h1>")
}

// RequireAuth wraps a function with another function that enforces
// HTTP Basic Auth.
func RequireAuth(handler func(conn http.ResponseWriter, req *http.Request)) func(conn http.ResponseWriter, req *http.Request) {
	return func(conn http.ResponseWriter, req *http.Request) {
		if mode.IsAuthorized(req) {
			handler(conn, req)
		} else {
			SendUnauthorized(conn)
		}
	}
}

type AuthHandler interface {
	http.Handler
	Unauthorized(http.ResponseWriter, *http.Request)
}

type Handler struct {
	AuthHandler
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if mode.IsAuthorized(r) {
		h.AuthHandler.ServeHTTP(w, r)
	} else {
		h.AuthHandler.Unauthorized(w, r)
	}
}
