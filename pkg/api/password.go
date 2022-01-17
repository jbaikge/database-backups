package api

import (
	"encoding/hex"
	"fmt"
	"os"

	"golang.org/x/crypto/nacl/secretbox"
)

func (s Server) DecryptPassword() (decrypted string, err error) {
	var key [32]byte
	var nonce [24]byte

	keyRaw, err := hex.DecodeString(os.Getenv("DATABASE_BACKUP_KEY"))
	if err != nil {
		err = fmt.Errorf("decoding DATABASE_BACKUP_KEY: %s", err)
		return
	}
	copy(key[:], keyRaw)

	encrypted, err := hex.DecodeString(s.Password)
	if err != nil {
		err = fmt.Errorf("decoding password: %s", err)
		return
	}
	copy(nonce[:], encrypted[:24])

	decryptedBytes, ok := secretbox.Open(nil, encrypted[24:], &nonce, &key)
	if !ok {
		err = fmt.Errorf("decryption error")
		return
	}
	decrypted = string(decryptedBytes)

	return
}
