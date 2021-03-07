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
