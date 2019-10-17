package main

type LDAPSync struct {
	lc *LDAPCaller
}

func NewLDAPSync() *LDAPSync {
	sync := new(LDAPSync)
	sync.lc = NewLDAPCaller().
		SetHost("e191.suse.de").
		SetPort(10389).
		SetUser("uid=admin,ou=system").
		SetPassword("admin")
	return sync
}

func (sync LDAPSync) Start() {
	sync.lc.Connect()
}

func (sync LDAPSync) Finish() {
	sync.lc.Disconnect()
}

// This will fetch users from the LDAP *and* Uyuni,
// intersect them and return only those that needed to be added.
func (sync *LDAPSync) GetUsersToSync() {

}

func (sync *LDAPSync) getExistingLDAPUsers() {

}
