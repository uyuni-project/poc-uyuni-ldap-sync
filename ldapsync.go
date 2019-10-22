package main

import (
	"fmt"
	"github.com/go-ldap/ldap"
	"log"
	"strings"
)

type LDAPSync struct {
	lc         *LDAPCaller
	uc         *UyuniCaller
	cr         *ConfigReader
	ldapusers  []*UyuniUser
	uyuniusers []*UyuniUser
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
	sync.ldapusers = make([]*UyuniUser, 0)
	sync.uyuniusers = make([]*UyuniUser, 0)

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
func (sync LDAPSync) in(user UyuniUser, users []*UyuniUser) bool {
	for _, u := range users {
		if u.Uid == user.Uid {
			return true
		}
	}
	return false
}

// GetUsersToSync will return a list of users that still needs to be added to the Uyuni.
func (sync *LDAPSync) GetUsersToSync() []*UyuniUser {
	users := make([]*UyuniUser, 0)
	for _, user := range sync.ldapusers {
		if !sync.in(*user, sync.uyuniusers) && user.IsValid() {
			users = append(users, user)
		}
	}
	return users
}

// GetFailedUsers returns a list of users that matches the search criteria
// and belong to the given group, but cannot be added due to missing data.
func (sync *LDAPSync) GetFailedUsers() []*UyuniUser {
	users := make([]*UyuniUser, 0)
	for _, user := range sync.ldapusers {
		if sync.in(*user, sync.uyuniusers) || !user.IsValid() {
			users = append(users, user)
		}
	}

	return users
}

func (sync *LDAPSync) SyncUsers() []*UyuniUser {
	failed := make([]*UyuniUser, 0)
	for _, user := range sync.GetUsersToSync() {
		// The 1 is PAM authentication usage
		fmt.Println("Synchronising user", user.Uid)
		_, user.Err = sync.uc.Call("user.create", sync.uc.Session(), user.Uid, "", user.Name, user.Secondname, user.Email, 1)
		if !user.IsValid() {
			failed = append(failed, user)
		}
	}
	return failed
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
func (sync *LDAPSync) refreshExistingUyuniUsers() []*UyuniUser {
	res, err := sync.uc.Call("user.listUsers", sync.uc.Session())
	if err != nil {
		log.Fatal(err)
	}
	for _, usrdata := range res.([]interface{}) {
		user := NewUyuniUser()
		user.Uid = usrdata.(map[string]interface{})["login"].(string)

		res, err = sync.uc.Call("user.getDetails", sync.uc.Session(), user.Uid)
		if err != nil {
			log.Fatal(err)
		}
		userDetails := res.(map[string]interface{})

		user.Email = userDetails["email"].(string)
		user.Name = userDetails["first_name"].(string)
		user.Secondname = userDetails["last_name"].(string)

		sync.uyuniusers = append(sync.uyuniusers, user)
	}
	return sync.uyuniusers
}

// Get existing LDAP users, including those that are in Uyuni registry
func (sync *LDAPSync) refreshExistingLDAPUsers() []*UyuniUser {
	request := ldap.NewSearchRequest(sync.lc.usersdn, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)", []string{}, nil)

	for _, entry := range sync.lc.Search(request).Entries {
		user := NewUyuniUser()
		user.Dn = entry.DN
		user.Uid = entry.GetAttributeValue("uid")
		user.Email = entry.GetAttributeValue("mail")

		cn := strings.Split(entry.GetAttributeValue("cn"), " ")
		if len(cn) == 2 {
			user.Name, user.Secondname = cn[0], cn[1]
		} else {
			user.Name = sync.getAttributes(entry, "name", "givenName")
			user.Secondname = entry.GetAttributeValue("sn")
		}

		if user.Uid != "" {
			sync.ldapusers = append(sync.ldapusers, user)
		}
	}

	for _, user := range sync.ldapusers {
		sync.updateLDAPUserRoles(user)
	}

	return sync.ldapusers
}

func (sync *LDAPSync) mergeRolesByAttributes(dn string, user *UyuniUser, filter string, attribute string, uyuniRoles []string) {
	req := ldap.NewSearchRequest(dn, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, filter, []string{}, nil)
	for _, entry := range sync.lc.Search(req).Entries {
		for _, roleDn := range entry.GetAttributeValues(attribute) {
			if roleDn == user.Dn {
				user.AddRoles(uyuniRoles...)
			}
		}
	}
}

// Get LDAP organizationalRole based on configuration
func (sync *LDAPSync) updateLDAPUserRoles(user *UyuniUser) {
	type SearchConfig struct {
		config    *[]map[string][]string
		filter    string
		attribute string
	}

	roleConfigs := [...]SearchConfig{
		SearchConfig{config: &sync.cr.Config().Directory.Roles,
			filter: "(objectClass=organizationalRole)", attribute: "roleOccupant"},
		SearchConfig{config: &sync.cr.Config().Directory.Groups,
			filter: "(|(objectClass=groupOfNames)(objectClass=group))", attribute: "member"},
	}

	for _, searchConfig := range roleConfigs {
		for _, roleConfig := range *searchConfig.config {
			for dn, uyuniRoles := range roleConfig {
				sync.mergeRolesByAttributes(dn, user, searchConfig.filter, searchConfig.attribute, uyuniRoles)
			}
		}
	}
}

func (sync *LDAPSync) TestBed() {
	fmt.Println("Checking updated users:")
	for _, user := range sync.ldapusers {
		if len(user.GetRoles()) > 0 {
			fmt.Println("User", user.Name, user.Secondname, "has roles:")
			for _, role := range user.GetRoles() {
				fmt.Println("  -", role)
			}
		}
	}
}
