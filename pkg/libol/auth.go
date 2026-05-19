package libol

import "encoding/base64"

type Auth struct {
	Type     string
	Username string
	Password string
}

func BasicAuth(username, password string) string {
	auth := username + ":"
	if password != "" {
		auth += password
	}
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
