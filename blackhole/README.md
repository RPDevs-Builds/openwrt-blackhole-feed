# Blackhole Webserver for OpenWRT

A lightweight, high-performance Go-based utility designed to capture, log, and mirror HTTP requests. Optimized for OpenWRT routers (ARM64).

## Features
- **Catch-all Routing**: Responds to all requests.
- **Request Mirroring**: Replicates URL paths in a tracked directory.
- **Dedicated Content Serving**: Supports a separate content directory to serve "real" content (e.g., connectivity check files) when present.
- **JSON Logging**: Detailed metadata (Method, URL, IP, Headers) stored in a local log file.
- **Tracking Pixel**: Responds with a 1x1 transparent GIF for unknown/untracked paths.
- **Network Binding**: Fully configurable IP binding (supports multi-homed routers) and port selection.

## Installation

### Via Custom Package Feed (Recommended)
1. Add the feed to your OpenWRT router configuration by adding the following to `/etc/apk/repositories.d/customfeeds.list`:
   ```text
   https://iamrpdev.github.io/openwrt-blackhole-feed
   ```
2. Update and install:
   ```bash
   apk update
   apk add blackhole luci-app-blackhole
   ```

### Manual (Static Binary)
1. Download the ARM64 binary from the [blackhole-server](https://github.com/IamRPDev/blackhole-server) repository.
2. Transfer to your router and run: `./blackhole-server -ip <IP> -port <PORT> -root <MIRROR_DIR> -content <CONTENT_DIR>`

## Configuration & Management
The application is fully integrated with the OpenWRT system:
- **Web UI**: Access management via **Services -> Blackhole Webserver** in your router's LuCI web interface.
- **Config File**: Located at `/etc/config/blackhole` (UCI).
- **Service Management**: Managed via standard init scripts:
  ```bash
  /etc/init.d/blackhole start|stop|restart|enable
  ```

## Repository Architecture
- [**blackhole-server**](https://github.com/IamRPDev/blackhole-server): The core Go binary and OpenWRT initialization scripts.
- [**luci-app-blackhole**](https://github.com/IamRPDev/luci-app-blackhole): The client-side rendered Web UI for LuCI.
- [**openwrt-blackhole-feed**](https://github.com/IamRPDev/openwrt-blackhole-feed): The automated package repository index built via GitHub Actions.

## License
MIT
