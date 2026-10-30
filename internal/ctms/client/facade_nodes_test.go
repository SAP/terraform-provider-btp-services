package ctmsclient

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

// Literal API JSON keeps the response contract independent of model serialization.
const nodeResponseJSON = `{"id":9007199254740993,"name":"node","uploadAllowed":false,"description":"","targets":[]}`

func TestNodesFacade_Create(t *testing.T) {
	for _, namedUser := range []string{"", "terraform-automation"} {
		t.Run("named_user="+namedUser, func(t *testing.T) {
			t.Run("constructs the request correctly and handles the response", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodPost, "/v2/nodes")
					assert.Empty(t, r.URL.RawQuery)
					assert.Equal(t, namedUser, r.Header.Get("tms-named-user"))
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					var body json.RawMessage
					assertRequestBody(t, r, &body)
					assert.JSONEq(t, `{"name":"node"}`, string(body))

					w.WriteHeader(http.StatusCreated)
					_, _ = w.Write([]byte(nodeResponseJSON))
				})
				got, err := uut.Nodes.Create(context.Background(), ctmsmodels.NodeCreateRequest{Name: "node"})
				require.True(t, srvCalled)
				require.NoError(t, err)
				assertResponseBody(t, got, nodeResponseJSON)
			})

			t.Run("returns NotFoundError on 404", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodPost, "/v2/nodes")
					assert.Empty(t, r.URL.RawQuery)
					assert.Equal(t, namedUser, r.Header.Get("tms-named-user"))
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					var body json.RawMessage
					assertRequestBody(t, r, &body)
					assert.JSONEq(t, `{"name":"node"}`, string(body))

					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"errorType":"NotFound","message":"gone"}`))
				})
				got, err := uut.Nodes.Create(context.Background(), ctmsmodels.NodeCreateRequest{Name: "node"})
				require.True(t, srvCalled)
				require.True(t, ctmsmodels.IsNotFound(err), "%v", err)
				require.Nil(t, got)
			})
		})
	}

	t.Run("does not adopt an existing resource after an uncertain create", func(t *testing.T) {
		calls := 0
		transport := &retryTransport{
			maxRetries: 3,
			wrapped: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				defer func() { _ = r.Body.Close() }()
				if r.URL.Host == "auth.test" {
					return response(200, `{"access_token":"token","expires_in":3600}`), nil
				}
				calls++
				if calls == 1 {
					return response(500, "failed after create"), nil
				}
				return response(422, `{"errorType":"DuplicateName","message":"already exists"}`), nil
			}),
		}
		cfg := CtmsClientConfig{Endpoint: "https://api.test", TokenURL: "https://auth.test/token"}
		c := NewCtmsClientFacadeWithHTTP(cfg, &http.Client{Transport: transport})
		node, err := c.Nodes.Create(context.Background(), ctmsmodels.NodeCreateRequest{Name: "node"})
		require.Nil(t, node)
		require.Equal(t, 2, calls)
		var api *ctmsmodels.APIError
		require.ErrorAs(t, err, &api)
		require.Equal(t, 422, api.StatusCode)
	})
}

func TestNodesFacade_Get(t *testing.T) {
	for _, namedUser := range []string{"", "terraform-automation"} {
		t.Run("named_user="+namedUser, func(t *testing.T) {
			t.Run("constructs the request correctly and handles the response", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/nodes/9007199254740993")
					assert.Empty(t, r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(nodeResponseJSON))
				})
				got, err := uut.Nodes.Get(context.Background(), 9007199254740993)
				require.True(t, srvCalled)
				require.NoError(t, err)
				assertResponseBody(t, got, nodeResponseJSON)
			})

			t.Run("returns NotFoundError on 404", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/nodes/9007199254740993")
					assert.Empty(t, r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"errorType":"NotFound","message":"gone"}`))
				})
				got, err := uut.Nodes.Get(context.Background(), 9007199254740993)
				require.True(t, srvCalled)
				require.True(t, ctmsmodels.IsNotFound(err), "%v", err)
				require.Nil(t, got)
			})
		})
	}

	t.Run("rejects malformed responses and missing or mismatched identity", func(t *testing.T) {
		for _, body := range []string{"", "null", "[]", "{broken", "{}", `{"id":null}`, `{"id":2}`} {
			t.Run(body, func(t *testing.T) {
				uut := prepareClientFacadeForTest(t, "", func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(body))
				})
				got, err := uut.Nodes.Get(context.Background(), 1)
				require.Error(t, err)
				require.Nil(t, got)
			})
		}
	})
}

