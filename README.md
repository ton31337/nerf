```
sequenceDiagram
    nerf->>nerf: Download Nebula to /opt/nebula/nebula
    nerf->>GitHub: Authorize
    GitHub-->>nerf: Authorized
    nerf->>nerf: Autodiscover all VPN endpoints through DNS SRV record
    nerf->>nerf: Probe all VPN endpoints via gRPC to find the fastest endpoint
    nerf->>nerf-server: Get generated config.yml for Nebula with appropriate IP and Groups
    nerf-server->>nerf-server: Generate config.yml
    nerf-server-->>nerf: Send config.yml
    nerf->>nebula: Start Nebula
```

![](/sequence.png)

## Compile

### Server

```
export OAUTH_MASTER_TOKEN=<masterToken>        # OAuth token with 'read:org' scope
export OAUTH_ORGANIZATION=<githubOrganization> # Github organization
make check                                     # Run linters, formatters, etc.
make server                                    # For Linux amd64 only
```

### Client

```
export OAUTH_CLIENT_ID=<clientID>              # OAuth application client id with 'user:email' scope
export OAUTH_CLIENT_SECRET=<clientSecret>      # OAuth application client secret
export DNS_AUTODISCOVER_ZONE=<dnsZone>         # DNS zone to discover VPN endpoints. E.g.: example.org
make check                                     # Run linters, formatters, etc.
make all                                       # For multiple ARCHs
make client                                    # For Linux amd64 only
```

## Start

### Server

```
Usage of ./nerf-server:
  -help
    	Print command line usage
  -lighthouse string
    	Set the lighthouse. E.g.: <NebulaIP>:<PublicIP>
  -log-level string
    	Set the logging level - values are 'debug', 'info', 'warn', and 'error' (default "info")
```

The server is needed to generate config.yml for Nebula. To start a server type:
```
./nerf-server -lighthouse 172.16.0.1:193.219.12.13
```

### Client

```
Usage of ./nerf:
  -help
    	Print command line usage
  -log-level string
    	Set the logging level - values are 'debug', 'info', 'warn', and 'error' (default "info")
  -redirect-all
    	Redirect all traffic through Nebula (default true)
```

The client MUST be with SUID bit set because of privileged user permissions to handle routes, DNS,
interfaces:

```
chown root ./nerf
chmod +s ./nerf
./nerf
```
