package ldapsync

import (
	"fmt"

	"github.com/go-ldap/ldap"
)

type LDAPCaller struct {
	user     string
	password string
	host     string
	proto    string
	port     int64
	conn     *ldap.Conn
}

// Constructor of the LDAP caller with default options
func NewLDAPCaller() *LDAPCaller {
	lc := new(LDAPCaller)
	lc.proto = "tcp"
	lc.port = 389

	return lc
}

// SetUser for LDAP caller
func (lc *LDAPCaller) SetUser(user string) *LDAPCaller {
	lc.user = user
	return lc
}

// SetPassword sets a password authentication for the LDAP caller
func (lc *LDAPCaller) SetPassword(password string) *LDAPCaller {
	lc.password = password
	return lc
}

// SetPort sets a port for the LDAP caller
func (lc *LDAPCaller) SetPort(port int64) *LDAPCaller {
	lc.port = port
	return lc
}

// SetProto sets a protocol scheme for the LDAP caller
func (lc *LDAPCaller) SetProto(proto string) *LDAPCaller {
	lc.proto = proto
	return lc
}

// SetHost sets a hostname FQDSN for the LDAP caller
func (lc *LDAPCaller) SetHost(host string) *LDAPCaller {
	lc.host = host
	return lc
}

// Connect  to the LDAP
func (lc *LDAPCaller) Connect() {
	var err error
	if lc.conn == nil {
		lc.conn, err = ldap.Dial(lc.proto, fmt.Sprintf("%s:%d", lc.host, lc.port))
		if err != nil {
			Log.Fatal(err)
		}
	}
}

// Disconnect from the LDAP and drain the connection
func (lc *LDAPCaller) Disconnect() {
	if lc.conn != nil {
		lc.conn.Close()
		lc.conn = nil
	}
}

// Search LDAP by request
func (lc *LDAPCaller) Search(request *ldap.SearchRequest) *ldap.SearchResult {
	res, err := lc.conn.Search(request)
	if err != nil {
		Log.Fatal(err)
	}
	return res
}
