# gocql for Astra

This provides a custom host dialer that can be used to allow gocql to connect to DataStax Astra. The goal is to provide
native support for gocql on Astra.

This library was made possible by the `gocql.HostDialer` interface added here: https://github.com/gocql/gocql/pull/1629

## How to use it:

Using an Astra bundle:

```go
cluster, err := gocqlastra.NewClusterFromBundle("/path/to/your/bundle.zip", 
	"<username>", "<password>", 10 * time.Second)

if err != nil {
    panic("unable to load the bundle")
}

session, err := gocql.NewSession(*cluster)

// ...
```

Using an Astra token:

```go
cluster, err = gocqlastra.NewClusterFromURL(gocqlastra.ProdAstraURL, 
	"<astra-database-id>", "<astra-token>", 10 * time.Second)

if err != nil {
panic("unable to load the bundle")
}

session, err := gocql.NewSession(*cluster)

// ...
```

Also, look at the [example](example) for more information.

### Running the example:

```
cd example
go build

# Using a bundle
./example --astra-bundle /path/to/bundle.zip --username <username> --password <password>

# Using a token
./example --astra-token <astra-token> --astra-database-id <astra-database-id> 
```

### Issues

* Astra uses Stargate which doesn't current support the system table `system.peers_v2`. Also, the underlying storage 
  system for Astra is returns `4.0.0.6816` for the `release_version` column, but it doesn't actually support Apache
  Cassandra 4.0 (which includes `system.peers_v2`).  This is currently using a hack that replaces the `HostInfo` 
  version with a custom `gocql.HostFilter`. See [hack.go](hack.go) for more information.
* Need to verify that topology/status events correct update the driver when using Astra.