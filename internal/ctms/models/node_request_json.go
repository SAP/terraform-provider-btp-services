package ctmsmodels

import "encoding/json"

// MarshalJSON encodes an explicit empty node description as null because CTMS
// rejects empty strings. A nil description remains omitted.
func (r NodeCreateRequest) MarshalJSON() ([]byte, error) {
	type request NodeCreateRequest
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
func (r NodeUpdateRequest) MarshalJSON() ([]byte, error) {
	type request NodeUpdateRequest
	if r.Description == nil || *r.Description != "" {
		return json.Marshal(request(r))
	}
	return json.Marshal(struct {
		request
		Description *string `json:"description"`
	}{request: request(r)})
}
