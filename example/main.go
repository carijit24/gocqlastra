package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/gocql/gocql"
	gocqlastra "gocql-astra"
)

type runConfig struct {
	AstraBundle     string        `yaml:"astra-bundle" help:"Path to secure connect bundle for an Astra database. Requires '--username' and '--password'. Ignored if using the token or contact points option." short:"b" env:"ASTRA_BUNDLE"`
	AstraToken      string        `yaml:"astra-token" help:"Token used to authenticate to an Astra database. Requires '--astra-database-id'. Ignored if using the bundle path or contact points option." short:"t" env:"ASTRA_TOKEN"`
	AstraDatabaseID string        `yaml:"astra-database-id" help:"Database ID of the Astra database. Requires '--astra-token'" short:"i" env:"ASTRA_DATABASE_ID"`
	AstraApiURL     string        `yaml:"astra-api-url" help:"URL for the Astra API" default:"https://api.astra.datastax.com" env:"ASTRA_API_URL"`
	AstraTimeout    time.Duration `yaml:"astra-timeout" help:"Timeout for contacting Astra when retrieving the bundle and metadata" default:"10s" env:"ASTRA_TIMEOUT"`
	Username        string        `yaml:"username" help:"Username to use for authentication" short:"u" env:"USERNAME"`
	Password        string        `yaml:"password" help:"Password to use for authentication" short:"p" env:"PASSWORD"`
}

func main() {
	var cfg runConfig

	parser, err := kong.New(&cfg)
	if err != nil {
		panic(err)
	}

	var cliCtx *kong.Context
	if cliCtx, err = parser.Parse(os.Args[1:]); err != nil {
		parser.Fatalf("error parsing flags: %v", err)
	}

	var dialer *gocqlastra.Dialer
	if len(cfg.AstraBundle) > 0 {
		dialer, err = gocqlastra.NewDialerFromBundle(cfg.AstraBundle, cfg.AstraTimeout)
		if err != nil {
			cliCtx.Fatalf("unable to open bundle %s from file: %v", cfg.AstraBundle, err)
		}
	} else if len(cfg.AstraToken) > 0 {
		if len(cfg.AstraDatabaseID) == 0 {
			cliCtx.Fatalf("database ID is required when using a token")
		}
		dialer, err = gocqlastra.NewDialerFromURL(cfg.AstraApiURL, cfg.AstraDatabaseID, cfg.AstraToken, cfg.AstraTimeout)

		if err != nil {
			cliCtx.Fatalf("unable to load bundle %s from astra: %v", cfg.AstraBundle, err)
		}
		cfg.Username = "token"
		cfg.Password = cfg.AstraToken
	} else {
		cliCtx.Fatalf("must provide either bundle path or token")
	}

	cluster := gocql.NewCluster("127.0.0.1")

	cluster.ConnectTimeout = 20 * time.Second
	cluster.ProtoVersion = 4
	cluster.Timeout = 20 * time.Second

	cluster.HostDialer = dialer
	cluster.PoolConfig = gocql.PoolConfig{HostSelectionPolicy: gocql.RoundRobinHostPolicy()}
	cluster.Authenticator = &gocql.PasswordAuthenticator{
		Username: cfg.Username,
		Password: cfg.Password,
	}

	start := time.Now()
	session, err := gocql.NewSession(*cluster)
	elapsed := time.Now().Sub(start)
	if err != nil {
		log.Fatalf("unable to connect session: %v", err)
	}

	iter := session.Query("SELECT version FROM system.local").Iter()

	var version string
	for iter.Scan(&version) {
		fmt.Println(version)
	}

	if err = iter.Close(); err != nil {
		log.Printf("error running query: %v", err)
	}

	fmt.Printf("Connection process took %s\n", elapsed)
}
