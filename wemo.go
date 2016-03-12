package wemo

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Switch represents a WeMo switch.
type Switch struct {
	host string
}

var binaryStateRegex = regexp.MustCompile(`<BinaryState>(0|1)</BinaryState>`)

type state uint

const (
	off state = iota
	on
)

func (s state) String() string {
	if s == on {
		return "on"
	}
	return "off"
}

// NewSwitch creates and returns a new switch. host should be the IP address and
// port for the switch.
func NewSwitch(host string) *Switch {
	return &Switch{
		host: host,
	}
}

// IsOn returns true if the switch is currently on and false otherwise.
func (s *Switch) IsOn() (bool, error) {
	switchState, err := s.getState()
	if err != nil {
		return false, err
	}
	return switchState == on, nil
}

// TurnOff turns the switch off.
func (s *Switch) TurnOff() error {
	return s.setState(off)
}

// TurnOn turns the switch on.
func (s *Switch) TurnOn() error {
	return s.setState(on)
}

// Toggle toggles the switch.
func (s *Switch) Toggle() error {
	switchState, err := s.getState()
	if err != nil {
		return err
	}
	switch switchState {
	case on:
		return s.TurnOff()
	default:
		return s.TurnOn()
	}
}

// getState returns the state of the WeMo switch.
func (s *Switch) getState() (state, error) {
	requestBody := `<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
  <s:Body>
    <u:GetBinaryState xmlns:u="urn:Belkin:service:basicevent:1"></u:GetBinaryState>
  </s:Body>
</s:Envelope>`
	responseBody, err := s.post("GetBinaryState", requestBody)
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
	return state(uint(stateInt)), nil
}

// setState sets the state of the WeMo switch.
func (s *Switch) setState(state state) error {
	requestBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
  <s:Body>
    <u:SetBinaryState xmlns:u="urn:Belkin:service:basicevent:1">
    	<BinaryState>%d</BinaryState>
    </u:SetBinaryState>
  </s:Body>
</s:Envelope>`, state)
	if _, err := s.post("SetBinaryState", requestBody); err != nil {
		return err
	}
	return nil
}

func (s *Switch) post(action string, body string) ([]byte, error) {
	url := "http://" + s.host + "/upnp/control/basicevent1"
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
