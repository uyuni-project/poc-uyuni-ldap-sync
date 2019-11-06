package main

import (
	"fmt"
	"github.com/isbm/uyuni-ldap-sync"
	"github.com/sirupsen/logrus"
	"github.com/t-tomalak/logrus-easy-formatter"
	"github.com/urfave/cli"
	"os"
	"strings"
	"time"
)

var log *logrus.Logger

type SyncApp struct {
	ldapSync   *ldapsync.LDAPSync
	cliContext *cli.Context
	output     *os.File
}

func NewSyncApp(ctx *cli.Context) *SyncApp {
	sa := new(SyncApp)
	sa.output = os.Stdout
	sa.cliContext = ctx

	return sa
}

// SetupLogger is used to setup all the preferences for the logging
func (sa *SyncApp) setupLogger(cr *ldapsync.ConfigReader) {
	log = ldapsync.Log
	if !sa.cliContext.Bool("verbose") {
		fmtr := new(easy.Formatter)
		fmtr.TimestampFormat = "2006-01-02 15:04:05"
		fmtr.LogFormat = "[%lvl%]: %time% - %msg%\n"
		log.SetFormatter(fmtr)
		fout, err := os.OpenFile(cr.Config().Common.Logpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
		if err == nil {
			sa.output = fout
		} else {
			log.Info("Failed to log to file, using default stderr")
		}
	} else {
		fmtr := new(logrus.TextFormatter)
		fmtr.TimestampFormat = time.RFC822Z
		fmtr.FullTimestamp = true
		fmtr.DisableLevelTruncation = true
		fmtr.DisableColors = false
		//fmtr.ForceQuote = false
		log.SetFormatter(fmtr)
	}

	log.SetOutput(sa.output)
	log.SetLevel(logrus.TraceLevel)
}

// GetLDAPSync returns a pointer to the LDAPSync object instance.
// Creates new, if not yet initialised.
func (sa *SyncApp) GetLDAPSync() *ldapsync.LDAPSync {
	if sa.ldapSync == nil {
		sa.ldapSync = ldapsync.NewLDAPSync(sa.cliContext.String("config"))
		sa.setupLogger(sa.ldapSync.ConfigReader())
		sa.ldapSync.Start()
	}

	return sa.ldapSync
}

// Finish the sync and close all connections
func (sa *SyncApp) Finish() {
	if sa.ldapSync != nil {
		sa.ldapSync.Finish()
	}
}

// Print users
func PrintUsers(title string, users []*ldapsync.UyuniUser) {
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
		for idx, uid := range lc.GetLDAPSync().ConfigReader().Config().Directory.Frozen {
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
		cli.BoolFlag{
			Name:   "verbose, d",
			Usage:  "Verbose (debug) mode",
			Hidden: false,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
