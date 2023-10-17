package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type StubCustomerStore struct {
	customers []GetCustomerResponse
}

func (s *StubCustomerStore) GetCustomerInfo(id int) (*GetCustomerResponse, error) {
	if len(s.customers) < id {
		return nil, fmt.Errorf("no customer with id %v", id)
	}

	customerInfo := s.customers[id]
	return &customerInfo, nil
}

func TestCreateUser(t *testing.T) {
	server := &CustomerServer{}
	t.Run("stores customer on POST and returns JWT", func(t *testing.T) {
		customer := CreateCustomerRequest{
			FirstName:   "Peter",
			LastName:    "Smith",
			PhoneNumber: "+359 88 576 5981",
			Email:       "petesmith@gmail.com",
			Password:    "firefirefire",
		}

		body := bytes.NewBuffer([]byte{})
		json.NewEncoder(body).Encode(customer)

		request, _ := http.NewRequest(http.MethodPost, "/customer/", body)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := response.Code
		want := http.StatusAccepted

		assertStatus(t, got, want)

	})
}

func TestGetUser(t *testing.T) {
	store := &StubCustomerStore{
		customers: []GetCustomerResponse{
			{
				FirstName:   "Peter",
				LastName:    "Smith",
				PhoneNumber: "+359 88 576 5981",
				Email:       "petesmith@gmail.com",
			},
			{
				FirstName:   "Alice",
				LastName:    "Johnson",
				PhoneNumber: "+359 88 444 2222",
				Email:       "alicejohn@gmail.com",
			},
		},
	}

	godotenv.Load("test.env")
	secretKey := []byte(os.Getenv("SECRET"))
	server := &CustomerServer{secretKey, store}

	t.Run("returns Peter's customer information", func(t *testing.T) {
		peterJWT, _ := generateJWT(secretKey, 0, time.Now().Add(time.Second))
		request := newGetCustomerRequest(peterJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		var got GetCustomerResponse
		json.NewDecoder(response.Body).Decode(&got)

		want := store.customers[0]

		assertStatus(t, response.Code, http.StatusOK)
		assertGetCustomerResponse(t, got, want)
	})

	t.Run("returns Alice's customer information", func(t *testing.T) {
		aliceJWT, _ := generateJWT(secretKey, 1, time.Now().Add(time.Second))
		request := newGetCustomerRequest(aliceJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		var got GetCustomerResponse
		json.NewDecoder(response.Body).Decode(&got)

		want := store.customers[1]

		assertStatus(t, response.Code, http.StatusOK)
		assertGetCustomerResponse(t, got, want)
	})

	t.Run("returns Unauthorized on expired JWT", func(t *testing.T) {
		expiredJWT, _ := generateJWT(secretKey, 0, time.Now())
		request := newGetCustomerRequest(expiredJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnauthorized)

	})

	t.Run("returns Unauthorized on invalid JWT", func(t *testing.T) {
		invalidJWT := "thisIsAnInvalidJWT"
		request := newGetCustomerRequest(invalidJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("returns Unauthorized on missing JWT", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/customer/", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("returns Not Found on nonexistent customer", func(t *testing.T) {
		noCustomerJWT, _ := generateJWT(secretKey, 3, time.Now().Add(time.Second))
		request := newGetCustomerRequest(noCustomerJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response.Code, http.StatusNotFound)
	})
}

func newGetCustomerRequest(jwt string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, "/customer/", nil)
	request.Header.Add("Token", jwt)

	return request
}

func generateJWT(secretKey []byte, subject int, expiresAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(int64(subject), 10),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	})

	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("got status %d, want %d", got, want)
	}
}

func assertGetCustomerResponse(t testing.TB, got, want GetCustomerResponse) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}
