package ctmsclient

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

func TestDestinationsFacade_List(t *testing.T) {
	for _, namedUser := range []string{"", "terraform-automation"} {
		t.Run("named_user="+namedUser, func(t *testing.T) {
			t.Run("constructs the request correctly and handles the response", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/destinations")
					assert.Empty(t, r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"destinations":[{"name":"dest","url":"https://example.test","platformTypes":[]}]}`))
				})
				got, err := uut.Destinations.List(context.Background())
				require.True(t, srvCalled)
				require.NoError(t, err)
				assertResponseBody(t, got, `[{"name":"dest","url":"https://example.test","platformTypes":[]}]`)
			})

			t.Run("returns NotFoundError on 404", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/destinations")
					assert.Empty(t, r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"errorType":"NotFound","message":"gone"}`))
				})
				got, err := uut.Destinations.List(context.Background())
				require.True(t, srvCalled)
				require.True(t, ctmsmodels.IsNotFound(err), "%v", err)
				require.Nil(t, got)
			})
		})
	}

	t.Run("returns an empty slice for absent null or empty collections", func(t *testing.T) {
		for _, body := range []string{`{}`, `{"destinations":null}`, `{"destinations":[]}`} {
			t.Run(body, func(t *testing.T) {
				uut := prepareClientFacadeForTest(t, "", func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(body))
				})
				got, err := uut.Destinations.List(context.Background())
				require.NoError(t, err)
				require.NotNil(t, got)
				require.Empty(t, got)
			})
		}
	})
}

func TestDestinationsFacade_Get(t *testing.T) {
	for _, namedUser := range []string{"", "terraform-automation"} {
		t.Run("named_user="+namedUser, func(t *testing.T) {
			t.Run("constructs the request correctly and handles the response", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/destinations/"+url.PathEscape("dest /?#+%"))
					assert.Equal(t, "includeTenantId=true", r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"name":"dest /?#+%","tenantId":"tenant"}`))
				})
				got, err := uut.Destinations.Get(context.Background(), "dest /?#+%", true)
				require.True(t, srvCalled)
				require.NoError(t, err)
				assertResponseBody(t, got, `{"name":"dest /?#+%","tenantId":"tenant"}`)
			})

			t.Run("returns NotFoundError on 404", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/destinations/"+url.PathEscape("dest /?#+%"))
					assert.Equal(t, "includeTenantId=true", r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"errorType":"NotFound","message":"gone"}`))
				})
				got, err := uut.Destinations.Get(context.Background(), "dest /?#+%", true)
				require.True(t, srvCalled)
				require.True(t, ctmsmodels.IsNotFound(err), "%v", err)
				require.Nil(t, got)
			})
		})
	}

	t.Run("omits includeTenantId when false", func(t *testing.T) {
		uut := prepareClientFacadeForTest(t, "", func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodGet, "/v2/destinations/dest")
			assert.Empty(t, r.URL.RawQuery)
			_, _ = w.Write([]byte(`{"name":"dest"}`))
		})
		_, err := uut.Destinations.Get(context.Background(), "dest", false)
		require.NoError(t, err)
	})
}
