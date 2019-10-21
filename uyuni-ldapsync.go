package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
)

// RunSync is a main sync runner
func RunSync(ctx *cli.Context) {
	lc := NewLDAPSync(ctx.String("config")).Start()
	if ctx.Bool("show") || ctx.Bool("failed") {
		var users []UyuniUser
		var msg string
		if ctx.Bool("show") {
			msg = "Users in your LDAP that matches your criteria and should be synchronised:"
			users = lc.GetUsersToSync()
		} else {
			msg = "Users in your LDAP that will not be synchronised due to missing data or duplicates:"
			users = lc.GetFailedUsers()
		}

		if len(users) > 0 {
			fmt.Println(msg)
			for idx, user := range users {
				idx++
				fmt.Printf("  %d. %s (%s %s) at %s\n", idx, user.uid, user.name, user.secondname, user.email)
			}
			fmt.Println()
		} else {
			fmt.Println("No users found for this criteria")
		}
	} else if ctx.Bool("sync") {
		fmt.Println("Synchronising...")
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
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "./ldapsync.conf", // TODO: change that
			Usage: "Configuration file",
		},
		cli.BoolFlag{
			Name:   "show, w",
			Usage:  "Show users that will be synchronised",
			Hidden: false,
		},
		cli.BoolFlag{
			Name:   "failed, f",
			Usage:  "Show users that will not be synchronised",
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
