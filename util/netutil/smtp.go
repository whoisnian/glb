package netutil

import (
	"crypto/tls"
	"errors"
	"net/smtp"
)

// https://stackoverflow.com/questions/57783841
type LoginAuth struct {
	username, password string
}

func (a *LoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *LoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("unkown fromServer")
		}
	}
	return nil, nil
}

func OutlookAccountVerify(username, password string) error {
	c, err := smtp.Dial("smtp.office365.com:587")
	if err != nil {
		return err
	}
	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: "smtp.office365.com"}
		if err = c.StartTLS(config); err != nil {
			return err
		}
	}
	if ok, _ := c.Extension("AUTH"); ok {
		if err = c.Auth(&LoginAuth{username, password}); err != nil {
			return err
		}
	}
	return c.Quit()
}
