package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetDNSDomains(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": [
				{"id": 1, "name": "example.com", "status": "ACTIVE", "geoip": true, "gtd": false}
			],
			"meta": {
				"pagination": {"total": 1, "count": 1, "perPage": 10, "currentPage": 1, "totalPages": 1},
				"links": {"next": ""}
			}
		}`))
	}))
	defer ts.Close()

	originalDNSRESTAPIBaseURL := dnsRESTAPIBaseURL
	defer func() { dnsRESTAPIBaseURL = originalDNSRESTAPIBaseURL }()
	dnsRESTAPIBaseURL = ts.URL

	domains, err := GetDNSDomains()
	if err != nil {
		t.Fatal(err)
	}
	if len(domains) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(domains))
	}
	if domains[0].Name != "example.com" {
		t.Errorf("expected name %q, got %q", "example.com", domains[0].Name)
	}
	if domains[0].ID != 1 {
		t.Errorf("expected id %d, got %d", 1, domains[0].ID)
	}
	if !domains[0].GeoIPEnabled {
		t.Errorf("expected GeoIPEnabled true, got false")
	}
}

func TestGetDNSDomains_unexpected_status(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	originalDNSRESTAPIBaseURL := dnsRESTAPIBaseURL
	defer func() { dnsRESTAPIBaseURL = originalDNSRESTAPIBaseURL }()
	dnsRESTAPIBaseURL = ts.URL

	_, err := GetDNSDomains()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
