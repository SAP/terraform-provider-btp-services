// internal/cicd/client/facade_credentials_test.go

package cicdclient

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
)

func TestCredentialsFacade_Create(t *testing.T) {
	t.Run("constructs the request correctly", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			assertRequest(t, r, http.MethodPost, "/v2/credentials")
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var body cicdmodels.CreateCredentialRequest
			assertRequestBody(t, r, &body)
			assert.Equal(t, "my-cred", body.Name)
			assert.Equal(t, "test description", body.Description)
			assert.Equal(t, "user", body.Basic.Username)
			assert.Equal(t, "pass", body.Basic.Password)

			w.WriteHeader(http.StatusCreated)
		}))
		defer srv.Close()

		err := uut.Credentials.Create(context.TODO(), cicdmodels.CreateCredentialRequest{
			Name:        "my-cred",
			Description: "test description",
			Basic:       &cicdmodels.BasicAuth{Username: "user", Password: "pass"},
		})

		if assert.True(t, srvCalled) {
			assert.NoError(t, err)
		}
	})

	t.Run("returns error on API failure", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`invalid credential name`))
		}))
		defer srv.Close()

		err := uut.Credentials.Create(context.TODO(), cicdmodels.CreateCredentialRequest{Name: "bad!"})

		if assert.True(t, srvCalled) {
			assert.Error(t, err)
		}
	})
}

func TestCredentialsFacade_Get(t *testing.T) {
	t.Run("constructs the request correctly and decodes response", func(t *testing.T) {
		var srvCalled bool

		want := cicdmodels.Credential{
			ID:          "abc-123",
			Name:        "my-cred",
			Description: "test cred",
			Basic:       &cicdmodels.BasicAuthModel{Username: "user"},
		}

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			assertRequest(t, r, http.MethodGet, "/v2/credentials/my-cred")

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(want)
		}))
		defer srv.Close()

		got, err := uut.Credentials.Get(context.TODO(), "my-cred")

		if assert.True(t, srvCalled) && assert.NoError(t, err) {
			assert.Equal(t, want.ID, got.ID)
			assert.Equal(t, want.Name, got.Name)
			assert.Equal(t, want.Description, got.Description)
			assert.Equal(t, want.Basic.Username, got.Basic.Username)
		}
	})

	t.Run("returns NotFoundError on 404", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		_, err := uut.Credentials.Get(context.TODO(), "missing")

		if assert.True(t, srvCalled) && assert.Error(t, err) {
			assert.True(t, cicdmodels.IsNotFound(err))
		}
	})
}

func TestCredentialsFacade_Update(t *testing.T) {
	t.Run("constructs the request correctly", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			assertRequest(t, r, http.MethodPut, "/v2/credentials/abc-123")
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var body cicdmodels.UpdateCredentialRequest
			assertRequestBody(t, r, &body)
			assert.Equal(t, "my-cred", body.Name)
			assert.Equal(t, "updated description", body.Description)
			assert.Equal(t, "user", body.Basic.Username)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		err := uut.Credentials.Update(context.TODO(), "abc-123", cicdmodels.UpdateCredentialRequest{
			Name:        "my-cred",
			Description: "updated description",
			Basic:       &cicdmodels.BasicAuth{Username: "user", Password: "newpass"},
		})

		if assert.True(t, srvCalled) {
			assert.NoError(t, err)
		}
	})

	t.Run("returns NotFoundError on 404", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		err := uut.Credentials.Update(context.TODO(), "missing", cicdmodels.UpdateCredentialRequest{})

		if assert.True(t, srvCalled) && assert.Error(t, err) {
			assert.True(t, cicdmodels.IsNotFound(err))
		}
	})
}

func TestCredentialsFacade_Patch(t *testing.T) {
	t.Run("constructs the request correctly - description only", func(t *testing.T) {
		var srvCalled bool

		desc := "patched description"

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			assertRequest(t, r, http.MethodPatch, "/v2/credentials/abc-123")
			assert.Equal(t, "application/merge-patch+json", r.Header.Get("Content-Type"))

			var body cicdmodels.PatchCredentialRequest
			assertRequestBody(t, r, &body)
			assert.Equal(t, &desc, body.Description)
			assert.Nil(t, body.Name)
			assert.Nil(t, body.Basic)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		err := uut.Credentials.Patch(context.TODO(), "abc-123", cicdmodels.PatchCredentialRequest{
			Description: &desc,
		})

		if assert.True(t, srvCalled) {
			assert.NoError(t, err)
		}
	})

	t.Run("constructs the request correctly - name only", func(t *testing.T) {
		var srvCalled bool

		name := "new-name"

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			assertRequest(t, r, http.MethodPatch, "/v2/credentials/abc-123")

			var body cicdmodels.PatchCredentialRequest
			assertRequestBody(t, r, &body)
			assert.Equal(t, &name, body.Name)
			assert.Nil(t, body.Description)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		err := uut.Credentials.Patch(context.TODO(), "abc-123", cicdmodels.PatchCredentialRequest{
			Name: &name,
		})

		if assert.True(t, srvCalled) {
			assert.NoError(t, err)
		}
	})

	t.Run("returns NotFoundError on 404", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		err := uut.Credentials.Patch(context.TODO(), "missing", cicdmodels.PatchCredentialRequest{})

		if assert.True(t, srvCalled) && assert.Error(t, err) {
			assert.True(t, cicdmodels.IsNotFound(err))
		}
	})
}

func TestCredentialsFacade_Delete(t *testing.T) {
	t.Run("constructs the request correctly", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			assertRequest(t, r, http.MethodDelete, "/v2/credentials/abc-123")

			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		err := uut.Credentials.Delete(context.TODO(), "abc-123")

		if assert.True(t, srvCalled) {
			assert.NoError(t, err)
		}
	})

	t.Run("returns NotFoundError on 404", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		err := uut.Credentials.Delete(context.TODO(), "missing")

		if assert.True(t, srvCalled) && assert.Error(t, err) {
			assert.True(t, cicdmodels.IsNotFound(err))
		}
	})
}

func TestCredentialsFacade_List(t *testing.T) {
	t.Run("constructs the request correctly and decodes response", func(t *testing.T) {
		var srvCalled bool

		creds := []cicdmodels.Credential{
			{ID: "1", Name: "cred-one", Basic: &cicdmodels.BasicAuthModel{Username: "user1"}},
			{ID: "2", Name: "cred-two", Basic: &cicdmodels.BasicAuthModel{Username: "user2"}},
		}

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			assertRequest(t, r, http.MethodGet, "/v2/credentials")

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cicdmodels.CredentialListResponse{
				Embedded: &cicdmodels.CredentialListEmbedded{Credentials: creds},
			})
		}))
		defer srv.Close()

		got, err := uut.Credentials.List(context.TODO())

		if assert.True(t, srvCalled) && assert.NoError(t, err) {
			assert.Len(t, got, 2)
			assert.Equal(t, "cred-one", got[0].Name)
			assert.Equal(t, "cred-two", got[1].Name)
		}
	})

	t.Run("returns empty slice when no credentials exist", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()

		got, err := uut.Credentials.List(context.TODO())

		if assert.True(t, srvCalled) && assert.NoError(t, err) {
			assert.Empty(t, got)
		}
	})

	t.Run("returns error on API failure", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		_, err := uut.Credentials.List(context.TODO())

		if assert.True(t, srvCalled) {
			assert.Error(t, err)
		}
	})
}
