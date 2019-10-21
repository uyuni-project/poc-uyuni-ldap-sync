package main

import (
	"strings"
)

type UyuniUser struct {
	Uid        string
	Name       string
	Secondname string
	Email      string
	Err        error
	roles      []string
}

func NewUyuniUser() *UyuniUser {
	uu := new(UyuniUser)
	uu.roles = make([]string, 0)

	return uu
}

func (u *UyuniUser) AddRole(role string) {
	role = strings.ToLower(role)
	for _, userRole := range u.roles {
		if userRole == role {
			return
		}
	}

	u.roles = append(u.roles, role)
}

// IsValid validates if the user data is compliant to the synchronised
func (u *UyuniUser) IsValid() bool {
	return u.Uid != "" && u.Email != "" && u.Name != "" && u.Secondname != "" && u.Err == nil
}
