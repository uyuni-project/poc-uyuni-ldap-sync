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
}

// Constructor
func NewUyuniUser() *UyuniUser {
	uu := new(UyuniUser)
	uu.roles = make([]string, 0)

	return uu
}

// AddRole allows add distinct roles to the user
func (u *UyuniUser) AddRole(role string) {
	role = strings.ToLower(role)
	for _, userRole := range u.roles {
		if userRole == role {
			return
		}
	}

	u.roles = append(u.roles, role)
}

// GetRoles returns all roles, assigned to the user
func (u *UyuniUser) GetRoles() []string {
	return u.roles
}

// IsValid validates if the user data is compliant to the synchronised
func (u *UyuniUser) IsValid() bool {
	return u.Uid != "" && u.Email != "" && u.Name != "" && u.Secondname != "" && u.Err == nil
}
