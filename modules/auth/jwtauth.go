package auth

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/appservR/appservR/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var randomSecret = uuid.New().String()

func getSecretKey() string {
	secret := os.Getenv("APPSERVR_AUTH_SECRET")
	if secret == "" {
		secret = randomSecret
	}
	return secret
}

type authCustomClaims struct {
	Username          string `json:"username"`
	DisplayedUsername string `json:"name"`
	Groups            string `json:"groups"`
	jwt.RegisteredClaims
}

func GenerateToken(user models.User) string {
	groups := []string{}
	for _, g := range user.Groups {
		groups = append(groups, g.Name)
	}
	claims := &authCustomClaims{
		user.Username,
		user.DisplayedName,
		strings.Join(groups, ","),
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
			Issuer:    "AppservR",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(getSecretKey()))
	if err != nil {
		panic(err)
	}
	return t
}

func ValidateToken(encodedToken string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			return nil, fmt.Errorf("Invalid token %s", token.Header["alg"])
		}
		return []byte(getSecretKey()), nil
	})
}
