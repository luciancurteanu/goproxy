package main

import (
	"database/sql"
	"flag"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/luciancurteanu/goproxy/middleware"
	"github.com/luciancurteanu/goproxy/routes"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"golang.org/x/net/proxy"
)

var (
	// database connection
	db *sql.DB
	// http clients
	clients routes.Clients
	// cleanup actions
	cleanup struct {
		lock       sync.Mutex
		called     bool
		registered []func()
		register   func(func())
		cleanup    func()
	}
)

func init() {

	var (
		config string
		err    error
	)

	// Make sure we have somewhere to store cleanup functions
	cleanup.register = func(f func()) {
		cleanup.lock.Lock()
		defer cleanup.lock.Unlock()
		cleanup.registered = append(cleanup.registered, f)
	}

	cleanup.cleanup = func() {
		// Check if clean up has already been performed
		if cleanup.called {
			return
		}
		cleanup.lock.Lock()
		defer cleanup.lock.Unlock()
		log.Debug("cleaning up")
		for _, f := range cleanup.registered {
			f()
		}
		// Make sure functions don't get called twice
		cleanup.called = true
	}

	// Allow configuration file to be specified on the command line
	flag.StringVar(&config, "config", "", "configuration file to use")
	flag.Parse()

	// Set up viper
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	if config != "" {
		viper.SetConfigFile(config)
	}

	// Read in config
	if err = viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read reading configuration file; %v", err)
	}

	// Make sure server address is specified
	if !viper.IsSet("server.address") {
		log.Fatal("invalid configuration; server.address must be specified")
	}

	/* DEFAULTS */
	if viper.GetString("server.mode") == "debug" {
		viper.SetDefault("logging.level", "DEBUG")
	} else {
		viper.SetDefault("logging.level", "INFO")
	}
	viper.SetDefault("logging.stderr", true)
	viper.SetDefault("clients.timeout", 30)
	viper.SetDefault("clients.proxy.protocol", "tcp")
	viper.SetDefault("clients.proxy.test", "https://icanhazip.com")
	viper.SetDefault("clients.proxy.test", "https://icanhazip.com")

	// Synchronization
	wg := sync.WaitGroup{}

	/* LOGGING */

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Debug("configuring logging")

		// If we log.Fatal or log.Exit, make sure cleanup is performed
		log.RegisterExitHandler(cleanup.cleanup)

		// Configure logging level
		level, err := log.ParseLevel(viper.GetString("logging.level"))
		if err != nil {
			log.Fatalf("%v", err)
		}
		log.SetLevel(level)

		// Configure logging / gin output
		if logpath := viper.GetString("logging.filepath"); logpath != "" {

			var wr io.Writer

			if logdir := path.Dir(logpath); logdir != "" {
				if err := os.MkdirAll(logdir, 0750); err != nil {
					log.Fatalf("failed to create directory structure for logs; %v", err)
				}
			}

			f, err := os.OpenFile(logpath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatalf("failed to open logging.filepath %q for writing; %v", logpath, err)
			}

			// Register cleanup function
			cleanup.register(func() {
				log.Debug("closing log file")
				f.Close()
			})

			// Should we log to standard error too?
			if viper.GetBool("logging.stderr") {
				wr = io.MultiWriter(f, os.Stderr)
			} else {
				wr = f
			}

			gin.DefaultWriter = wr
			gin.DefaultErrorWriter = wr
			log.SetOutput(wr)
		}
	}()

	/* DATABASE */

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Debug("configuring database")

		// Configure database connection
		dsn := mysql.NewConfig()
		dsn.User = viper.GetString("database.username")
		dsn.Passwd = viper.GetString("database.password")
		dsn.Addr = viper.GetString("database.host")
		dsn.DBName = viper.GetString("database.database")
		dsn.Net = "tcp"

		// Open database connection
		db, err = sql.Open("mysql", dsn.FormatDSN())
		if err != nil {
			log.Fatalf("failed to open connection to database; %v", err)
		}

		// Register cleanup
		cleanup.register(func() {
			log.Debug("closing database connection")
			db.Close()
			db = nil
		})
	}()

	/* HTTP CLIENTS */

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Debug("configuring http clients")

		// Unproxied "normal" client
		clients.Normal = &http.Client{
			Timeout: time.Duration(viper.GetInt("clients.timeout")) * time.Second,
		}

		// Check if proxy should be configured
		if !viper.IsSet("clients.proxy.address") {
			log.Warn("clients.proxy.address not specified, not creating a proxied client")
			clients.Proxy = clients.Normal
		} else {

			log.Debug("configuring proxied http client")
			prot := viper.GetString("clients.proxy.protocol")
			addr := viper.GetString("clients.proxy.address")

			// Create dialler
			dialer, err := proxy.SOCKS5(prot, addr, nil, proxy.Direct)
			if err != nil {
				log.Fatalf("failed to create proxy dialler; %v", err)
			}
			clients.Proxy = &http.Client{
				Transport: &http.Transport{
					Dial: dialer.Dial,
				},
				Timeout: clients.Normal.Timeout,
			}

			// Make sure proxy works
			test := viper.GetString("clients.proxy.test")
			log.Debugf("testing proxy against %s", test)
			if _, err = clients.Proxy.Get(test); err != nil {
				log.Fatalf("invalid proxy configuration; %v", err)
			}
		}

		// Should cookies persist? Default is false
		if viper.GetBool("clients.cookies.persist") {
			log.Warn("persisting client cookies between requests")
		} else {
			clients.Proxy.Jar = nil
			clients.Normal.Jar = nil
		}

	}()

	// Configure gin
	gin.SetMode(viper.GetString("server.mode"))

	wg.Wait()
	log.Debug("finished initializing successfully")
}

func main() {
	defer cleanup.cleanup()
	r := gin.Default()

	/* MIDDLEWARE */

	r.Use(middleware.Headers(viper.GetStringMapString("server.headers")))
	if viper.GetBool("server.whitelist.enable") {
		r.Use(middleware.Whitelist(viper.GetStringSlice("server.whitelist.addresses")))
	}

	/* ROUTES */

	// Custom request
	r.GET("/custom", routes.Custom(clients))

	// Simple GET
	r.GET("/get", routes.Get(clients))

	// IP Address
	r.GET("/ip", routes.IPAddress(clients))

	r.Run(viper.GetString("server.address"))
}
