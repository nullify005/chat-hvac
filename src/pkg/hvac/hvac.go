package hvac

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
)

type HVACSet struct {
	Param string `json:"param"`
	Value string `json:"value"`
}

type HVACStatus struct {
	Device struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		FamilyID       int    `json:"familyId"`
		ModelID        int    `json:"modelId"`
		InstallationID int    `json:"installationId"`
		ZoneID         int    `json:"zoneId"`
		Order          int    `json:"order"`
		Widgets        []int  `json:"widgets"`
	} `json:"device"`
	Status struct {
		Num187           int `json:"187"`
		Num188           int `json:"188"`
		Num189           int `json:"189"`
		Num190           int `json:"190"`
		Num50008         int `json:"50008"`
		Num50010         int `json:"50010"`
		AlarmStatus      int `json:"alarm_status"`
		ConfigConfirmOff int `json:"config_confirm_off"`
		ConfigFanMap     struct {
			Num0 string `json:"0"`
			Num1 string `json:"1"`
			Num2 string `json:"2"`
			Num3 string `json:"3"`
			Num4 string `json:"4"`
		} `json:"config_fan_map"`
		ConfigModeMap             int    `json:"config_mode_map"`
		ConfigQuiet               int    `json:"config_quiet"`
		ConfigVerticalVanes       int    `json:"config_vertical_vanes"`
		CoolTemperatureMax        int    `json:"cool_temperature_max"`
		CoolTemperatureMin        int    `json:"cool_temperature_min"`
		ErrorAddress              int    `json:"error_address"`
		ErrorCode                 int    `json:"error_code"`
		ExternalLed               string `json:"external_led"`
		FanSpeed                  int    `json:"fan_speed"`
		FilterClean               int    `json:"filter_clean"`
		FilterDueHours            int    `json:"filter_due_hours"`
		HeatTemperatureMin        int    `json:"heat_temperature_min"`
		InternalLed               string `json:"internal_led"`
		InternalTemperatureOffset int    `json:"internal_temperature_offset"`
		MainenanceWReset          int    `json:"mainenance_w_reset"`
		MainenanceWoReset         int    `json:"mainenance_wo_reset"`
		Mode                      string `json:"mode"`
		Power                     string `json:"power"`
		QuietMode                 string `json:"quiet_mode"`
		RemoteControllerLock      int    `json:"remote_controller_lock"`
		Rssi                      int    `json:"rssi"`
		RuntimeModeRestrictions   int    `json:"runtime_mode_restrictions"`
		Setpoint                  int    `json:"setpoint"`
		SetpointMax               int    `json:"setpoint_max"`
		SetpointMin               int    `json:"setpoint_min"`
		TempLimitation            string `json:"temp_limitation"`
		Temperature               int    `json:"temperature"`
		UID185                    int    `json:"uid_185"`
		UID186                    int    `json:"uid_186"`
		Vvane                     string `json:"vvane"`
		WorkingHours              int    `json:"working_hours"`
	} `json:"status"`
}

type Hvac struct {
	logger *log.Logger
	api    string
	device string
}

const (
	defaultApi     string = "http://127.0.0.1:2112"
	defaultDevice  string = "127934703953"
	deviceEndpoint string = "%s/hvac/%s"
	postMethod     string = "post"
	getMethod      string = "get"
)

type HvacOption func(h *Hvac)

func WithLogger(l *log.Logger) HvacOption {
	return func(h *Hvac) {
		h.logger = l
	}
}

func WithApi(a string) HvacOption {
	return func(h *Hvac) {
		h.api = a
	}
}

func WithDevice(d string) HvacOption {
	return func(h *Hvac) {
		h.device = d
	}
}

func New(opts ...HvacOption) *Hvac {
	h := &Hvac{
		logger: log.New(os.Stdout, "Hvac: ", log.Ldate|log.Ltime|log.Lshortfile),
		api:    defaultApi,
		device: defaultDevice,
	}
	for _, opt := range opts {
		opt(h)
	}
	h.logger.Printf("using api: %s device: %s", h.api, h.device)
	return h
}

// return the full api endpoint for the device status
func (h *Hvac) deviceEndpoint() string {
	return fmt.Sprintf(deviceEndpoint, h.api, h.device)
}

// fetch the status from the service-intesis endpoint
func (h *Hvac) Status() string {
	body, err := httpCall(h.deviceEndpoint(), getMethod, nil)
	if err != nil {
		return fmt.Sprintf(":x: %v", err)
	}
	status := &HVACStatus{}
	if err = json.Unmarshal([]byte(body), &status); err != nil {
		e := fmt.Sprintf(":x: unable to decode: %s", string(body))
		h.logger.Printf(e)
		return e
	}
	return status.String()
}

// performs a set for a key value pair against the API
func (h *Hvac) Set(key, value string) string {
	payload := &HVACSet{Param: key, Value: value}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		e := fmt.Sprintf("unable to encode: %v to json. cause: %v", payload, err)
		h.logger.Printf(e)
		return e
	}
	body, err := httpCall(h.deviceEndpoint(), postMethod, &buf)
	if err != nil {
		return fmt.Sprintf(":x: %v", err)
	}
	return fmt.Sprintf(":+1: `%s`", string(body))
}

// enumerate the fields of Status & return them as a new line delimited key: value pair string
func (h *HVACStatus) String() string {
	ret := ""
	v := reflect.ValueOf(h.Status)
	s := v.Type()
	for i := 0; i < v.NumField(); i++ {
		ret += fmt.Sprintf("%v: %v\n", s.Field(i).Name, v.Field(i).Interface())
	}
	return strings.TrimRight(ret, "\n")
}

// common method for http calls
func httpCall(endpoint, method string, payload *bytes.Buffer) (string, error) {
	var (
		resp *http.Response
		err  error
	)
	if method == postMethod {
		resp, err = http.Post(endpoint, "application/json", payload)
	} else {
		resp, err = http.Get(endpoint)
	}
	if err != nil {
		return "", fmt.Errorf("http %s failed. url: %s cause: %v", method, endpoint, err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("cannot read response body: %s from: %s cause: %v", resp.Body, endpoint, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("invalid status code: %d url: %s body: %s", resp.StatusCode, endpoint, string(body))
	}
	return string(body), nil
}
