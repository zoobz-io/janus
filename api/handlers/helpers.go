package handlers

import "github.com/zoobz-io/rocco"

// pathID extracts a string path parameter.
func pathID(params *rocco.Params, name string) string {
	return params.Path[name]
}
