package csob

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	version        = "v1.9"
	baseUrlTesting = "https://iapi.iplatebnibrana.csob.cz/api/" + version
	baseUrl        = "https://api.platebnibrana.csob.cz/api/" + version
	//csobError      = errors.New("CSOB connection error")
	Debug = false
)

func (c *CSOB) baseUrl() string {
	if c.testingEnvironment {
		return baseUrlTesting
	} else {
		return baseUrl
	}
}

func (c *CSOB) sign(args ...interface{}) (string, error) {
	ret := ""
	for i := 0; i < len(args); i++ {
		if i > 0 {
			ret += "|"
		}
		ret += fmt.Sprintf("%s", args[i])
	}
	return signData(c.key, ret)
}

func (c *CSOB) paymentStatusTypeCall(payId string, method, urlFragment string) (*PaymentStatus, error) {
	dttm := timestamp()
	signature, err := c.sign(c.merchantId, payId, dttm)
	if err != nil {
		return nil, err
	}

	urlStr := fmt.Sprintf("/%s/%s/%s/%s/%s",
		urlFragment,
		url.QueryEscape(c.merchantId),
		url.QueryEscape(payId),
		dttm,
		url.QueryEscape(signature),
	)

	resp, err := c.apiRequest(method, urlStr, nil)
	if err != nil {
		return nil, err
	}

	return parseStatusResponse(resp)
}

func (c *CSOB) paymentStatusTypePutCall(payId string, urlFragment string) error {
	data := map[string]interface{}{
		"merchantId": c.merchantId,
		"payId":      payId,
		"dttm":       timestamp(),
	}

	signature, err := c.sign(data["merchantId"], data["payId"], data["dttm"])
	if err != nil {
		return err
	}

	data["signature"] = signature

	resp, err := c.apiRequest("PUT", "/payment/"+urlFragment, data)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("CSOB cant' put call, code: %d", resp.StatusCode)
	}

	return nil
}

func (c *CSOB) apiRequest(method, urlStr string, data map[string]interface{}) (resp *http.Response, err error) {
	urlStr = c.baseUrl() + urlStr

	marshaled, err := json.Marshal(data)
	if err != nil {
		return
	}

	if Debug {
		fmt.Println("-----")
		fmt.Println(method, " ", urlStr)
		fmt.Println(string(marshaled))
	}

	req, err := http.NewRequest(method, urlStr, ioutil.NopCloser(bytes.NewReader(marshaled)))
	if err != nil {
		return resp, err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err = c.client.Do(req)

	return
}

func parseStatusResponse(response *http.Response) (*PaymentStatus, error) {
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("CSOB cant' parse status, code: %d", response.StatusCode)
	}

	respBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var paymentStatus PaymentStatus
	err = json.Unmarshal(respBytes, &paymentStatus)
	return &paymentStatus, err
}
