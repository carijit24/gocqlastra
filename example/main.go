package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
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
	Parallel        string        `yaml:"parrallel" help:"number of times to call DB" default:"100" env:"PARALLEL"`
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

	var cluster *gocql.ClusterConfig
	if len(cfg.AstraBundle) > 0 {
		cluster, err = gocqlastra.NewClusterFromBundle("/Users/arijit.chakraborty/go/src/github.com/riptano/gocqlastra/example/scb.zip", "IhLuIxdmnJTGqythGPLhTPtc", "3CMhoXZi9H4DYJjBqr.q-ns7z1GfEg5ZkoBS2f9i4zdhGESbsGK8HEtu6QRD5yEJ4_WCiHT0YcXjTPGgs8GIhQG40kajWr6ZZXPfYI,MUzTJhZZQYiHrbWbPheEsPsps", cfg.AstraTimeout)
		if err != nil {
			cliCtx.Fatalf("unable to open bundle %s from file: %v", cfg.AstraBundle, err)
		}
	} else if len(cfg.AstraToken) > 0 {
		if len(cfg.AstraDatabaseID) == 0 {
			cliCtx.Fatalf("database ID is required when using a token")
		}
		cluster, err = gocqlastra.NewClusterFromURL(cfg.AstraApiURL, cfg.AstraDatabaseID, cfg.AstraToken, cfg.AstraTimeout)
		if err != nil {
			cliCtx.Fatalf("unable to load bundle %s from astra: %v", cfg.AstraBundle, err)
		}
	} else {
		cliCtx.Fatalf("must provide either bundle path or token")
	}
	size, _ := strconv.Atoi(cfg.Parallel)
	fmt.Printf("number of thread - %d", size)
	var waitGroup sync.WaitGroup
	waitGroup.Add(size)
	for i := 0; i < size; i++ {
		callAstra(i, err, cluster, &waitGroup)
	}
	waitGroup.Wait()
}

func callAstra(thread int, err error, cluster *gocql.ClusterConfig, wg *sync.WaitGroup) {
	start := time.Now()
	session, err := gocql.NewSession(*cluster)
	elapsed := time.Now().Sub(start)
	if err != nil {
		log.Fatalf("unable to connect session: %v", err)
	}

	iter := session.Query("SELECT release_version FROM system.local").Iter()

	var version string
	for iter.Scan(&version) {
		fmt.Printf("cassandra version - %s\n", version)
	}

	if err = iter.Close(); err != nil {
		log.Printf("error running query: %v", err)
	}

	fmt.Printf("Thread %d Connection process took %s\n", thread+1, elapsed)
	wg.Done()
}
