package auth

import (
	"crypto/sha256"
	"os"
	"sync"
	"time"

	"menedger_paroley/crypto"
)

var (
	rememberedHash []byte
	rememberUntil  = time.Now()
	hashMutex      sync.RWMutex
)

func Verify(password string) bool {
	hashMutex.RLock()
	if time.Now().Before(rememberUntil) && constantTimeEqual(generateHash(password), rememberedHash) {
		hashMutex.RUnlock()
		return true
	}
	hashMutex.RUnlock()

	data, err := os.ReadFile("token.enc")
	if err != nil {
		return false
	}

	decrypted, err := crypto.Decrypt(data, []byte(password))
	if err != nil {
		return false
	}

	if string(decrypted) == "MASTER_PASSWORD_VERIFIED" {
		// Запоминаем на 10 минут
		hashMutex.Lock()
		rememberedHash = generateHash(password)
		rememberUntil = time.Now().Add(10 * time.Minute)
		hashMutex.Unlock()
		return true
	}

	return false
}

func Reset() {
	hashMutex.Lock()
	rememberedHash = nil
	rememberUntil = time.Time{}
	hashMutex.Unlock()
}

func generateHash(p string) []byte {
	h := sha256.Sum256([]byte(p))
	return h[:]
}

func constantTimeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

func SetMasterPassword(password string) error {
    data := []byte("MASTER_PASSWORD_VERIFIED")
    encrypted, err := crypto.Encrypt(data, []byte(password))
    if err != nil {
        return err
    }
    return os.WriteFile("token.enc", encrypted, 0600)
}