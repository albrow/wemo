package wemo

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var binaryStateRegex = regexp.MustCompile(`<BinaryState>(0|1)</BinaryState>`)

// State represents the state of the WeMo switch and is either On or Off.
type State uint

const (
	Off State = iota
	On
)

func (s State) String() string {
	if s == On {
		return "on"
	}
	return "off"
}

// GetState returns the state of the WeMo switch.
func GetState() (State, error) {
	requestBody := `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
  <s:Body>
    <u:GetBinaryState xmlns:u="urn:Belkin:service:basicevent:1"></u:GetBinaryState>
  </s:Body>
</s:Envelope>`
	responseBody, err := post("GetBinaryState", requestBody)
	if err != nil {
		return 0, err
	}
	stateStr := binaryStateRegex.FindStringSubmatch(string(responseBody))
	if len(stateStr) == 0 {
		return 0, fmt.Errorf("Unexpected response: %s", string(responseBody))
	}
	stateInt, err := strconv.Atoi(stateStr[1])
	if err != nil {
		return 0, err
	}
	return State(uint(stateInt)), nil
}

// SetState sets the state of the WeMo switch.
func SetState(state State) error {
	requestBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
  <s:Body>
    <u:SetBinaryState xmlns:u="urn:Belkin:service:basicevent:1">
    	<BinaryState>%d</BinaryState>
    </u:SetBinaryState>
  </s:Body>
</s:Envelope>`, state)
	if _, err := post("SetBinaryState", requestBody); err != nil {
		return err
	}
	return nil
}

func post(action string, body string) ([]byte, error) {
	wemoHost := os.Getenv("WEMO_HOST")
	if wemoHost == "" {
		return nil, errors.New("Missing required env var: WEMO_HOST")
	}
	url := "http://" + wemoHost + "/upnp/control/basicevent1"
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("SOAPACTION", fmt.Sprintf("\"urn:Belkin:service:basicevent:1#%s\"", action))
	req.Header.Add("Content-type", "text/xml")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}
