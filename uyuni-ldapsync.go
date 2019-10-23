package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
)

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
	fmt.Println(title)
	if len(users) > 0 {
		for idx, user := range users {
			idx++
			fmt.Printf("  %d. %s (%s %s) at %s\n", idx, user.Uid, user.Name, user.Secondname, user.Email)
		}
		fmt.Println()
	} else {
		fmt.Println("  No users found for this criteria")
	}
	fmt.Println()
}

// RunSync is a main sync runner
func RunSync(ctx *cli.Context) {
	lc := NewSyncApp(ctx)
	if ctx.Bool("overview") {
		PrintUsers("New users:", lc.GetLDAPSync().GetNewUsers())
		PrintUsers("Outdated users:", lc.GetLDAPSync().GetOutdatedUsers())
	} else if ctx.Bool("sync") {
		fmt.Println("Synchronising...")
		for _, user := range lc.GetLDAPSync().SyncUsers() {
			fmt.Printf("Adding new user as \"%s\" has been failed: %s\n", user.Uid, user.Err.Error())
		}
		for _, user := range lc.GetLDAPSync().SyncUserRoles() {
			fmt.Printf("Updating roles for \"%s\" has been failed: %s\n", user.Uid, user.Err.Error())
		}
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
			Value: "./ldapsync.conf", // TODO: change that
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
