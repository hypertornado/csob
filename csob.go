package csob

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

//TODO: overit string
//http://localhost:8585/darkovy-poukaz/zaplacene?payId=4860adcd41654b3&dttm=20151102133207&resultCode=0&resultMessage=OK&paymentStatus=3&signature=tduLxxMIfXGTeBj36Zny5kpnz63ACxT4PDvaZMhSimh%2BGLHtBDMJmMuePbw5U9RzdaCSkmAHY%2B8uvVzgl5xtfgTNlmm1xUTJphWrGkaC4rR1u9hTXKAmZyWuS0hvmq6ej1mh4bzMoGEYYl9420xHGCteVMAO4eU%2FcVrzMZ9f86AaW3RUkPl0%2F6vKuqgwaXQQixLjXNEJrqoDTtzzspKMLIygHFgCMSoCKMR8hSOLE71QYKFS6PckP4IRgV%2FxEFzP7NKEDvzt%2FYGOtgkFauRTlqni%2BxkCZESOpg5Me7lP9tl504O5qS8Lk7%2BTB2%2BOUn8ZOsDPzDvOCjFUmJqorVmxDg%3D%3D

/*func NewCSOBTestingEnvironment(merchantId, privateKeyPath string) (*CSOB, error) {
	csob, err := NewCSOB(merchantId, privateKeyPath)
	if err == nil {
		csob.testingEnvironment = true
	}
	csob.ReturnUrl("GET", "http://www.example.com")
	return csob, err
}*/

func HumanReadableStatusCzech(id int) string {
	switch id {
	case 1:
		return "Platba založena"
	case 2:
		return "Platba probíhá"
	case 3:
		return "Platba zrušena"
	case 4:
		return "Platba potvrzena"
	case 5:
		return "Platba odvolána"
	case 6:
		return "Platba zamítnuta"
	case 7:
		return "Čekání na zůčtování"
	case 8:
		return "Platba zůčtována"
	case 9:
		return "Zpracování vrácení"
	case 10:
		return "Platba vrácena"
	}
	return "Žádný stav"
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

func (c *CSOB) TestingEnvironment() {
	c.testingEnvironment = true
}

func (c *CSOB) EchoGet() error {
	dttm := timestamp()
	signature, err := c.sign(c.merchantId, dttm)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/echo/%s/%s/%s",
		c.merchantId,
		dttm,
		url.QueryEscape(signature),
	)
	resp, err := c.apiRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return csobError
	}

	return nil
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

type PaymentStatus struct {
	PayId         string `json:"payId"`
	Dttm          string `json:"dttm"`
	ResultCode    int    `json:"resultCode"`
	ResultMessage string `json:"resultMessage"`
	PaymentStatus int    `json:"paymentStatus"`
	AuthCode      string `json:"authCode"`
	Signature     string `json:"signature"`
}

func (c *CSOB) ProcessURL(paymentStatus *PaymentStatus) (string, error) {
	dttm := timestamp()
	signature, err := c.sign(c.merchantId, paymentStatus.PayId, dttm)
	if err != nil {
		return "", err
	}

	ret := fmt.Sprintf("%s/payment/process/%s/%s/%s/%s",
		c.baseUrl(),
		c.merchantId,
		paymentStatus.PayId,
		dttm,
		url.QueryEscape(signature),
	)
	return ret, nil
}

func (c *CSOB) Status(payId string) (*PaymentStatus, error) {
	return c.paymentStatusTypeCall(payId, "GET", "payment/status")
}

func (c *CSOB) Reverse(payId string) (*PaymentStatus, error) {
	return c.paymentStatusTypeCall(payId, "PUT", "payment/reverse")
}

func (c *CSOB) Close(payId string) (*PaymentStatus, error) {
	return c.paymentStatusTypeCall(payId, "PUT", "payment/close")
}

func (c *CSOB) Refund(payId string) (*PaymentStatus, error) {
	return c.paymentStatusTypeCall(payId, "PUT", "payment/refund")
}

type CSOB struct {
	merchantId         string
	key                *rsa.PrivateKey
	testingEnvironment bool
	client             *http.Client
	returnUrl          string
	returnMethod       string
}

func (c *CSOB) ReturnUrl(returnMethod, returnUrl string) {
	c.returnMethod = returnMethod
	c.returnUrl = returnUrl
}

type orderItem struct {
	name     string
	quantity uint
	amount   uint
}

type order struct {
	orderNo      uint
	name         string
	description  string
	quantity     uint
	amount       uint
	closePayment bool
	orderItems   []orderItem
	payOperation string
	payMethod    string
	language     string
	currency     string
}

func (c *CSOB) NewOrder(orderNo uint, name, description string) *order {
	return &order{
		orderNo:      orderNo,
		name:         name,
		description:  description,
		closePayment: false,
		orderItems:   []orderItem{},
		payOperation: "payment",
		payMethod:    "card",
		language:     "EN",
		currency:     "USD",
	}
}

func (o *order) Close() {
	o.closePayment = true
}

func (o *order) Language(language string) {
	o.language = language
}

func (o *order) Currency(currency string) {
	o.currency = currency
}

func (o *order) AddItem(name string, quantity, amount uint) {
	var item = orderItem{
		name:     name,
		quantity: quantity,
		amount:   amount,
	}
	o.orderItems = append(o.orderItems, item)

}

func (c *CSOB) Init(order *order) (*PaymentStatus, error) {

	if len(order.orderItems) != 1 {
		return nil, errors.New("there should be 1 order item")
	}

	amountStr := fmt.Sprintf("%d", order.orderItems[0].amount)
	quantityStr := fmt.Sprintf("%d", order.orderItems[0].quantity)

	total := amountStr

	closePaymentStr := "true"
	if order.closePayment {
		closePaymentStr = "false"
	}

	orderNoStr := fmt.Sprintf("%d", order.orderNo)

	params := map[string]interface{}{
		"merchantId":   c.merchantId,
		"orderNo":      orderNoStr,
		"dttm":         timestamp(),
		"payOperation": order.payOperation,
		"payMethod":    order.payMethod,
		"totalAmount":  total,
		"currency":     order.currency,
		"closePayment": closePaymentStr,
		"returnUrl":    c.returnUrl,
		"returnMethod": c.returnMethod,
		"cart": []interface{}{
			map[string]interface{}{
				"name":     order.orderItems[0].name,
				"quantity": quantityStr,
				"amount":   amountStr,
			},
		},
		"description":  order.description,
		"merchantData": nil,
		"customerId":   nil,
		"language":     order.language,
		"signature":    "",
	}

	signature, err := c.sign(
		c.merchantId,
		orderNoStr,
		params["dttm"],
		order.payOperation,
		order.payMethod,
		total,
		order.currency,
		closePaymentStr,
		c.returnUrl,
		c.returnMethod,
		order.orderItems[0].name,
		quantityStr,
		amountStr,
		order.description,
		order.language,
	)
	if err != nil {
		return nil, err
	}

	params["signature"] = signature

	resp, err := c.apiRequest("POST", "/payment/init", params)
	if err != nil {
		return nil, err
	}

	return parseStatusResponse(resp)

}
