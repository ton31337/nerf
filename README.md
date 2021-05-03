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

```
export OAUTH_CLIENT_ID=<clientID>              # OAuth application client id with 'user:email' scope
export OAUTH_CLIENT_SECRET=<clientSecret>      # OAuth application client secret
export OAUTH_MASTER_TOKEN=<masterToken>        # OAuth token with 'read:org' scope
export OAUTH_ORGANIZATION=<githubOrganization> # Github organization
export DNS_AUTODISCOVER_ZONE=<dnsZone>         # DNS zone to discover VPN endpoints. E.g.: example.org
make check                                     # Run linters, formatters, etc.
make all                                       # For multiple ARCHs (clients)
make                                           # For Linux amd64 only (client)
make client                                    # For Linux amd64 only (client)
make server                                    # For Linux amd64 only (server)
```

## Start a gRPC server

The server is needed to generate config.yml for Nebula. To start a server type:
```
./nerf-server -lighthouse 172.16.0.1:193.219.12.13
```

## Start a client

```
chown root ./nerf
chmod +s ./nerf
./nerf
```
