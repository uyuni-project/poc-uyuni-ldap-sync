package main

import (
	"fmt"
	"github.com/go-ldap/ldap"
	"log"
	"strings"
)

type SearchConfig struct {
	config    *map[string][]string
	filter    string
	attribute string
}

type LDAPSync struct {
	lc          *LDAPCaller
	uc          *UyuniCaller
	cr          *ConfigReader
	ldapusers   []*UyuniUser
	uyuniusers  []*UyuniUser
	roleConfigs [2]*SearchConfig
}

func NewLDAPSync(cfgpath string) *LDAPSync {
	sync := new(LDAPSync)
	sync.cr = NewConfigReader(cfgpath)

	sync.lc = NewLDAPCaller().
		SetHost(sync.cr.Config().Directory.Host).
		SetPort(sync.cr.Config().Directory.Port).
		SetUser(sync.cr.Config().Directory.User).
		SetPassword(sync.cr.Config().Directory.Password).
		SetUsersDn(sync.cr.Config().Directory.Users)

	sync.uc = NewUyuniCaller(sync.cr.Config().Spacewalk.Url, !sync.cr.Config().Spacewalk.Checkssl).
		SetUser(sync.cr.Config().Spacewalk.User).
		SetPassword(sync.cr.Config().Spacewalk.Password)
	sync.ldapusers = make([]*UyuniUser, 0)
	sync.uyuniusers = make([]*UyuniUser, 0)

	sync.roleConfigs = [2]*SearchConfig{
		&SearchConfig{config: &sync.cr.Config().Directory.Roles,
			filter: "(objectClass=organizationalRole)", attribute: "roleOccupant"},
		&SearchConfig{config: &sync.cr.Config().Directory.Groups,
			filter: "(|(objectClass=groupOfNames)(objectClass=group))", attribute: "member"},
	}
	return sync
}

