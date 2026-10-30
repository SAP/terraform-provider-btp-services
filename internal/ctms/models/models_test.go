package ctmsmodels

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatchPresence(t *testing.T) {
	for _, tc := range []struct {
		name string
		body any
		want string
	}{
		{"omitted node fields", NodeUpdateRequest{}, `{}`},
		{"explicit node resets", NodeUpdateRequest{
			Description:          new(""),
			UploadAllowed:        new(false),
			NotificationEnabled:  new(false),
			ImportDisabled:       new(false),
			ImportDisabledReason: new(""),
			Virtual:              new(false),
			AutoImport:           new(false),
			Targets:              new([]DeployTarget{}),
		}, `{
  "description": "",
  "uploadAllowed": false,
  "notificationEnabled": false,
  "importDisabled": false,
  "importDisabledReason": "",
  "virtual": false,
  "autoImport": false,
  "targets": []
}`},
		{"omitted route fields", RoutePatchRequest{}, `{}`},
		{
			name: "explicit route values",
			body: RoutePatchRequest{
				Description:  new(""),
				SourceNodeId: new(int64(0)),
			},
			want: `{"description":"","sourceNodeId":0}`,
		},
		{
			name: "create explicit defaults",
			body: NodeCreateRequest{
				Name:          "node",
				UploadAllowed: new(false),
				ForwardMode:   new(AUTO_ForwardMode),
			},
			want: `{"name":"node","uploadAllowed":false,"forwardMode":"AUTO"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.body)
			require.NoError(t, err)
			require.JSONEq(t, tc.want, string(body))
		})
	}
}

func TestNodeResponseInheritedFieldsAndPresence(t *testing.T) {
	const fixture = `{
  "id": 9007199254740993,
  "name": "node",
  "jobId": 123,
  "description": "description",
  "uploadAllowed": false,
  "forwardMode": "AUTO",
  "controlledBy": "CHARM",
  "notificationEnabled": false,
  "importDisabled": false,
  "importDisabledReason": "",
  "virtual": false,
  "autoImport": false,
  "targets": [
    {
      "contentType": "MTA",
      "destinationName": "target",
      "importOptions": {
        "strategy": "blue-green"
      }
    }
  ]
}`
	var node NodeResponseTarget
	require.NoError(t, json.Unmarshal([]byte(fixture), &node))
	require.Equal(t, int64(9007199254740993), *node.Id)
	roundtrip, err := json.Marshal(node)
	require.NoError(t, err)
	require.JSONEq(t, fixture, string(roundtrip))
	var missing NodeResponseTarget
	require.NoError(t, json.Unmarshal([]byte(`{"id":1}`), &missing))
	require.Nil(t, missing.UploadAllowed)
	require.Nil(t, missing.Description)
	require.Nil(t, missing.Targets)
	var explicit NodeResponseTarget
	require.NoError(t, json.Unmarshal([]byte(`{"id":1,"uploadAllowed":false,"description":"","targets":[]}`), &explicit))
	require.NotNil(t, explicit.UploadAllowed)
	require.False(t, *explicit.UploadAllowed)
	require.NotNil(t, explicit.Description)
	require.NotNil(t, explicit.Targets)
	require.Empty(t, *explicit.Targets)
}
