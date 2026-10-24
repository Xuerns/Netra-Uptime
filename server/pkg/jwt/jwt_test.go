package jwt

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func init() {
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-secret-key-1234567890")
	}
}

func TestGenerateToken_ContainsUserID(t *testing.T) {
	os.Setenv("JWT_SECRET", "my-test-secret")
	userID := "user-12345"

	token, err := GenerateToken(userID, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected token to be non-empty")
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed on generated token: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, claims.UserID)
	}
	if claims.Issuer != "netra-uptime" {
		t.Errorf("expected issuer netra-uptime, got %s", claims.Issuer)
	}
}

func TestValidateToken_Valid(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret-xyz")
	token, err := GenerateToken("user-abc", 10*time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken returned error for valid token: %v", err)
	}
	if claims.UserID != "user-abc" {
		t.Errorf("expected user-abc, got %s", claims.UserID)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret-xyz")
	// Expiry in past (-1 second)
	token, err := GenerateToken("user-expired", -1*time.Second)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	claims, err := ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
	if claims != nil {
		t.Errorf("expected nil claims on error, got %+v", claims)
	}
}

func TestValidateToken_TamperedPayload(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret-xyz")
	token, err := GenerateToken("user-safe", time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected token structure: %s", token)
	}
	// Tamper payload part
	tamperedToken := parts[0] + ".eyJ1c2VyX2lkIjoidGFtcGVyZWQifQ." + parts[2]

	claims, err := ValidateToken(tamperedToken)
	if err == nil {
		t.Fatal("expected error for tampered token, got nil")
	}
	if claims != nil {
		t.Errorf("expected nil claims for tampered token, got %+v", claims)
	}
}

func TestValidateToken_WrongSignatureKey(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret-A")
	token, err := GenerateToken("user-key", time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	// Change secret key for validation
	os.Setenv("JWT_SECRET", "secret-B")
	claims, err := ValidateToken(token)
	if err == nil {
		t.Fatal("expected error when validating with different secret, got nil")
	}
	if claims != nil {
		t.Errorf("expected nil claims with wrong secret, got %+v", claims)
	}
}

func TestValidateToken_EmptyString(t *testing.T) {
	claims, err := ValidateToken("")
	if err == nil {
		t.Fatal("expected error for empty token string, got nil")
	}
	if claims != nil {
		t.Errorf("expected nil claims, got %+v", claims)
	}
}

func TestValidateToken_NonHMACMethod(t *testing.T) {
	// Token signed with 'none' alg
	token := jwt.NewWithClaims(jwt.SigningMethodNone, &JWTClaims{
		UserID: "user-none",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "netra-uptime",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed creating unsigned token: %v", err)
	}

	claims, err := ValidateToken(tokenString)
	if err == nil {
		t.Fatal("expected error for non-HMAC signing method, got nil")
	}
	if claims != nil {
		t.Errorf("expected nil claims, got %+v", claims)
	}
}
