package main

type UyuniUser struct {
	id         string
	uid        string
	name       string
	secondname string
	email      string
}

func NewUyuniUser() *UyuniUser {
	return new(UyuniUser)
}

// IsValid validates if the user data is compliant to the synchronised
func (u *UyuniUser) IsValid() bool {
	return u.id != "" && u.email != "" && u.name != "" && u.secondname != ""
}
