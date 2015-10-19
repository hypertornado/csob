package csob

import (
	"crypto/rsa"
	"net/http"
)

type CSOB struct {
	merchantId         string
	key                *rsa.PrivateKey
	testingEnvironment bool
	client             *http.Client
}

func NewCSOBTestingEnvironment(merchantId, privateKeyPath string) (*CSOB, error) {
	csob, err := NewCSOB(merchantId, privateKeyPath)
	if err == nil {
		csob.testingEnvironment = true
	}
	return csob, err
}

func NewCSOB(merchantId, privateKeyPath string) (*CSOB, error) {
	key, err := loadKey(privateKeyPath)
	if err != nil {
		return nil, err
	}
	return &CSOB{
		merchantId:         merchantId,
		key:                key,
		testingEnvironment: false,
		client:             &http.Client{},
	}, nil
}

func (c *CSOB) Echo() error {
	dateStr := timestamp()

	signature, err := c.sign(c.merchantId, dateStr)
	if err != nil {
		return err
	}

	params := map[string]interface{}{
		"dttm":       dateStr,
		"signature":  signature,
		"merchantId": c.merchantId,
	}

	resp, err := c.apiRequest("POST", "/echo", params)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return csobError
	}

	return nil
}

func (c *CSOB) Init(orderNo, name string, quantity, amount uint, description string) error {
	params := map[string]interface{}{
		"merchantId":   c.merchantId,
		"orderNo":      orderNo,
		"dttm":         timestamp(),
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

	signature, err := c.sign(
		c.merchantId,
		params["orderNo"],
		params["dttm"],
		"payment",
		"card",
		"1",
		"CZK",
		"true",
		"https://example.com",
		"POST",
		"a",
		"1",
		"1",
		"a",
		"a",
		"CZ",
	)
	if err != nil {
		return err
	}

	params["signature"] = signature

	resp, err := c.apiRequest("POST", "/payment/init", params)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return csobError
	}

	return nil
}
