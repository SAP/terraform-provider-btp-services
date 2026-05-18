// internal/cicd/client/facade_repositories_test.go

package cicdclient

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	cicdmodels "github.com/SAP/terraform-provider-sap-btp-services/internal/cicd/models"
)

func TestRepositoriesFacade_Create(t *testing.T) {
	t.Run("constructs the request correctly", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			assertRequest(t, r, http.MethodPost, "/v2/repositories")
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var body cicdmodels.CreateRepositoryRequest
			assertRequestBody(t, r, &body)
			assert.Equal(t, "my-repo", body.Name)
			assert.Equal(t, "https://github.com/example/my-repo", body.CloneURL)

			w.WriteHeader(http.StatusCreated)
		}))
		defer srv.Close()

		err := uut.Repositories.Create(context.TODO(), cicdmodels.CreateRepositoryRequest{
			Name:     "my-repo",
			CloneURL: "https://github.com/example/my-repo",
		})

		if assert.True(t, srvCalled) {
			assert.NoError(t, err)
		}
	})

	t.Run("includes optional credential IDs when set", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			var body cicdmodels.CreateRepositoryRequest
			assertRequestBody(t, r, &body)
			if assert.NotNil(t, body.CloneCredentialID) {
				assert.Equal(t, "cred-id-123", *body.CloneCredentialID)
			}
			if assert.NotNil(t, body.CloudConnectorCredentialID) {
				assert.Equal(t, "cc-cred-id-456", *body.CloudConnectorCredentialID)
			}

			w.WriteHeader(http.StatusCreated)
		}))
		defer srv.Close()

		cloneCredID := "cred-id-123"
		ccCredID := "cc-cred-id-456"
		err := uut.Repositories.Create(context.TODO(), cicdmodels.CreateRepositoryRequest{
			Name:                       "my-repo",
			CloneURL:                   "https://github.com/example/my-repo",
			CloneCredentialID:          &cloneCredID,
			CloudConnectorCredentialID: &ccCredID,
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
			_, _ = w.Write([]byte(`invalid repository name`))
		}))
		defer srv.Close()

		err := uut.Repositories.Create(context.TODO(), cicdmodels.CreateRepositoryRequest{Name: "bad!"})

		if assert.True(t, srvCalled) {
			assert.Error(t, err)
		}
	})
}

func TestRepositoriesFacade_Get(t *testing.T) {
	t.Run("constructs the request correctly and decodes response", func(t *testing.T) {
		var srvCalled bool

		want := cicdmodels.Repository{
			ID:       "pb091fd5-845b-4146-9bfs-d8cb74be04f8",
			Name:     "my-repo",
			CloneURL: "https://github.com/example/my-repo",
		}

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			assertRequest(t, r, http.MethodGet, "/v2/repositories/my-repo")

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(want)
		}))
		defer srv.Close()

		got, err := uut.Repositories.Get(context.TODO(), "my-repo")

		if assert.True(t, srvCalled) && assert.NoError(t, err) {
			assert.Equal(t, want.ID, got.ID)
			assert.Equal(t, want.Name, got.Name)
			assert.Equal(t, want.CloneURL, got.CloneURL)
		}
	})

	t.Run("decodes event_receiver when present", func(t *testing.T) {
		var srvCalled bool

		want := cicdmodels.Repository{
			ID:       "pb091fd5-845b-4146-9bfs-d8cb74be04f8",
			Name:     "my-repo",
			CloneURL: "https://github.com/example/my-repo",
			EventReceiver: &cicdmodels.EventReceiverModel{
				Active:                   true,
				SCMType:                  "GITHUB",
				WebhookID:                "webhook-id-123",
				WebhookTokenCredentialID: "cred-id-456",
			},
		}

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(want)
		}))
		defer srv.Close()

		got, err := uut.Repositories.Get(context.TODO(), "my-repo")

		if assert.True(t, srvCalled) && assert.NoError(t, err) {
			if assert.NotNil(t, got.EventReceiver) {
				assert.True(t, got.EventReceiver.Active)
				assert.Equal(t, "GITHUB", got.EventReceiver.SCMType)
				assert.Equal(t, "webhook-id-123", got.EventReceiver.WebhookID)
				assert.Equal(t, "cred-id-456", got.EventReceiver.WebhookTokenCredentialID)
			}
		}
	})

	t.Run("returns NotFoundError on 404", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		_, err := uut.Repositories.Get(context.TODO(), "missing")

		if assert.True(t, srvCalled) && assert.Error(t, err) {
			assert.True(t, cicdmodels.IsNotFound(err))
		}
	})
}

func TestRepositoriesFacade_Update(t *testing.T) {
	t.Run("constructs the request correctly - PUT to /v2/repositories with id in body", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			assertRequest(t, r, http.MethodPut, "/v2/repositories")
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var body cicdmodels.UpdateRepositoryRequest
			assertRequestBody(t, r, &body)
			assert.Equal(t, "pb091fd5-845b-4146-9bfs-d8cb74be04f8", body.ID)
			assert.Equal(t, "my-repo", body.Name)
			assert.Equal(t, "https://github.com/example/my-repo-updated", body.CloneURL)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		err := uut.Repositories.Update(context.TODO(), cicdmodels.UpdateRepositoryRequest{
			ID:       "pb091fd5-845b-4146-9bfs-d8cb74be04f8",
			Name:     "my-repo",
			CloneURL: "https://github.com/example/my-repo-updated",
		})

		if assert.True(t, srvCalled) {
			assert.NoError(t, err)
		}
	})

	t.Run("returns error on API failure", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		err := uut.Repositories.Update(context.TODO(), cicdmodels.UpdateRepositoryRequest{Name: "missing"})

		if assert.True(t, srvCalled) && assert.Error(t, err) {
			assert.True(t, cicdmodels.IsNotFound(err))
		}
	})
}

func TestRepositoriesFacade_Delete(t *testing.T) {
	t.Run("constructs the request correctly", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			assertRequest(t, r, http.MethodDelete, "/v2/repositories/pb091fd5-845b-4146-9bfs-d8cb74be04f8")

			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		err := uut.Repositories.Delete(context.TODO(), "pb091fd5-845b-4146-9bfs-d8cb74be04f8")

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

		err := uut.Repositories.Delete(context.TODO(), "missing")

		if assert.True(t, srvCalled) && assert.Error(t, err) {
			assert.True(t, cicdmodels.IsNotFound(err))
		}
	})
}

func TestRepositoriesFacade_List(t *testing.T) {
	t.Run("constructs the request correctly and decodes response", func(t *testing.T) {
		var srvCalled bool

		repos := []cicdmodels.Repository{
			{ID: "1", Name: "repo-one", CloneURL: "https://github.com/example/repo-one"},
			{ID: "2", Name: "repo-two", CloneURL: "https://github.com/example/repo-two"},
		}

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			assertRequest(t, r, http.MethodGet, "/v2/repositories")

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cicdmodels.RepositoryListResponse{
				Embedded: &cicdmodels.RepositoryListEmbedded{Repositories: repos},
			})
		}))
		defer srv.Close()

		got, err := uut.Repositories.List(context.TODO())

		if assert.True(t, srvCalled) && assert.NoError(t, err) {
			assert.Len(t, got, 2)
			assert.Equal(t, "repo-one", got[0].Name)
			assert.Equal(t, "repo-two", got[1].Name)
		}
	})

	t.Run("returns empty slice when no repositories exist", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()

		got, err := uut.Repositories.List(context.TODO())

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

		_, err := uut.Repositories.List(context.TODO())

		if assert.True(t, srvCalled) {
			assert.Error(t, err)
		}
	})
}
