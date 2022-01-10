package api

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"os"
)

func (s Server) DecryptPassword() (decrypted string, err error) {
	key, err := base64.StdEncoding.DecodeString(os.Getenv("DATABASE_BACKUP_KEY"))
	if err != nil {
		return
	}

	data, err := base64.RawStdEncoding.DecodeString(s.Password)
	if err != nil {
		return
	}

	if len(data) < 16 {
		err = errors.New("data must be at least 16 bytes")
		return
	}
	iv, msg := data[0:16], data[16:]

	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	cipher.NewCBCDecrypter(cipherBlock, iv).CryptBlocks(msg, msg)
	decrypted = string(msg)

	return
}
