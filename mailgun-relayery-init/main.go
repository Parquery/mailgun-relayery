package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Parquery/mailgun-relayery/database"
	ver "github.com/Parquery/mailgun-relayery/version"
)

var version = flag.Bool("version", false,
	"print the version to STDOUT and exit immediately")

var databaseDir = flag.String("database_dir", "",
	"Path to the directory where the database should be initialized")

func main() {
	os.Exit(func() (retcode int) {
		flag.Parse()

		if *version {
			fmt.Println(ver.Latest)
			return 0
		}

		logOut := log.New(os.Stdout, "[init] ", log.Ldate|log.Ltime)
		logErr := log.New(os.Stderr, "[init] ", log.Ldate|log.Ltime)

		if *databaseDir == "" {
			logErr.Println("-database_dir is mandatory")
			flag.PrintDefaults()
			return 1
		}

		var err error

		// Set up the database
		err = database.Initialize(database.ControlAccess, *databaseDir)
		if err != nil {
			logErr.Printf("failed to initialize the "+
				"database %#v: %s\n", *databaseDir, err.Error())
			return 1
		}

		logOut.Println("Database succesfully initialized.")
		return 0
	}())
}
