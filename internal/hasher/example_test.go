package hasher_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/MPoline/alert_service_yp/internal/hasher"
)

func ExampleInitHasher() {
	h := hasher.InitHasher("sha256")

	data := []byte("test data")
	key := []byte("secret key")
	hash, err := h.CalculateHash(data, key)
	if err != nil {
		fmt.Println("Hash calculation error:", err)
		return
	}

	fmt.Printf("Calculated hash: %x\n", hash)
}

func ExampleNewSHA265Hasher() {
	h := hasher.NewSHA265Hasher()

	hash, err := h.CalculateHash([]byte("data"), []byte("key"))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("HMAC-SHA256 hash: %x\n", hash)
}

func TestHasher(t *testing.T) {
	t.Run("Valid input", func(t *testing.T) {
		h := hasher.NewSHA265Hasher()
		hash, err := h.CalculateHash([]byte("test"), []byte("key"))
		if err != nil {
			t.Fatalf("CalculateHash failed: %v", err)
		}
		if len(hash) != sha256.Size {
			t.Errorf("Expected hash length %d, got %d", sha256.Size, len(hash))
		}
	})

	t.Run("Empty data", func(t *testing.T) {
		h := hasher.NewSHA265Hasher()
		_, err := h.CalculateHash([]byte(""), []byte("key"))
		if err == nil {
			t.Error("Expected error for empty data")
		}
	})

	t.Run("Empty key", func(t *testing.T) {
		h := hasher.NewSHA265Hasher()
		_, err := h.CalculateHash([]byte("data"), []byte(""))
		if err == nil {
			t.Error("Expected error for empty key")
		}
	})
}
