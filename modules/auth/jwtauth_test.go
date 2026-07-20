package auth

import (
	"os"
	"testing"
	"time"

	"github.com/appservR/appservR/models"
	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateToken(t *testing.T) {
	user := models.User{
		Username:      "alice",
		DisplayedName: "Alice A.",
		Groups:        []models.Group{{Name: "admins"}, {Name: "editors"}},
	}

	tokenString := GenerateToken(user)
	if tokenString == "" {
		t.Fatal("expected non-empty token string")
	}

	token, err := ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}
	if !token.Valid {
		t.Fatal("expected token.Valid to be true")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("expected claims to be jwt.MapClaims")
	}
	if claims["username"] != "alice" {
		t.Errorf("expected username claim %q, got %q", "alice", claims["username"])
	}
	if claims["name"] != "Alice A." {
		t.Errorf("expected name claim %q, got %q", "Alice A.", claims["name"])
	}
	groups, _ := claims["groups"].(string)
	if groups != "admins,editors" {
		t.Errorf("expected groups claim %q, got %q", "admins,editors", groups)
	}
	if claims["iss"] != "AppservR" {
		t.Errorf("expected issuer %q, got %q", "AppservR", claims["iss"])
	}
}

func TestGenerateTokenNoGroups(t *testing.T) {
	user := models.User{Username: "bob", DisplayedName: "Bob"}
	tokenString := GenerateToken(user)

	token, err := ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}
	claims := token.Claims.(jwt.MapClaims)
	if claims["groups"] != "" {
		t.Errorf("expected empty groups claim, got %q", claims["groups"])
	}
}

func TestValidateTokenRejectsGarbage(t *testing.T) {
	_, err := ValidateToken("not-a-token")
	if err == nil {
		t.Fatal("expected error validating garbage token string")
	}
}

func TestValidateTokenRejectsTamperedSignature(t *testing.T) {
	user := models.User{Username: "carol"}
	tokenString := GenerateToken(user)

	// Flip a character in the middle of the signature segment. The very
	// last base64url character of a 256-bit HMAC-SHA256 signature encodes
	// 2 bits of padding slack that don't affect the decoded bytes, so
	// flipping it there can non-deterministically fail to actually change
	// the signature; a middle character doesn't have that problem.
	i := len(tokenString) / 2
	tampered := tokenString[:i]
	if tokenString[i] == 'a' {
		tampered += "b"
	} else {
		tampered += "a"
	}
	tampered += tokenString[i+1:]

	token, err := ValidateToken(tampered)
	if err == nil && token.Valid {
		t.Fatal("expected tampered token to fail validation")
	}
}

func TestValidateTokenRejectsExpiredToken(t *testing.T) {
	claims := &authCustomClaims{
		Username:          "dave",
		DisplayedUsername: "Dave",
		Groups:            "",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			Issuer:    "AppservR",
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(getSecretKey()))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}

	token, err := ValidateToken(signed)
	if err == nil && token.Valid {
		t.Fatal("expected expired token to fail validation")
	}
}

func TestValidateTokenRejectsUnexpectedSigningMethod(t *testing.T) {
	claims := &authCustomClaims{
		Username: "eve",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to sign 'none'-alg test token: %v", err)
	}

	token, err := ValidateToken(signed)
	if err == nil && token.Valid {
		t.Fatal("expected token signed with alg=none to be rejected")
	}
}

func TestGetSecretKeyUsesEnvVarWhenSet(t *testing.T) {
	os.Setenv("APPSERVR_AUTH_SECRET", "my-fixed-test-secret")
	defer os.Unsetenv("APPSERVR_AUTH_SECRET")

	if got := getSecretKey(); got != "my-fixed-test-secret" {
		t.Errorf("expected secret from env var, got %q", got)
	}

	// A token signed under the env-configured secret should validate.
	user := models.User{Username: "frank"}
	tokenString := GenerateToken(user)
	token, err := ValidateToken(tokenString)
	if err != nil || !token.Valid {
		t.Fatalf("expected token signed with env secret to validate, err=%v", err)
	}
}

func TestGetSecretKeyFallsBackToRandomSecret(t *testing.T) {
	os.Unsetenv("APPSERVR_AUTH_SECRET")
	if got := getSecretKey(); got != randomSecret {
		t.Errorf("expected fallback to randomSecret, got %q", got)
	}
}
