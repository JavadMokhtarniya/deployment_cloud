package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	// "github.com/alabarjasteh/deployment/order/repo/authentication"
	"github.com/gin-gonic/gin"
	"repo/authentication"
	// "order/vendor/github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	// "order/vendor/google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	// "order/vendor/google.golang.org/grpc/credentials/insecure"
)

type authMiddleware struct {
	AuthClient authentication.AuthServiceClient
	methodMap  map[string]authentication.Resource_Method
}

func NewAuthMiddleware(addr string) *authMiddleware {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	authClient := authentication.NewAuthServiceClient(conn)
	methodMap := initMethodMap()
	return &authMiddleware{authClient, methodMap}
}

func initMethodMap() map[string]authentication.Resource_Method {
	mm := map[string]authentication.Resource_Method{
		"GET":     authentication.Resource_GET,
		"POST":    authentication.Resource_POST,
		"PUT":     authentication.Resource_PUT,
		"DELETE":  authentication.Resource_DELETE,
		"HEAD":    authentication.Resource_HEAD,
		"CONNECT": authentication.Resource_CONNECT,
		"OPTIONS": authentication.Resource_OPTIONS,
		"TRACE":   authentication.Resource_TRACE,
		"PATCH":   authentication.Resource_PATCH,
	}
	return mm
}

type JwtClaims struct {
	UserID          string `json:"user_id,omitempty"`
	UserAccessLevel int    `json:"user_access_level,omitempty"`
	TokenUseCase    string `json:"token_use_case,omitempty"`
	Exp             int64  `json:"exp,omitempty"`
	Issued          int64  `json:"issued,omitempty"`
}

func (mw *authMiddleware) hasAccess(c *gin.Context) {
	tk, err := extractToken(c)
	if err != nil {
		log.Println(err)
		if errors.Is(err, ErrJWTIsMissing) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "jwt is missing",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal error",
		})
		return
	}
	m, ok := mw.methodMap[c.Request.Method]
	if !ok {
		m = authentication.Resource_INVALID
	}

	acc, err := mw.AuthClient.HasAccess(c.Request.Context(), &authentication.Resource{
		Method: m,
		Path:   c.Request.URL.Path,
		Jwt:    tk,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrJWTIsMissing):
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API token required"})
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		}
		return
	}
	if !acc.HasAccess {
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		return
	}
	claims, err := extractClaims(tk)
	log.Printf("claims: %v", claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal error",
		})
		return
	}
	c.Set("userID", claims.UserID)
	c.Next()
}

// Note: this is not the correct way of extracting claims.
// In a formal way, first jwt signature must be verified
//  against the public key of the Auth server.
func extractClaims(jwt string) (*JwtClaims, error) {
	parts := strings.Split(jwt, ".")
	claims, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var c *JwtClaims
	json.Unmarshal(claims, &c)
	return c, nil
}

func extractToken(c *gin.Context) (string, error) {
	tk := c.Request.Header.Get("Authorization")
	if tk == "" || !strings.HasPrefix(tk, "Bearer ") {
		return "", ErrJWTIsMissing
	}
	tk = strings.TrimPrefix(tk, "Bearer")
	return tk, nil
}
