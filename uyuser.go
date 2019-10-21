package main

type UyuniUser struct {
	Uid        string
	Name       string
	Secondname string
	Email      string
	Err        error
}

func NewUyuniUser() *UyuniUser {
	return new(UyuniUser)
}

// IsValid validates if the user data is compliant to the synchronised
func (u *UyuniUser) IsValid() bool {
	return u.Uid != "" && u.Email != "" && u.Name != "" && u.Secondname != "" && u.Err == nil
}
