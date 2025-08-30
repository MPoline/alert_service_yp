// Пакет crypto предоставляет функции для работы с асимметричным шифрованием RSA
package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"
)

var OAEPHash = sha256.New()
var OAEPLabel []byte = nil

const (
	ChunkProtocolVersion uint8 = 1
	ChunkHeaderSize      int   = 5 // version(1) + chunk_count(4)
	ChunkSizeHeaderSize  int   = 4 // uint32 для размера каждого chunk
)

// ChunkProtocol представляет зашифрованные данные с метаинформацией
type ChunkProtocol struct {
	Version    uint8
	ChunkCount uint32
	ChunkSizes []uint32
	ChunkData  [][]byte
}

// LoadPublicKey загружает публичный ключ из файла
// filename - путь к файлу с публичным ключом
// Возвращает *rsa.PublicKey или ошибку
func LoadPublicKey(filename string) (*rsa.PublicKey, error) {
	if filename == "" {
		return nil, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	if block.Type != "PUBLIC KEY" && block.Type != "RSA PUBLIC KEY" {
		return nil, errors.New("unexpected key type: " + block.Type)
	}

	var pub interface{}
	pub, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		pub, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPub, nil
}

// LoadPrivateKey загружает приватный ключ из файла
// filename - путь к файлу с приватным ключом
// Возвращает *rsa.PrivateKey или ошибку
func LoadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	if filename == "" {
		return nil, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the private key")
	}

	if block.Type != "PRIVATE KEY" && block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("unexpected key type: " + block.Type)
	}

	var priv *rsa.PrivateKey
	priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		var ok bool
		priv, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
	}

	return priv, nil
}

// EncryptOAEP шифрует данные с использованием RSA-OAEP
func EncryptOAEP(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	if publicKey == nil {
		return data, nil // Режим без шифрования
	}

	maxSize := publicKey.Size() - 2*OAEPHash.Size() - 2
	if len(data) > maxSize {
		return nil, fmt.Errorf("data too large for RSA-OAEP encryption: %d > %d",
			len(data), maxSize)
	}

	encrypted, err := rsa.EncryptOAEP(OAEPHash, rand.Reader, publicKey, data, OAEPLabel)
	if err != nil {
		return nil, fmt.Errorf("OAEP encryption failed: %w", err)
	}

	return encrypted, nil
}

// DecryptOAEP расшифровывает данные с использованием RSA-OAEP
func DecryptOAEP(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	if privateKey == nil {
		return data, nil // Режим без расшифровки
	}

	if len(data) != privateKey.PublicKey.Size() {
		return nil, errors.New("encrypted data size doesn't match key size")
	}

	decrypted, err := rsa.DecryptOAEP(OAEPHash, rand.Reader, privateKey, data, OAEPLabel)
	if err != nil {
		return nil, fmt.Errorf("OAEP decryption failed: %w", err)
	}

	return decrypted, nil
}

// EncryptLargeData шифрует большие данные с использованием chunk protocol
func EncryptLargeData(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	if publicKey == nil {
		return data, nil
	}

	hashSize := OAEPHash.Size()
	maxChunkSize := publicKey.Size() - 2*hashSize - 2

	zap.L().Debug("Chunk encryption parameters",
		zap.Int("key_size", publicKey.Size()),
		zap.Int("hash_size", hashSize),
		zap.Int("max_chunk_size", maxChunkSize),
		zap.Int("data_size", len(data)))

	if len(data) <= maxChunkSize {
		zap.L().Debug("Data fits in single chunk, encrypting directly")
		return EncryptOAEP(publicKey, data)
	}

	zap.L().Debug("Data requires multiple chunks, creating chunk protocol")
	var encryptedChunks [][]byte
	var chunkSizes []uint32

	for i := 0; i < len(data); i += maxChunkSize {
		end := i + maxChunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[i:end]
		encryptedChunk, err := EncryptOAEP(publicKey, chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt chunk %d: %w", len(encryptedChunks), err)
		}

		encryptedChunks = append(encryptedChunks, encryptedChunk)
		chunkSizes = append(chunkSizes, uint32(len(encryptedChunk)))
	}

	protocolData, err := createChunkProtocol(encryptedChunks, chunkSizes)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunk protocol: %w", err)
	}

	zap.L().Debug("Chunk protocol created",
		zap.Int("chunks_count", len(encryptedChunks)),
		zap.Int("protocol_size", len(protocolData)))

	return protocolData, nil
}

