package testutils

import (
	"context"
	"testing"

	ctmsmodels "github.com/SAP/terraform-provider-btp-services/internal/ctms/models"
)

// CreateVirtualNode creates a disposable discovery-test node and records its cleanup.
func (r *Recording) CreateVirtualNode(t *testing.T, name string) int64 {
	t.Helper()
	client := r.Client("terraform-provider-ctms-test")
	virtual := true
	node, err := client.Nodes.Create(context.Background(), ctmsmodels.NodeCreateRequest{
		Name:    name,
		Virtual: &virtual,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err := client.Nodes.Delete(context.Background(), *node.Id)
		if err != nil && !ctmsmodels.IsNotFound(err) {
			t.Errorf("cleanup CTMS node: %v", err)
		}
	})
	return *node.Id
}
