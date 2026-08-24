# WindLock

WindLock coordinates mechanical, hydraulic and electrical isolation for offshore wind-turbine maintenance. It exposes status and command APIs for isolation, pitch, yaw, turning gear and safety incidents while persisting events to local files.

## Build and run

```sh
go build -mod=vendor ./...
go run -mod=vendor ./cmd/windlock -listen 127.0.0.1:21210 -data ./data
```

The service exposes `/healthz` and the operational resources under `/api`.
