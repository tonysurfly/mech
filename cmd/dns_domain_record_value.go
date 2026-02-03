package cmd

import (
	"fmt"
	"strings"
)

type DNSStandardItemValue struct {
	Value   string `json:"value" yaml:"value"`
	Enabled bool   `json:"enabled" yaml:"enabled"`
}

type DNSFailoverValue struct {
	Mode    string                  `json:"mode" yaml:"mode"`
	Enabled bool                    `json:"enabled" yaml:"enabled"`
	Values  []*DNSFailoverItemValue `json:"values" yaml:"values"`
}

type DNSFailoverItemValue struct {
	Enabled      bool   `json:"enabled" yaml:"enabled"`
	Order        int    `json:"order" yaml:"order"`
	SonarCheckID int    `json:"sonarCheckId" yaml:"sonarCheckId"`
	Value        string `json:"value" yaml:"value"`
}

type DNSMXStandardItemValue struct {
	Server   string `json:"server" yaml:"server"`
	Priority int    `json:"priority" yaml:"priority"`
	Enabled  bool   `json:"enabled" yaml:"enabled"`
}

type DNSCAAStandardItemValue struct {
	Tag     string `json:"tag" yaml:"tag"`
	Data    string `json:"data" yaml:"data"`
	Flags   int    `json:"flags" yaml:"flags"`
	Enabled bool   `json:"enabled" yaml:"enabled"`
}

type DNSHTTPStandardItemValue struct {
	Hard         bool   `json:"hard" yaml:"hard"`
	RedirectType string `json:"redirectType" yaml:"redirectType"`
	Title        string `json:"title" yaml:"title"`
	Keywords     string `json:"keywords" yaml:"keywords"`
	Description  string `json:"description" yaml:"description"`
	URL          string `json:"url" yaml:"url"`
}

type aliasDNSRecord DNSRecord

