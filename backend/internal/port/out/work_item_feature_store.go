package out

import "context"

// WorkItemFeatureStore is an optional extension point for feature-specific data.
// A future MongoDB adapter can implement it without changing the base entity.
type WorkItemFeatureStore interface {
	GetFeatures(ctx context.Context, workItemID string) (map[string]any, error)
	SaveFeatures(ctx context.Context, workItemID string, features map[string]any) error
}
