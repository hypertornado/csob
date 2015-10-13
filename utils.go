package csob

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	version        = "v1.5"
	baseUrlTesting = "https://iapi.iplatebnibrana.csob.cz/api/" + version
	baseUrl        = "https://api.platebnibrana.csob.cz/api/" + version
	csobError      = errors.New("CSOB connection error")
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

func (c *CSOB) apiRequest(method, urlStr string, data map[string]interface{}) (resp *http.Response, err error) {
	urlStr = c.baseUrl() + urlStr

	marshaled, err := json.Marshal(data)
	if err != nil {
		return
	}

	req, err := http.NewRequest(method, urlStr, ioutil.NopCloser(bytes.NewReader(marshaled)))
	if err != nil {
		return resp, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err = c.client.Do(req)
	return
}
