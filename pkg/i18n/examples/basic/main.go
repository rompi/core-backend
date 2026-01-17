// Package main demonstrates basic usage of the i18n package.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rompi/core-backend/pkg/i18n"
)

func main() {
	// Create i18n instance with JSON catalog
	i, err := i18n.New(i18n.Config{
		DefaultLocale:  "en",
		FallbackLocale: "en",
		CatalogType:    "json",
		Path:           "./locales",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Basic translation with default locale (English)
	fmt.Println("=== English (default) ===")
	fmt.Println(i.T(ctx, "greeting", "World"))
	fmt.Println(i.T(ctx, "welcome.title"))
	fmt.Println(i.T(ctx, "welcome.message"))

	// Switch to Spanish
	fmt.Println("\n=== Spanish ===")
	ctx = i.WithLocale(ctx, "es")
	fmt.Println(i.T(ctx, "greeting", "Mundo"))
	fmt.Println(i.T(ctx, "welcome.title"))
	fmt.Println(i.T(ctx, "welcome.message"))

	// Using named arguments
	fmt.Println("\n=== Named Arguments ===")
	ctx = i.WithLocale(ctx, "en")
	msg := i.Tf(ctx, "order.confirmation", map[string]interface{}{
		"OrderID": "12345",
		"Name":    "John",
	})
	fmt.Println(msg)

	// Using the Localizer directly
	fmt.Println("\n=== Using Localizer ===")
	en := i.L("en")
	es := i.L("es")

	fmt.Println("English:", en.T("greeting", "World"))
	fmt.Println("Spanish:", es.T("greeting", "Mundo"))

	// Check available locales
	fmt.Println("\n=== Available Locales ===")
	fmt.Println(i.Locales())
}
