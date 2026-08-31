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

type GeoProximity struct {
	ID        int     `json:"id"`
	Name      string  `json:"name" yaml:"name"`
	Country   string  `json:"country" yaml:"country"` // Optional, long/lat autocomplete
	Region    string  `json:"region" yaml:"region"`   // Optional
	City      int     `json:"city" yaml:"city"`       // City is optional and works as long/lat autocomplete
	Longitude float64 `json:"longitude" yaml:"longitude"`
	Latitude  float64 `json:"latitude" yaml:"latitude"`
}

func (ac *GeoProximity) GetResource() interface{} {
	return ac
}

func (ac *GeoProximity) GetResourceID() string {
	return ac.Name
}

func (ac *GeoProximity) GetConstellixID() int {
	return ac.ID
}

func (ac *GeoProximity) SyncResourceDelete(constellixID int) error {
	logger.Printf("  removing resource %q\n", ac.GetResourceID())
	endpoint, err := url.JoinPath(dnsRESTAPIBaseURL, "geoproximities", fmt.Sprint(constellixID))
	if err != nil {
		return fmt.Errorf("build GeoProximity delete endpoint: %w", err)
	}
	data, err := makev4APIRequest("DELETE", endpoint, nil, 204)
	if err != nil {
		var details string
		for _, item := range data {
			details += string(item)
		}
		L.Warn("unexpected response", slog.String("details", details))
		return fmt.Errorf("unable to delete GeoProximity: %w", err)
	}
	return nil
}

type ExpectedGeoProximity struct {
	// Mapping of defined fields from parsed data to struct Field Names
	definedFieldsMap map[string]string
	// List of immutable fields which can't be updated via API
	immutableFields []string
	// List of mandatory fields which must be defined, used for validation
	mandatoryFields []string
	GeoProximity
}

// UnmarshalYAML unmarshals the mesage and stores original fields
func (ex *ExpectedGeoProximity) UnmarshalYAML(value *yaml.Node) error {
	ex.immutableFields = []string{}
	ex.mandatoryFields = []string{"name", "longitude", "latitude"}

	// Unmarshall data into GeoProximity struct
	var s GeoProximity
	err := value.Decode(&s)
	if err != nil {
		return fmt.Errorf("decode GeoProximity: %w", err)
	}
	ex.GeoProximity = s

	// Save specified fields
	dm := make(map[string]interface{})
	err = value.Decode(&dm)
	if err != nil {
		return fmt.Errorf("decode GeoProximity fields: %w", err)
	}

	definedFields := make([]string, len(dm))
	i := 0
	for k := range dm {
		definedFields[i] = k
		i++
	}
	ex.definedFieldsMap = getFieldNamesMap(&ex.GeoProximity, "yaml", definedFields...)
	return nil
}

// Validate performs simple validation of user provided data
func (ex *ExpectedGeoProximity) Validate() error {
	// Validate that all mandatory fields are present
	for _, f := range ex.mandatoryFields {
		if _, ok := ex.definedFieldsMap[f]; !ok {
			return fmt.Errorf("%s: mandatory field %q is not defined", ex.Name, f)
		}
	}
	return nil
}

// GetDefinedStructFieldNames returns list of defined struct fields from local configuration
func (ex *ExpectedGeoProximity) GetDefinedStructFieldNames() []string {
	return slices.Collect(maps.Values(ex.definedFieldsMap))
}

// GetImmutableStructFields returns list of immutable struct fields
func (ex *ExpectedGeoProximity) GetImmutableStructFields() []string {
	var imf []string
	for k, v := range ex.definedFieldsMap {
		if slices.Contains(ex.immutableFields, k) {
			imf = append(imf, v)
		}
	}
	return imf
}

func (ex *ExpectedGeoProximity) GetResource() interface{} {
	return ex.GeoProximity
}

func (ex *ExpectedGeoProximity) GetResourceID() string {
	return ex.Name
}

func (ex *ExpectedGeoProximity) SyncResourceUpdate(constellixID int) error {
	logger.Printf("  updating resource %q\n", ex.GetResourceID())
	endpoint, err := url.JoinPath(dnsRESTAPIBaseURL, "geoproximities", fmt.Sprint(constellixID))
	if err != nil {
		return fmt.Errorf("build GeoProximity update endpoint: %w", err)
	}
	payload, err := generatePayload(ex, slices.Collect(maps.Keys(ex.definedFieldsMap)), nil)
	if err != nil {
		return err
	}
	payloadReader := bytes.NewReader(payload)
	data, err := makev4APIRequest("PUT", endpoint, payloadReader, 200)
	if err != nil {
		var details string
		for _, item := range data {
			details += string(item)
		}
		L.Warn("unexpected response", slog.String("details", details))
		return fmt.Errorf("unable to update GeoProximity: %w", err)
	}
	return nil
}

func (ex *ExpectedGeoProximity) SyncResourceCreate() error {
	logger.Printf("  creating new resource %q\n", ex.GetResourceID())
	endpoint, err := url.JoinPath(dnsRESTAPIBaseURL, "geoproximities")
	if err != nil {
		return fmt.Errorf("build GeoProximity create endpoint: %w", err)
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
		return fmt.Errorf("unable to create GeoProximity: %w", err)
	}
	return nil
}

// GetGeoProximities returns active geo proximities
var GetGeoProximities = func() ([]*GeoProximity, error) {
	L.Debug("retrieving GeoProximities")
	if cachedGeoProximities != nil {
		L.Debug("using cached GeoProximities")
		return cachedGeoProximities, nil
	}
	endpoint, err := url.JoinPath(dnsRESTAPIBaseURL, "geoproximities")
	if err != nil {
		return nil, fmt.Errorf("build GeoProximities endpoint: %w", err)
	}
	data, err := makev4APIRequest("GET", endpoint, nil, 200)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve GeoProximities: %w", err)
	}

	geops := make([]*GeoProximity, 0)
	for _, item := range data {
		var tmpGP []*GeoProximity
		err = json.Unmarshal(item, &tmpGP)
		if err != nil {
			return nil, fmt.Errorf("parse GeoProximities response: %w", err)
		}
		if len(tmpGP) > 0 {
			geops = append(geops, tmpGP...)
		}
	}

	cachedGeoProximities = geops
	return geops, nil
}
