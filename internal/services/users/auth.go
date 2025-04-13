package users

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/morzisorn/gofermart/config"
)

func generateToken(login string) (string, error) {
	claims := jwt.MapClaims{
		"login": login,
		"exp":   time.Now().Add(7 * time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	cnfg := config.GetConfig()
	signedToken, err := token.SignedString([]byte(cnfg.SecretKey))
	if err != nil {
		return "", fmt.Errorf("generate token error: %w", err)
	}

	return signedToken, nil
}
