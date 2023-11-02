package customer

import (
	"net/http"
	"time"

	"github.com/VitoNaychev/bt-customer-svc/handlers/auth"
	"github.com/VitoNaychev/bt-customer-svc/models"
)

type CustomerServer struct {
	secretKey []byte
	expiresAt time.Duration
	store     models.CustomerStore
	http.Handler
}

func NewCustomerServer(secretKey []byte, expiresAt time.Duration, store models.CustomerStore) *CustomerServer {
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

func (c *CustomerServer) CustomerHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		c.createCustomer(w, r)
	case http.MethodGet:
		auth.AuthenticationMiddleware(c.getCustomer, c.secretKey)(w, r)
	case http.MethodDelete:
		auth.AuthenticationMiddleware(c.deleteCustomer, c.secretKey)(w, r)
	case http.MethodPut:
		auth.AuthenticationMiddleware(c.updateCustomer, c.secretKey)(w, r)
	}
}