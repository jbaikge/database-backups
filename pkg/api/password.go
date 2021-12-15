package api

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"os"
)

func (s Server) DecryptPassword() (decrypted string, err error) {
	iv, err := base64.StdEncoding.DecodeString(os.Getenv("DATABASE_BACKUP_IV"))
	if err != nil {
		return
	}

	key, err := base64.StdEncoding.DecodeString(os.Getenv("DATABASE_BACKUP_KEY"))
	if err != nil {
		return
	}

	data, err := base64.RawStdEncoding.DecodeString(s.Password)
	if err != nil {
		return
	}

	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	cipher.NewCBCDecrypter(cipherBlock, iv).CryptBlocks(data, data)
	decrypted = string(data)

	return
}
