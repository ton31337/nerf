```
sequenceDiagram
    nerf->>nerf: Download Nebula to /opt/nebula/nebula
    nerf->>GitHub: Authorize
    GitHub-->>nerf: Authorized
    nerf->>nerf-server: Get generated config.yml for Nebula with appropriate IP and Groups
    nerf-server->>nerf-server: Generate config.yml
    nerf-server-->>nerf: Send config.yml
    nerf->>nebula: Start Nebula
```

![](/sequence.png)

#### Start a gRPC server

The server is needed to generate config.yml for Nebula. To start a server type:
```
OAUTH_MASTER_TOKEN=<token> ./nerf -server -lighthouse 172.16.0.1:193.219.12.13
```
