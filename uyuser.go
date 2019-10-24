package main

import (
	"strings"
)

type UyuniUser struct {
	Dn         string
	Uid        string
	Name       string
	Secondname string
	Email      string
	Err        error
	roles      []string
	new        bool
	outdated   bool
}

// Constructor
func NewUyuniUser() *UyuniUser {
	uu := new(UyuniUser)
	uu.roles = make([]string, 0)
	uu.new, uu.outdated = false, false

	return uu
}

// AddRole allows add distinct roles to the user
func (u *UyuniUser) AddRoles(roles ...string) {
	for _, role := range roles {
		role = strings.ToLower(role)
		for _, userRole := range u.roles {
			if userRole == role {
				goto Skip
			}
		}
		u.roles = append(u.roles, role)
	Skip:
	}
}

func (u *UyuniUser) FlushRoles() *UyuniUser {
	u.roles = nil
	return u
}

// GetRoles returns all roles, assigned to the user
func (u *UyuniUser) GetRoles() []string {
	return u.roles
}

// IsValid validates if the user data is compliant
// for the synchronisation
func (u *UyuniUser) IsValid() bool {
	return u.Uid != "" && u.Email != "" && u.Name != "" && u.Secondname != "" && u.Err == nil
}

// IsNew resturns a flag, indicating if that user
// is new to Uyuni (i.e. is not yet created)
func (u *UyuniUser) IsNew() bool {
	return u.new
}

// IsOutdated returns a flag, indicating that user's
// data has been changed in the LDAP and it needs to be updated.
func (u *UyuniUser) IsOutdated() bool {
	return u.outdated
}

// Clone user creates a new instance with the same data
func (u *UyuniUser) Clone() *UyuniUser {
	user := NewUyuniUser()
	user.Dn = u.Dn
	user.Uid = u.Uid
	user.Email = u.Email
	user.Name = u.Name
	user.Secondname = u.Secondname
	user.Err = u.Err
	user.new = u.new
	user.outdated = u.outdated
	user.AddRoles(u.GetRoles()...)

	return user
}
