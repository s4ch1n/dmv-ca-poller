package notification

import (
	"fmt"
	"net/smtp"
	"strings"
)

// outlook.com needs custermized stmp.Auth
type loginAuth struct {
	username, password string
}

// OutlookSmtpAuth returns an Auth that implements the LOGIN authentication
// mechanism as defined in RFC 4616.
func OutlookSmtpAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

//Start to satisfy smtp.Auth interface
func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", nil, nil
}

//Next to satisfy smtp.Auth interface
func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	command := string(fromServer)
	command = strings.TrimSpace(command)
	command = strings.TrimSuffix(command, ":")
	command = strings.ToLower(command)

	if !more {
		return nil, nil
	}
	if command == "username" {
		return []byte(a.username), nil
	} else if command == "password" {
		return []byte(a.password), nil
	} else {
		// We've already sent everything.
		return nil, fmt.Errorf("unexpected server challenge: %s", command)
	}

}
