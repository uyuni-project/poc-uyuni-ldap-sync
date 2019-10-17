package main

import (
	"crypto/tls"
	"fmt"
	"github.com/kolo/xmlrpc"
	"log"
	"net/http"
)

type UyuniCaller struct {
	client   *xmlrpc.Client
	user     string
	password string
	session  string
}

// NewUyuniCaller is a constructor for the UyuniCaller object
func NewUyuniCaller(url string, skipSslCheck bool) *UyuniCaller {
	uc := new(UyuniCaller)
	uc.client, _ = xmlrpc.NewClient(url,
		&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipSslCheck,
			},
		})
	return uc
}

// SetUser sets the username for the authentication
func (c *UyuniCaller) SetUser(user string) *UyuniCaller {
	c.user = user
	return c
}

// SetPassword sets the password for the authentication
func (c *UyuniCaller) SetPassword(password string) *UyuniCaller {
	c.password = password
	return c
}

// Obtain an authentication token
func (c *UyuniCaller) authenticate() {
	if c.user != "" && c.password != "" {
		c.session = c.Call("auth.login", c.user, c.password).(string)
	} else {
		log.Fatalf("User and/or password for Uyuni required!")
	}
}

// Session returns a token after the authentication
func (c *UyuniCaller) Session() string {
	if c.session == "" {
		c.authenticate()
	}
	return c.session
}

// Call any XML-RPC function
func (c *UyuniCaller) Call(name string, args ...interface{}) interface{} {
	var res interface{}
	err := c.client.Call(name, args, &res)
	if err != nil {
		log.Fatal(err)
	}

	return res
}
