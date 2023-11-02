package unittest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/VitoNaychev/bt-customer-svc/handlers"
	"github.com/VitoNaychev/bt-customer-svc/handlers/address"
	"github.com/VitoNaychev/bt-customer-svc/handlers/auth"
	"github.com/VitoNaychev/bt-customer-svc/models"
	td "github.com/VitoNaychev/bt-customer-svc/tests/testdata"
	"github.com/VitoNaychev/bt-customer-svc/tests/testutil"
)

func TestUpdateCustomerAddress(t *testing.T) {
	addressData := []models.Address{td.PeterAddress1, td.PeterAddress2, td.AliceAddress}
	customerData := []models.Customer{td.PeterCustomer, td.AliceCustomer}
	stubAddressStore := testutil.NewStubAddressStore(addressData)
	stubCustomerStore := testutil.NewStubCustomerStore(customerData)
	server := address.NewCustomerAddressServer(stubAddressStore, stubCustomerStore, testEnv.SecretKey)

	t.Run("returns Unauthorized on invalid JWT", func(t *testing.T) {
		invalidJWT := "thisIsAnInvalidJWT"
		request, _ := http.NewRequest(http.MethodPut, "/customer/address/", nil)
		request.Header.Add("Token", invalidJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("updates address on valid body and credentials", func(t *testing.T) {
		updatedAddress := td.PeterAddress2
		updatedAddress.City = "Varna"

		peterJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, td.PeterCustomer.Id)

		request := address.NewUpdateAddressRequest(updatedAddress, peterJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertUpdatedAddress(t, stubAddressStore, updatedAddress)
	})

	t.Run("returns Bad Request on invalid request", func(t *testing.T) {
		invalidAddress := models.Address{}

		peterJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, td.PeterCustomer.Id)

		request := address.NewUpdateAddressRequest(invalidAddress, peterJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusBadRequest)
		testutil.AssertErrorResponse(t, response.Body, handlers.ErrInvalidRequestField)
	})

	t.Run("returns Not Found on missing customer", func(t *testing.T) {
		updatedAddress := td.PeterAddress2
		updatedAddress.City = "Varna"

		missingJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, 10)

		request := address.NewUpdateAddressRequest(updatedAddress, missingJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusNotFound)
		testutil.AssertErrorResponse(t, response.Body, handlers.ErrMissingCustomer)
	})

	t.Run("returns Not Found on missing address", func(t *testing.T) {
		updatedAddress := td.PeterAddress2
		updatedAddress.Id = 10

		missingJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, td.PeterCustomer.Id)

		request := address.NewUpdateAddressRequest(updatedAddress, missingJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusNotFound)
		testutil.AssertErrorResponse(t, response.Body, handlers.ErrMissingAddress)
	})

	t.Run("returns Unauthorized on update on another customer's address", func(t *testing.T) {
		updatedAddress := td.PeterAddress2
		updatedAddress.City = "Varna"

		peterJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, td.AliceCustomer.Id)

		request := address.NewUpdateAddressRequest(updatedAddress, peterJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusUnauthorized)
		testutil.AssertErrorResponse(t, response.Body, handlers.ErrUnathorizedAction)
	})
}

func TestDeleteCustomerAddress(t *testing.T) {
	addressData := []models.Address{td.PeterAddress1, td.PeterAddress2, td.AliceAddress}
	customerData := []models.Customer{td.PeterCustomer, td.AliceCustomer}
	stubAddressStore := testutil.NewStubAddressStore(addressData)
	stubCustomerStore := testutil.NewStubCustomerStore(customerData)
	server := address.NewCustomerAddressServer(stubAddressStore, stubCustomerStore, testEnv.SecretKey)

	t.Run("returns Unauthorized on invalid JWT", func(t *testing.T) {
		invalidJWT := "thisIsAnInvalidJWT"
		request := address.NewDeleteAddressRequest(invalidJWT, td.PeterAddress1.Id)
		request.Header.Add("Token", invalidJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("returns Bad Request on inavlid request", func(t *testing.T) {
		body := bytes.NewBuffer([]byte{})
		peterJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, td.PeterCustomer.Id)

		request, _ := http.NewRequest(http.MethodDelete, "/customer/address", body)
		request.Header.Add("Token", peterJWT)

		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("returns Not Found on missing user", func(t *testing.T) {
		missingJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, 10)

		request := address.NewDeleteAddressRequest(missingJWT, td.PeterAddress1.Id)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusNotFound)
		testutil.AssertErrorResponse(t, response.Body, handlers.ErrMissingCustomer)
	})

	t.Run("returns Not Found on missing address", func(t *testing.T) {
		peterJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, td.PeterCustomer.Id)

		request := address.NewDeleteAddressRequest(peterJWT, 10)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusNotFound)
		testutil.AssertErrorResponse(t, response.Body, handlers.ErrMissingAddress)
	})

	t.Run("returns Unathorized on delete on another customer's address", func(t *testing.T) {
		peterJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, td.PeterCustomer.Id)

		request := address.NewDeleteAddressRequest(peterJWT, td.AliceAddress.Id)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusUnauthorized)
		testutil.AssertErrorResponse(t, response.Body, handlers.ErrUnathorizedAction)
	})

	t.Run("deletes address on valid body and credentials", func(t *testing.T) {
		peterJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, td.PeterCustomer.Id)

		request := address.NewDeleteAddressRequest(peterJWT, td.PeterAddress1.Id)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertDeletedAddress(t, stubAddressStore, td.PeterAddress1)
	})
}

