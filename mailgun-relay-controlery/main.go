package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/Parquery/mailgun-relayery/database"
	"github.com/Parquery/mailgun-relayery/mailgun-relay-controlery/control"
	"github.com/Parquery/mailgun-relayery/siger"
	ver "github.com/Parquery/mailgun-relayery/version"
)

var version = flag.Bool("version", false,
	"print the version to STDOUT and exit immediately")

var databaseDir = flag.String("database_dir", "",
	"Path to the LMDB containing channel and timestamps data")

var address = flag.String("address", ":8300",
	"address to be used for the control server")

var quiet = flag.Bool("quiet", false,
	"If set, outputs as little messages as possible")

func routeTableAsString(r *mux.Router) (string, error) {
	var lines []string
	err := r.Walk(func(route *mux.Route, router *mux.Router,
		ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}

		methods, err := route.GetMethods()
		if err != nil {
			return err
		}
		t = strings.Join(methods, ", ") + " " + t
		lines = append(lines, t)
		return nil
	})

	if err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}

func main() {
	os.Exit(func() (retcode int) {
		flag.Parse()

		if *version {
			fmt.Println(ver.Latest)
			return 0
		}

		var logOut *log.Logger
		if *quiet {
			logOut = log.New(ioutil.Discard, "[control server]", log.Ldate|log.Ltime)
		} else {
			logOut = log.New(os.Stdout, "[control server]", log.Ldate|log.Ltime)
		}

		logErr := log.New(os.Stderr, "[control server]", log.Ldate|log.Ltime)

		if *databaseDir == "" {
			logErr.Println("-database_dir is mandatory")
			flag.PrintDefaults()
			return 1
		}

		logOut.Println("Hi from control server.")

		var err error

		////
		// Set up the database
		////
		var env *database.Env
		env, err = database.NewEnv(database.ControlAccess, *databaseDir)
		if err != nil {
			logErr.Printf("failed to open the database "+
				"%#v: %s\n", *databaseDir, err.Error())
			return 1
		}

		defer func() {
			closeErr := env.Close()
			if closeErr != nil {
				logErr.Printf("failed to close the database "+
					"%#v: %s\n", *databaseDir, closeErr.Error())
			}
		}()

		ctlSrver := http.Server{Addr: *address,
			ReadTimeout:       60 * time.Second,
			ReadHeaderTimeout: 60 * time.Second}

		go func() {
			h := &control.HandlerImpl{
				Env:    env,
				LogOut: logOut,
				LogErr: logErr}

			r := control.SetupRouter(h)

			routeTable, err := routeTableAsString(r)
			if err != nil {
				logErr.Printf("Failed to produce the route table for the "+
					"control server: %s\n", err.Error())
				return
			}

			logOut.Printf("Control server listening on %s\n", *address)
			logOut.Printf("Available routes:\n%s\n", routeTable)

			ctlSrver.Handler = r
			err = ctlSrver.ListenAndServe()
			if err != http.ErrServerClosed {
				logErr.Printf("Failed to listen and serve on %s: %s\n",
					*address, err.Error())
			}

			logOut.Println("Goodbye from the control server.")
		}()

		siger.RegisterHandler()
		for !siger.Done() {
			time.Sleep(time.Second)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err = ctlSrver.Shutdown(ctx)
		if err != nil {
			logErr.Printf(
				"failed to shut down the control server correctly: %s",
				err.Error())
			return 1
		}

		logOut.Println("Goodbye.")
		return 0
	}())
}
