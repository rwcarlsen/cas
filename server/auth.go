package server

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
	OrLocalhost        bool // if true, allow localhost ident auth too
}

func (up *UserPass) IsAuthorized(req *http.Request) bool {
	if up.OrLocalhost && localhostAuthorized(req) {
		return true
	}
	user, pass, err := basicAuth(req)
	if err != nil {
		return false
	}
	return user == up.Username && pass == up.Password
}

func (up *UserPass) AddAuthHeader(req *http.Request) {
	req.SetBasicAuth(up.Username, up.Password)
}
