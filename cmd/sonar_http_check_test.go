package cmd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func TestSonarHTTPCheck_GetResourceID(t *testing.T) {
	c := &SonarHTTPCheck{Name: "prod"}
	if got := c.GetResourceID(); got != "prod" {
		t.Errorf("expected %q, got %q", "prod", got)
	}
}

func TestSonarHTTPCheck_GetConstellixID(t *testing.T) {
	c := &SonarHTTPCheck{ID: 42}
	if got := c.GetConstellixID(); got != 42 {
		t.Errorf("expected %d, got %d", 42, got)
	}
}

func TestSonarHTTPCheck_SyncResourceDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/http/42" {
			t.Errorf("expected path %q, got %q", "/http/42", r.URL.Path)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	originalSonarRESTAPIBaseURL := sonarRESTAPIBaseURL
	defer func() { sonarRESTAPIBaseURL = originalSonarRESTAPIBaseURL }()
	sonarRESTAPIBaseURL = ts.URL

	c := &SonarHTTPCheck{ID: 42, Name: "prod"}
	if err := c.SyncResourceDelete(42); err != nil {
		t.Fatal(err)
	}
}

func TestSonarHTTPCheck_SyncResourceDelete_unexpected_status(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	originalSonarRESTAPIBaseURL := sonarRESTAPIBaseURL
	defer func() { sonarRESTAPIBaseURL = originalSonarRESTAPIBaseURL }()
	sonarRESTAPIBaseURL = ts.URL

	c := &SonarHTTPCheck{ID: 42, Name: "prod"}
	if err := c.SyncResourceDelete(42); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExpectedSonarHTTPCheck_SyncResourceCreate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	originalSonarRESTAPIBaseURL := sonarRESTAPIBaseURL
	defer func() { sonarRESTAPIBaseURL = originalSonarRESTAPIBaseURL }()
	sonarRESTAPIBaseURL = ts.URL

	data := `
name: prod
host: 1.2.3.4
ipVersion: IPV4
port: 443
protocolType: HTTPS
interval: ONEMINUTE
checkSites: [1]
`
	var obj ExpectedSonarHTTPCheck
	if err := yaml.Unmarshal([]byte(data), &obj); err != nil {
		t.Fatal(err)
	}
	if err := obj.SyncResourceCreate(); err != nil {
		t.Fatal(err)
	}
}

func TestGetSonarHTTPChecks(t *testing.T) {
	cachedSonarHTTPChecks = nil
	defer func() { cachedSonarHTTPChecks = nil }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id": 1, "name": "prod", "host": "1.2.3.4", "port": 443}]`))
	}))
	defer ts.Close()

	originalSonarRESTAPIBaseURL := sonarRESTAPIBaseURL
	defer func() { sonarRESTAPIBaseURL = originalSonarRESTAPIBaseURL }()
	sonarRESTAPIBaseURL = ts.URL

	checks, err := GetSonarHTTPChecks()
	if err != nil {
		t.Fatal(err)
	}
	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks))
	}
	if checks[0].Name != "prod" {
		t.Errorf("expected name %q, got %q", "prod", checks[0].Name)
	}
}

func TestGetSonarHTTPChecks_cached(t *testing.T) {
	cachedSonarHTTPChecks = []*SonarHTTPCheck{{ID: 1, Name: "cached"}}
	defer func() { cachedSonarHTTPChecks = nil }()

	// No server configured; a cache miss would fail to connect.
	originalSonarRESTAPIBaseURL := sonarRESTAPIBaseURL
	defer func() { sonarRESTAPIBaseURL = originalSonarRESTAPIBaseURL }()
	sonarRESTAPIBaseURL = "http://127.0.0.1:0"

	checks, err := GetSonarHTTPChecks()
	if err != nil {
		t.Fatal(err)
	}
	if len(checks) != 1 || checks[0].Name != "cached" {
		t.Fatalf("expected cached result, got %v", checks)
	}
}

func TestGetSonarHTTPCheckStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/http/42/status" {
			t.Errorf("expected path %q, got %q", "/http/42/status", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "UP"}`))
	}))
	defer ts.Close()

	originalSonarRESTAPIBaseURL := sonarRESTAPIBaseURL
	defer func() { sonarRESTAPIBaseURL = originalSonarRESTAPIBaseURL }()
	sonarRESTAPIBaseURL = ts.URL

	status, err := GetSonarHTTPCheckStatus(42)
	if err != nil {
		t.Fatal(err)
	}
	if status != StatusUp {
		t.Errorf("expected %q, got %q", StatusUp, status)
	}
}

func TestExpectedSonarHTTPCheck_UnmarshalYAML(t *testing.T) {
	data := `
name: prod
port: 80
alien: yes
`
	var obj ExpectedSonarHTTPCheck
	err := yaml.Unmarshal([]byte(data), &obj)
	if err != nil {
		t.Error(err)
		return
	}
	if len(obj.definedFieldsMap) != 2 {
		t.Errorf("wrong length: got %d, want %d", len(obj.definedFieldsMap), 2)
		return
	}
	if obj.definedFieldsMap["name"] != "Name" {
		t.Errorf("expected %q to be mapped to %q, got %q", "name", "Name", obj.definedFieldsMap["name"])
		return
	}
	if obj.definedFieldsMap["port"] != "Port" {
		t.Errorf("expected %q to be mapped to %q, got %q", "port", "Port", obj.definedFieldsMap["port"])
		return
	}
}

func TestExpectedSonarHTTPCheck_Validate_no_mandatory(t *testing.T) {
	data := `
name: prod
port: 80
`
	var obj ExpectedSonarHTTPCheck
	// Stub mandatory fields
	obj.mandatoryFields = []string{"name", "port", "host"}
	err := yaml.Unmarshal([]byte(data), &obj)
	if err != nil {
		t.Error(err)
		return
	}
	err = obj.Validate()
	if err == nil {
		t.Error("expected error, got nil")
		return
	}
	expected := "prod: mandatory field \"host\" is not defined"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
		return
	}
}

func TestExpectedSonarHTTPCheck_SyncResourceUpdate_exclude_immutable(t *testing.T) {
	// Set up test environment
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		expected := `{"name":"prod","port":80}`
		if string(body) != expected {
			t.Errorf("expected %q, got %q", expected, string(body))
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer ts.Close()

	originalSonarRESTAPIBaseURL := sonarRESTAPIBaseURL
	defer func() {
		sonarRESTAPIBaseURL = originalSonarRESTAPIBaseURL
	}()
	sonarRESTAPIBaseURL = ts.URL
	data := `
name: prod
port: 80
ipVersion: IPV4
`
	var obj ExpectedSonarHTTPCheck
	err := yaml.Unmarshal([]byte(data), &obj)
	if err != nil {
		t.Error(err)
		return
	}
	err = obj.SyncResourceUpdate(999)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestGetSonarHTTPChecks_cachedEmpty(t *testing.T) {
	cachedSonarHTTPChecks = nil
	defer func() { cachedSonarHTTPChecks = nil }()

	requests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer ts.Close()

	originalSonarRESTAPIBaseURL := sonarRESTAPIBaseURL
	defer func() { sonarRESTAPIBaseURL = originalSonarRESTAPIBaseURL }()
	sonarRESTAPIBaseURL = ts.URL

	for i := range 2 {
		checks, err := GetSonarHTTPChecks()
		if err != nil {
			t.Fatal(err)
		}
		if len(checks) != 0 {
			t.Fatalf("call %d: expected 0 checks, got %d", i, len(checks))
		}
	}
	if requests != 1 {
		t.Fatalf("expected 1 request, got %d", requests)
	}
}
