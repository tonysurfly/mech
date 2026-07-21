package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// SOA is ignored for now
type DNSDomain struct {
	ID               int         `json:"id"`
	Name             string      `json:"name"`
	Note             string      `json:"note"`
	Status           string      `json:"status"`
	GeoIPEnabled     bool        `json:"geoip"`
	GTDEnabled       bool        `json:"gtd"`
	Nameservers      []string    `json:"nameservers"`
	Tags             []string    `json:"tags"`
	Template         int         `json:"template"`
	VanityNameserver interface{} `json:"vanityNameserver"`
	Contacts         []int       `json:"contacts"`
	CreatedAt        string      `json:"createdAt"`
	UpdatedAt        string      `json:"updatedAt"`
}

// GetDNSDomains returns active DNS domains in Constellix
func GetDNSDomains() ([]*DNSDomain, error) {
	L.Debug("retrieving DNS domains")
	endpoint, err := url.JoinPath(dnsRESTAPIBaseURL, "domains")
	if err != nil {
		return nil, fmt.Errorf("build DNS domains endpoint: %w", err)
	}
	data, err := makev4APIRequest("GET", endpoint, nil, 200)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve DNS domains list: %w", err)
	}
	var domains []*DNSDomain
	for _, item := range data {
		tmpDomains := []*DNSDomain{}
		err = json.Unmarshal(item, &tmpDomains)
		if err != nil {
			return nil, fmt.Errorf("parse DNS domains response: %w", err)
		}
		domains = append(domains, tmpDomains...)
	}
	return domains, nil
}