func TestNodesFacade_Update(t *testing.T) {
	for _, namedUser := range []string{"", "terraform-automation"} {
		t.Run("named_user="+namedUser, func(t *testing.T) {
			t.Run("constructs the request correctly and handles the response", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodPatch, "/v2/nodes/9007199254740993")
					assert.Empty(t, r.URL.RawQuery)
					assert.Equal(t, namedUser, r.Header.Get("tms-named-user"))
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					var body json.RawMessage
					assertRequestBody(t, r, &body)
					assert.JSONEq(t, `{"uploadAllowed":false,"description":"","targets":[]}`, string(body))

					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(nodeResponseJSON))
				})
				got, err := uut.Nodes.Update(context.Background(), 9007199254740993, ctmsmodels.NodeUpdateRequest{
					UploadAllowed: new(false),
					Description:   new(""),
					Targets:       new([]ctmsmodels.DeployTarget{}),
				})
				require.True(t, srvCalled)
				require.NoError(t, err)
				assertResponseBody(t, got, nodeResponseJSON)
			})

			t.Run("returns NotFoundError on 404", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodPatch, "/v2/nodes/9007199254740993")
					assert.Empty(t, r.URL.RawQuery)
					assert.Equal(t, namedUser, r.Header.Get("tms-named-user"))
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					var body json.RawMessage
					assertRequestBody(t, r, &body)
					assert.JSONEq(t, `{"uploadAllowed":false,"description":"","targets":[]}`, string(body))

					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"errorType":"NotFound","message":"gone"}`))
				})
				got, err := uut.Nodes.Update(context.Background(), 9007199254740993, ctmsmodels.NodeUpdateRequest{
					UploadAllowed: new(false),
					Description:   new(""),
					Targets:       new([]ctmsmodels.DeployTarget{}),
				})
				require.True(t, srvCalled)
				require.True(t, ctmsmodels.IsNotFound(err), "%v", err)
				require.Nil(t, got)
			})
		})
	}
}

func TestNodesFacade_Delete(t *testing.T) {
	for _, namedUser := range []string{"", "terraform-automation"} {
		t.Run("named_user="+namedUser, func(t *testing.T) {
			t.Run("constructs the request correctly and handles the response", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodDelete, "/v2/nodes/9007199254740993")
					assert.Empty(t, r.URL.RawQuery)
					assert.Equal(t, namedUser, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusNoContent)
					_, _ = w.Write([]byte(""))
				})
				err := uut.Nodes.Delete(context.Background(), 9007199254740993)
				require.True(t, srvCalled)
				require.NoError(t, err)
			})

			t.Run("returns NotFoundError on 404", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodDelete, "/v2/nodes/9007199254740993")
					assert.Empty(t, r.URL.RawQuery)
					assert.Equal(t, namedUser, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"errorType":"NotFound","message":"gone"}`))
				})
				err := uut.Nodes.Delete(context.Background(), 9007199254740993)
				require.True(t, srvCalled)
				require.True(t, ctmsmodels.IsNotFound(err), "%v", err)
			})
		})
	}
}

// Local lifecycle coverage includes ID-only import reads, external drift, and resets.

func TestNodesFacade_List(t *testing.T) {
	for _, namedUser := range []string{"", "terraform-automation"} {
		t.Run("named_user="+namedUser, func(t *testing.T) {
			t.Run("constructs the request correctly and handles the response", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/nodes")
					assert.Empty(t, r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"nodes":[` + nodeResponseJSON + `]}`))
				})
				got, err := uut.Nodes.List(context.Background())
				require.True(t, srvCalled)
				require.NoError(t, err)
				assertResponseBody(t, got, `[`+nodeResponseJSON+`]`)
			})

			t.Run("returns NotFoundError on 404", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/nodes")
					assert.Empty(t, r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"errorType":"NotFound","message":"gone"}`))
				})
				got, err := uut.Nodes.List(context.Background())
				require.True(t, srvCalled)
				require.True(t, ctmsmodels.IsNotFound(err), "%v", err)
				require.Nil(t, got)
			})
		})
	}

	t.Run("returns an empty slice for absent null or empty collections", func(t *testing.T) {
		for _, body := range []string{`{}`, `{"nodes":null}`, `{"nodes":[]}`} {
			t.Run(body, func(t *testing.T) {
				uut := prepareClientFacadeForTest(t, "", func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(body))
				})
				got, err := uut.Nodes.List(context.Background())
				require.NoError(t, err)
				require.NotNil(t, got)
				require.Empty(t, got)
			})
		}
	})
}

func TestNodesFacade_Lifecycle(t *testing.T) {
	ctx := context.Background()
	uut, setDescription := prepareLandscapeFacadeForTest(t, "nodes")
	created, err := uut.Nodes.Create(ctx, ctmsmodels.NodeCreateRequest{
		Name:          "node",
		Description:   new("initial"),
		UploadAllowed: new(true),
		Targets: new([]ctmsmodels.DeployTarget{
			{
				ContentType:     new(ctmsmodels.MTA_ContentType),
				DestinationName: "dest",
			},
		}),
	})
	require.NoError(t, err)
	require.Equal(t, new(int64(1)), created.Id)

	imported := newFacadeForImportTest(uut)
	observed, err := imported.Nodes.Get(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, new("initial"), observed.Description)

	setDescription("external change")
	observed, err = imported.Nodes.Get(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, new("external change"), observed.Description)

	updated, err := uut.Nodes.Update(ctx, 1, ctmsmodels.NodeUpdateRequest{
		Description:   new(""),
		UploadAllowed: new(false),
		Targets:       new([]ctmsmodels.DeployTarget{}),
	})
	require.NoError(t, err)
	require.Equal(t, new(int64(1)), updated.Id)
	observed, err = imported.Nodes.Get(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, new(""), observed.Description)
	require.Equal(t, new(false), observed.UploadAllowed)
	require.NotNil(t, observed.Targets)
	require.Empty(t, *observed.Targets)

	require.NoError(t, uut.Nodes.Delete(ctx, 1))
	_, err = imported.Nodes.Get(ctx, 1)
	require.True(t, ctmsmodels.IsNotFound(err))
	require.True(t, ctmsmodels.IsNotFound(uut.Nodes.Delete(ctx, 1)))
}
