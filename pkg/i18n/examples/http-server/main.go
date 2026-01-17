// Package main demonstrates HTTP middleware integration with the i18n package.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/rompi/core-backend/pkg/i18n"
	"github.com/rompi/core-backend/pkg/i18n/middleware"
)

var i18nInstance i18n.I18n

func main() {
	// Create i18n instance
	var err error
	i18nInstance, err = i18n.New(i18n.Config{
		DefaultLocale:  "en",
		FallbackLocale: "en",
		CatalogType:    "json",
		Path:           "./locales",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create HTTP mux
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/api/greeting", greetingHandler)
	mux.HandleFunc("/api/items", itemsHandler)

	// Apply i18n middleware
	// Locale is detected from:
	// 1. Query param: ?lang=es
	// 2. Cookie: lang=es
	// 3. Accept-Language header
	handler := middleware.HTTP(i18nInstance,
		middleware.WithQueryParam("lang"),
		middleware.WithCookie("lang"),
		middleware.WithAcceptLanguage(),
		middleware.WithDefaultLocale("en"),
		middleware.WithSetCookie(true),
	)(mux)

	fmt.Println("Server starting on :8080")
	fmt.Println("Try:")
	fmt.Println("  curl http://localhost:8080/")
	fmt.Println("  curl http://localhost:8080/?lang=es")
	fmt.Println("  curl -H 'Accept-Language: de' http://localhost:8080/")
	fmt.Println("  curl http://localhost:8080/api/greeting?name=World")
	fmt.Println("  curl http://localhost:8080/api/items?count=5")

	log.Fatal(http.ListenAndServe(":8080", handler))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	locale := i18nInstance.Locale(r.Context())

	response := map[string]interface{}{
		"locale":  locale,
		"title":   i18nInstance.T(r.Context(), "welcome.title"),
		"message": i18nInstance.T(r.Context(), "welcome.message"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func greetingHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "Guest"
	}

	locale := i18nInstance.Locale(r.Context())
	localizer := i18nInstance.L(locale)

	response := map[string]interface{}{
		"locale":   locale,
		"greeting": i18nInstance.Tf(r.Context(), "greeting", map[string]interface{}{"Name": name}),
		"direction": string(localizer.Direction()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	countStr := r.URL.Query().Get("count")
	count := 1
	if countStr != "" {
		fmt.Sscanf(countStr, "%d", &count)
	}

	locale := i18nInstance.Locale(r.Context())

	response := map[string]interface{}{
		"locale": locale,
		"count":  count,
		"text":   i18nInstance.Tn(r.Context(), "items", count),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
