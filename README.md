<p><img src="https://img.shields.io/badge/EXPERIMENTAL-WIP-red" /></p>

# LDAP sync for Uyuni

LDAP sync tool is a reimplementation of
[`sw-ldap-user-sync`](https://github.com/uyuni-project/uyuni/blob/master/utils/sw-ldap-user-sync)
script.

# Motivation

Single Uyuni server has very little of LDAP synchronisation
ideas. This tool is covering most of the use cases of typical LDAP
setup.

# Build

Typical build procedure:

	git clone https://github.com/isbm/uyuni-ldap-sync.git
	cd uyuni-ldap-sync/cmd
	make deps
	make

# ToDO

- Add automatic PAM setup
- Unit tests
