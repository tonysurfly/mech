package cmd

import (
	"encoding/json"
	"fmt"
	"testing"
)

type testExpectedResource struct {
	Name            string
	Port            int
	definedFields   []string
	immutableFields []string
	syncCalls       []string
}

func (ter *testExpectedResource) GetDefinedStructFieldNames() []string {
	return ter.definedFields
}

func (ter *testExpectedResource) GetImmutableStructFields() []string {
	return ter.immutableFields
}

func (ter *testExpectedResource) GetResource() interface{} {
	return ter
}

func (ter *testExpectedResource) GetResourceID() string {
	return ter.Name
}

func (ter *testExpectedResource) SyncResourceCreate() error {
	ter.syncCalls = append(ter.syncCalls, "create")
	return nil
}

func (ter *testExpectedResource) SyncResourceUpdate(id int) error {
	ter.syncCalls = append(ter.syncCalls, "update:"+fmt.Sprint(id))
	return nil
}

type testActiveResource struct {
	Name         string
	Port         int
	constellixID int
	syncCalls    []string
}

func (tar *testActiveResource) GetConstellixID() int {
	return tar.constellixID
}

func (tar *testActiveResource) GetResource() interface{} {
	return tar
}

func (tar *testActiveResource) GetResourceID() string {
	return tar.Name
}

func (tar *testActiveResource) SyncResourceDelete(id int) error {
	tar.syncCalls = append(tar.syncCalls, "delete:"+fmt.Sprint(id))
	return nil
}

func Test_Sync_create_dry(t *testing.T) {
	er := &testExpectedResource{
		Name: "Field1",
	}
	expCol := toResourceMatcher([]*testExpectedResource{er})

	reportToTestBuffer = true
	defer func() {
		reportToTestBuffer = false
		testBuffer.Reset()
	}()
	err := Sync(expCol, nil, false, false, "")

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	output := stripBashColors(testBuffer.String())
	expected := "create,Field1,\n"
	if output != expected {
		t.Errorf("want %q, got %q", expected, output)
		return
	}
	if len(er.syncCalls) != 0 {
		t.Errorf("want 0 sync calls, got %d", len(er.syncCalls))
	}
}

func Test_Sync_delete_dry(t *testing.T) {
	ar := &testActiveResource{
		Name:         "Field1",
		constellixID: 999,
	}
	actCol := toResourceMatcher([]*testActiveResource{ar})

	reportToTestBuffer = true
	defer func() {
		reportToTestBuffer = false
		testBuffer.Reset()
	}()
	err := Sync(nil, actCol, false, false, "")

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	output := stripBashColors(testBuffer.String())
	expected := "delete,Field1,Resource ID 999\n"
	if output != expected {
		t.Errorf("want %q, got %q", expected, output)
		return
	}
	if len(ar.syncCalls) != 0 {
		t.Errorf("want 0 sync calls, got %d", len(ar.syncCalls))
	}
}

func Test_Sync_ok_dry(t *testing.T) {
	er := &testExpectedResource{
		Name: "Field1",
	}
	expCol := toResourceMatcher([]*testExpectedResource{er})

	ar := &testActiveResource{
		Name:         "Field1",
		constellixID: 999,
	}
	actCol := toResourceMatcher([]*testActiveResource{ar})

	reportToTestBuffer = true
	defer func() {
		reportToTestBuffer = false
		testBuffer.Reset()
	}()
	err := Sync(expCol, actCol, false, false, "")

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	output := stripBashColors(testBuffer.String())
	expected := "ok,Field1,\n"
	if output != expected {
		t.Errorf("want %q, got %q", expected, output)
		return
	}
	if len(er.syncCalls) != 0 {
		t.Errorf("want 0 sync calls, got %d", len(er.syncCalls))
		return
	}
	if len(ar.syncCalls) != 0 {
		t.Errorf("want 0 sync calls, got %d", len(ar.syncCalls))
		return
	}
}

