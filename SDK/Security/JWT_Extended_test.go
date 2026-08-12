package Security

import (
	"strings"
	"testing"
	"time"
)

func newExtendedJWTTestProcessor(t *testing.T) *JWTProcessor {
	t.Helper()
	processor := &JWTProcessor{}
	if !processor.LoadRSAKey(nil, nil) {
		t.Fatal("failed to initialize RSA key")
	}
	if !processor.LoadAESKey([]byte("0123456789abcdef")) {
		t.Fatal("failed to initialize AES key")
	}
	return processor
}

func TestExtendedJWTCreateAndVerify(t *testing.T) {
	processor := newExtendedJWTTestProcessor(t)
	claims := map[string]interface{}{
		"sub": "account-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	extensions := []map[string]interface{}{
		{"department": "factory"},
		{"features": []interface{}{"audit", "terminal"}},
	}

	token, err := processor.CreateExtendedToken(claims, extensions)
	if err != nil {
		t.Fatalf("CreateExtendedToken failed: %v", err)
	}
	if got := len(strings.Split(token, ".")); got != 5 {
		t.Fatalf("expected 5 segments, got %d", got)
	}
	verified, err := processor.VerifyExtendedToken(token, false)
	if err != nil {
		t.Fatalf("VerifyExtendedToken failed: %v", err)
	}
	if verified.Claims["sub"] != "account-1" {
		t.Fatalf("unexpected claims: %#v", verified.Claims)
	}
	if len(verified.Extensions) != 2 || verified.Extensions[0]["department"] != "factory" {
		t.Fatalf("unexpected extensions: %#v", verified.Extensions)
	}
	claimsObject := decryptTokenWithProcessor(processor, token, false, true)
	if claimsObject == nil || !claimsObject.Has(extendedJWTClaimsKey) {
		t.Fatalf("Security verifier did not expose verified extensions: %#v", claimsObject)
	}
}

func TestExtendedJWTRejectsTamperAndTruncation(t *testing.T) {
	processor := newExtendedJWTTestProcessor(t)
	token, err := processor.CreateExtendedToken(
		map[string]interface{}{"sub": "account-1", "exp": time.Now().Add(time.Hour).Unix()},
		[]map[string]interface{}{{"scope": "read"}, {"scope": "write"}},
	)
	if err != nil {
		t.Fatalf("CreateExtendedToken failed: %v", err)
	}

	parts := strings.Split(token, ".")
	parts[3] = tamperExtendedJWTSegment(parts[3])
	if _, err := processor.VerifyExtendedToken(strings.Join(parts, "."), false); err == nil {
		t.Fatal("expected tampered extension to be rejected")
	}
	if _, err := processor.VerifyExtendedToken(strings.Join(strings.Split(token, ".")[:4], "."), false); err == nil {
		t.Fatal("expected truncated extension to be rejected")
	}

	parts = strings.Split(token, ".")
	parts[1] = tamperExtendedJWTSegment(parts[1])
	if _, err := processor.VerifyExtendedToken(strings.Join(parts, "."), false); err == nil {
		t.Fatal("expected tampered JWT payload to be rejected")
	}

	parts = strings.Split(token, ".")
	parts[3], parts[4] = parts[4], parts[3]
	if _, err := processor.VerifyExtendedToken(strings.Join(parts, "."), false); err == nil {
		t.Fatal("expected swapped extensions to be rejected")
	}
}

func tamperExtendedJWTSegment(segment string) string {
	replacement := byte('A')
	if segment[len(segment)-1] == replacement {
		replacement = 'B'
	}
	return segment[:len(segment)-1] + string(replacement)
}

func TestExtendedJWTDoesNotChangeCompactJWE(t *testing.T) {
	processor := newExtendedJWTTestProcessor(t)
	token := processor.CreateToken(TM_AES.Value(), map[string]interface{}{
		"iss": "legacy-jwe",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	if token == "" || len(strings.Split(token, ".")) != 5 {
		t.Fatalf("failed to create compact JWE: %q", token)
	}
	if processor.IsExtendedToken(token) {
		t.Fatal("compact JWE must not be classified as Extended JWT")
	}
	claims := processor.DecryptToken(token, false)
	if claims == nil || claims.OptString("iss", "") != "legacy-jwe" {
		t.Fatalf("compact JWE compatibility failed: %#v", claims)
	}
}

func TestExtendedJWTThreeAndFourSegments(t *testing.T) {
	processor := newExtendedJWTTestProcessor(t)
	claims := map[string]interface{}{"sub": "account-1", "exp": time.Now().Add(time.Hour).Unix()}

	for extensionCount := 0; extensionCount <= 1; extensionCount++ {
		extensions := make([]map[string]interface{}, extensionCount)
		if extensionCount == 1 {
			extensions[0] = map[string]interface{}{"tenant": "factory"}
		}
		token, err := processor.CreateExtendedToken(claims, extensions)
		if err != nil {
			t.Fatalf("CreateExtendedToken(%d) failed: %v", extensionCount, err)
		}
		if got := len(strings.Split(token, ".")); got != 3+extensionCount {
			t.Fatalf("expected %d segments, got %d", 3+extensionCount, got)
		}
		if _, err := processor.VerifyExtendedToken(token, false); err != nil {
			t.Fatalf("VerifyExtendedToken(%d) failed: %v", extensionCount, err)
		}
	}
}

func TestExtendedJWTRejectsExpiredClaims(t *testing.T) {
	processor := newExtendedJWTTestProcessor(t)
	token, err := processor.CreateExtendedToken(
		map[string]interface{}{"sub": "account-1", "exp": time.Now().Add(-time.Minute).Unix()},
		nil,
	)
	if err != nil {
		t.Fatalf("CreateExtendedToken failed: %v", err)
	}
	if _, err := processor.VerifyExtendedToken(token, false); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
	if _, err := processor.VerifyExtendedToken(token, true); err != nil {
		t.Fatalf("ignoreExp should accept the token: %v", err)
	}
}
