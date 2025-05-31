# immich-share

Share Immich photos and albums securely on the internet without exposing your Immich instance.

## Installation

### Tailscale Mode

```bash
docker run -d -e MODE=tailscale -e TS_AUTHKEY=<YOUR_TAILSCALE_AUTHKEY> -e IMMICH_BASE_URL=<YOUR_IMMICH_BASE_URL> ghcr.io/j4n-e4t/immich-public-proxy:latest
```

### Local Mode

```bash
docker run -d -e MODE=local -e IMMICH_BASE_URL=<YOUR_IMMICH_BASE_URL> -p 8080:8080 ghcr.io/j4n-e4t/immich-public-proxy:latest
```

## Configuration

* `MODE`: Set to `tailscale` to use Tailscale Funnels or `local` to run the proxy locally.
* `TS_AUTHKEY`: Your Tailscale Auth Key. (Only required if `MODE=tailscale`.)   
* `TS_HOSTNAME`: Your Tailscale Hostname. (Only required if `MODE=tailscale`, default: `immich-share`).
* `IMMICH_BASE_URL`: The base URL of your Immich instance. (default: `http://immich:2283/`)