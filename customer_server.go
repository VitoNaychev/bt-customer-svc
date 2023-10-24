package bt_customer_svc

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
)

type CustomerStore interface {
	GetCustomerById(id int) (*Customer, error)
	GetCustomerByEmail(email string) (*Customer, error)
	StoreCustomer(customer Customer) int
	DeleteCustomer(id int) error
	UpdateCustomer(customer Customer) error
}

type CustomerServer struct {
	secretKey []byte
	expiresAt time.Time
	store     CustomerStore
	http.Handler
}

func NewCustomerServer(secretKey []byte, expiresAt time.Time, store CustomerStore, addressStore CustomerAddressStore) *CustomerServer {
	govalidator.TagMap["phonenumber"] = govalidator.Validator(func(str string) bool {
		matched, _ := regexp.Match(`^\+[\d ]+$`, []byte(str))
		return matched
	})

	c := new(CustomerServer)

	c.secretKey = secretKey
	c.expiresAt = expiresAt
	c.store = store

	router := http.NewServeMux()
	router.HandleFunc("/customer/", c.CustomerHandler)
	router.HandleFunc("/customer/login/", c.LoginHandler)

	c.Handler = router

	return c
}

type Customer struct {
	Id          int
	FirstName   string
	LastName    string
	PhoneNumber string
	Email       string
	Password    string
}

type GetCustomerResponse struct {
	FirstName   string
	LastName    string
	PhoneNumber string
	Email       string
}

type CreateCustomerRequest struct {
	FirstName   string `valid:"stringlength(2|20),required"`
	LastName    string `valid:"stringlength(2|20),required"`
	PhoneNumber string `valid:"phonenumber,required"`
	Email       string `valid:"email,required"`
	Password    string `valid:"stringlength(2|72),required"`
}

type LoginCustomerRequest struct {
	Email    string `valid:"email,required"`
	Password string `valid:"stringlength(2|72),required"`
}

type UpdateCustomerRequest struct {
	FirstName   string `valid:"stringlength(2|20),required"`
	LastName    string `valid:"stringlength(2|20),required"`
	PhoneNumber string `valid:"phonenumber,required"`
	Email       string `valid:"email,required"`
	Password    string `valid:"stringlength(2|72),required"`
}

type ErrorResponse struct {
	Message string
}

func (c *CustomerServer) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var loginCustomerRequest LoginCustomerRequest
	err := ValidateRequest(r.Body, &loginCustomerRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
		return
	}

	customer, err := c.store.GetCustomerByEmail(loginCustomerRequest.Email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Message: ErrMissingCustomer.Error()})
		return
	}

	if customer.Password != loginCustomerRequest.Password {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Message: ErrInvalidCredentials.Error()})
		return
	}

	loginJWT, _ := GenerateJWT(c.secretKey, c.expiresAt, customer.Id)

	w.WriteHeader(http.StatusAccepted)
	w.Header().Add("Token", loginJWT)
}

func (c *CustomerServer) CustomerHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		c.storeCustomer(w, r)
	case http.MethodGet:
		AuthenticationMiddleware(c.getCustomer, c.secretKey)(w, r)
	case http.MethodDelete:
		AuthenticationMiddleware(c.deleteCustomer, c.secretKey)(w, r)
	case http.MethodPut:
		AuthenticationMiddleware(c.updateCustomer, c.secretKey)(w, r)
	}
}

func (c *CustomerServer) updateCustomer(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.Header["Subject"][0])
	customer, _ := c.store.GetCustomerById(id)

	var updateCustomerRequest UpdateCustomerRequest
	err := ValidateRequest(r.Body, &updateCustomerRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
		return
	}

	customer.FirstName = updateCustomerRequest.FirstName
	customer.LastName = updateCustomerRequest.LastName
	customer.Email = updateCustomerRequest.Email
	customer.PhoneNumber = updateCustomerRequest.PhoneNumber
	customer.Password = updateCustomerRequest.Password

	c.store.UpdateCustomer(*customer)
}

func (c *CustomerServer) deleteCustomer(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.Header["Subject"][0])
	err := c.store.DeleteCustomer(id)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Message: ErrMissingCustomer.Error()})
	}
}

func (c *CustomerServer) storeCustomer(w http.ResponseWriter, r *http.Request) {
	var createCustomerRequest CreateCustomerRequest
	err := ValidateRequest(r.Body, &createCustomerRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
		return
	}

	if _, err := c.store.GetCustomerByEmail(createCustomerRequest.Email); err == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: ErrExistingUser.Error()})
		return
	}

	customer := getCustomerFromCreateCustomerRequest(createCustomerRequest)
	customerId := c.store.StoreCustomer(*customer)

	customerJWT, _ := GenerateJWT(c.secretKey, c.expiresAt, customerId)

	w.WriteHeader(http.StatusAccepted)
	w.Header().Add("Token", customerJWT)
}

func getCustomerFromCreateCustomerRequest(createCustomerRequest CreateCustomerRequest) *Customer {
	customer := &Customer{
		Id:          0,
		FirstName:   createCustomerRequest.FirstName,
		LastName:    createCustomerRequest.LastName,
		PhoneNumber: createCustomerRequest.PhoneNumber,
		Email:       createCustomerRequest.Email,
		Password:    createCustomerRequest.Password,
	}
	return customer
}

func (c *CustomerServer) getCustomer(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.Header["Subject"][0])
	customer, err := c.store.GetCustomerById(id)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Message: ErrMissingCustomer.Error()})
		return
	}

	getCustomerResponse := newGetCustomerResponse(*customer)
	json.NewEncoder(w).Encode(getCustomerResponse)

}

func newGetCustomerResponse(customer Customer) *GetCustomerResponse {
	getCustomerResponse := GetCustomerResponse{
		FirstName:   customer.FirstName,
		LastName:    customer.LastName,
		PhoneNumber: customer.PhoneNumber,
		Email:       customer.Email,
	}

	return &getCustomerResponse
}