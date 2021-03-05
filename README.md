```
sequenceDiagram
    nerf->>nerf: Download Nebula to /opt/nebula/nebula
    nerf->>GitHub: Authorize
    GitHub-->>nerf: Authorized
    nerf->>nerf-server: Get certificates for Nebula with appropriate IP and Groups
    nerf-server-->>nerf: Download ca.crt and generate <username.crt>, <username.key>
    nerf->>nerf: Generate /opt/nebula/config.yml
    nerf->>nebula: Start Nebula
```
