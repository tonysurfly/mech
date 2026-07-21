package cmd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func TestSonarTCPCheck_GetResourceID(t *testing.T) {
	c := &SonarTCPCheck{Name: "prod"}
	if got := c.GetResourceID(); got != "prod" {
		t.Errorf("expected %q, got %q", "prod", got)
	}
}

func TestSonarTCPCheck_GetConstellixID(t *testing.T) {
	c := &SonarTCPCheck{ID: 42}
	if got := c.GetConstellixID(); got != 42 {
		t.Errorf("expected %d, got %d", 42, got)
	}
}

func TestSonarTCPCheck_SyncResourceDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/tcp/42" {
			t.Errorf("expected path %q, got %q", "/tcp/42", r.URL.Path)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	originalSonarRESTAPIBaseURL := sonarRESTAPIBaseURL
	defer func() { sonarRESTAPIBaseURL = originalSonarRESTAPIBaseURL }()
	sonarRESTAPIBaseURL = ts.URL

	c := &SonarTCPCheck{ID: 42, Name: "prod"}
	if err := c.SyncResourceDelete(42); err != nil {
		t.Fatal(err)
	}
}

func TestSonarTCPCheck_SyncResourceDelete_unexpected_status(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	originalSonarRESTAPIBaseURL := sonarRESTAPIBaseURL
	defer func() { sonarRESTAPIBaseURL = originalSonarRESTAPIBaseURL }()
	sonarRESTAPIBaseURL = ts.URL

	c := &SonarTCPCheck{ID: 42, Name: "prod"}
	if err := c.SyncResourceDelete(42); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExpectedSonarTCPCheck_UnmarshalYAML(t *testing.T) {
	data := `
name: prod
port: 443
alien: yes
`
	var obj ExpectedSonarTCPCheck
	if err := yaml.Unmarshal([]byte(data), &obj); err != nil {
		t.Fatal(err)
	}
	if len(obj.definedFieldsMap) != 2 {
		t.Fatalf("expected 2 defined fields, got %d", len(obj.definedFieldsMap))
	}
	if obj.definedFieldsMap["name"] != "Name" {
		t.Errorf("expected %q to be mapped to %q, got %q", "name", "Name", obj.definedFieldsMap["name"])
	}
	if obj.definedFieldsMap["port"] != "Port" {
		t.Errorf("expected %q to be mapped to %q, got %q", "port", "Port", obj.definedFieldsMap["port"])
	}
}

func TestExpectedSonarTCPCheck_Validate_no_mandatory(t *testing.T) {
	data := `
name: prod
port: 443
`
	var obj ExpectedSonarTCPCheck
	obj.mandatoryFields = []string{"name", "port", "host"}
	if err := yaml.Unmarshal([]byte(data), &obj); err != nil {
		t.Fatal(err)
	}
	err := obj.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	expected := `prod: mandatory field "host" is not defined`
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestExpectedSonarTCPCheck_GetImmutableStructFields(t *testing.T) {
	data := `
name: prod
host: 1.2.3.4
ipVersion: IPV4
`
	var obj ExpectedSonarTCPCheck
	if err := yaml.Unmarshal([]byte(data), &obj); err != nil {
		t.Fatal(err)
	}
	got := obj.GetImmutableStructFields()
	if len(got) != 2 {
		t.Fatalf("expected 2 immutable fields, got %d: %v", len(got), got)
	}
}

func TestExpectedSonarTCPCheck_SyncResourceUpdate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		expected := `{"name":"prod","port":443}`
		if string(body) != expected {
			t.Errorf("expected %q, got %q", expected, string(body))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	originalSonarRESTAPIBaseURL := sonarRESTAPIBaseURL
	defer func() { sonarRESTAPIBaseURL = originalSonarRESTAPIBaseURL }()
	sonarRESTAPIBaseURL = ts.URL

	data := `
name: prod
port: 443
ipVersion: IPV4
`
	var obj ExpectedSonarTCPCheck
	if err := yaml.Unmarshal([]byte(data), &obj); err != nil {
		t.Fatal(err)
	}
	if err := obj.SyncResourceUpdate(999); err != nil {
		t.Fatal(err)
	}
}

func TestExpectedSonarTCPCheck_SyncResourceCreate(t *testing.T) {
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
interval: THIRTYSECONDS
checkSites: [1]
`
	var obj ExpectedSonarTCPCheck
	if err := yaml.Unmarshal([]byte(data), &obj); err != nil {
		t.Fatal(err)
	}
	if err := obj.SyncResourceCreate(); err != nil {
		t.Fatal(err)
	}
}

func TestGetSonarTCPChecks(t *testing.T) {
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

	checks, err := GetSonarTCPChecks()
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

func TestGetSonarTCPChecks_unexpected_status(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	originalSonarRESTAPIBaseURL := sonarRESTAPIBaseURL
	defer func() { sonarRESTAPIBaseURL = originalSonarRESTAPIBaseURL }()
	sonarRESTAPIBaseURL = ts.URL

	if _, err := GetSonarTCPChecks(); err == nil {
		t.Fatal("expected error, got nil")
	}
}