// populateDNSRecordValue populates the Value field of a DNSRecord based on the
// Mode field.
// TODO: be carefull with type casting, use similar to sonarCheckID everywhere
func populateDNSRecordValue(record interface{}) error {
	s, ok := record.(*DNSRecord)
	if !ok {
		return fmt.Errorf("unable to assert record to DNSRecord")
	}
	switch s.Type {
	case "A", "AAAA", "ANAME", "CNAME":
		switch s.Mode {
		case "standard":
			m, ok := s.Value.([]interface{})
			if !ok {
				return fmt.Errorf("unable to parse value for standard mode, expected an array")
			}
			valueObj := make([]*DNSStandardItemValue, 0)
			for _, el := range m {
				elMap, ok := el.(map[string]interface{})
				if !ok {
					return fmt.Errorf("unable to parse value for standard mode, expected an map")
				}
				valueEl := DNSStandardItemValue{
					Value:   elMap["value"].(string),
					Enabled: elMap["enabled"].(bool),
				}
				valueObj = append(valueObj, &valueEl)
			}
			s.Value = valueObj
		case "failover":
			valueObj := DNSFailoverValue{}
			m, ok := s.Value.(map[string]interface{})
			if !ok {
				return fmt.Errorf("unable to parse value for failover mode, expected an map")
			}
			valueObj.Mode = m["mode"].(string)
			valueObj.Enabled = m["enabled"].(bool)
			values := make([]*DNSFailoverItemValue, 0)
			for _, valueItem := range m["values"].([]interface{}) {
				valueItemMap, ok := valueItem.(map[string]interface{})
				if !ok {
					return fmt.Errorf("unable to parse value for value of failover mode, expected an map")
				}
				sonarCheckID, sonarCheckHost, err := getSonarCheckID(valueItemMap["sonarCheckId"])
				if err != nil {
					return err
				}
				if sonarCheckHost == "" {
					sonarCheckHost = valueItemMap["value"].(string)
				}
				valueItemObj := DNSFailoverItemValue{
					Enabled:      valueItemMap["enabled"].(bool),
					Order:        toInt(valueItemMap["order"]),
					Value:        sonarCheckHost,
					SonarCheckID: sonarCheckID,
				}
				values = append(values, &valueItemObj)
			}
			valueObj.Values = values
			s.Value = &valueObj
		case "roundrobin-failover":
			if s.Type == "CNAME" || s.Type == "ANAME" {
				return fmt.Errorf("roundrobin-failover is not supported for CNAME records")
			}
			m, ok := s.Value.([]interface{})
			if !ok {
				return fmt.Errorf("unable to parse value for roundrobin-failover mode, expected an array")
			}
			valueObj := make([]*DNSFailoverItemValue, 0)
			for _, el := range m {
				elMap, ok := el.(map[string]interface{})
				if !ok {
					return fmt.Errorf("unable to parse value for roundrobin-failover mode, expected an map")
				}
				sonarCheckID, sonarCheckHost, err := getSonarCheckID(elMap["sonarCheckId"])
				if err != nil {
					return err
				}
				if sonarCheckHost == "" {
					sonarCheckHost = elMap["value"].(string)
				}
				valueEl := DNSFailoverItemValue{
					Enabled:      elMap["enabled"].(bool),
					Order:        toInt(elMap["order"]),
					Value:        sonarCheckHost,
					SonarCheckID: sonarCheckID,
				}
				valueObj = append(valueObj, &valueEl)
			}
			s.Value = valueObj
		case "pools":
			m, ok := s.Value.([]interface{})
			if !ok {
				return fmt.Errorf("unable to parse value for pools mode, expected an array")
			}
			valueObj := make([]int, 0)
			for _, el := range m {
				valueObj = append(valueObj, toInt(el))
			}
			s.Value = valueObj
		default:
			return fmt.Errorf("unknown mode %q", s.Mode)
		}
	case "MX":
		if s.Mode != "standard" {
			return fmt.Errorf("unsupported mode %q for MX record", s.Mode)
		}
		m, ok := s.Value.([]interface{})
		if !ok {
			return fmt.Errorf("unable to parse value for MX record in standard mode, expected an array")
		}
		valueObj := make([]*DNSMXStandardItemValue, 0)
		for _, el := range m {
			elMap, ok := el.(map[string]interface{})
			if !ok {
				return fmt.Errorf("unable to parse value for standard mode, expected an map")
			}
			valueEl := DNSMXStandardItemValue{
				Server:   elMap["server"].(string),
				Priority: toInt(elMap["priority"]),
				Enabled:  elMap["enabled"].(bool),
			}
			valueObj = append(valueObj, &valueEl)
		}
		s.Value = valueObj
	case "TXT":
		if s.Mode != "standard" {
			return fmt.Errorf("unsupported mode %q for TXT record", s.Mode)
		}
		m, ok := s.Value.([]interface{})
		if !ok {
			return fmt.Errorf("unable to parse value for TXT record in standard mode, expected an array")
		}
		valueObj := make([]*DNSStandardItemValue, 0)
		for _, el := range m {
			elMap, ok := el.(map[string]interface{})
			if !ok {
				return fmt.Errorf("unable to parse value for TXT record in standard mode, expected an map")
			}
			valueEl := DNSStandardItemValue{
				Value:   elMap["value"].(string),
				Enabled: elMap["enabled"].(bool),
			}
			valueObj = append(valueObj, &valueEl)
			s.Value = valueObj
		}
	case "HTTP":
		if s.Mode != "standard" {
			return fmt.Errorf("unsupported mode %q for HTTP record", s.Mode)
		}
		m, ok := s.Value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unable to parse value for HTTP record in standard mode, expected a map")
		}
		valueObj := &DNSHTTPStandardItemValue{}
		valueObj.Hard, _ = m["hard"].(bool)
		valueObj.URL, _ = m["url"].(string)
		valueObj.RedirectType, _ = m["redirectType"].(string)
		valueObj.Title, _ = m["title"].(string)
		valueObj.Keywords, _ = m["keywords"].(string)
		valueObj.Description, _ = m["description"].(string)
		s.Value = valueObj
	case "CAA":
		if s.Mode != "standard" {
			return fmt.Errorf("unsupported mode %q for CAA record", s.Mode)
		}
		m, ok := s.Value.([]interface{})
		if !ok {
			return fmt.Errorf("unable to parse value for CAA record in standard mode, expected an array")
		}
		valueObj := make([]*DNSCAAStandardItemValue, 0)
		for _, el := range m {
			elMap, ok := el.(map[string]interface{})
			if !ok {
				return fmt.Errorf("unable to parse value for CAA record in standard mode, expected a map")
			}
			valueEl := DNSCAAStandardItemValue{
				Tag:     elMap["tag"].(string),
				Data:    elMap["data"].(string),
				Flags:   toInt(elMap["flags"]),
				Enabled: elMap["enabled"].(bool),
			}
			valueObj = append(valueObj, &valueEl)
		}
		s.Value = valueObj
	default:
		return fmt.Errorf("unsupported record type %q", s.Type)
	}
	return nil
}

func toInt(i interface{}) int {
	switch v := i.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}

func getSonarCheckID(i interface{}) (int, string, error) {
	switch v := i.(type) {
	case string:
		checkType, checkName, err := parseSonarCheckID(v)
		if err != nil {
			return 0, "", err
		}
		switch checkType {
		case "http":
			checks, err := GetSonarHTTPChecks()
			if err != nil {
				return 0, "", err
			}
			for _, check := range checks {
				if check.GetResourceID() == checkName {
					return check.ID, check.Host, nil
				}
			}
			return 0, "", fmt.Errorf("unable to find sonar check %s:%s", checkType, checkName)
		default:
			return 0, "", fmt.Errorf("unsupported check type: %s", checkType)
		}
	}
	return toInt(i), "", nil
}

// parseSonarCheckID parses a sonar check ID from a string. It assumes that the string
// will start with a @, followed by code word 'sonar' with specified check type and the
// name of the check itself
func parseSonarCheckID(s string) (string, string, error) {
	if !strings.HasPrefix(s, "@sonar,") {
		return "", "", fmt.Errorf("invalid sonar check ID. Expected @sonar,<check_type>:<check_name> or int")
	}
	s = strings.TrimPrefix(s, "@sonar,")
	split := strings.Split(s, ":")
	if len(split) != 2 {
		return "", "", fmt.Errorf("invalid sonar check ID. Expected @sonar,<check_type>:<check_name> or int")
	}
	return split[0], split[1], nil
}
