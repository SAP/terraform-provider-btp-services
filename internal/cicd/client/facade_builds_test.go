// internal/cicd/client/facade_builds_test.go

package cicdclient

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
)

func TestBuildsFacade_Trigger(t *testing.T) {
	t.Run("sends POST with full body and accepts 201", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			assertRequest(t, r, http.MethodPost, "/v2/jobs/my-job/builds")

			var body cicdmodels.BuildRequestDTO
			assertRequestBody(t, r, &body)
			assert.Equal(t, `W/"3"`, body.JobETag)
			assert.Equal(t, "main", body.CommitToBeBuilt)
			assert.Len(t, body.Parameters, 1)
			assert.Equal(t, "addon.yml", body.Parameters[0].Name)
			assert.Equal(t, "RESTRICTED", body.Parameters[0].Visibility)

			w.WriteHeader(http.StatusCreated)
		}))
		defer srv.Close()

		err := uut.Builds.Trigger(context.TODO(), "my-job", cicdmodels.BuildRequestDTO{
			JobETag:         `W/"3"`,
			CommitToBeBuilt: "main",
			Parameters: []cicdmodels.BuildParameter{
				{Name: "addon.yml", Value: "content", Visibility: "RESTRICTED"},
			},
		})

		assert.True(t, srvCalled)
		assert.NoError(t, err)
	})

	t.Run("sends POST with empty body (no guards)", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			assertRequest(t, r, http.MethodPost, "/v2/jobs/job-123/builds")
			w.WriteHeader(http.StatusCreated)
		}))
		defer srv.Close()

		err := uut.Builds.Trigger(context.TODO(), "job-123", cicdmodels.BuildRequestDTO{})
		assert.True(t, srvCalled)
		assert.NoError(t, err)
	})

	t.Run("URL-encodes job reference", func(t *testing.T) {
		var gotPath string

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.RequestURI
			w.WriteHeader(http.StatusCreated)
		}))
		defer srv.Close()

		_ = uut.Builds.Trigger(context.TODO(), "my job", cicdmodels.BuildRequestDTO{})
		assert.Equal(t, "/v2/jobs/my%20job/builds", gotPath)
	})

	t.Run("returns ConflictError on 409 (stale ETag)", func(t *testing.T) {
		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusConflict)
		}))
		defer srv.Close()

		err := uut.Builds.Trigger(context.TODO(), "my-job", cicdmodels.BuildRequestDTO{JobETag: `W/"1"`})
		assert.True(t, cicdmodels.IsConflict(err))
	})

	t.Run("returns error on API failure", func(t *testing.T) {
		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		err := uut.Builds.Trigger(context.TODO(), "my-job", cicdmodels.BuildRequestDTO{})
		assert.Error(t, err)
		assert.False(t, cicdmodels.IsConflict(err))
	})
}

func TestBuildsFacade_Abort(t *testing.T) {
	t.Run("sends POST to abort path and accepts 202", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			assertRequest(t, r, http.MethodPost, "/v2/jobs/my-job/builds/42/abort")
			w.WriteHeader(http.StatusAccepted)
		}))
		defer srv.Close()

		err := uut.Builds.Abort(context.TODO(), "my-job", "42")
		assert.True(t, srvCalled)
		assert.NoError(t, err)
	})

	t.Run("URL-encodes job and build references", func(t *testing.T) {
		var gotPath string

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.RequestURI
			w.WriteHeader(http.StatusAccepted)
		}))
		defer srv.Close()

		_ = uut.Builds.Abort(context.TODO(), "my job", "build id")
		assert.Equal(t, "/v2/jobs/my%20job/builds/build%20id/abort", gotPath)
	})

	t.Run("returns NotFoundError on 404", func(t *testing.T) {
		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		err := uut.Builds.Abort(context.TODO(), "my-job", "999")
		assert.True(t, cicdmodels.IsNotFound(err))
	})

	t.Run("returns error on API failure", func(t *testing.T) {
		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		err := uut.Builds.Abort(context.TODO(), "my-job", "1")
		assert.Error(t, err)
	})
}

func TestBuildsFacade_Delete(t *testing.T) {
	t.Run("sends DELETE and accepts 204", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			assertRequest(t, r, http.MethodDelete, "/v2/jobs/my-job/builds/7")
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		err := uut.Builds.Delete(context.TODO(), "my-job", "7")
		assert.True(t, srvCalled)
		assert.NoError(t, err)
	})

	t.Run("URL-encodes job and build references", func(t *testing.T) {
		var gotPath string

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.RequestURI
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		_ = uut.Builds.Delete(context.TODO(), "my job", "build id")
		assert.Equal(t, "/v2/jobs/my%20job/builds/build%20id", gotPath)
	})

	t.Run("returns NotFoundError on 404", func(t *testing.T) {
		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		err := uut.Builds.Delete(context.TODO(), "my-job", "999")
		assert.True(t, cicdmodels.IsNotFound(err))
	})

	t.Run("returns error on API failure", func(t *testing.T) {
		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		err := uut.Builds.Delete(context.TODO(), "my-job", "1")
		assert.Error(t, err)
	})
}
