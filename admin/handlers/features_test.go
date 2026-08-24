package handlers

import "testing"

func TestFeatureEndpointScopes(t *testing.T) {
	requireScope(t, listFeatures, "directory:read")
	requireScope(t, addFeature, "applications:manage")
	requireScope(t, removeFeature, "applications:manage")
}
