package services

import (
	"crypto/rsa"
	"sync"

	"go.uber.org/zap"
)

var (
	decryptionPrivateKey *rsa.PrivateKey
	decryptionMutex      sync.RWMutex
)

// InitDecryption инициализирует расшифровку с предоставленным приватным ключом
func InitDecryption(privateKey *rsa.PrivateKey) {
	decryptionMutex.Lock()
	defer decryptionMutex.Unlock()

	if privateKey == nil {
		zap.L().Warn("Private key is nil - decryption disabled")
		return
	}

	decryptionPrivateKey = privateKey
}

// GetPrivateKey возвращает текущий приватный ключ для расшифровки
func GetPrivateKey() *rsa.PrivateKey {
	decryptionMutex.RLock()
	defer decryptionMutex.RUnlock()
	return decryptionPrivateKey
}

// IsDecryptionEnabled проверяет, включена ли расшифровка
func IsDecryptionEnabled() bool {
	decryptionMutex.RLock()
	defer decryptionMutex.RUnlock()
	return decryptionPrivateKey != nil
}
