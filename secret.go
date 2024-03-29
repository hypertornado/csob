package csob

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io"
	"io/ioutil"
	"time"
)

func loadKey(path string) (*rsa.PrivateKey, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(bytes)

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	err = key.Validate()
	if err != nil {
		return nil, err
	}

	return key, nil

}

func signData(key *rsa.PrivateKey, data string) (string, error) {
	signBytes, err := signDataSHA256(key, data)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signBytes), nil

}

func decrypt(key *rsa.PrivateKey, ciphertext string) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		panic(err)
	}
	decrypted, err := key.Decrypt(rand.Reader, cipherBytes, crypto.SHA256.New())
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(decrypted), nil
}

func signDataSHA256(key *rsa.PrivateKey, data string) ([]byte, error) {
	hash := sha256.New()
	io.WriteString(hash, data)
	sum := hash.Sum(nil)

	signed, err := key.Sign(rand.Reader, sum, crypto.SHA256)
	if err != nil {
		return []byte{}, err
	}
	return signed, nil
}

func timestamp() string {
	t := time.Now()
	return t.Format("20060102150405")
}
