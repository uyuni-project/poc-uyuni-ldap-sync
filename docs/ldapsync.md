mgr-ldapsync(1) -- Utility to synchronise users between LDAP of your choice and the Uyuni server
====

## SYNOPSIS

`mgr-ldapsync` [option]

## DESCRIPTION

**mgr-ldapsync(1)** is a program that synchronises LDAP users with
the Uyuni server. It will add users from LDAP, update them if their
attributes changed or remove them from Uyuni.

## OPTIONS

* `-c`, `--config`=[value]:
  Specifiy an optional configuration file. Default configuration file
  is: `/etc/rhn/ldapsync.conf`

* `-o`, `--overview`:
  Overview or "dry run" mode. In this case the `mgr-ldapsync` will
  only show you what is going to change, but will not perform any real
  actions.

* `-s`, `--sync`:
  Perform an actual synchronisation.

* `-h`, `--help`:
  Shows help.

* `-v`, `--version`:
  Shows current version of the `mgr-ldapsync`.

## REQUIREMENTS

LDAP is very flexible and easy to customise. Because of this,
`mgr-ldapsync` allows users to adjust the configuration for the
existing LDAP schema. If not specified otherwise, default
configuration file is first looked up in `/etc/rhn/ldapsync.conf`.

There are certain requirements how to manage users in LDAP so they are
synchronised correctly with the Uyuni server:

1. There should be either at least one *group* of object class
   `groupOfNames` or at least one *role* of object class
   `organizationalRole`. In case of Active Directory, *group* object
   class should be called `group`.

2. Each group should have at least one attribute `member` with a valid
   DN of an actual user.

3. In case a groups are not used, a role has to have at least one
   attribute `roleOccupant` with the valid DN of an actual user.

4. There should be an accessible DN for all users in the LDAP
   database.

Each user should have the following **mandatory** attributes:

   - `uid`
   - `cn`
   - `name` or `givenName` (optional, if `cn` has name and second name)
   - `sn` (optional, if `cn` has name and second name)
   - `mail`

**IMPORTANT:** The `mgr-ldapsync` is not expected to work properly,
if the requirements above are not met.

## CONFIGURATION

Configuration file should be a valid YAML file.

It has two main sections at the root:

1. `directory`:
   Directory describes all the configuration, required by the LDAP server.

2. `rpc`:
   This section describes all the configuration, required by Uyuni server.

The **directory** section has the following attributes:

* `udn` (string):
  A user's DN for the directory to connect. Example: `uid=admin,ou=system`.

* `password` (string):
  LDAP authentication password for the `udn` above.

* `host` (string):
  Fully qualified domain name of the LDAP server.

* `port` (integer, optional):
  Port on which LDAP server is running. By default it is `389`.

* `allusers` (string):
  DN for all the users subtree. Example: `ou=users,dc=example,dc=com`.

The `directory` section has also the following directives:

1. `frozen` (map, mandatory). This is a list of Uyuni user IDs that
   `mgr-ldapsync` should _completely_ ignore. Any user ID, specified
   in this directive will always exclude any user in the Uyuni server
   from being affected by LDAP operations. This is usually the main
   "static" administrator account or an emergency login. This
   directive is mandatory and LDAP sync will refuse to work if you
   have no at least one frozen user with `org_admin` permissions
   assigned.

2. `groups` or `roles` map, at least one of them must be present.
   **Either** directive is **mandatory** to specify, in order to
   properly manage Uyuni roles. Both directives has the same structure
   a list of Uyuni roles, attached to a CN in the LDAP. See "examples"
   section below for more details:

```
   roles|groups:
     cn:
       - role
	   - role
	   - role
	   ...
```

The **rpc** section contains all the necessary information for XML-RPC
API of Uyuni server:

* `url` (string):
   URL of the XML-RPC API for the Uyuni server. It should contain a
   schema in order to specify if SSL/TLS is required.

* `checkssl` (boolean, optional, default `true`):
   A boolean true/false option to allow SSL connection to the URL
   without proper SSL certificate.

* `user` (string):
   Username of the Uyuni administrator with the `org_admin` or
   `satellite_admin` role. This is usually the user was created at the
   very first start.

* `password` (string):
   Password for the Uyuni administrator username.

## LIST OF UYUNI ROLES

Uyuni server supports the following roles:

* `org_admin`:
   Administrative role. Appears as "Organization Administrator". It
   can administer the entire Uyuni server across all the organisations.

* `satellite_admin`:
   Administrative role. Appears as "SUSE Manager Administrator". It
   can administer the entire Uyuni server, but only within the given
   organisation.

* `config_admin`:
   Appears as "Configuration Administrator" and gives user to
   configure individual system profiles, channels and certain
   configuration files.

* `channel_admin`:
   Appears as "Channel Administrator" and gives user ability to add,
   modify and delete channels.

* `system_group_admin`:
   Appears as "System Group Administrator" and gives user to access
   systems section.

* `activation_key_admin`:
   Appears as "Activation Key Administrator" and gives user the
   control over activation keys, subscriptions etc.

* `image_admin`:
   Appears as "Image Administrator" and is related to OS images
   administration to build them, store in the registry etc.

## EXAMPLES

**Example: map LDAP roles**

To map a `config_admin` and `channel_group_admin` Uyuni roles to a
`organizationalRole` object in the LDAP, do the following:

1. Create a role group in the LDAP with the class `organizationalRole`.

2. Add at least one user to that role that is supposed to have it with the
   attribute `roleOccupant`.

3. Create the following configuration in the `/etc/rhn/ldapsync.conf`
   file:

```
  directory:
    roles:
      cn=admins,ou=groups,dc=example,dc=com
        - config_admin
        - channel_group_admin
```

The configuration above will assign a `config_admin` and a
`channel_group_admin` Uyuni roles to the CN of a role group in LDAP.

**Example: map LDAP groups**

To map a `config_admin` and `channel_group_admin` Uyuni roles to a
group is very similar to `organizationalRole` scenario above, with few
differences:

1. Create a group in the LDAP with the class `groupOfNames` (POSIX) or
   `group` (Active Directory).

2. Add at least one user to that group with the attribute
   `member`. NOTE: User attributes should meet the requirements,
   described in "Requirements" section above.

3. Create the following configuration in the `/etc/rhn/ldapsync.conf`
   file:

```
   directory:
     groups:
       cn=admins,ou=groups,dc=example,dc=com
	     - config_admin
	     - channel_group_admin
```

The configuration above will assign a `config_admin` and a
`channel_group_admin` Uyuni roles to the CN of a group in LDAP.

For more information, look into the configuration file itself and
follow the examples there.

## DIAGNOSTICS

`mgr-ldapsync` returns zero on normal operation, non-zero otherwise.

## AUTHOR

Bo Maryniuk <bo@suse.de>

## SEE ALSO

pam
