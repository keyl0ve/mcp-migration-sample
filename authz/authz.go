package authz

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
)

var (
	jwtSecret = []byte("your-secret-key")

	apiKeysMu sync.RWMutex
	apiKeys   = map[string]*APIKey{
		"sk-1234567890abcdef": {
			Key:    "sk-1234567890abcdef",
			UserID: "user1",
			Scopes: []string{"read", "write"},
		},
		"sk-abcdef1234567890": {
			Key:    "sk-abcdef1234567890",
			UserID: "user2",
			Scopes: []string{"read"},
		},
	}
)

type JWTClaims struct {
	UserID string   `json:"user_id"`
	Scopes []string `json:"scopes"`
	jwt.RegisteredClaims
}

type APIKey struct {
	Key    string   `json:"key"`
	UserID string   `json:"user_id"`
	Scopes []string `json:"scopes"`
}

func VerifyJWT(ctx context.Context, tokenString string, _ *http.Request) (*mcpauth.TokenInfo, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", mcpauth.ErrInvalidToken, err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return &mcpauth.TokenInfo{
			Scopes:     claims.Scopes,
			Expiration: claims.ExpiresAt.Time,
		}, nil
	}

	return nil, fmt.Errorf("%w: invalid token claims", mcpauth.ErrInvalidToken)
}

func VerifyAPIKey(ctx context.Context, key string, _ *http.Request) (*mcpauth.TokenInfo, error) {
	apiKeysMu.RLock()
	stored, ok := apiKeys[key]
	apiKeysMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: API key not found", mcpauth.ErrInvalidToken)
	}

	return &mcpauth.TokenInfo{
		Scopes:     stored.Scopes,
		Expiration: time.Now().Add(24 * time.Hour),
	}, nil
}

func GenerateTokenHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "test-user"
	}

	scopes := strings.Split(r.URL.Query().Get("scopes"), ",")
	if len(scopes) == 1 && scopes[0] == "" {
		scopes = []string{"read", "write"}
	}

	expiresIn := 1 * time.Hour
	if expStr := r.URL.Query().Get("expires_in"); expStr != "" {
		if exp, err := time.ParseDuration(expStr); err == nil {
			expiresIn = exp
		}
	}

	claims := JWTClaims{
		UserID: userID,
		Scopes: scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "failed to sign token", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token": tokenString,
		"type":  "Bearer",
	})
}

func GenerateAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		http.Error(w, "failed to generate api key", http.StatusInternalServerError)
		return
	}
	apiKey := "sk-" + base64.URLEncoding.EncodeToString(bytes)

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "test-user"
	}

	scopes := strings.Split(r.URL.Query().Get("scopes"), ",")
	if len(scopes) == 1 && scopes[0] == "" {
		scopes = []string{"read"}
	}

	apiKeysMu.Lock()
	apiKeys[apiKey] = &APIKey{
		Key:    apiKey,
		UserID: userID,
		Scopes: scopes,
	}
	apiKeysMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]string{
		"api_key": apiKey,
		"type":    "Bearer",
	})
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
