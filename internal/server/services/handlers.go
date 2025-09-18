package services

import (
	"crypto/rsa"

	"github.com/MPoline/alert_service_yp/internal/storage"
)

type ServiceHandler struct {
	storage    storage.Storage
	privateKey *rsa.PrivateKey
	key        string
}

func NewServiceHandler(storage storage.Storage, privateKey *rsa.PrivateKey, key string) *ServiceHandler {
	return &ServiceHandler{
		storage:    storage,
		privateKey: privateKey,
		key:        key,
	}
}
