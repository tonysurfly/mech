package cmd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func TestGeoProximity_GetResourceID(t *testing.T) {
	gp := &GeoProximity{Name: "test"}
	if got := gp.GetResourceID(); got != "test" {
		t.Errorf("expected %q, got %q", "test", got)
	}
}

func TestGeoProximity_GetConstellixID(t *testing.T) {
	gp := &GeoProximity{ID: 42}
	if got := gp.GetConstellixID(); got != 42 {
		t.Errorf("expected %d, got %d", 42, got)
	}
}

func TestGeoProximity_GetResource(t *testing.T) {
	gp := &GeoProximity{Name: "test"}
	if got := gp.GetResource(); got != gp {
		t.Errorf("expected %v, got %v", gp, got)
	}
}

func TestGeoProximity_SyncResourceDelete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/geoproximities/42" {
			t.Errorf("expected path %q, got %q", "/geoproximities/42", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	originalDNSRESTAPIBaseURL := dnsRESTAPIBaseURL
	defer func() { dnsRESTAPIBaseURL = originalDNSRESTAPIBaseURL }()
	dnsRESTAPIBaseURL = ts.URL

	gp := &GeoProximity{ID: 42, Name: "test"}
	if err := gp.SyncResourceDelete(42); err != nil {
		t.Fatal(err)
	}
}

func TestGeoProximity_SyncResourceDelete_unexpected_status(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	originalDNSRESTAPIBaseURL := dnsRESTAPIBaseURL
	defer func() { dnsRESTAPIBaseURL = originalDNSRESTAPIBaseURL }()
	dnsRESTAPIBaseURL = ts.URL

	gp := &GeoProximity{ID: 42, Name: "test"}
	if err := gp.SyncResourceDelete(42); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExpectedGeoProximity_UnmarshalYAML(t *testing.T) {
	data := `
name: test
longitude: 1.23
latitude: 4.56
`
	var obj ExpectedGeoProximity
	if err := yaml.Unmarshal([]byte(data), &obj); err != nil {
		t.Fatal(err)
	}
	if len(obj.definedFieldsMap) != 3 {
		t.Fatalf("expected 3 defined fields, got %d", len(obj.definedFieldsMap))
	}
	if obj.definedFieldsMap["name"] != "Name" {
		t.Errorf("expected %q to be mapped to %q, got %q", "name", "Name", obj.definedFieldsMap["name"])
	}
	if obj.Name != "test" {
		t.Errorf("expected name %q, got %q", "test", obj.Name)
	}
}

func TestExpectedGeoProximity_Validate_no_mandatory(t *testing.T) {
	data := `
name: test
longitude: 1.23
`
	var obj ExpectedGeoProximity
	if err := yaml.Unmarshal([]byte(data), &obj); err != nil {
		t.Fatal(err)
	}
	err := obj.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	expected := `test: mandatory field "latitude" is not defined`
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestExpectedGeoProximity_Validate_ok(t *testing.T) {
	data := `
name: test
longitude: 1.23
latitude: 4.56
`
	var obj ExpectedGeoProximity
	if err := yaml.Unmarshal([]byte(data), &obj); err != nil {
		t.Fatal(err)
	}
	if err := obj.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestExpectedGeoProximity_GetImmutableStructFields(t *testing.T) {
	data := `
name: test
longitude: 1.23
latitude: 4.56
`
	var obj ExpectedGeoProximity
	if err := yaml.Unmarshal([]byte(data), &obj); err != nil {
		t.Fatal(err)
	}
	if got := obj.GetImmutableStructFields(); len(got) != 0 {
		t.Errorf("expected no immutable fields, got %v", got)
	}
}

func TestExpectedGeoProximity_SyncResourceUpdate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		expected := `{"name":"test"}`
		if string(body) != expected {
			t.Errorf("expected body %q, got %q", expected, string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	originalDNSRESTAPIBaseURL := dnsRESTAPIBaseURL
	defer func() { dnsRESTAPIBaseURL = originalDNSRESTAPIBaseURL }()
	dnsRESTAPIBaseURL = ts.URL

	data := `
name: test
`
	var obj ExpectedGeoProximity
	if err := yaml.Unmarshal([]byte(data), &obj); err != nil {
		t.Fatal(err)
	}
	if err := obj.SyncResourceUpdate(42); err != nil {
		t.Fatal(err)
	}
}

func TestExpectedGeoProximity_SyncResourceCreate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	originalDNSRESTAPIBaseURL := dnsRESTAPIBaseURL
	defer func() { dnsRESTAPIBaseURL = originalDNSRESTAPIBaseURL }()
	dnsRESTAPIBaseURL = ts.URL

	data := `
name: test
longitude: 1.23
latitude: 4.56
`
	var obj ExpectedGeoProximity
	if err := yaml.Unmarshal([]byte(data), &obj); err != nil {
		t.Fatal(err)
	}
	if err := obj.SyncResourceCreate(); err != nil {
		t.Fatal(err)
	}
}

func TestGetGeoProximities(t *testing.T) {
	cachedGeoProximities = nil
	defer func() { cachedGeoProximities = nil }()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": [
				{"id": 1, "name": "test", "longitude": 1.23, "latitude": 4.56}
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

	geops, err := GetGeoProximities()
	if err != nil {
		t.Fatal(err)
	}
	if len(geops) != 1 {
		t.Fatalf("expected 1 geoproximity, got %d", len(geops))
	}
	if geops[0].Name != "test" {
		t.Errorf("expected name %q, got %q", "test", geops[0].Name)
	}
}

func TestGetGeoProximities_cached(t *testing.T) {
	cachedGeoProximities = []*GeoProximity{{ID: 1, Name: "cached"}}
	defer func() { cachedGeoProximities = nil }()

	// No server configured; a cache miss would fail to connect.
	originalDNSRESTAPIBaseURL := dnsRESTAPIBaseURL
	defer func() { dnsRESTAPIBaseURL = originalDNSRESTAPIBaseURL }()
	dnsRESTAPIBaseURL = "http://127.0.0.1:0"

	geops, err := GetGeoProximities()
	if err != nil {
		t.Fatal(err)
	}
	if len(geops) != 1 || geops[0].Name != "cached" {
		t.Fatalf("expected cached result, got %v", geops)
	}
}

func TestGetGeoProximities_cachedEmpty(t *testing.T) {
	cachedGeoProximities = nil
	defer func() { cachedGeoProximities = nil }()

	requests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": [],
			"meta": {
				"pagination": {"total": 0, "count": 0, "perPage": 10, "currentPage": 1, "totalPages": 1},
				"links": {"next": ""}
			}
		}`))
	}))
	defer ts.Close()

	originalDNSRESTAPIBaseURL := dnsRESTAPIBaseURL
	defer func() { dnsRESTAPIBaseURL = originalDNSRESTAPIBaseURL }()
	dnsRESTAPIBaseURL = ts.URL

	for i := range 2 {
		geops, err := GetGeoProximities()
		if err != nil {
			t.Fatal(err)
		}
		if len(geops) != 0 {
			t.Fatalf("call %d: expected 0 geoproximities, got %d", i, len(geops))
		}
	}
	if requests != 1 {
		t.Fatalf("expected 1 request, got %d", requests)
	}
}
