# gocql for Astra

This provides a custom host dialer that can be used to allow gocql to connect to DataStax Astra. The goal is to provide
native support for gocql on Astra.

The is currently very close to working, but there is one issue preventing connection.

## How to use it:

```go
dialer, err = gocqlastra.NewDialerFromBundle("/path/to/your/bundle.zip", 10 * time.Second)

if err != nil {
	panic("unable to load the bundle")
}

cluster := gocql.NewCluster("127.0.0.1")

cluster.HostDialer = dialer
cluster.PoolConfig = gocql.PoolConfig{HostSelectionPolicy: gocql.RoundRobinHostPolicy()}
cluster.Authenticator = &gocql.PasswordAuthenticator{
Username: cfg.Username,
Password: cfg.Password,
}

session, err := gocql.NewSession(*cluster)

// ...
```

### Issues

* Astra uses Stargate which doesn't current support the system table `system.peers_v2`. Also, the underlying storage 
  system for Astra is returns `4.0.0.6816` for the `release_version` column, but it doesn't actually support Apache
  Cassandra 4.0 (which includes `system.peers_v2`). When connecting you'll see the following error.
  
  ```
  2022/08/11 12:29:38 unable to connect session: gocql: unable to create session: unable to fetch peer host info: table system.peers_v2 does not exist
  ```