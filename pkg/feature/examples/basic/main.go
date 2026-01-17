package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rompi/core-backend/pkg/feature"
	"github.com/rompi/core-backend/pkg/feature/provider"
)

func main() {
	// Create an in-memory provider with some initial flags
	memProvider := provider.NewMemoryProviderWithFlags(map[string]*feature.Flag{
		"new-dashboard": {
			Key:          "new-dashboard",
			Type:         feature.FlagTypeBool,
			DefaultValue: false,
			Enabled:      true,
			Variants: []feature.Variant{
				{Name: "off", Value: false},
				{Name: "on", Value: true},
			},
			Rules: []feature.Rule{
				{
					ID:        "enable-for-all",
					Variation: 1, // "on" variant
				},
			},
		},
		"theme": {
			Key:          "theme",
			Type:         feature.FlagTypeString,
			DefaultValue: "light",
			Enabled:      true,
			Variants: []feature.Variant{
				{Name: "light", Value: "light"},
				{Name: "dark", Value: "dark"},
			},
			Rules: []feature.Rule{
				{Variation: 0}, // Default to light
			},
		},
		"max-items": {
			Key:          "max-items",
			Type:         feature.FlagTypeInt,
			DefaultValue: 10,
			Enabled:      true,
			Variants: []feature.Variant{
				{Name: "default", Value: 10},
				{Name: "premium", Value: 100},
			},
			Rules: []feature.Rule{
				{Variation: 0},
			},
		},
	})

	// Create client with the provider
	client := feature.NewWithProvider(memProvider)
	defer client.Close()

	ctx := context.Background()

	// Evaluate boolean flag
	if client.Bool(ctx, "new-dashboard", false) {
		fmt.Println("New dashboard enabled!")
	} else {
		fmt.Println("Using old dashboard")
	}

	// Evaluate string flag
	theme := client.String(ctx, "theme", "light")
	fmt.Printf("Theme: %s\n", theme)

	// Evaluate int flag
	maxItems := client.Int(ctx, "max-items", 10)
	fmt.Printf("Max items: %d\n", maxItems)

	// Get all flag values
	allFlags := client.AllFlags(ctx)
	fmt.Printf("All flags: %v\n", allFlags)

	// Get detailed evaluation
	eval, err := client.Variation(ctx, "new-dashboard")
	if err != nil {
		log.Printf("Error evaluating flag: %v", err)
		return
	}

	fmt.Printf("Evaluation details:\n")
	fmt.Printf("  Key: %s\n", eval.Key)
	fmt.Printf("  Value: %v\n", eval.Value)
	fmt.Printf("  Reason: %s\n", eval.Reason)
	fmt.Printf("  Rule ID: %s\n", eval.RuleID)
}
