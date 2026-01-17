package main

import (
	"context"
	"fmt"

	"github.com/rompi/core-backend/pkg/feature"
	"github.com/rompi/core-backend/pkg/feature/provider"
)

func main() {
	// Create flags with targeting rules
	memProvider := provider.NewMemoryProviderWithFlags(map[string]*feature.Flag{
		"premium-features": {
			Key:          "premium-features",
			Type:         feature.FlagTypeBool,
			DefaultValue: false,
			Enabled:      true,
			Variants: []feature.Variant{
				{Name: "off", Value: false},
				{Name: "on", Value: true},
			},
			Rules: []feature.Rule{
				// Rule 1: Beta testers always get it
				{
					ID: "beta-testers",
					Clauses: []feature.Clause{
						{
							Attribute: "groups",
							Operator:  feature.OpContains,
							Values:    []interface{}{"beta-testers"},
						},
					},
					Variation: 1,
				},
				// Rule 2: Pro and enterprise plan users
				{
					ID: "paid-plans",
					Clauses: []feature.Clause{
						{
							Attribute: "plan",
							Operator:  feature.OpIn,
							Values:    []interface{}{"pro", "enterprise"},
						},
					},
					Variation: 1,
				},
				// Rule 3: US users with verified email
				{
					ID: "us-verified",
					Clauses: []feature.Clause{
						{
							Attribute: "country",
							Operator:  feature.OpEquals,
							Values:    []interface{}{"US"},
						},
						{
							Attribute: "email_verified",
							Operator:  feature.OpEquals,
							Values:    []interface{}{true},
						},
					},
					Variation: 1,
				},
			},
		},
		"experiment-button": {
			Key:          "experiment-button",
			Type:         feature.FlagTypeString,
			DefaultValue: "blue",
			Enabled:      true,
			Variants: []feature.Variant{
				{Name: "blue", Value: "blue"},
				{Name: "green", Value: "green"},
				{Name: "red", Value: "red"},
			},
			Rules: []feature.Rule{
				// Percentage rollout
				{
					ID: "ab-test",
					Rollout: &feature.Rollout{
						BucketBy: "key",
						Variations: []feature.WeightedVariation{
							{Variation: 0, Weight: 33333}, // ~33% blue
							{Variation: 1, Weight: 33333}, // ~33% green
							{Variation: 2, Weight: 33334}, // ~33% red
						},
					},
				},
			},
		},
	})

	client := feature.NewWithProvider(memProvider)
	defer client.Close()

	// Test different user contexts
	testUsers := []struct {
		name string
		ctx  *feature.Context
	}{
		{
			name: "Beta tester",
			ctx: feature.NewContext("user-1").
				WithGroup("beta-testers").
				WithAttribute("plan", "free"),
		},
		{
			name: "Pro user",
			ctx: feature.NewContext("user-2").
				WithAttribute("plan", "pro"),
		},
		{
			name: "US verified user",
			ctx: feature.NewContext("user-3").
				WithCountry("US").
				WithAttribute("email_verified", true).
				WithAttribute("plan", "free"),
		},
		{
			name: "Free user in UK",
			ctx: feature.NewContext("user-4").
				WithCountry("UK").
				WithAttribute("plan", "free"),
		},
	}

	fmt.Println("=== Premium Features Flag ===")
	for _, user := range testUsers {
		ctx := feature.WithContext(context.Background(), user.ctx)
		enabled := client.Bool(ctx, "premium-features", false)
		fmt.Printf("%s: premium-features = %v\n", user.name, enabled)
	}

	fmt.Println("\n=== A/B Test Button Color ===")
	// Test percentage rollout with multiple users
	for i := 1; i <= 10; i++ {
		userCtx := feature.NewContext(fmt.Sprintf("user-%d", i))
		ctx := feature.WithContext(context.Background(), userCtx)

		color := client.String(ctx, "experiment-button", "blue")
		eval, _ := client.Variation(ctx, "experiment-button")

		fmt.Printf("user-%d: button color = %s (variant %d)\n", i, color, eval.VariationIdx)
	}
}
