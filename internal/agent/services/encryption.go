package services

import (
	"crypto/rsa"
	"errors"
	"sync"
)

var (
	encryptionPublicKey *rsa.PublicKey
	encryptionMutex     sync.RWMutex
)

// InitEncryption инициализирует шифрование с предоставленным публичным ключом
func InitEncryption(publicKey *rsa.PublicKey) error {
	encryptionMutex.Lock()
	defer encryptionMutex.Unlock()

	if publicKey == nil {
		return errors.New("public key cannot be nil")
	}

	encryptionPublicKey = publicKey
	return nil
}

// GetPublicKey возвращает текущий публичный ключ для шифрования
func GetPublicKey() *rsa.PublicKey {
	encryptionMutex.RLock()
	defer encryptionMutex.RUnlock()
	return encryptionPublicKey
}

// IsEncryptionEnabled проверяет, включено ли шифрование
func IsEncryptionEnabled() bool {
	encryptionMutex.RLock()
	defer encryptionMutex.RUnlock()
	return encryptionPublicKey != nil
}
