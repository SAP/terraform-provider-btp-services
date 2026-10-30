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
const routeResponseJSON = `{"id":42,"name":"route","description":"","sourceNodeId":1,"targetNodeId":2}`

func TestRoutesFacade_Create(t *testing.T) {
	for _, namedUser := range []string{"", "terraform-automation"} {
		t.Run("named_user="+namedUser, func(t *testing.T) {
			t.Run("constructs the request correctly and handles the response", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodPost, "/v2/routes")
					assert.Empty(t, r.URL.RawQuery)
					assert.Equal(t, namedUser, r.Header.Get("tms-named-user"))
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					var body json.RawMessage
					assertRequestBody(t, r, &body)
					assert.JSONEq(t, `{"name":"route","sourceNodeId":1,"targetNodeId":2}`, string(body))

					w.WriteHeader(http.StatusCreated)
					_, _ = w.Write([]byte(routeResponseJSON))
				})
				got, err := uut.Routes.Create(context.Background(), ctmsmodels.RouteCreateRequest{
					Name:         "route",
					SourceNodeId: 1,
					TargetNodeId: 2,
				})
				require.True(t, srvCalled)
				require.NoError(t, err)
				assertResponseBody(t, got, routeResponseJSON)
			})

			t.Run("returns NotFoundError on 404", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodPost, "/v2/routes")
					assert.Empty(t, r.URL.RawQuery)
					assert.Equal(t, namedUser, r.Header.Get("tms-named-user"))
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					var body json.RawMessage
					assertRequestBody(t, r, &body)
					assert.JSONEq(t, `{"name":"route","sourceNodeId":1,"targetNodeId":2}`, string(body))

					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"errorType":"NotFound","message":"gone"}`))
				})
				got, err := uut.Routes.Create(context.Background(), ctmsmodels.RouteCreateRequest{
					Name:         "route",
					SourceNodeId: 1,
					TargetNodeId: 2,
				})
				require.True(t, srvCalled)
				require.True(t, ctmsmodels.IsNotFound(err), "%v", err)
				require.Nil(t, got)
			})
		})
	}
}

func TestRoutesFacade_Get(t *testing.T) {
	for _, namedUser := range []string{"", "terraform-automation"} {
		t.Run("named_user="+namedUser, func(t *testing.T) {
			t.Run("constructs the request correctly and handles the response", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/routes/42")
					assert.Empty(t, r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(routeResponseJSON))
				})
				got, err := uut.Routes.Get(context.Background(), 42)
				require.True(t, srvCalled)
				require.NoError(t, err)
				assertResponseBody(t, got, routeResponseJSON)
			})

			t.Run("returns NotFoundError on 404", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/routes/42")
					assert.Empty(t, r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"errorType":"NotFound","message":"gone"}`))
				})
				got, err := uut.Routes.Get(context.Background(), 42)
				require.True(t, srvCalled)
				require.True(t, ctmsmodels.IsNotFound(err), "%v", err)
				require.Nil(t, got)
			})
		})
	}
}