func TestSaveCustomerAddress(t *testing.T) {
	addressData := []models.Address{}
	customerData := []models.Customer{td.PeterCustomer, td.AliceCustomer}
	stubAddressStore := testutil.NewStubAddressStore(addressData)
	stubCustomerStore := testutil.NewStubCustomerStore(customerData)
	server := address.NewCustomerAddressServer(stubAddressStore, stubCustomerStore, testEnv.SecretKey)

	t.Run("returns Unauthorized on invalid JWT", func(t *testing.T) {
		invalidJWT := "thisIsAnInvalidJWT"
		request := address.NewAddAddressRequest(invalidJWT, td.AliceAddress)
		request.Header.Add("Token", invalidJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("returns Bad Request on inavlid request", func(t *testing.T) {
		peterJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, td.PeterCustomer.Id)

		request := address.NewAddAddressRequest(peterJWT, models.Address{})
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusBadRequest)
	})

	t.Run("returns Not Found on missing user", func(t *testing.T) {
		missingJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, 10)

		request := address.NewAddAddressRequest(missingJWT, td.AliceAddress)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("saves Peter's new address", func(t *testing.T) {
		peterJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, td.PeterCustomer.Id)

		request := address.NewAddAddressRequest(peterJWT, td.PeterAddress1)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertStoredAddress(t, stubAddressStore, td.PeterAddress1)
	})

	t.Run("saves Alice's new address", func(t *testing.T) {
		stubAddressStore.Empty()

		aliceJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, td.AliceCustomer.Id)

		request := address.NewAddAddressRequest(aliceJWT, td.AliceAddress)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertStoredAddress(t, stubAddressStore, td.AliceAddress)
	})
}

func TestGetCustomerAddress(t *testing.T) {
	addressData := []models.Address{td.PeterAddress1, td.PeterAddress2, td.AliceAddress}
	customerData := []models.Customer{td.PeterCustomer, td.AliceCustomer}
	stubAddressStore := testutil.NewStubAddressStore(addressData)
	stubCustomerStore := testutil.NewStubCustomerStore(customerData)
	server := address.NewCustomerAddressServer(stubAddressStore, stubCustomerStore, testEnv.SecretKey)

	t.Run("returns Peter's addresses", func(t *testing.T) {
		peterJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, td.PeterCustomer.Id)
		request := address.NewGetAddressRequest(peterJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusOK)

		want := []address.GetAddressResponse{
			address.AddressToGetAddressResponse(td.PeterAddress1),
			address.AddressToGetAddressResponse(td.PeterAddress2),
		}
		var got []address.GetAddressResponse
		json.NewDecoder(response.Body).Decode(&got)

		assertAddresses(t, got, want)
	})

	t.Run("returns Alice's addresses", func(t *testing.T) {
		aliceJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, td.AliceCustomer.Id)
		request := address.NewGetAddressRequest(aliceJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusOK)

		want := []address.GetAddressResponse{
			address.AddressToGetAddressResponse(td.AliceAddress),
		}
		var got []address.GetAddressResponse
		json.NewDecoder(response.Body).Decode(&got)

		assertAddresses(t, got, want)
	})

	t.Run("returns Unauthorized on invalid JWT", func(t *testing.T) {
		invalidJWT := "thisIsAnInvalidJWT"
		request := address.NewGetAddressRequest(invalidJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("returns Not Found on missing user", func(t *testing.T) {
		aliceJWT, _ := auth.GenerateJWT(testEnv.SecretKey, testEnv.ExpiresAt, 10)
		request := address.NewGetAddressRequest(aliceJWT)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusNotFound)
		testutil.AssertErrorResponse(t, response.Body, handlers.ErrMissingCustomer)
	})
}

func assertAddresses(t testing.TB, got, want []address.GetAddressResponse) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, td.AliceAddress)
	}
}