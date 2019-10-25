package main

import (
	"strings"
)

type UyuniUser struct {
	Dn             string
	Uid            string
	Name           string
	Secondname     string
	Email          string
	Err            error
	roles          []string
	new            bool
	removed        bool
	outdated       bool
	roleschanged   bool
	accountchanged bool

	POSSIBLE_ROLES [7]string
}

// Constructor
func NewUyuniUser() *UyuniUser {
	uu := new(UyuniUser)
	uu.roles = make([]string, 0)
	uu.new, uu.outdated = false, false
	uu.POSSIBLE_ROLES = [7]string{
		"satellite_admin",
		"org_admin",
		"channel_admin",
		"config_admin",
		"system_group_admin",
		"activation_key_admin",
		"image_admin",
	}

	return uu
}

// AddRole allows add distinct roles to the user
func (u *UyuniUser) AddRoles(newRoles ...string) {
	for _, role := range newRoles {
		role = strings.ToLower(role)

		// Org admin gets everything
		if role == "org_admin" {
			u.FlushRoles()
			for _, sr := range u.POSSIBLE_ROLES {
				u.roles = append(u.roles, sr)
			}
			return
		}

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

// IsRemoved returns a flag, indicating that the user was removed
// from the LDAP (exists in Uyuni, does not exists in LDAP)
func (u *UyuniUser) IsRemoved() bool {
	return u.removed
}

// IsAccontDataChanged returns a flag, indicated that account data,
// such as email, name/second name etc has been changed.
func (u *UyuniUser) IsAccountDataChanged() bool {
	return u.accountchanged
}

// IsRolesChanged returns a flag, indicated that roles data has been changed.
func (u *UyuniUser) IsRolesChanged() bool {
	return u.roleschanged
}

// Clone creates a new user instance with the same data
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
	user.removed = u.removed
	user.accountchanged = u.accountchanged
	user.roleschanged = u.roleschanged
	user.AddRoles(u.GetRoles()...)

	return user
}