func Test_Sync_update_dry(t *testing.T) {
	er := &testExpectedResource{
		Name:          "Field1",
		Port:          80,
		definedFields: []string{"Port"},
	}
	expCol := toResourceMatcher([]*testExpectedResource{er})

	ar := &testActiveResource{
		Name:         "Field1",
		Port:         443,
		constellixID: 999,
	}
	actCol := toResourceMatcher([]*testActiveResource{ar})

	reportToTestBuffer = true
	defer func() {
		reportToTestBuffer = false
		testBuffer.Reset()
	}()
	err := Sync(expCol, actCol, false, false, "")

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	output := stripBashColors(testBuffer.String())
	expected := "update,Field1,\"Port:\n  443\n  80\"\n"
	if output != expected {
		t.Errorf("want %q, got %q", expected, output)
		return
	}
	if len(er.syncCalls) != 0 {
		t.Errorf("want 0 sync calls, got %d", len(er.syncCalls))
		return
	}
	if len(ar.syncCalls) != 0 {
		t.Errorf("want 0 sync calls, got %d", len(ar.syncCalls))
		return
	}
}

func Test_Sync_create_doit(t *testing.T) {
	er := &testExpectedResource{
		Name: "Field1",
	}
	expCol := toResourceMatcher([]*testExpectedResource{er})

	reportToTestBuffer = true
	defer func() {
		reportToTestBuffer = false
		testBuffer.Reset()
	}()
	err := Sync(expCol, nil, true, false, "")

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	output := stripBashColors(testBuffer.String())
	expected := "create,Field1,\n"
	if output != expected {
		t.Errorf("want %q, got %q", expected, output)
		return
	}
	if len(er.syncCalls) != 1 {
		t.Errorf("want 1 sync calls, got %d", len(er.syncCalls))
		return
	}
	if er.syncCalls[0] != "create" {
		t.Errorf("want create call, got %s", er.syncCalls[0])
		return
	}
}

func Test_Sync_delete_dry_doit(t *testing.T) {
	ar := &testActiveResource{
		Name:         "Field1",
		constellixID: 999,
	}
	actCol := toResourceMatcher([]*testActiveResource{ar})

	reportToTestBuffer = true
	defer func() {
		reportToTestBuffer = false
		testBuffer.Reset()
	}()
	err := Sync(nil, actCol, true, false, "")
	expectedErr := "resource deletion is not allowed. Use --remove flag to allow it"

	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if err.Error() != expectedErr {
		t.Errorf("want error %q, got %q", expectedErr, err.Error())
		return
	}
	output := stripBashColors(testBuffer.String())
	expected := "delete,Field1,Resource ID 999\n"
	if output != expected {
		t.Errorf("want %q, got %q", expected, output)
		return
	}
	if len(ar.syncCalls) != 0 {
		t.Errorf("want 0 sync calls, got %d", len(ar.syncCalls))
	}
}

func Test_Sync_delete_dry_doit_remove(t *testing.T) {
	ar := &testActiveResource{
		Name:         "Field1",
		constellixID: 999,
	}
	actCol := toResourceMatcher([]*testActiveResource{ar})

	reportToTestBuffer = true
	defer func() {
		reportToTestBuffer = false
		testBuffer.Reset()
	}()
	err := Sync(nil, actCol, true, true, "")

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	output := stripBashColors(testBuffer.String())
	expected := "delete,Field1,Resource ID 999\n"
	if output != expected {
		t.Errorf("want %q, got %q", expected, output)
		return
	}
	if len(ar.syncCalls) != 1 {
		t.Errorf("want 1 sync calls, got %d", len(ar.syncCalls))
	}
	if ar.syncCalls[0] != "delete:999" {
		t.Errorf("want delete:999 call, got %s", ar.syncCalls[0])
		return
	}
}

func Test_Sync_ok_doit(t *testing.T) {
	er := &testExpectedResource{
		Name: "Field1",
	}
	expCol := toResourceMatcher([]*testExpectedResource{er})

	ar := &testActiveResource{
		Name:         "Field1",
		constellixID: 999,
	}
	actCol := toResourceMatcher([]*testActiveResource{ar})

	reportToTestBuffer = true
	defer func() {
		reportToTestBuffer = false
		testBuffer.Reset()
	}()
	err := Sync(expCol, actCol, true, false, "")

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	output := stripBashColors(testBuffer.String())
	expected := "ok,Field1,\n"
	if output != expected {
		t.Errorf("want %q, got %q", expected, output)
		return
	}
	if len(er.syncCalls) != 0 {
		t.Errorf("want 0 sync calls, got %d", len(er.syncCalls))
		return
	}
	if len(ar.syncCalls) != 0 {
		t.Errorf("want 0 sync calls, got %d", len(ar.syncCalls))
		return
	}
}

