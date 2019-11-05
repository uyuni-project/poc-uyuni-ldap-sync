package ldapsync

import (
	"github.com/thoas/go-funk"
)

// CompareRoles of two users
func CompareRoles(a *UyuniUser, b *UyuniUser) bool {
	ra := a.GetRoles()
	rb := b.GetRoles()

	if len(ra) != len(rb) {
		return false
	}

	for _, r := range ra {
		if !funk.Contains(rb, r) {
			return false
		}
	}

	return true
}
