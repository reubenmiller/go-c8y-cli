package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/crypto/scrypt"
)

// Source:
// https://itnext.io/encrypt-data-with-a-password-in-go-b5366384e291

type SecureData struct {
	Prefix []byte
}

// NewSecureData is used to securely read and write encrypted data
func NewSecureData(prefix string) *SecureData {
	return &SecureData{
		Prefix: []byte(prefix),
	}
}

// DeriveKey creates a secure key from a given password. It also accepts a salt which is used to increase the security of the key to prevent Rainbow tables attacks.
// If the salt is nil, then a secure salt will be generated using scrypt.
func (s *SecureData) DeriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}
	// 16384 iterations (ok for logins)
	key, err := scrypt.Key(password, salt, 16384, 8, 1, 32)
	if err != nil {
		return nil, nil, err
	}
	return key, salt, nil
}

// TryEncryptString encrypts a string only if it is not already encrypted
func (s *SecureData) TryEncryptString(data string, passphrase string) (string, error) {
	if len(s.Prefix) > 0 && strings.HasPrefix(data, string(s.Prefix)) {
		// string already encrypted
		return data, nil
	}
	return s.EncryptString(data, passphrase)
}

// EncryptString encryptes a string to a hex encoded string using a passphrase and salt.
// A prefix is also added to it (if defined)
func (s *SecureData) EncryptString(data string, passphrase string) (string, error) {
	encryptedData, err := s.Encrypt([]byte(data), passphrase)

	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s", s.Prefix, s.ToHexString(encryptedData)), nil
}

func (s *SecureData) ToHexString(data []byte) string {
	return fmt.Sprintf("%x", data)
}

func (s *SecureData) FromHex(data []byte) ([]byte, error) {
	decoded := make([]byte, 0)
	_, err := hex.Decode(decoded, data)
	return decoded, err
}
func (s *SecureData) FromHexString(data string) ([]byte, error) {
	decoded, err := hex.DecodeString(data)
	return decoded, err
}

// TryDecryptString tries to decrypt a string. If a encryption prefex string is used, then
// the data will only be decrypted if the input data has the prefix, otherwise it will be returned as is.
func (s *SecureData) TryDecryptString(data string, passphrase string) (string, error) {

	if len(s.Prefix) > 0 && !bytes.HasPrefix([]byte(data), s.Prefix) {
		return data, nil
	}

	decryptedData, err := s.DecryptString(data, passphrase)

	if err != nil {
		return "", err
	}
	return decryptedData, nil
}

func (s *SecureData) DecryptString(data string, passphrase string) (string, error) {

	if len(s.Prefix) > 0 {
		data = strings.TrimPrefix(data, string(s.Prefix))
	}

	decodedData, err := s.FromHexString(data)

	if err != nil {
		return "", err
	}

	v, err := s.Decrypt(decodedData, passphrase)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (s *SecureData) DecryptHex(data string, passphrase string) (string, error) {
	hexVal, err := hex.DecodeString(data)
	if err != nil {
		return "", err
	}
	return s.DecryptString(string(hexVal), passphrase)
}

// Encrypt encryptes the data with a passphrase. Salt is used on the passphrase and stored in the last 32 bytes of the data (appended unencrypted)
// Example: asdfadsfasdfasdfasdf<salt>
func (s *SecureData) Encrypt(data []byte, passphrase string) ([]byte, error) {
	key, salt, err := s.DeriveKey([]byte(passphrase), nil)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	ciphertext = append(ciphertext, salt...)

	return ciphertext, nil
}

// Decrypt decrypt the data. The data should be stored with a 32 byte salt which is append at the end of the data
func (s *SecureData) Decrypt(data []byte, passphrase string) ([]byte, error) {
	// Get salt from encrypted data
	if data == nil || len(data) <= 33 {
		return nil, fmt.Errorf("encrypted data is in an unexpected format")
	}
	salt, data := data[len(data)-32:], data[:len(data)-32]
	key, _, err := s.DeriveKey([]byte(passphrase), salt)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("encrypted data is in an unexpected format")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func (s *SecureData) EncryptFile(filename string, data []byte, passphrase string) {
	f, _ := os.Create(filename)
	defer f.Close()

	encryptedData, err := s.Encrypt(data, passphrase)

	if err != nil {
		panic("Failed to encrypt file")
	}
	f.Write(encryptedData)
}

func (s *SecureData) DecryptFile(filename string, passphrase string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return s.Decrypt(data, passphrase)
}
