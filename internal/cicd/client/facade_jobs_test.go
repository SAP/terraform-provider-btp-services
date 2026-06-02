// internal/cicd/client/facade_jobs_test.go

package cicdclient

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	cicdmodels "github.com/SAP/terraform-provider-btp-services/internal/cicd/models"
)

func TestJobsFacade_ListByRepository(t *testing.T) {
	t.Run("constructs the request correctly and decodes response", func(t *testing.T) {
		var srvCalled bool

		jobs := []cicdmodels.Job{
			{
				ID:                 "job-id-1",
				Name:               "my-job",
				Active:             true,
				Pipeline:           "sap-cloud-sdk",
				PipelineVersion:    "3.0",
				PipelineParameters: map[string]any{"runTestStage": true},
				BuildRetentionDays: 28,
				MaxBuildsToKeep:    15,
				Branch:             "main",
				RepositoryID:       "repo-id-1",
			},
			{
				ID:                 "job-id-2",
				Name:               "another-job",
				Active:             false,
				Pipeline:           "cpi",
				PipelineVersion:    "1.0",
				PipelineParameters: map[string]any{},
				BuildRetentionDays: 14,
				MaxBuildsToKeep:    10,
			},
		}

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			assertRequest(t, r, http.MethodGet, "/v2/repositories/my-repo/jobs")

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cicdmodels.JobListResponse{
				Embedded: &cicdmodels.JobListEmbedded{Jobs: jobs},
			})
		}))
		defer srv.Close()

		got, err := uut.Jobs.ListByRepository(context.TODO(), "my-repo")

		if assert.True(t, srvCalled) && assert.NoError(t, err) {
			assert.Len(t, got, 2)
			assert.Equal(t, "my-job", got[0].Name)
			assert.Equal(t, true, got[0].Active)
			assert.Equal(t, "sap-cloud-sdk", got[0].Pipeline)
			assert.Equal(t, "main", got[0].Branch)
			assert.Equal(t, "another-job", got[1].Name)
		}
	})

	t.Run("accepts repository ID as reference", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true

			assertRequest(t, r, http.MethodGet, "/v2/repositories/abc-123-uuid/jobs")

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cicdmodels.JobListResponse{
				Embedded: &cicdmodels.JobListEmbedded{Jobs: []cicdmodels.Job{{ID: "j1", Name: "job-one"}}},
			})
		}))
		defer srv.Close()

		got, err := uut.Jobs.ListByRepository(context.TODO(), "abc-123-uuid")

		if assert.True(t, srvCalled) && assert.NoError(t, err) {
			assert.Len(t, got, 1)
		}
	})

	t.Run("returns empty slice when repository has no jobs", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()

		got, err := uut.Jobs.ListByRepository(context.TODO(), "empty-repo")

		if assert.True(t, srvCalled) && assert.NoError(t, err) {
			assert.Empty(t, got)
		}
	})

	t.Run("returns NotFoundError when repository does not exist", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		_, err := uut.Jobs.ListByRepository(context.TODO(), "missing-repo")

		if assert.True(t, srvCalled) && assert.Error(t, err) {
			assert.True(t, cicdmodels.IsNotFound(err))
		}
	})

	t.Run("returns error on API failure", func(t *testing.T) {
		var srvCalled bool

		uut, srv := prepareClientFacadeForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srvCalled = true
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		_, err := uut.Jobs.ListByRepository(context.TODO(), "my-repo")

		if assert.True(t, srvCalled) {
			assert.Error(t, err)
		}
	})
}
