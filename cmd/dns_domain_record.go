package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"slices"

	"gopkg.in/yaml.v3"
)

var dnsRecordResourceIDTemplate = "%s %q (%s, %d)"

// Missing fields: lastValues, skipLookup, contacts
type DNSRecord struct {
	ID                   int         `json:"id"`
	Name                 string      `json:"name" yaml:"name"`
	Type                 string      `json:"type" yaml:"type"`
	TTL                  int         `json:"ttl" yaml:"ttl"`
	Mode                 string      `json:"mode" yaml:"mode"`
	Region               string      `json:"region" yaml:"region"`
	IPFilter             interface{} `json:"ipfilter" yaml:"ipfilter"`
	IPFilterDrop         bool        `json:"ipfilterDrop" yaml:"ipfilterDrop"`
	GeoFailover          bool        `json:"geoFailover" yaml:"geoFailover"`
	GeoProximity         interface{} `json:"geoproximity" yaml:"geoproximity"`
	Enabled              bool        `json:"enabled" yaml:"enabled"`
	Value                interface{} `json:"value" yaml:"value"`
	Notes                string      `json:"notes" yaml:"notes"`
	domainIDInConstellix int
}

func (ac *DNSRecord) UnmarshalJSON(b []byte) error {
	var alias aliasDNSRecord
	err := json.Unmarshal(b, &alias)
	if err != nil {
		return fmt.Errorf("unmarshal DNS record: %w", err)
	}
	s := DNSRecord(alias)
	err = populateDNSRecordValue(&s)
	if err != nil {
		return fmt.Errorf("record %s (%s): %w", s.Name, s.Type, err)
	}
	err = populateDNSRecordIPFilterForJSON(&s)
	if err != nil {
		return fmt.Errorf("record %s (%s): %w", s.Name, s.Type, err)
	}
	err = populateDNSRecordGeoproximityForJSON(&s)
	if err != nil {
		return fmt.Errorf("record %s (%s): %w", s.Name, s.Type, err)
	}
	*ac = s
	return nil
}

func (ac *DNSRecord) GetResource() interface{} {
	return ac
}

func (ac *DNSRecord) GetResourceID() string {
	if ac.GeoProximity != nil {
		return fmt.Sprintf(dnsRecordResourceIDTemplate, ac.Type, ac.Name, ac.Region, ac.GeoProximity)
	}
	return fmt.Sprintf(dnsRecordResourceIDTemplate, ac.Type, ac.Name, ac.Region, 0)
}

func (ac *DNSRecord) GetConstellixID() int {
	return ac.ID
}

func (ac *DNSRecord) SyncResourceDelete(constellixID int) error {
	logger.Printf("  removing resource %q\n", ac.GetResourceID())
	if ac.domainIDInConstellix == 0 {
		return fmt.Errorf("unable to delete DNS record: domain ID is not defined (internal error)")
	}
	endpoint, err := url.JoinPath(
		dnsRESTAPIBaseURL,
		"domains",
		fmt.Sprintf("%d", ac.domainIDInConstellix),
		"records",
		fmt.Sprintf("%d", constellixID),
	)
	if err != nil {
		return fmt.Errorf("build DNS record delete endpoint: %w", err)
	}
	data, err := makev4APIRequest("DELETE", endpoint, nil, 204)
	if err != nil {
		var details string
		for _, item := range data {
			details += string(item)
		}
		L.Warn("unexpected response", slog.String("details", details))
		return fmt.Errorf("unable to delete DNS record: %w", err)
	}
	return nil
}

type ExpectedDNSRecord struct {
	// Mapping of defined fields from parsed data to struct Field Names
	definedFieldsMap map[string]string
	// List of immutable fields which can't be updated via API
	immutableFields []string
	// List of mandatory fields which must be defined, used for validation
	mandatoryFields []string
	DNSRecord
}

// UnmarshalYAML unmarshals the mesage and stores original fields
func (ex *ExpectedDNSRecord) UnmarshalYAML(value *yaml.Node) error {
	ex.immutableFields = []string{"type"}
	ex.mandatoryFields = []string{"type", "value"}

	// Unmarshall data into DNSRecord struct
	var s DNSRecord
	err := value.Decode(&s)
	if err != nil {
		return fmt.Errorf("decode DNS record: %w", err)
	}

	err = populateDNSRecordValue(&s)
	if err != nil {
		return fmt.Errorf("record %s (%s): %w", s.Name, s.Type, err)
	}
	err = populateDNSRecordIPFilterForYAML(&s)
	if err != nil {
		return fmt.Errorf("record %s (%s): %w", s.Name, s.Type, err)
	}
	err = populateDNSRecordGeoproximityForYAML(&s)
	if err != nil {
		return fmt.Errorf("record %s (%s): %w", s.Name, s.Type, err)
	}
	ex.DNSRecord = s

	// Save specified fields
	dm := make(map[string]interface{})
	err = value.Decode(&dm)
	if err != nil {
		return fmt.Errorf("decode DNS record fields: %w", err)
	}

	definedFields := make([]string, len(dm))
	i := 0
	for k := range dm {
		definedFields[i] = k
		i++
	}
	ex.definedFieldsMap = getFieldNamesMap(&ex.DNSRecord, "yaml", definedFields...)
	return nil
}

