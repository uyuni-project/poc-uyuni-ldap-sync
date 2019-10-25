package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"strings"
)

var log = logrus.New()

type SyncApp struct {
	ldapSync   *LDAPSync
	cliContext *cli.Context
}

func NewSyncApp(ctx *cli.Context) *SyncApp {
	sa := new(SyncApp)
	sa.cliContext = ctx

	return sa
}

func (sa *SyncApp) GetLDAPSync() *LDAPSync {
	if sa.ldapSync == nil {
		sa.ldapSync = NewLDAPSync(sa.cliContext.String("config")).Start()
	}
	return sa.ldapSync
}

func (sa *SyncApp) Finish() {
	if sa.ldapSync != nil {
		sa.ldapSync.Finish()
	}
}

// Print users
func PrintUsers(title string, users []*UyuniUser) {
	if len(users) > 0 {
		fmt.Printf("%s:\n", title)
		for idx, user := range users {
			idx++
			fmt.Printf("  %d. %s (%s %s) at %s\n", idx, user.Uid, user.Name, user.Secondname, user.Email)
		}
		fmt.Println()
	} else {
		fmt.Printf("No %s has been found for this criteria\n", strings.ToLower(title))
	}
}

// RunSync is a main sync runner
func RunSync(ctx *cli.Context) {
	lc := NewSyncApp(ctx)
	if ctx.Bool("overview") {
		// TODO: add reporting facility instead of this
		fmt.Println("Ignored users:")
		for idx, uid := range lc.GetLDAPSync().cr.Config().Directory.Frozen {
			idx++
			fmt.Printf("  %d. %s\n", idx, uid)
		}
		fmt.Println()

		PrintUsers("New users", lc.GetLDAPSync().GetNewUsers())
		PrintUsers("Outdated users", lc.GetLDAPSync().GetOutdatedUsers())
		PrintUsers("Removed users", lc.GetLDAPSync().GetDeletedUsers())
	} else if ctx.Bool("sync") {
		lc.GetLDAPSync().SyncUsers()
	} else {
		cli.ShowAppHelpAndExit(ctx, 1)
	}
	lc.Finish()
}

// Main function
func main() {
	app := cli.NewApp()
	app.Name = "LDAP Sync"
	app.Usage = "Synchronise users between Uyuni/SUSE Manager and LDAP of your choice"
	app.Action = RunSync
	app.Version = "0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "/etc/rhn/ldapsync.conf", // TODO: change that
			Usage: "Configuration file",
		},
		cli.BoolFlag{
			Name:   "overview, o",
			Usage:  "Users overview",
			Hidden: false,
		},
		cli.BoolFlag{
			Name:   "sync, s",
			Usage:  "Synchronise users",
			Hidden: false,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
