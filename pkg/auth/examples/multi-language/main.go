package main

import (
	"log"
	"net/http"

	"github.com/rompi/core-backend/pkg/auth"
)

func main() {
	translator := auth.NewTranslator("en")
	translator.Register("es", map[string]string{
		auth.CodeInvalidCredentials: "Credenciales inválidas",
		auth.CodeAccountLocked:      "Cuenta bloqueada por intentos fallidos",
	})
	auth.DefaultTranslator = translator

	err := auth.NewAuthError(auth.CodeInvalidCredentials, http.StatusUnauthorized, "es", nil)
	log.Printf("Spanish message: %s", err.Message)
}
