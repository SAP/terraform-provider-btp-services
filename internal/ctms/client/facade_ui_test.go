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

func TestUIFacade_GetURL(t *testing.T) {
	for _, namedUser := range []string{"", "terraform-automation"} {
		t.Run("named_user="+namedUser, func(t *testing.T) {
			t.Run("constructs the request correctly and handles the response", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/tmsUiUrl")
					assert.Empty(t, r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"tmsUiUrl":"https://ui.example.test"}`))
				})
				got, err := uut.UI.GetURL(context.Background())
				require.True(t, srvCalled)
				require.NoError(t, err)
				assertResponseBody(t, got, `{"tmsUiUrl":"https://ui.example.test"}`)
			})

			t.Run("returns NotFoundError on 404", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/tmsUiUrl")
					assert.Empty(t, r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"errorType":"NotFound","message":"gone"}`))
				})
				got, err := uut.UI.GetURL(context.Background())
				require.True(t, srvCalled)
				require.True(t, ctmsmodels.IsNotFound(err), "%v", err)
				require.Nil(t, got)
			})
		})
	}
}

func TestUIFacade_GetTransportRequestLogsURL(t *testing.T) {
	for _, namedUser := range []string{"", "terraform-automation"} {
		t.Run("named_user="+namedUser, func(t *testing.T) {
			t.Run("constructs the request correctly and handles the response", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/tmsUiUrl/transportRequestLogs")
					assert.Equal(t, url.Values{"transportRequestId": {"req /?&+"}, "nodeId": {"node /?&+"}}.Encode(), r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"tmsUiLogUrl":"https://ui.example.test/log"}`))
				})
				got, err := uut.UI.GetTransportRequestLogsURL(context.Background(), "req /?&+", new("node /?&+"))
				require.True(t, srvCalled)
				require.NoError(t, err)
				assertResponseBody(t, got, `{"tmsUiLogUrl":"https://ui.example.test/log"}`)
			})

			t.Run("returns NotFoundError on 404", func(t *testing.T) {
				var srvCalled bool
				uut := prepareClientFacadeForTest(t, namedUser, func(w http.ResponseWriter, r *http.Request) {
					srvCalled = true
					assertRequest(t, r, http.MethodGet, "/v2/tmsUiUrl/transportRequestLogs")
					assert.Equal(t, url.Values{"transportRequestId": {"req /?&+"}, "nodeId": {"node /?&+"}}.Encode(), r.URL.RawQuery)
					assert.Empty(t, r.Header.Get("tms-named-user"))
					assertEmptyRequestBody(t, r)

					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"errorType":"NotFound","message":"gone"}`))
				})
				got, err := uut.UI.GetTransportRequestLogsURL(context.Background(), "req /?&+", new("node /?&+"))
				require.True(t, srvCalled)
				require.True(t, ctmsmodels.IsNotFound(err), "%v", err)
				require.Nil(t, got)
			})
		})
	}

	t.Run("omits nodeId when unset", func(t *testing.T) {
		uut := prepareClientFacadeForTest(t, "", func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodGet, "/v2/tmsUiUrl/transportRequestLogs")
			assert.Equal(t, url.Values{"transportRequestId": {"1"}}, r.URL.Query())
			_, _ = w.Write([]byte(`{"tmsUiLogUrl":"https://example.test"}`))
		})
		_, err := uut.UI.GetTransportRequestLogsURL(context.Background(), "1", nil)
		require.NoError(t, err)
	})
	t.Run("rejects a missing transport request ID before making a request", func(t *testing.T) {
		uut := prepareClientFacadeForTest(t, "", func(w http.ResponseWriter, r *http.Request) {
			t.Error("unexpected API request")
		})
		_, err := uut.UI.GetTransportRequestLogsURL(context.Background(), "", nil)
		require.Error(t, err)
	})
}
