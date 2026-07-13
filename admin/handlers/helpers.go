package handlers

import (
	"strconv"

	"github.com/zoobz-io/rocco"

	"github.com/zoobz-io/janus/models"
)

// pathID extracts a string path parameter.
func pathID(params *rocco.Params, name string) string {
	return params.Path[name]
}

// offsetPage builds an offset-based page from the "limit" and "offset" query
// parameters. Missing or invalid values fall back to store defaults.
func offsetPage(params *rocco.Params) models.OffsetPage {
	limit, _ := strconv.Atoi(params.Query["limit"])
	offset, _ := strconv.Atoi(params.Query["offset"])
	if offset < 0 {
		offset = 0
	}
	return models.OffsetPage{Limit: limit, Offset: offset}
}