func Test_Sync_update_doit(t *testing.T) {
	er := &testExpectedResource{
		Name:          "Field1",
		Port:          80,
		definedFields: []string{"Port"},
	}
	expCol := toResourceMatcher([]*testExpectedResource{er})

	ar := &testActiveResource{
		Name:         "Field1",
		Port:         443,
		constellixID: 999,
	}
	actCol := toResourceMatcher([]*testActiveResource{ar})

	reportToTestBuffer = true
	defer func() {
		reportToTestBuffer = false
		testBuffer.Reset()
	}()
	err := Sync(expCol, actCol, true, false, "")

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	output := stripBashColors(testBuffer.String())
	expected := "update,Field1,\"Port:\n  443\n  80\"\n"
	if output != expected {
		t.Errorf("want %q, got %q", expected, output)
		return
	}
	if len(er.syncCalls) != 1 {
		t.Errorf("want 1 sync calls, got %d", len(er.syncCalls))
		return
	}
	if er.syncCalls[0] != "update:999" {
		t.Errorf("want update:999 call, got %s", er.syncCalls[0])
		return
	}
	if len(ar.syncCalls) != 0 {
		t.Errorf("want 0 sync calls, got %d", len(ar.syncCalls))
		return
	}
}

func Test_Sync_update_dry_immutable(t *testing.T) {
	er := &testExpectedResource{
		Name:            "Field1",
		Port:            80,
		definedFields:   []string{"Port"},
		immutableFields: []string{"Port"},
	}
	expCol := toResourceMatcher([]*testExpectedResource{er})

	ar := &testActiveResource{
		Name:         "Field1",
		Port:         443,
		constellixID: 999,
	}
	actCol := toResourceMatcher([]*testActiveResource{ar})

	reportToTestBuffer = true
	defer func() {
		reportToTestBuffer = false
		testBuffer.Reset()
	}()
	err := Sync(expCol, actCol, false, false, "")

	if err == nil {
		t.Errorf("expected error")
		return
	}
	expected := `resource "Field1": found change in immutable field: Port`
	if err.Error() != expected {
		t.Errorf("want error %q, got %q", expected, err.Error())
		return
	}
}

func Test_Generate_payload_full(t *testing.T) {
	er := &testExpectedResource{
		Name:          "Field1",
		Port:          80,
		definedFields: []string{"Port"},
	}
	payload, err := generatePayload(er, er.definedFields, nil)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	expected := `{"Port":80}`
	payloadStr := string(payload)
	if payloadStr != expected {
		t.Errorf("want %q, got %q", expected, payloadStr)
		return
	}
}

func Test_Generate_payload_immutable(t *testing.T) {
	er := &testExpectedResource{
		Name:            "Field1",
		Port:            80,
		definedFields:   []string{"Port", "Name"},
		immutableFields: []string{"Port"},
	}
	payload, err := generatePayload(er, er.definedFields, nil)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	expected := `{"Name":"Field1","Port":80}`
	payloadStr := string(payload)
	if payloadStr != expected {
		t.Errorf("want %q, got %q", expected, payloadStr)
		return
	}
}

func Test_Generate_payload_excluded(t *testing.T) {
	er := &testExpectedResource{
		Name:            "Field1",
		Port:            80,
		definedFields:   []string{"Port", "Name"},
		immutableFields: []string{"Port"},
	}
	payload, err := generatePayload(er, er.definedFields, []string{"Name"})
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	expected := `{"Port":80}`
	payloadStr := string(payload)
	if payloadStr != expected {
		t.Errorf("want %q, got %q", expected, payloadStr)
		return
	}
}