// GetDefinedStructFieldNames returns list of defined struct fields from local configuration
func (ex *ExpectedDNSRecord) GetDefinedStructFieldNames() []string {
	return slices.Collect(maps.Values(ex.definedFieldsMap))
}

// GetImmutableStructFields returns list of immutable struct fields
func (ex *ExpectedDNSRecord) GetImmutableStructFields() []string {
	var imf []string
	for k, v := range ex.definedFieldsMap {
		if slices.Contains(ex.immutableFields, k) {
			imf = append(imf, v)
		}
	}
	return imf
}

func (ex *ExpectedDNSRecord) GetResource() interface{} {
	return ex.DNSRecord
}

func (ex *ExpectedDNSRecord) GetResourceID() string {
	if ex.GeoProximity != nil {
		return fmt.Sprintf(dnsRecordResourceIDTemplate, ex.Type, ex.Name, ex.Region, ex.GeoProximity)
	}
	return fmt.Sprintf(dnsRecordResourceIDTemplate, ex.Type, ex.Name, ex.Region, 0)
}

func (ex *ExpectedDNSRecord) SyncResourceUpdate(constellixID int) error {
	logger.Printf("  updating resource %q\n", ex.GetResourceID())
	if ex.domainIDInConstellix == 0 {
		return fmt.Errorf("unable to update DNS record: domain ID is not defined (internal error)")
	}
	endpoint, err := url.JoinPath(
		dnsRESTAPIBaseURL,
		"domains",
		fmt.Sprintf("%d", ex.domainIDInConstellix),
		"records",
		fmt.Sprintf("%d", constellixID),
	)
	if err != nil {
		return fmt.Errorf("build DNS record update endpoint: %w", err)
	}
	payload, err := generatePayload(ex, slices.Collect(maps.Keys(ex.definedFieldsMap)), nil)
	if err != nil {
		return err
	}
	payloadReader := bytes.NewReader(payload)
	data, err := makev4APIRequest("PATCH", endpoint, payloadReader, 200)
	if err != nil {
		var details string
		for _, item := range data {
			details += string(item)
		}
		L.Warn("unexpected response", slog.String("details", details))
		return fmt.Errorf("unable to update DNS record: %w", err)
	}
	return nil
}

func (ex *ExpectedDNSRecord) SyncResourceCreate() error {
	logger.Printf("  creating new resource %q\n", ex.GetResourceID())
	if ex.domainIDInConstellix == 0 {
		return fmt.Errorf("unable to create DNS record: domain ID is not defined (internal error)")
	}
	endpoint, err := url.JoinPath(dnsRESTAPIBaseURL, "domains", fmt.Sprintf("%d", ex.domainIDInConstellix), "records")
	if err != nil {
		return fmt.Errorf("build DNS record create endpoint: %w", err)
	}
	payload, err := generatePayload(ex, slices.Collect(maps.Keys(ex.definedFieldsMap)), nil)
	if err != nil {
		return err
	}
	payloadReader := bytes.NewReader(payload)
	data, err := makev4APIRequest("POST", endpoint, payloadReader, 202)
	if err != nil {
		var details string
		for _, item := range data {
			details += string(item)
		}
		L.Warn("unexpected response", slog.String("details", details))
		return fmt.Errorf("unable to create DNS record: %w", err)
	}
	return nil
}

// GetDNSRecords retrieves domain's DNS records
func GetDNSRecords(id int) ([]*DNSRecord, error) {
	L.Debug("retrieving DNS records", slog.Int("domain_id", id))
	endpoint, err := url.JoinPath(dnsRESTAPIBaseURL, "domains", fmt.Sprintf("%d", id), "records")
	if err != nil {
		return nil, fmt.Errorf("build DNS records endpoint: %w", err)
	}

	data, err := makev4APIRequest("GET", endpoint, nil, 200)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve DNS records: %w", err)
	}

	var records []*DNSRecord
	for _, item := range data {
		var tmpRecords []*DNSRecord
		err = json.Unmarshal(item, &tmpRecords)
		if err != nil {
			return nil, fmt.Errorf("parse DNS records response: %w", err)
		}
		if len(tmpRecords) > 0 {
			records = append(records, tmpRecords...)
		}
	}

	for _, item := range records {
		item.domainIDInConstellix = id
	}
	return records, nil
}
