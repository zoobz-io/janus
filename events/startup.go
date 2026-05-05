// Package events provides event definitions for the application.
package events

import "github.com/zoobzio/capitan"

// Startup signals for server lifecycle.
// These are direct capitan signals (not sum.Event) since they're
// operational events, not domain lifecycle events for consumers.
var (
	StartupDatabaseConnected = capitan.NewSignal("janus.startup.database.connected", "Database connection established")
	StartupStorageConnected  = capitan.NewSignal("janus.startup.storage.connected", "Object storage connection established")
	StartupServicesReady     = capitan.NewSignal("janus.startup.services.ready", "All services registered")
	StartupOTELReady         = capitan.NewSignal("janus.startup.otel.ready", "OpenTelemetry providers initialized")
	StartupApertureReady     = capitan.NewSignal("janus.startup.aperture.ready", "Aperture observability bridge initialized")
	StartupServerListening   = capitan.NewSignal("janus.startup.server.listening", "HTTP server listening")
	StartupFailed            = capitan.NewSignal("janus.startup.failed", "Server startup failed")
)

// Startup field keys for direct emission.
var (
	StartupPortKey    = capitan.NewIntKey("port")
	StartupWorkersKey = capitan.NewIntKey("workers")
	StartupErrorKey   = capitan.NewErrorKey("error")
)
