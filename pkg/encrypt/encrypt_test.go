package encrypt

import (
	"strings"
	"testing"
)

func TestEncryptString(t *testing.T) {
	secureData := NewSecureData("{encrypted}")
	data := "Hello World"
	passphrase := "s3cure"
	encryptedData, err := secureData.EncryptString(data, passphrase)

	if err != nil {
		t.Errorf("Failed to encrypt string. expected nil, got %s", err)
	}

	if !strings.HasPrefix(encryptedData, "{encrypted}") {
		t.Error("Missing prefix in encrypted data")
	}

	decryptedData, err := secureData.DecryptString(encryptedData, passphrase)
	if err != nil {
		t.Errorf("Failed to decrypt data. expected nil, got %s", err)
	}

	if decryptedData != data {
		t.Errorf("Encryption round trip did not work. expected=%s, got=%s", data, decryptedData)
	}
}
