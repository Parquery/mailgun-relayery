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
	"github.com/Parquery/mailgun-relayery/mailgun-relayery/relay"
	"github.com/Parquery/mailgun-relayery/siger"
	ver "github.com/Parquery/mailgun-relayery/version"
)

var version = flag.Bool("version", false,
	"print the version to STDOUT and exit immediately")

var databaseDir = flag.String("database_dir", "",
	"Path to the LMDB containing channel and timestamps data")

var apiKeyPath = flag.String("api_key_path", "",
	"Path to where the MailGun API key is stored")

var mailgunAddress = flag.String("mailgun_address",
	relay.DefaultMailgunAddress, "MailGun server address")

var address = flag.String("address", ":8200",
	"address to be used for the relay server")

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
			logOut = log.New(ioutil.Discard,
				"[relay server]", log.Ldate|log.Ltime)
		} else {
			logOut = log.New(os.Stdout,
				"[relay server]", log.Ldate|log.Ltime)
		}

		logErr := log.New(os.Stderr,
			"[relay server]", log.Ldate|log.Ltime)

		if *databaseDir == "" {
			logErr.Println("-database_dir is mandatory")
			flag.PrintDefaults()
			return 1
		}

		if *apiKeyPath == "" {
			logErr.Println("-api_key_path is mandatory")
			flag.PrintDefaults()
			return 1
		}

		logOut.Println("Hi from relay server.")

		var err error

		////
		// Read the API key and create a MailgunData object
		////
		mailgunData := relay.MailgunData{Address: *mailgunAddress}
		buffer, err := ioutil.ReadFile(*apiKeyPath)
		if err != nil {
			logErr.Printf("failed to open the file containing the "+
				"API key %#v: %s\n", *apiKeyPath, err.Error())
			return 1
		}

		mailgunData.APIKey = string(buffer)

		////
		// Set up the database
		////
		var env *database.Env
		env, err = database.NewEnv(database.RelayAccess, *databaseDir)
		if err != nil {
			logErr.Printf("failed to open the "+
				"database %#v: %s\n", *databaseDir, err.Error())
			return 1
		}

		defer func() {
			closeErr := env.Close()
			if closeErr != nil {
				logErr.Printf("failed to close the read-only channel "+
					"database %#v: %s\n", *databaseDir, closeErr.Error())
			}
		}()

		srver := http.Server{Addr: *address,
			ReadTimeout:       60 * time.Second,
			ReadHeaderTimeout: 60 * time.Second}

		go func() {
			h := &relay.Handler{
				Env:         env,
				MailgunData: mailgunData,
				LogOut:      logOut,
				LogErr:      logErr}

			r := relay.SetupRouter(h)

			r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter,
				r *http.Request) {
				logErr.Printf("URL not handled: %#v for %#v",
					r.URL.String(), r.RemoteAddr)
				http.Error(w, "404 Not found", http.StatusNotFound)
			})

			routeTable, err := routeTableAsString(r)
			if err != nil {
				logErr.Printf("Failed to produce the route table for "+
					"the server: %s\n", err.Error())
				return
			}

			logOut.Printf("Relay server listening on %s\n", *address)
			logOut.Printf("Available routes:\n%s\n", routeTable)

			srver.Handler = r
			err = srver.ListenAndServe()
			if err != http.ErrServerClosed {
				logErr.Printf("Failed to listen and serve on %s: %s\n",
					*address, err.Error())
			}
			logOut.Println("Goodbye from the relay server.")
		}()

		siger.RegisterHandler()
		for !siger.Done() {
			time.Sleep(time.Second)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err = srver.Shutdown(ctx)
		if err != nil {
			logErr.Printf(
				"failed to shut down the relay server correctly: %s",
				err.Error())
			return 1
		}

		logOut.Println("Goodbye.")
		return 0
	}())
}
