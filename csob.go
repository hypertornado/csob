package csob

import (
	"bytes"
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"
)

type CSOB struct {
	merchantId string
	key        *rsa.PrivateKey
}

func NewCSOB(merchantId, privateKeyPath string) *CSOB {
	key := loadKey(privateKeyPath)
	return &CSOB{
		merchantId,
		key,
	}
}

func (c *CSOB) Echo() error {
	dateStr := timestamp()
	strToSign := c.merchantId + "|" + dateStr
	signBytes, err := signData(c.key, strToSign)
	signature := base64.StdEncoding.EncodeToString(signBytes)

	params := map[string]interface{}{
		"dttm":       dateStr,
		"signature":  signature,
		"merchantId": c.merchantId,
	}

	url := "https://iapi.iplatebnibrana.csob.cz/api/v1.5/echo/"

	marshaled, err := json.MarshalIndent(params, " ", " ")
	if err != nil {
		panic(err)
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", url, ioutil.NopCloser(bytes.NewReader(marshaled)))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	fmt.Println("REQUEST:")
	oo, err := httputil.DumpRequest(req, true)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(oo))

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nRESPONSE:")
	out, err := httputil.DumpResponse(resp, true)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))

	return nil
}

func (c *CSOB) Init(price uint, title, desc string) {
	dateStr := timestamp()

	params := map[string]interface{}{
		"merchantId":   c.merchantId,
		"orderNo":      "1",
		"dttm":         dateStr,
		"payOperation": "payment",
		"payMethod":    "card",
		"totalAmount":  1,
		"currency":     "CZK",
		"closePayment": true,
		"returnUrl":    "https://example.com",
		"returnMethod": "POST",
		"cart": []interface{}{
			map[string]interface{}{
				"name":        "a",
				"quantity":    1,
				"amount":      1,
				"description": "a",
			},
		},
		"description":  "a",
		"merchantData": nil,
		"customerId":   nil,
		"language":     "CZ",
		"signature":    "",
	}

	signString := c.merchantId + "|1|" + dateStr + "|payment|card|1|CZK|true|https://example.com|POST|a|1|1|a|a|CZ"

	fmt.Println("STRING TO SIGN:\n" + signString + "\n")

	signed, err := signData(c.key, signString)
	if err != nil {
		panic(err)
	}

	params["signature"] = signed

	marshaled, err := json.MarshalIndent(params, " ", " ")
	if err != nil {
		panic(err)
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", "https://iapi.iplatebnibrana.csob.cz/api/v1.5/payment/init", ioutil.NopCloser(bytes.NewReader(marshaled)))
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	fmt.Println("REQUEST:")
	oo, err := httputil.DumpRequest(req, true)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(oo))

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nRESPONSE:")
	out, err := httputil.DumpResponse(resp, true)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))

}

func loadKey(path string) *rsa.PrivateKey {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(bytes)

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	err = key.Validate()
	if err != nil {
		panic(err)
	}

	return key

}

func signData(key *rsa.PrivateKey, data string) ([]byte, error) {
	return signDataSHA1(key, data)

}

func signDataSHA1(key *rsa.PrivateKey, data string) ([]byte, error) {
	hash := sha1.New()
	io.WriteString(hash, data)
	sum := hash.Sum(nil)

	signed, err := key.Sign(rand.Reader, sum, crypto.SHA1)
	if err != nil {
		return []byte{}, err
	}
	return signed, nil
}

func timestamp() string {
	t := time.Now()
	return t.Format("20060102150405")
}

func signDataMD5(key *rsa.PrivateKey, data string) ([]byte, error) {
	hash := md5.New()
	io.WriteString(hash, data)
	sum := hash.Sum(nil)

	signed, err := key.Sign(rand.Reader, sum, crypto.MD5)
	if err != nil {
		return []byte{}, err
	}
	return signed, nil
}
