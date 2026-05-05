// Package events provides event definitions for the application.
package events

import "github.com/zoobzio/capitan"

// Startup signals for server lifecycle.
// These are direct capitan signals (not sum.Event) since they're
// operational events, not domain lifecycle events for consumers.
var (
	StartupDatabaseConnected = capitan.NewSignal("Janus.startup.database.connected", "Database connection established")
	StartupStorageConnected  = capitan.NewSignal("Janus.startup.storage.connected", "Object storage connection established")
	StartupServicesReady     = capitan.NewSignal("Janus.startup.services.ready", "All services registered")
	StartupOTELReady         = capitan.NewSignal("Janus.startup.otel.ready", "OpenTelemetry providers initialized")
	StartupApertureReady     = capitan.NewSignal("Janus.startup.aperture.ready", "Aperture observability bridge initialized")
	StartupServerListening   = capitan.NewSignal("Janus.startup.server.listening", "HTTP server listening")
	StartupFailed            = capitan.NewSignal("Janus.startup.failed", "Server startup failed")
)

// Startup field keys for direct emission.
var (
	StartupPortKey    = capitan.NewIntKey("port")
	StartupWorkersKey = capitan.NewIntKey("workers")
	StartupErrorKey   = capitan.NewErrorKey("error")
)