func TestRoutesFacade_Update(t *testing.T) {
	for _, namedUser := range []string{"", "terraform-automation"} {
		t.Run("named_user="+namedUser, func(t *testing.T) {
			t.Run("constructs the request correctly and handles the response", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodPatch, "/v2/routes/42")
					assert.Empty(t, r.URL.RawQuery)
					assert.Equal(t, namedUser, r.Header.Get("tms-named-user"))
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					var body json.RawMessage
					assertRequestBody(t, r, &body)
					assert.JSONEq(t, `{"description":""}`, string(body))

					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(routeResponseJSON))
				})
				got, err := uut.Routes.Update(context.Background(), 42, ctmsmodels.RoutePatchRequest{Description: new("")})
				require.True(t, srvCalled)
				require.NoError(t, err)
				assertResponseBody(t, got, routeResponseJSON)
			})

			t.Run("returns NotFoundError on 404", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodPatch, "/v2/routes/42")
					assert.Empty(t, r.URL.RawQuery)
					assert.Equal(t, namedUser, r.Header.Get("tms-named-user"))
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					var body json.RawMessage
					assertRequestBody(t, r, &body)
					assert.JSONEq(t, `{"description":""}`, string(body))

					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"errorType":"NotFound","message":"gone"}`))
				})
				got, err := uut.Routes.Update(context.Background(), 42, ctmsmodels.RoutePatchRequest{Description: new("")})
				require.True(t, srvCalled)
				require.True(t, ctmsmodels.IsNotFound(err), "%v", err)
				require.Nil(t, got)
			})
		})
	}
}

func TestRoutesFacade_Delete(t *testing.T) {
	for _, namedUser := range []string{"", "terraform-automation"} {
		t.Run("named_user="+namedUser, func(t *testing.T) {
			t.Run("constructs the request correctly and handles the response", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodDelete, "/v2/routes/42")
					assert.Empty(t, r.URL.RawQuery)
					assert.Equal(t, namedUser, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusNoContent)
					_, _ = w.Write([]byte(""))
				})
				err := uut.Routes.Delete(context.Background(), 42)
				require.True(t, srvCalled)
				require.NoError(t, err)
			})

			t.Run("returns NotFoundError on 404", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodDelete, "/v2/routes/42")
					assert.Empty(t, r.URL.RawQuery)
					assert.Equal(t, namedUser, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"errorType":"NotFound","message":"gone"}`))
				})
				err := uut.Routes.Delete(context.Background(), 42)
				require.True(t, srvCalled)
				require.True(t, ctmsmodels.IsNotFound(err), "%v", err)
			})
		})
	}
}

// Local lifecycle coverage includes ID-only import reads, external drift, and resets.

func TestRoutesFacade_List(t *testing.T) {
	for _, namedUser := range []string{"", "terraform-automation"} {
		t.Run("named_user="+namedUser, func(t *testing.T) {
			t.Run("constructs the request correctly and handles the response", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/routes")
					assert.Empty(t, r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"routes":[` + routeResponseJSON + `]}`))
				})
				got, err := uut.Routes.List(context.Background())
				require.True(t, srvCalled)
				require.NoError(t, err)
				assertResponseBody(t, got, `[`+routeResponseJSON+`]`)
			})

			t.Run("returns NotFoundError on 404", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/routes")
					assert.Empty(t, r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"errorType":"NotFound","message":"gone"}`))
				})
				got, err := uut.Routes.List(context.Background())
				require.True(t, srvCalled)
				require.True(t, ctmsmodels.IsNotFound(err), "%v", err)
				require.Nil(t, got)
			})
		})
	}

	t.Run("returns an empty slice for absent null or empty collections", func(t *testing.T) {
		for _, body := range []string{`{}`, `{"routes":null}`, `{"routes":[]}`} {
			t.Run(body, func(t *testing.T) {
				uut := prepareClientFacadeForTest(t, "", func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(body))
				})
				got, err := uut.Routes.List(context.Background())
				require.NoError(t, err)
				require.NotNil(t, got)
				require.Empty(t, got)
			})
		}
	})
}

func TestRoutesFacade_Lifecycle(t *testing.T) {
	ctx := context.Background()
	uut, setDescription := prepareLandscapeFacadeForTest(t, "routes")
	created, err := uut.Routes.Create(ctx, ctmsmodels.RouteCreateRequest{
		Name:         "route",
		Description:  new("initial"),
		SourceNodeId: 10,
		TargetNodeId: 20,
	})
	require.NoError(t, err)
	require.Equal(t, new(int64(1)), created.Id)

	imported := newFacadeForImportTest(uut)
	observed, err := imported.Routes.Get(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, new("initial"), observed.Description)

	setDescription("external change")
	observed, err = imported.Routes.Get(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, new("external change"), observed.Description)

	updated, err := uut.Routes.Update(ctx, 1, ctmsmodels.RoutePatchRequest{
		Description: new(""),
	})
	require.NoError(t, err)
	require.Equal(t, new(int64(1)), updated.Id)
	observed, err = imported.Routes.Get(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, new(""), observed.Description)
	require.Equal(t, new(int64(10)), observed.SourceNodeId)
	require.Equal(t, new(int64(20)), observed.TargetNodeId)

	require.NoError(t, uut.Routes.Delete(ctx, 1))
	_, err = imported.Routes.Get(ctx, 1)
	require.True(t, ctmsmodels.IsNotFound(err))
	require.True(t, ctmsmodels.IsNotFound(uut.Routes.Delete(ctx, 1)))
}