func Test_Generate_payload_ipfilter(t *testing.T) {
	inRecord := `{"enabled":true,"geoFailover":false,"geoproximity":null,"ipfilter":{"id":1,"name":"World (Default)"},"ipfilterDrop":false,"mode":"failover","name":"","notes":"","region":"europe","ttl":60,"type":"A","value":{"enabled":true,"mode":"normal","values":[{"enabled":true,"order":1,"sonarCheckId":42040,"value":"159.69.18.28"},{"enabled":true,"order":2,"sonarCheckId":84732,"value":"5.161.66.36"}]}}`
	var recordObj DNSRecord
	err := json.Unmarshal([]byte(inRecord), &recordObj)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	// Check parsing
	if recordObj.IPFilter != 1 {
		t.Errorf("want 1, got %d", recordObj.IPFilter)
		return
	}

	payload, err := generatePayload(&recordObj, []string{"ipfilter"}, nil)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if string(payload) != `{"ipfilter":1}` {
		t.Errorf("want %q, got %q", `{"ipfilter":{"id":1}`, string(payload))
		return
	}
}

func Test_Generate_payload_ipfilter_empty(t *testing.T) {
	inRecord := `{"enabled":true,"geoFailover":false,"geoproximity":null,"ipfilter":null,"ipfilterDrop":false,"mode":"failover","name":"","notes":"","region":"europe","ttl":60,"type":"A","value":{"enabled":true,"mode":"normal","values":[{"enabled":true,"order":1,"sonarCheckId":42040,"value":"159.69.18.28"},{"enabled":true,"order":2,"sonarCheckId":84732,"value":"5.161.66.36"}]}}`
	var recordObj ExpectedDNSRecord
	err := json.Unmarshal([]byte(inRecord), &recordObj)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	// Check parsing
	if recordObj.IPFilter != nil {
		t.Errorf("want nil, got %v", recordObj.IPFilter)
		return
	}

	payload, err := generatePayload(&recordObj, []string{"ipfilter"}, nil)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if string(payload) != `{"ipfilter":null}` {
		t.Errorf("want %q, got %q", `{"ipfilter":null}`, string(payload))
		return
	}
}

func Test_Generate_payload_geoproximity_integer(t *testing.T) {
	inRecord := `{"enabled":true,"geoFailover":false,"geoproximity":{"id":10,"name":"gp 1"},"ipfilterDrop":false,"mode":"failover","name":"","notes":"","region":"europe","ttl":60,"type":"A","value":{"enabled":true,"mode":"normal","values":[{"enabled":true,"order":1,"sonarCheckId":42040,"value":"159.69.18.28"},{"enabled":true,"order":2,"sonarCheckId":84732,"value":"5.161.66.36"}]}}`
	var recordObj DNSRecord
	err := json.Unmarshal([]byte(inRecord), &recordObj)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	// Check parsing
	if recordObj.GeoProximity != 10 {
		t.Errorf("want 10, got %d", recordObj.GeoProximity)
		return
	}

	payload, err := generatePayload(&recordObj, []string{"geoproximity"}, nil)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if string(payload) != `{"geoproximity":10}` {
		t.Errorf("want %q, got %q", `{"geoproximity":10`, string(payload))
		return
	}
}

func Test_Generate_payload_geoproximity_empty(t *testing.T) {
	inRecord := `{"enabled":true,"geoFailover":false,"ipfilterDrop":false,"mode":"failover","name":"","notes":"","region":"europe","ttl":60,"type":"A","value":{"enabled":true,"mode":"normal","values":[{"enabled":true,"order":1,"sonarCheckId":42040,"value":"159.69.18.28"},{"enabled":true,"order":2,"sonarCheckId":84732,"value":"5.161.66.36"}]}}`
	var recordObj DNSRecord
	err := json.Unmarshal([]byte(inRecord), &recordObj)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	// Check parsing
	if recordObj.GeoProximity != nil {
		t.Errorf("want nil, got %d", recordObj.GeoProximity)
		return
	}

	payload, err := generatePayload(&recordObj, []string{"geoproximity"}, nil)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if string(payload) != `{"geoproximity":null}` {
		t.Errorf("want %q, got %q", `{"geoproximity":null`, string(payload))
		return
	}
}
