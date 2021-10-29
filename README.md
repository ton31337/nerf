```
sequenceDiagram
    nerf (GUI)->>nerf (GUI): Download Nebula to /opt/nebula/nebula
    nerf (GUI)->>GitHub: Authorize
    GitHub-->>nerf (GUI): Authorized
    nerf (GUI)->>nerf-api: Connect (gRPC over UNIX socket)
    nerf-api->>nerf-server: Connect (gRPC over TCP socket)
    nerf-api->>nerf-api: Autodiscover all VPN endpoints through DNS SRV record
    nerf-api->>nerf-api: Probe all VPN endpoints via gRPC to find the fastest endpoint
    nerf-api->>nerf-server: Get generated config.yml for Nebula with appropriate IP and Groups
    nerf-server->>nerf-server: Generate config.yml
    nerf-server-->>nerf-api: Send config.yml
    nerf-api->>nebula: Start Nebul
    nerf (GUI)->>nerf-api: Disconnect
    nerf-api->>nerf-server: Disconnect
    nerf (GUI)->>nerf-api: Quit
    nerf-api->>nerf-server: Disconnect
```

![](/doc/img/sequence.png)

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
sudo apt install libappindicator3-dev gir1.2-appindicator3-0.1
export OAUTH_CLIENT_ID=<clientID>              # OAuth application client id with 'user:email' scope
export OAUTH_CLIENT_SECRET=<clientSecret>      # OAuth application client secret
export DNS_AUTODISCOVER_ZONE=<dnsZone>         # DNS zone to discover VPN endpoints. E.g.: example.org
make check                                     # Run linters, formatters, etc.
make darwin-client                             # For MacOS
make linux-client                              # For Linux
make deb                                       # Build Ubuntu/Debian .deb package
```

## Start

### Server

```
Usage of ./nerf-server:
  -gaidysUrl string
    	Set URL for Gaidys service (IPAM)
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

#### API for GUI

This is the gRPC API for GUI to talk

```
sudo chown root ./nerf-api
sudo chmod +s ./nerf-api
./nerf-api -log-level debug
```

#### Start GUI

```
./nerf
```

## Building application for OSX

To build an application for OSX it's recommended to use [Packages](http://s.sudre.free.fr/Software/Packages/about.html) application (easiest).

After running `make darwin-client`, the binary is copied to `osx/Nerf.app/Contents/MacOS/nerf`, and the whole structure is created for the app.

In payload section right click on `Applications` and `Add files`. Add `osx/Nerf.app`.

Also add `osx/LaunchDaemons/com.ton31337.nerf.app.launchd.plist` under `/Library/LaunchDaemons`.

![](/doc/img/payload1.png)

Below, under `/Library`, create a new directory `Services/Nerf` and put `./nerf-api`.

It's IMPORTANT to set `SetUID` bit for the owner (root:wheel).

![](/doc/img/payload2.png)

Put pre-install and post-install scripts located in `osx/scripts` accordingly.

![](/doc/img/scripts.png)

### Installing a package

```
./osx/scripts/install.sh
```

### Uninstall a package

```
./osx/scripts/uninstall.sh
```
