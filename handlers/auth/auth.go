package auth

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/VitoNaychev/bt-customer-svc/handlers"
)

func GenerateJWT(secretKey []byte, expiresAt time.Duration, subject int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(int64(subject), 10),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresAt)),
	})

	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateJWTWithStringSubject(secretKey []byte, expiresAt time.Duration, subject string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresAt)),
	})

	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateJWTWithoutSubject(secretKey []byte, expiresAt time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresAt)),
	})

	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyJWT(jwtString string, secretKey []byte) (*jwt.Token, error) {
	token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	if err != nil {
		return nil, err
	}

	return token, nil
}

func AuthenticationMiddleware(endpointHandler func(w http.ResponseWriter, r *http.Request), secretKey []byte) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] == nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(handlers.ErrorResponse{Message: handlers.ErrMissingToken.Error()})
			return
		}

		token, err := VerifyJWT(r.Header["Token"][0], secretKey)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(handlers.ErrorResponse{Message: err.Error()})
			return
		}

		id, err := getIDFromToken(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(handlers.ErrorResponse{Message: err.Error()})
			return
		}

		r.Header.Add("Subject", strconv.Itoa(id))

		endpointHandler(w, r)
	})
}

func getIDFromToken(token *jwt.Token) (int, error) {
	subject, err := token.Claims.GetSubject()
	if err != nil || subject == "" {
		return -1, handlers.ErrMissingSubject
	}

	id, err := strconv.Atoi(subject)
	if err != nil {
		return -1, handlers.ErrNonIntegerSubject
	}

	return id, nil
}