func (sync *LDAPSync) Start() *LDAPSync {
	sync.lc.Connect()
	sync.refreshExistingUyuniUsers()
	sync.refreshExistingLDAPUsers()
	sync.refreshUyuniUsersStatus()

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

// Match a given user by a DN, compare all metadata.
func (sync LDAPSync) sameAsIn(user *UyuniUser, users []*UyuniUser) (bool, error) {
	for _, u := range users {
		if u.Uid == user.Uid {
			same := u.Email == user.Email
			if same {
				same = u.Name == user.Name
			}

			if same {
				same = u.Secondname == user.Secondname
			}

			if same {
				same = CompareRoles(user, u)
			}

			return same, nil
		}
	}
	return false, fmt.Errorf("User UID %s was not found", user.Uid)
}

// Returns a copy of LDAP user by Uyuni user
func (sync *LDAPSync) updateFromLDAPUser(uyuniUser *UyuniUser) {
	for _, ldapUser := range sync.ldapusers {
		if ldapUser.Uid == uyuniUser.Uid {
			uyuniUser.Name, uyuniUser.Secondname, uyuniUser.Email = ldapUser.Name, ldapUser.Secondname, ldapUser.Email
			uyuniUser.FlushRoles()
			for _, role := range ldapUser.GetRoles() {
				uyuniUser.AddRoles(role)
			}
		}
	}
}

// GetNewUsers returns LDAP users that are not yet in the Uyuni
func (sync *LDAPSync) GetNewUsers() []*UyuniUser {
	var users []*UyuniUser
	for _, user := range sync.uyuniusers {
		if user.IsNew() {
			sync.updateFromLDAPUser(user)
			users = append(users, user)
		}
	}

	return users
}

// GetOutdatedUsers returns LDAP users that are in the Uyuni, but needs refresh
func (sync *LDAPSync) GetOutdatedUsers() []*UyuniUser {
	var users []*UyuniUser
	for _, user := range sync.uyuniusers {
		if !user.IsNew() && user.IsOutdated() {
			users = append(users, user)
		}
	}

	return users
}

// SyncUsers is creating new users in Uyuni by their names and emails.
func (sync *LDAPSync) SyncUsers() []*UyuniUser {
	failed := make([]*UyuniUser, 0)
	newUsers := sync.GetNewUsers()
	if len(newUsers) > 0 {
		fmt.Println("Adding new users...")
		for idx, user := range newUsers {
			idx++
			fmt.Printf("  %d. %s\n", idx, user.Uid)
			// The 1 is for PAM authentication usage
			_, user.Err = sync.uc.Call("user.create", sync.uc.Session(), user.Uid, "", user.Name, user.Secondname, user.Email, 1)

			if !user.IsValid() {
				failed = append(failed, user)
			}
		}
	} else {
		fmt.Println("No new users to be added")
	}

	existingUsers := sync.GetOutdatedUsers()
	if len(existingUsers) > 0 {
		fmt.Println("Updating existing users...")
		for idx, user := range existingUsers {
			idx++
			fmt.Printf("  %d. %s", idx, user.Uid)
			// The 1 is for PAM authentication usage
			//_, user.Err = sync.uc.Call("user.create", sync.uc.Session(), user.Uid, "", user.Name, user.Secondname, user.Email, 1)

			//if !user.IsValid() {
			//	failed = append(failed, user)
			//}
		}
	} else {
		fmt.Println("No users to be updated")
	}

	return failed
}

// SyncUserRoles synchronises roles of each user.
func (sync *LDAPSync) SyncUserRoles() []*UyuniUser {
	uyuniStaticRoles := [...]string{
		"satellite_admin", "org_admin", "channel_admin", "config_admin",
		"system_group_admin", "activation_key_admin", "image_admin",
	}
	failed := make([]*UyuniUser, 0)
	var hasFailures bool
	// Cleanup all roles
	for _, user := range sync.GetNewUsers() { // XXX: Here must be actually the same users as in LDAP, because this will only clean each user roles and update from LDAP.
		hasFailures = false
		fmt.Println("Synchronising roles for user", user.Uid)
		for _, role := range uyuniStaticRoles {
			_, err := sync.uc.Call("user.removeRole", sync.uc.Session(), user.Uid, role)
			if err != nil {
				fmt.Println("Failed to remove existing role", role, "due to", err.Error())
				hasFailures = true
			}
		}

		for _, newRole := range user.GetRoles() {
			_, err := sync.uc.Call("user.addRole", sync.uc.Session(), user.Uid, newRole)
			if err != nil {
				fmt.Println("Failed to set a new role", newRole, "due to", err.Error())
				hasFailures = true
			}
		}

		if hasFailures {
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

// Pick a user from the array of those
func (sync *LDAPSync) pickUserByUid(uid string, users []*UyuniUser) *UyuniUser {
	for _, user := range users {
		if user.Uid == uid {
			return user
		}
	}
	return nil
}

// Refresh what users are new and what needs update
func (sync *LDAPSync) refreshUyuniUsersStatus() []*UyuniUser {
	for _, user := range sync.ldapusers {
		user.new = !sync.in(*user, sync.uyuniusers)
		if !user.IsNew() {
			isSame, err := sync.sameAsIn(user, sync.uyuniusers)
			if err != nil {
				user.outdated = !isSame
			}
		} else {
			sync.uyuniusers = append(sync.uyuniusers, user)
		}
	}

	return sync.uyuniusers
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

		// Get user roles
		res, err = sync.uc.Call("user.listRoles", sync.uc.Session(), user.Uid)
		if err != nil {
			log.Fatal(err)
		}

		for _, roleItf := range res.([]interface{}) {
			user.AddRoles(roleItf.(string))
		}

		sync.uyuniusers = append(sync.uyuniusers, user)
	}
	return sync.uyuniusers
}

func (sync *LDAPSync) newUserFromDN(dn string) *UyuniUser {
	user := NewUyuniUser()
	request := ldap.NewSearchRequest(dn, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)", []string{}, nil)

	entries := sync.lc.Search(request).Entries
	if len(entries) == 1 {
		entry := entries[0]
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
	} else {
		fmt.Println("DN does not find one exact user:", dn)
	}

	return user
}

// Get existing LDAP users, based on the groups mapping
func (sync *LDAPSync) refreshExistingLDAPUsers() []*UyuniUser {
	sync.ldapusers = nil
	udns := make(map[string]bool)

	// Get all *distinct* user DNs from the "member" attiribute across all the groups
	for gdn := range sync.cr.Config().Directory.Groups {
		request := ldap.NewSearchRequest(gdn, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			"(objectClass=*)", []string{}, nil)
		for _, entry := range sync.lc.Search(request).Entries {
			for _, udn := range entry.GetAttributeValues("member") {
				udns[udn] = true
			}
		}
	}

	// Collect users data
	for udn := range udns {
		user := sync.newUserFromDN(udn)
		if user.Uid != "" {
			sync.updateLDAPUserRoles(user)
			sync.ldapusers = append(sync.ldapusers, user)
		}
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
	for _, searchConfig := range sync.roleConfigs {
		for dn, uyuniRoles := range *searchConfig.config {
			sync.mergeRolesByAttributes(dn, user, searchConfig.filter, searchConfig.attribute, uyuniRoles)
		}
	}
}
