package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"slices"

	yaml "gopkg.in/yaml.v3"
)

// Example:
//
//	{
//	  "id": 83645,
//	  "name": "tcp-test",
//	  "host": "159.69.18.28",
//	  "port": 443,
//	  "ipVersion": "IPV4",
//	  "stringToSend": "",
//	  "stringToReceive": "",
//	  "note": "",
//	  "runTraceroute": "DISABLED",
//	  "userId": 300003895,
//	  "interval": "THIRTYSECONDS",
//	  "monitorIntervalPolicy": "PARALLEL",
//	  "checkSites": [
//	    4
//	  ],
//	  "notificationGroups": [],
//	  "scheduleId": 0,
//	  "notificationReportTimeout": 1440,
//	  "verificationPolicy": "SIMPLE"
//	}
type SonarTCPCheck struct {
	ID                        int    `json:"id"`
	Name                      string `json:"name" yaml:"name"`
	Host                      string `json:"host" yaml:"host"`
	IPVersion                 string `json:"ipVersion" yaml:"ipVersion"`
	Port                      int    `json:"port" yaml:"port"`
	Interval                  string `json:"interval" yaml:"interval"`
	CheckSites                []int  `json:"checkSites" yaml:"checkSites"`
	RunTraceroute             string `json:"runTraceroute" yaml:"runTraceroute"`
	StringToSend              string `json:"stringToSend" yaml:"stringToSend"`
	StringToReceive           string `json:"stringToReceive" yaml:"stringToReceive"`
	Note                      string `json:"note" yaml:"note"`
	UserID                    int    `json:"userId" yaml:"userId"`
	MonitorIntervalPolicy     string `json:"monitorIntervalPolicy" yaml:"monitorIntervalPolicy"`
	NotificationGroups        []int  `json:"notificationGroups" yaml:"notificationGroups"`
	ScheduleID                int    `json:"scheduleId" yaml:"scheduleId"`
	NotificationReportTimeout int    `json:"notificationReportTimeout" yaml:"notificationReportTimeout"`
	VerificationPolicy        string `json:"verificationPolicy" yaml:"verificationPolicy"`
}

func (ac *SonarTCPCheck) GetResource() interface{} {
	return ac
}

func (ac *SonarTCPCheck) GetResourceID() string {
	return ac.Name
}

func (ac *SonarTCPCheck) GetConstellixID() int {
	return ac.ID
}

func (ac *SonarTCPCheck) SyncResourceDelete(constellixID int) error {
	logger.Printf("  removing resource %q\n", ac.GetResourceID())
	endpoint, err := url.JoinPath(sonarRESTAPIBaseURL, "tcp", fmt.Sprint(constellixID))
	if err != nil {
		return fmt.Errorf("build Sonar TCP check delete endpoint: %w", err)
	}
	body, err := makeSimpleAPIRequest("DELETE", endpoint, nil, 202)
	if err != nil {
		L.Warn("unexpected response", slog.String("details", string(body)))
		return fmt.Errorf("unable to delete Sonar TCP check: %w", err)
	}
	return nil
}

type ExpectedSonarTCPCheck struct {
	// Mapping of defined fields from parsed data to struct Field Names
	definedFieldsMap map[string]string
	// List of immutable fields which can't be updated via API
	immutableFields []string
	// List of mandatory fields which must be defined, used for validation
	mandatoryFields []string
	SonarTCPCheck
}

// UnmarshalYAML unmarshals the mesage and stores original fields
func (ex *ExpectedSonarTCPCheck) UnmarshalYAML(value *yaml.Node) error {
	ex.immutableFields = []string{"host", "ipVersion"}
	ex.mandatoryFields = []string{"name", "host", "ipVersion", "port", "interval", "checkSites"}

	// Unmarshall data into SonarTCPCheck struct
	var s SonarTCPCheck
	err := value.Decode(&s)
	if err != nil {
		return fmt.Errorf("decode Sonar TCP check: %w", err)
	}
	ex.SonarTCPCheck = s

	// Save specified fields
	dm := make(map[string]interface{})
	err = value.Decode(&dm)
	if err != nil {
		return fmt.Errorf("decode Sonar TCP check fields: %w", err)
	}

	definedFields := make([]string, len(dm))
	i := 0
	for k := range dm {
		definedFields[i] = k
		i++
	}
	ex.definedFieldsMap = getFieldNamesMap(&ex.SonarTCPCheck, "yaml", definedFields...)
	return nil
}

