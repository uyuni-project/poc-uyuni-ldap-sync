package main

import (
	"fmt"
	"github.com/go-ldap/ldap"
	"log"
)

type LDAPCaller struct {
	user     string
	password string
	host     string
	proto    string
	port     int64
	usersdn  string
	groupsdn string
	conn     *ldap.Conn
}

// Constructor of the LDAP caller with default options
func NewLDAPCaller() *LDAPCaller {
	lc := new(LDAPCaller)
	lc.proto = "tcp"
	lc.port = 389

	return lc
}

func (lc *LDAPCaller) SetUser(user string) *LDAPCaller {
	lc.user = user
	return lc
}

func (lc *LDAPCaller) SetPassword(password string) *LDAPCaller {
	lc.password = password
	return lc
}

func (lc *LDAPCaller) SetPort(port int64) *LDAPCaller {
	lc.port = port
	return lc
}

func (lc *LDAPCaller) SetProto(proto string) *LDAPCaller {
	lc.proto = proto
	return lc
}

func (lc *LDAPCaller) SetHost(host string) *LDAPCaller {
	lc.host = host
	return lc
}

func (lc *LDAPCaller) SetGroupsDn(dn string) *LDAPCaller {
	lc.groupsdn = dn
	return lc
}

func (lc *LDAPCaller) SetUsersDn(dn string) *LDAPCaller {
	lc.usersdn = dn
	return lc
}

func (lc *LDAPCaller) Connect() {
	var err error
	lc.conn, err = ldap.Dial(lc.proto, fmt.Sprintf("%s:%d", lc.host, lc.port))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Bound successfully.")
}

func (lc *LDAPCaller) Disconnect() {
	if lc.conn != nil {
		lc.conn.Close()
		lc.conn = nil
		fmt.Println("Closed")
	}
}

func (lc *LDAPCaller) Search(request *ldap.SearchRequest) *ldap.SearchResult {
	res, err := lc.conn.Search(request)
	if err != nil {
		log.Fatal(err)
	}
	return res
}
