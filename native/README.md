# Draqun Native Client

A Native SDK WebView shell for the existing Go-served Draqun app. It
loads `http://localhost:3000/`; it does not duplicate or bundle the
web client.

## Run locally

From the repository root, start the Go server in one terminal:

```powershell
go run .
```

Then start the native client in another:

```powershell
bun run native:dev
```

The client defaults to the server's local port 3000. To connect to a
different server, update `server_url` and `server_origin` in
`src/main.zig`, and set the matching origin in `app.json` under
`security.navigation.allowed_origins`. Origins are intentionally
allowlisted rather than permitting arbitrary navigation.

On Android and iOS, `localhost` means the mobile device itself; it will
not reach a server running on a separate development computer.

## Build and targets

Zig 0.16.0 is required. The Native SDK supports desktop builds on
macOS, Windows, and Linux. Its iOS and Android targets use experimental
mobile hosts and require the platform-specific toolchains; iOS builds
require macOS/Xcode.

```powershell
bun run native:check
bun run native:build
bun run native:dev -- --target android
bun run native:dev -- --target ios
```