// createChunkProtocol создает бинарный protocol для chunks
func createChunkProtocol(chunks [][]byte, chunkSizes []uint32) ([]byte, error) {
	if len(chunks) != len(chunkSizes) {
		return nil, errors.New("chunks and sizes arrays must have same length")
	}

	chunkCount := uint32(len(chunks))

	totalSize := ChunkHeaderSize + (len(chunkSizes) * ChunkSizeHeaderSize)
	for _, chunk := range chunks {
		totalSize += len(chunk)
	}

	buf := make([]byte, 0, totalSize)
	protocolBuffer := bytes.NewBuffer(buf)

	if err := binary.Write(protocolBuffer, binary.BigEndian, ChunkProtocolVersion); err != nil {
		return nil, fmt.Errorf("failed to write version: %w", err)
	}
	if err := binary.Write(protocolBuffer, binary.BigEndian, chunkCount); err != nil {
		return nil, fmt.Errorf("failed to write chunk count: %w", err)
	}

	for _, size := range chunkSizes {
		if err := binary.Write(protocolBuffer, binary.BigEndian, size); err != nil {
			return nil, fmt.Errorf("failed to write chunk size: %w", err)
		}
	}

	for _, chunk := range chunks {
		if _, err := protocolBuffer.Write(chunk); err != nil {
			return nil, fmt.Errorf("failed to write chunk data: %w", err)
		}
	}

	return protocolBuffer.Bytes(), nil
}

// parseChunkProtocol парсит бинарный protocol chunks
func parseChunkProtocol(data []byte) (*ChunkProtocol, error) {
	if len(data) < ChunkHeaderSize {
		return nil, errors.New("protocol data too short")
	}

	protocolBuffer := bytes.NewReader(data)
	var protocol ChunkProtocol

	if err := binary.Read(protocolBuffer, binary.BigEndian, &protocol.Version); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}
	if protocol.Version != ChunkProtocolVersion {
		return nil, fmt.Errorf("unsupported protocol version: %d", protocol.Version)
	}

	if err := binary.Read(protocolBuffer, binary.BigEndian, &protocol.ChunkCount); err != nil {
		return nil, fmt.Errorf("failed to read chunk count: %w", err)
	}

	protocol.ChunkSizes = make([]uint32, protocol.ChunkCount)
	for i := uint32(0); i < protocol.ChunkCount; i++ {
		if err := binary.Read(protocolBuffer, binary.BigEndian, &protocol.ChunkSizes[i]); err != nil {
			return nil, fmt.Errorf("failed to read chunk size %d: %w", i, err)
		}
	}

	protocol.ChunkData = make([][]byte, protocol.ChunkCount)
	for i := uint32(0); i < protocol.ChunkCount; i++ {
		chunkSize := int(protocol.ChunkSizes[i])
		protocol.ChunkData[i] = make([]byte, chunkSize)

		if n, err := protocolBuffer.Read(protocol.ChunkData[i]); err != nil || n != chunkSize {
			return nil, fmt.Errorf("failed to read chunk %d data: %w", i, err)
		}
	}

	return &protocol, nil
}

// DecryptLargeData расшифровывает данные, используя chunk protocol
func DecryptLargeData(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	if privateKey == nil {
		return data, nil
	}

	zap.L().Debug("Decrypting data",
		zap.Int("data_size", len(data)),
		zap.Int("key_size", privateKey.Size()))

	if IsChunkProtocol(data) {
		zap.L().Debug("Detected chunk protocol format")
		protocol, err := parseChunkProtocol(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse chunk protocol: %w", err)
		}

		var decryptedData []byte
		for i, encryptedChunk := range protocol.ChunkData {
			if len(encryptedChunk) != privateKey.Size() {
				return nil, fmt.Errorf("chunk %d size %d doesn't match key size %d",
					i, len(encryptedChunk), privateKey.Size())
			}

			decryptedChunk, err := DecryptOAEP(privateKey, encryptedChunk)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt chunk %d: %w", i, err)
			}
			decryptedData = append(decryptedData, decryptedChunk...)
		}

		zap.L().Debug("Chunk protocol decrypted successfully",
			zap.Int("chunks_count", len(protocol.ChunkData)),
			zap.Int("decrypted_size", len(decryptedData)))

		return decryptedData, nil
	}

	expectedSize := privateKey.Size()
	if len(data) == expectedSize {
		zap.L().Debug("Detected single encrypted chunk")
		return DecryptOAEP(privateKey, data)
	}

	if len(data) > expectedSize && len(data)%expectedSize == 0 {
		zap.L().Debug("Attempting manual chunk splitting")
		chunkCount := len(data) / expectedSize
		var decryptedData []byte

		for i := 0; i < chunkCount; i++ {
			start := i * expectedSize
			end := start + expectedSize
			chunk := data[start:end]

			decryptedChunk, err := DecryptOAEP(privateKey, chunk)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt manual chunk %d: %w", i, err)
			}
			decryptedData = append(decryptedData, decryptedChunk...)
		}

		zap.L().Debug("Manual chunk decryption successful",
			zap.Int("chunks_count", chunkCount),
			zap.Int("decrypted_size", len(decryptedData)))

		return decryptedData, nil
	}

	return nil, fmt.Errorf("invalid encrypted data format: size %d, expected %d or multiple thereof",
		len(data), expectedSize)
}

func IsChunkProtocol(data []byte) bool {
	if len(data) < ChunkHeaderSize {
		return false
	}

	if data[0] != ChunkProtocolVersion {
		return false
	}

	if len(data) < ChunkHeaderSize+ChunkSizeHeaderSize {
		return false
	}

	return true
}
