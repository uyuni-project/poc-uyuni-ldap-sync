package ldapsync

import (
	"crypto/tls"
	"net/http"

	"github.com/kolo/xmlrpc"
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
	var err error
	var res interface{}
	if c.user != "" && c.password != "" {
		res, err = c.Call("auth.login", c.user, c.password)
		if err != nil {
			Log.Fatal(err)
		}
		c.session = res.(string)
	} else {
		Log.Fatalf("User and/or password for Uyuni required!")
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
func (c *UyuniCaller) Call(name string, args ...interface{}) (interface{}, error) {
	var res interface{}
	err := c.client.Call(name, args, &res)
	return res, err
}
