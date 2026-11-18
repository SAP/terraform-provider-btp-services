package ctmsmodels

import "encoding/json"

// MarshalJSON encodes an explicit empty route description as null because CTMS
// rejects empty strings. A nil description remains omitted.
func (r RouteCreateRequest) MarshalJSON() ([]byte, error) {
	type request RouteCreateRequest
	if r.Description == nil || *r.Description != "" {
		return json.Marshal(request(r))
	}
	return json.Marshal(struct {
		request
		Description *string `json:"description"`
	}{request: request(r)})
}

// MarshalJSON preserves PATCH field presence while encoding a description
// reset as null. A nil description leaves the existing description untouched.
func (r RoutePatchRequest) MarshalJSON() ([]byte, error) {
	type request RoutePatchRequest
	if r.Description == nil || *r.Description != "" {
		return json.Marshal(request(r))
	}
	return json.Marshal(struct {
		request
		Description *string `json:"description"`
	}{request: request(r)})
}
