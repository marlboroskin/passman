package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

const debug = false

// Encrypt шифрует данные с помощью пароля
func Encrypt(data, password []byte) ([]byte, error) {
	if debug {
		fmt.Println("🔐 DEBUG: Начало шифрования")
	}
	os.Stdout.Sync()

	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	key := pbkdf2.Key(password, salt, 100000, 32, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, data, nil)
	return append(append(salt, nonce...), ciphertext...), nil
}

// Decrypt расшифровывает данные с помощью пароля
func Decrypt(data, password []byte) ([]byte, error) {
	salt, nonceSize := data[:32], 12
	nonce, ciphertext := data[32:32+nonceSize], data[32+nonceSize:]

	key := pbkdf2.Key(password, salt, 100000, 32, sha256.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}
