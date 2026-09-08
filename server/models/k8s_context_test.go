package models

import (
	"testing"

	"github.com/meshery/meshery/server/internal/sql"
	"github.com/meshery/schemas/models/core"
	"github.com/gofrs/uuid"
)

func TestK8sContextGenerateID(t *testing.T) {
	instanceID, _ := uuid.NewV4()

	tests := []struct {
		name     string
		context  K8sContext
		wantSame bool // whether ID should remain same after token change
	}{
		{
			name: "regular context with token",
			context: K8sContext{
				Name:  "test-context",
				Auth:  sql.Map{"user": map[string]interface{}{"token": "original-token"}},
				Cluster: sql.Map{"server": "https://k8s.example.com"},
				MesheryInstanceID: &instanceID,
			},
			wantSame: true, // ID should be same after token change
		},
		{
			name: "in-cluster context with token",
			context: K8sContext{
				Name:  "in-cluster-context",
				Auth:  sql.Map{"user": map[string]interface{}{"token": "service-account-token"}},
				Cluster: sql.Map{"server": "https://kubernetes.default.svc"},
				MesheryInstanceID: &instanceID,
			},
			wantSame: true, // ID should be same after token rotation
		},
		{
			name: "context without token",
			context: K8sContext{
				Name:  "no-token-context",
				Auth:  sql.Map{"user": map[string]interface{}{"client-certificate": "cert"}},
				Cluster: sql.Map{"server": "https://k8s.example.com"},
				MesheryInstanceID: &instanceID,
			},
			wantSame: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate initial ID
			id1, err := K8sContextGenerateID(tt.context)
			if err != nil {
				t.Fatalf("K8sContextGenerateID() error = %v", err)
			}

			// Simulate token rotation by changing the token
			if tt.context.Auth != nil {
				if user, ok := tt.context.Auth["user"].(map[string]interface{}); ok {
					if _, hasToken := user["token"]; hasToken {
						user["token"] = "rotated-token"
					}
				}
			}

			// Generate ID after token change
			id2, err := K8sContextGenerateID(tt.context)
			if err != nil {
				t.Fatalf("K8sContextGenerateID() error = %v", err)
			}

			// Check if IDs remain the same
			if tt.wantSame && id1 != id2 {
				t.Errorf("ID changed after token rotation, got %v, want %v", id2, id1)
			}
		})
	}
}
