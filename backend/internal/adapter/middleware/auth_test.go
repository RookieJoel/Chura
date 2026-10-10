package middleware_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/RookieJoel/Chura/backend/internal/adapter/middleware"
	"github.com/RookieJoel/Chura/backend/internal/domain"
)

func newKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

func signToken(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

// authApp mounts the middleware on a tiny app whose handler echoes the Actor.
func authApp(key *rsa.PrivateKey) *fiber.App {
	app := fiber.New()
	keyfunc := func(*jwt.Token) (any, error) { return &key.PublicKey, nil }
	app.Get("/me", middleware.KeycloakAuth(keyfunc), func(c *fiber.Ctx) error {
		actor, _ := c.Locals(middleware.ActorLocalsKey).(domain.Actor)
		return c.JSON(actor)
	})
	return app
}

func call(t *testing.T, app *fiber.App, bearer string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(nethttp.MethodGet, "/me", nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	var obj map[string]any
	_ = json.Unmarshal(raw, &obj)
	return resp.StatusCode, obj
}

func claims(overrides jwt.MapClaims) jwt.MapClaims {
	c := jwt.MapClaims{
		"sub":          "u1",
		"email":        "u1@example.com",
		"exp":          time.Now().Add(time.Hour).Unix(),
		"realm_access": map[string]any{"roles": []any{"member"}},
	}
	for k, v := range overrides {
		if v == nil {
			delete(c, k)
			continue
		}
		c[k] = v
	}
	return c
}

func realm(roles ...any) map[string]any { return map[string]any{"roles": roles} }

func TestKeycloakAuth_Claims(t *testing.T) {
	key := newKey(t)
	cases := []struct {
		name       string
		claims     jwt.MapClaims
		wantStatus int
		wantError  string
		wantRole   string
	}{
		{"member", claims(nil), 200, "", "team_member"},
		{"auditor", claims(jwt.MapClaims{"realm_access": realm("auditor")}), 200, "", "auditor"},
		{"member among other roles", claims(jwt.MapClaims{"realm_access": realm("offline_access", "member")}), 200, "", "team_member"},
		{"non-string role entries ignored", claims(jwt.MapClaims{"realm_access": realm(123, "member")}), 200, "", "team_member"},
		{"no roles", claims(jwt.MapClaims{"realm_access": realm()}), 403, "forbidden", ""},
		{"unrelated roles only", claims(jwt.MapClaims{"realm_access": realm("offline_access")}), 403, "forbidden", ""},
		{"both roles", claims(jwt.MapClaims{"realm_access": realm("member", "auditor")}), 403, "forbidden", ""},
		{"only non-string roles", claims(jwt.MapClaims{"realm_access": realm(1, true)}), 403, "forbidden", ""},
		{"realm_access wrong type", claims(jwt.MapClaims{"realm_access": "member"}), 403, "forbidden", ""},
		{"roles wrong type", claims(jwt.MapClaims{"realm_access": map[string]any{"roles": "member"}}), 403, "forbidden", ""},
		{"realm_access missing", claims(jwt.MapClaims{"realm_access": nil}), 403, "forbidden", ""},
		{"missing sub", claims(jwt.MapClaims{"sub": nil}), 401, "unauthenticated", ""},
		{"sub is number", claims(jwt.MapClaims{"sub": 42}), 401, "unauthenticated", ""},
		{"blank sub", claims(jwt.MapClaims{"sub": "  "}), 401, "unauthenticated", ""},
		{"missing email", claims(jwt.MapClaims{"email": nil}), 401, "unauthenticated", ""},
		{"email not a string", claims(jwt.MapClaims{"email": []any{"a@b.c"}}), 401, "unauthenticated", ""},
		{"blank email", claims(jwt.MapClaims{"email": ""}), 401, "unauthenticated", ""},
		{"expired", claims(jwt.MapClaims{"exp": time.Now().Add(-time.Hour).Unix()}), 401, "unauthenticated", ""},
	}
	app := authApp(key)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := call(t, app, signToken(t, key, tc.claims))
			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body %v)", status, tc.wantStatus, body)
			}
			if tc.wantError != "" && body["error"] != tc.wantError {
				t.Fatalf("error = %v, want %q", body["error"], tc.wantError)
			}
			if tc.wantRole != "" {
				if body["UserID"] != "u1" || body["Email"] != "u1@example.com" || body["Role"] != tc.wantRole {
					t.Fatalf("actor = %v", body)
				}
			}
		})
	}
}

func TestKeycloakAuth_RejectsBadTokens(t *testing.T) {
	key := newKey(t)
	app := authApp(key)

	t.Run("signed by another key", func(t *testing.T) {
		status, body := call(t, app, signToken(t, newKey(t), claims(nil)))
		if status != 401 || body["error"] != "unauthenticated" {
			t.Fatalf("status = %d, body = %v", status, body)
		}
	})
	t.Run("no Authorization header", func(t *testing.T) {
		status, body := call(t, app, "")
		if status != 401 || body["error"] != "unauthenticated" {
			t.Fatalf("status = %d, body = %v", status, body)
		}
	})
	t.Run("garbage token", func(t *testing.T) {
		status, body := call(t, app, "not-a-jwt")
		if status != 401 || body["error"] != "unauthenticated" {
			t.Fatalf("status = %d, body = %v", status, body)
		}
	})
}

func TestKeycloakAuth_StoresAuthUserForWhoAmI(t *testing.T) {
	key := newKey(t)
	app := fiber.New()
	keyfunc := func(*jwt.Token) (any, error) { return &key.PublicKey, nil }
	app.Get("/me", middleware.KeycloakAuth(keyfunc), func(c *fiber.Ctx) error {
		user, _ := c.UserContext().Value(middleware.UserKey).(middleware.AuthUser)
		return c.JSON(user)
	})

	status, body := call(t, app, signToken(t, key, claims(jwt.MapClaims{"realm_access": realm("member", 7, "offline_access")})))

	roles, _ := body["Roles"].([]any)
	if status != 200 || body["ID"] != "u1" || body["Email"] != "u1@example.com" || len(roles) != 2 || roles[0] != "member" || roles[1] != "offline_access" {
		t.Fatalf("status = %d, body = %v", status, body)
	}
}
