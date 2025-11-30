package auth

import "testing"

func TestValidatePassword(t *testing.T) {
	cfg := defaultConfig()
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "strong password", password: "Str0ng!Pass"},
		{name: "missing uppercase", password: "weakpass1!", wantErr: true},
		{name: "missing digit", password: "Weak!Pass", wantErr: true},
		{name: "missing special", password: "WeakPass1", wantErr: true},
		{name: "too short", password: "S1!a", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePassword(tt.password, cfg); (err != nil) != tt.wantErr {
				t.Fatalf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHashAndComparePassword(t *testing.T) {
	cfg := defaultConfig()
	hash, err := HashPassword("Str0ng!Pass", cfg.BcryptCost)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if err := ComparePassword(hash, "Str0ng!Pass"); err != nil {
		t.Fatalf("ComparePassword() error = %v", err)
	}
}
