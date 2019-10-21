package main

type UyuniUser struct {
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
	return u.uid != "" && u.email != "" && u.name != "" && u.secondname != ""
}
