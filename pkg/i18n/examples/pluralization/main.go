// Package main demonstrates pluralization with the i18n package.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rompi/core-backend/pkg/i18n"
)

func main() {
	// Create i18n instance
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

	// English pluralization (simple: one/other)
	fmt.Println("=== English Pluralization ===")
	for _, count := range []int{0, 1, 2, 5, 10, 21} {
		fmt.Printf("%d: %s\n", count, i.Tn(ctx, "items", count))
	}

	// Russian pluralization (complex: one/few/many/other)
	fmt.Println("\n=== Russian Pluralization ===")
	ctx = i.WithLocale(ctx, "ru")
	for _, count := range []int{0, 1, 2, 5, 10, 21, 22, 25} {
		fmt.Printf("%d: %s\n", count, i.Tn(ctx, "items", count))
	}

	// Arabic pluralization (zero/one/two/few/many/other)
	fmt.Println("\n=== Arabic Pluralization ===")
	ctx = i.WithLocale(ctx, "ar")
	for _, count := range []int{0, 1, 2, 3, 10, 11, 100} {
		fmt.Printf("%d: %s\n", count, i.Tn(ctx, "items", count))
	}

	// Using localizer directly
	fmt.Println("\n=== Using Localizer ===")
	en := i.L("en")
	ru := i.L("ru")

	fmt.Println("English - 1 item:", en.Tn("items", 1))
	fmt.Println("English - 5 items:", en.Tn("items", 5))
	fmt.Println("Russian - 1 товар:", ru.Tn("items", 1))
	fmt.Println("Russian - 2 товара:", ru.Tn("items", 2))
	fmt.Println("Russian - 5 товаров:", ru.Tn("items", 5))
}
