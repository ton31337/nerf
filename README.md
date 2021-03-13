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
export OAUTH_CLIENT_ID=<clientID>
export OAUTH_CLIENT_SECRET=<clientSecret>
export OAUTH_MASTER_TOKEN=<masterToken>
export OAUTH_ORGANIZATION=<githubOrganization>
export DNS_AUTODISCOVER_ZONE=example.com
make all # For multiple ARCHs
make # For Linux amd64 only
```

## Start a gRPC server

The server is needed to generate config.yml for Nebula. To start a server type:
```
./nerf -server -lighthouse 172.16.0.1:193.219.12.13
```

## Start a client

```
./nerf
```
