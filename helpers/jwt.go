package helpers

import (
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID uint, email, role string) (string, int64, error) {
	expiresHours, _ := strconv.Atoi(os.Getenv("JWT_EXPIRES_HOURS"))
	if expiresHours == 0 {
		expiresHours = 24
	}
	expiresAt := time.Now().Add(time.Duration(expiresHours) * time.Hour)
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	return signed, int64(time.Until(expiresAt).Seconds()), err
}

func ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return token.Claims.(jwt.MapClaims), nil
}
