package main

import (
	"github.com/go-ldap/ldap"
	"strings"
)

type LDAPSync struct {
	lc         *LDAPCaller
	uc         *UyuniCaller
	cr         *ConfigReader
	ldapusers  []UyuniUser
	uyuniusers []UyuniUser
}

func NewLDAPSync(cfgpath string) *LDAPSync {
	sync := new(LDAPSync)
	sync.cr = NewConfigReader(cfgpath)

	sync.lc = NewLDAPCaller().
		SetHost(sync.cr.Config().Directory.Host).
		SetPort(sync.cr.Config().Directory.Port).
		SetUser(sync.cr.Config().Directory.User).
		SetPassword(sync.cr.Config().Directory.Password).
		SetGroupsDn(sync.cr.Config().Directory.Group).
		SetUsersDn(sync.cr.Config().Directory.Users)

	sync.uc = NewUyuniCaller(sync.cr.Config().Spacewalk.Url, !sync.cr.Config().Spacewalk.Checkssl).
		SetUser(sync.cr.Config().Spacewalk.User).
		SetPassword(sync.cr.Config().Spacewalk.Password)
	sync.ldapusers = make([]UyuniUser, 0)
	sync.uyuniusers = make([]UyuniUser, 0)

	return sync
}

func (sync *LDAPSync) Start() *LDAPSync {
	sync.lc.Connect()
	sync.refreshExistingLDAPUsers()
	sync.refreshExistingUyuniUsers()

	return sync
}

func (sync *LDAPSync) Finish() {
	sync.lc.Disconnect()
}

// Helper function that looks for the same user or at least its ID
func (sync LDAPSync) in(user UyuniUser, users []UyuniUser) bool {
	for _, u := range users {
		if u.uid == user.uid {
			return true
		}
	}
	return false
}

// GetUsersToSync will return a list of users that still needs to be added to the Uyuni.
func (sync *LDAPSync) GetUsersToSync() []UyuniUser {
	users := make([]UyuniUser, 0)
	for _, user := range sync.ldapusers {
		if !sync.in(user, sync.uyuniusers) && user.IsValid() {
			users = append(users, user)
		}
	}
	return users
}

// GetFailedUsers returns a list of users that matches the search criteria
// and belong to the given group, but cannot be added due to missing data.
func (sync *LDAPSync) GetFailedUsers() []UyuniUser {
	users := make([]UyuniUser, 0)
	for _, user := range sync.ldapusers {
		if sync.in(user, sync.uyuniusers) || !user.IsValid() {
			users = append(users, user)
		}
	}

	return users
}

// Iterate over possible attribute aliases
func (sync LDAPSync) getAttributes(entry *ldap.Entry, attr ...string) string {
	for _, a := range attr {
		obj := entry.GetAttributeValue(a)
		if obj != "" {
			return obj
		}
	}

	return ""
}

// Get all existing users in Uyuni.
func (sync *LDAPSync) refreshExistingUyuniUsers() []UyuniUser {
	for _, usrdata := range sync.uc.Call("user.listUsers", sync.uc.Session()).([]interface{}) {
		user := NewUyuniUser()
		user.uid = usrdata.(map[string]interface{})["login"].(string)

		userDetails := sync.uc.Call("user.getDetails", sync.uc.Session(), user.uid).(map[string]interface{})

		user.email = userDetails["email"].(string)
		user.name = userDetails["first_name"].(string)
		user.secondname = userDetails["last_name"].(string)

		sync.uyuniusers = append(sync.uyuniusers, *user)
	}
	return sync.uyuniusers
}

// Get existing LDAP users, including those that are in Uyuni registry
func (sync *LDAPSync) refreshExistingLDAPUsers() []UyuniUser {
	request := ldap.NewSearchRequest(sync.lc.usersdn, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)", []string{}, nil)

	for _, entry := range sync.lc.Search(request).Entries {
		user := NewUyuniUser()
		user.uid = entry.GetAttributeValue("uid")
		user.email = entry.GetAttributeValue("mail")

		cn := strings.Split(entry.GetAttributeValue("cn"), " ")
		if len(cn) == 2 {
			user.name, user.secondname = cn[0], cn[1]
		} else {
			user.name = sync.getAttributes(entry, "name", "givenName")
			user.secondname = entry.GetAttributeValue("sn")
		}

		if user.uid != "" {
			sync.ldapusers = append(sync.ldapusers, *user)
		}
	}

	return sync.ldapusers
}
