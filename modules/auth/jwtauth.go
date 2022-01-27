package auth

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/appservR/appservR/models"
	"github.com/golang-jwt/jwt"
    uuid "github.com/satori/go.uuid"
)

var randomSecret = uuid.NewV4().String()

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
	jwt.StandardClaims
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
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
			Issuer:    "AppservR",
			IssuedAt:  time.Now().Unix(),
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
