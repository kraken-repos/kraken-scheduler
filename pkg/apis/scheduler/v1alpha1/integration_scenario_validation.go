package v1alpha1

import (
	"context"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmp"
)

func (r *IntegrationScenario) Validate(ctx context.Context) *apis.FieldError {
	if apis.IsInUpdate(ctx) {
		original := apis.GetBaseline(ctx).(*IntegrationScenario)
		if diff, err := kmp.ShortDiff(original.Spec, r.Spec); err != nil {
			return &apis.FieldError{
				Message: "Failed to diff IntegrationScenario",
				Paths:   []string{"spec"},
				Details: err.Error(),
			}
		} else if diff != "" {
			return &apis.FieldError{
				Message: "Immutable fields changed (-old +new)",
				Paths:   []string{"spec"},
				Details: diff,
			}
		}
	}
	return nil
}