// Validate performs simple validation of user provided data
func (ex *ExpectedSonarTCPCheck) Validate() error {
	// Validate that all mandatory fields are present
	for _, f := range ex.mandatoryFields {
		if _, ok := ex.definedFieldsMap[f]; !ok {
			return fmt.Errorf("%s: mandatory field %q is not defined", ex.Name, f)
		}
	}
	return nil
}

// GetDefinedStructFieldNames returns list of defined struct fields from local configuration
func (ex *ExpectedSonarTCPCheck) GetDefinedStructFieldNames() []string {
	return slices.Collect(maps.Values(ex.definedFieldsMap))
}

// GetImmutableStructFields returns list of immutable struct fields
func (ex *ExpectedSonarTCPCheck) GetImmutableStructFields() []string {
	var imf []string
	for k, v := range ex.definedFieldsMap {
		if slices.Contains(ex.immutableFields, k) {
			imf = append(imf, v)
		}
	}
	return imf
}

func (ex *ExpectedSonarTCPCheck) GetResource() interface{} {
	return ex.SonarTCPCheck
}

func (ex *ExpectedSonarTCPCheck) GetResourceID() string {
	return ex.Name
}

func (ex *ExpectedSonarTCPCheck) SyncResourceUpdate(constellixID int) error {
	logger.Printf("  updating resource %q\n", ex.GetResourceID())
	endpoint, err := url.JoinPath(sonarRESTAPIBaseURL, "tcp", fmt.Sprint(constellixID))
	if err != nil {
		return fmt.Errorf("build Sonar TCP check update endpoint: %w", err)
	}
	payload, err := generatePayload(ex, slices.Collect(maps.Keys(ex.definedFieldsMap)), ex.immutableFields)
	if err != nil {
		return err
	}
	payloadReader := bytes.NewReader(payload)
	body, err := makeSimpleAPIRequest("PUT", endpoint, payloadReader, 200)
	if err != nil {
		L.Warn("unexpected response", slog.String("details", string(body)))
		return fmt.Errorf("unable to update Sonar TCP check: %w", err)
	}
	return nil
}

func (ex *ExpectedSonarTCPCheck) SyncResourceCreate() error {
	logger.Printf("  creating new resource %q\n", ex.GetResourceID())
	endpoint, err := url.JoinPath(sonarRESTAPIBaseURL, "tcp")
	if err != nil {
		return fmt.Errorf("build Sonar TCP check create endpoint: %w", err)
	}
	payload, err := generatePayload(ex, slices.Collect(maps.Keys(ex.definedFieldsMap)), nil)
	if err != nil {
		return err
	}
	payloadReader := bytes.NewReader(payload)
	body, err := makeSimpleAPIRequest("POST", endpoint, payloadReader, 201)
	if err != nil {
		L.Warn("unexpected response", slog.String("details", string(body)))
		return fmt.Errorf("unable to create Sonar TCP check: %w", err)
	}
	return nil
}

// GetSonarTCPChecks returns active Sonar Checks
func GetSonarTCPChecks() ([]*SonarTCPCheck, error) {
	L.Debug("retrieving Sonar TCP checks")
	endpoint, err := url.JoinPath(sonarRESTAPIBaseURL, "tcp")
	if err != nil {
		return nil, fmt.Errorf("build Sonar TCP checks endpoint: %w", err)
	}
	data, err := makeSimpleAPIRequest("GET", endpoint, nil, 200)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sonar TCP checks: %w", err)
	}

	checks := make([]*SonarTCPCheck, 0)
	err = json.Unmarshal(data, &checks)
	if err != nil {
		return nil, fmt.Errorf("parse Sonar TCP checks response: %w", err)
	}
	return checks, nil
}
