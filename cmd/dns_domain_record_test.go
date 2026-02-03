package cmd

import (
	"encoding/json"
	"reflect"
	"testing"

	"golang.org/x/exp/slices"
	yaml "gopkg.in/yaml.v3"
)

func TestExpectedDNSRecord_Standard_UnmarshalYAML(t *testing.T) {
	data := `
name: abc
type: A
ttl: 60
mode: standard
region: default
enabled: true
value:
  - value: 1.1.1.1
    enabled: true
`
	var obj ExpectedDNSRecord
	err := yaml.Unmarshal([]byte(data), &obj)
	if err != nil {
		t.Error(err)
		return
	}
	if obj.Name != "abc" {
		t.Errorf("expected %q, got %q", "abc", obj.Name)
		return
	}
	if obj.Mode != "standard" {
		t.Errorf("expected %q, got %q", "standard", obj.Mode)
		return
	}
	expected := DNSStandardItemValue{
		Value:   "1.1.1.1",
		Enabled: true,
	}
	res, ok := obj.Value.([]*DNSStandardItemValue)
	if !ok {
		t.Errorf("unexpected type %T", obj.Value)
		return
	}
	if len(res) != 1 {
		t.Errorf("expected 1, got %d", len(res))
		return
	}
	value := res[0]
	if value.Value != expected.Value {
		t.Errorf("expected %q, got %q", expected.Value, value.Value)
		return
	}
	if value.Enabled != expected.Enabled {
		t.Errorf("expected %t, got %t", expected.Enabled, value.Enabled)
		return
	}
}

func TestExpectedDNSRecord_Standard_UnmarshalJSON(t *testing.T) {
	data := `{"id":31847357,"name":"abc","type":"A","ttl":600,"mode":"standard","region":"default","ipfilter":null,"ipfilterDrop":false,"geoFailover":false,"geoproximity":null,"enabled":true,"value":[{"value":"1.1.1.1","enabled":true}],"lastValues":{"roundRobinFailover":[],"standard":[{"value":"8.8.8.8","enabled":true}],"failover":{"enabled":false,"mode":"normal","values":[]},"pools":[]},"notes":"","skipLookup":null,"domain":{"id":1004580,"name":"surfly.gratis","status":"ACTIVE","geoip":false,"gtd":false,"tags":[],"createdAt":"2022-12-28T15:13:57+00:00","updatedAt":"2022-12-29T19:35:29+00:00","links":{"self":"http:\/\/api.dns.constellix.com\/v4\/domains\/1004580","records":"http:\/\/api.dns.constellix.com\/v4\/domains\/1004580\/records","history":"http:\/\/api.dns.constellix.com\/v4\/domains\/1004580\/history","nameservers":"http:\/\/api.dns.constellix.com\/v4\/domains\/1004580\/nameservers","analytics":"http:\/\/api.dns.constellix.com\/v4\/domains\/1004580\/analytics"}},"contacts":[],"links":{"self":"http:\/\/api.dns.constellix.com\/v4\/domains\/1004580\/records\/31847357","domain":"http:\/\/api.dns.constellix.com\/v4\/domains\/1004580"}}`
	var obj DNSRecord
	err := json.Unmarshal([]byte(data), &obj)
	if err != nil {
		t.Error(err)
		return
	}
	if obj.Name != "abc" {
		t.Errorf("expected %q, got %q", "abc", obj.Name)
		return
	}
	if obj.Mode != "standard" {
		t.Errorf("expected %q, got %q", "standard", obj.Mode)
		return
	}
	expected := DNSStandardItemValue{
		Value:   "1.1.1.1",
		Enabled: true,
	}
	res, ok := obj.Value.([]*DNSStandardItemValue)
	if !ok {
		t.Errorf("unexpected type %T", obj.Value)
		return
	}
	if len(res) != 1 {
		t.Errorf("expected 1, got %d", len(res))
		return
	}
	value := res[0]
	if value.Value != expected.Value {
		t.Errorf("expected %q, got %q", expected.Value, value.Value)
		return
	}
	if value.Enabled != expected.Enabled {
		t.Errorf("expected %t, got %t", expected.Enabled, value.Enabled)
		return
	}
}

func TestExpectedDNSRecord_Failover_UnmarshalYAML(t *testing.T) {
	data := `
name: abc
type: A
ttl: 60
mode: failover
region: default
enabled: true
value:
  enabled: true
  mode: normal
  values:
    - value: 1.1.1.1
      enabled: true
      order: 1
      sonarCheckId: 123
      active: false
      failed: true
      status: DOWN
    - value: 1.1.1.2
      enabled: true
      order: 2
      sonarCheckId: null
      active: false
      failed: true
      status: N/A
`
	var obj ExpectedDNSRecord
	err := yaml.Unmarshal([]byte(data), &obj)
	if err != nil {
		t.Error(err)
		return
	}
	if obj.Name != "abc" {
		t.Errorf("expected %q, got %q", "abc", obj.Name)
		return
	}
	if obj.Mode != "failover" {
		t.Errorf("expected %q, got %q", "failover", obj.Mode)
		return
	}
	expected := DNSFailoverValue{
		Enabled: true,
		Mode:    "normal",
		Values: []*DNSFailoverItemValue{
			{
				Value:        "1.1.1.1",
				Enabled:      true,
				Order:        1,
				SonarCheckID: 123,
			},
			{
				Value:        "1.1.1.2",
				Enabled:      true,
				Order:        2,
				SonarCheckID: 0,
			},
		},
	}
	res, ok := obj.Value.(*DNSFailoverValue)
	if !ok {
		t.Errorf("unexpected type %T", obj.Value)
		return
	}
	if res.Enabled != expected.Enabled {
		t.Errorf("expected %t, got %t", expected.Enabled, res.Enabled)
		return
	}
	if res.Mode != expected.Mode {
		t.Errorf("expected %q, got %q", expected.Mode, res.Mode)
		return
	}
	if len(res.Values) != 2 {
		t.Errorf("expected 2, got %d", len(res.Values))
		return
	}
	res1 := res.Values[0]
	if res1.Value != expected.Values[0].Value {
		t.Errorf("expected %q, got %q", expected.Values[0].Value, res1.Value)
		return
	}
	if res1.Enabled != expected.Values[0].Enabled {
		t.Errorf("expected %t, got %t", expected.Values[0].Enabled, res1.Enabled)
		return
	}
	if res1.Order != expected.Values[0].Order {
		t.Errorf("expected %d, got %d", expected.Values[0].Order, res1.Order)
		return
	}
	if res1.SonarCheckID != expected.Values[0].SonarCheckID {
		t.Errorf("expected %d, got %d", expected.Values[0].SonarCheckID, res1.SonarCheckID)
		return
	}
	res2 := res.Values[1]
	if res2.SonarCheckID != expected.Values[1].SonarCheckID {
		t.Errorf("expected %d, got %d", expected.Values[1].SonarCheckID, res2.SonarCheckID)
		return
	}
}

func TestExpectedDNSRecord_Failover_UnmarshalJSON(t *testing.T) {
	data := `{"id":31847262,"name":"abc","type":"A","ttl":60,"mode":"failover","region":"default","ipfilter":null,"ipfilterDrop":false,"geoFailover":false,"geoproximity":null,"enabled":true,"value":{"enabled":true,"mode":"normal","values":[{"value":"159.69.18.28","order":1,"sonarCheckId":84874,"enabled":true,"active":false,"failed":true,"status":"DOWN"},{"value":"1.1.1.1","order":2,"sonarCheckId":null,"enabled":true,"active":false,"failed":false,"status":"N\/A"}]}}`
	var obj ExpectedDNSRecord
	err := json.Unmarshal([]byte(data), &obj)
	if err != nil {
		t.Error(err)
		return
	}
	if obj.Name != "abc" {
		t.Errorf("expected %q, got %q", "abc", obj.Name)
		return
	}
	if obj.Mode != "failover" {
		t.Errorf("expected %q, got %q", "failover", obj.Mode)
		return
	}
	expected := DNSFailoverValue{
		Enabled: true,
		Mode:    "normal",
		Values: []*DNSFailoverItemValue{
			{
				Value:        "159.69.18.28",
				Enabled:      true,
				Order:        1,
				SonarCheckID: 84874,
			},
			{
				Value:        "1.1.1.2",
				Enabled:      true,
				Order:        2,
				SonarCheckID: 0,
			},
		},
	}
	res, ok := obj.Value.(*DNSFailoverValue)
	if !ok {
		t.Errorf("unexpected type %T", obj.Value)
		return
	}
	if res.Enabled != expected.Enabled {
		t.Errorf("expected %t, got %t", expected.Enabled, res.Enabled)
		return
	}
	if res.Mode != expected.Mode {
		t.Errorf("expected %q, got %q", expected.Mode, res.Mode)
		return
	}
	if len(res.Values) != 2 {
		t.Errorf("expected 2, got %d", len(res.Values))
		return
	}
	res1 := res.Values[0]
	if res1.Value != expected.Values[0].Value {
		t.Errorf("expected %q, got %q", expected.Values[0].Value, res1.Value)
		return
	}
	if res1.Enabled != expected.Values[0].Enabled {
		t.Errorf("expected %t, got %t", expected.Values[0].Enabled, res1.Enabled)
		return
	}
	if res1.Order != expected.Values[0].Order {
		t.Errorf("expected %d, got %d", expected.Values[0].Order, res1.Order)
		return
	}
	if res1.SonarCheckID != expected.Values[0].SonarCheckID {
		t.Errorf("expected %d, got %d", expected.Values[0].SonarCheckID, res1.SonarCheckID)
		return
	}
	res2 := res.Values[1]
	if res2.SonarCheckID != expected.Values[1].SonarCheckID {
		t.Errorf("expected %d, got %d", expected.Values[1].SonarCheckID, res2.SonarCheckID)
		return
	}
}

func TestExpectedDNSRecord_RRFailover_UnmarshalYAML(t *testing.T) {
	data := `
name: abc
type: A
ttl: 60
mode: roundrobin-failover
region: default
enabled: true
value:
  - value: 1.1.1.1
    enabled: true
    order: 1
    sonarCheckId: 123
  - value: 1.1.1.2
    enabled: true
    order: 2
    sonarCheckId: null
`
	var obj ExpectedDNSRecord
	err := yaml.Unmarshal([]byte(data), &obj)
	if err != nil {
		t.Error(err)
		return
	}
	if obj.Name != "abc" {
		t.Errorf("expected %q, got %q", "abc", obj.Name)
		return
	}
	if obj.Mode != "roundrobin-failover" {
		t.Errorf("expected %q, got %q", "roundrobin-failover", obj.Mode)
		return
	}
	expected := []*DNSFailoverItemValue{
		{
			Value:        "1.1.1.1",
			Enabled:      true,
			Order:        1,
			SonarCheckID: 123,
		},
		{
			Value:        "1.1.1.2",
			Enabled:      true,
			Order:        2,
			SonarCheckID: 0,
		},
	}
	res, ok := obj.Value.([]*DNSFailoverItemValue)
	if !ok {
		t.Errorf("unexpected type %T", obj.Value)
		return
	}
	if len(res) != 2 {
		t.Errorf("expected 2, got %d", len(res))
		return
	}
	res1 := res[0]
	if res1.Value != expected[0].Value {
		t.Errorf("expected %q, got %q", expected[0].Value, res1.Value)
		return
	}
	if res1.Enabled != expected[0].Enabled {
		t.Errorf("expected %t, got %t", expected[0].Enabled, res1.Enabled)
		return
	}
	if res1.Order != expected[0].Order {
		t.Errorf("expected %d, got %d", expected[0].Order, res1.Order)
		return
	}
	if res1.SonarCheckID != expected[0].SonarCheckID {
		t.Errorf("expected %d, got %d", expected[0].SonarCheckID, res1.SonarCheckID)
		return
	}
	res2 := res[1]
	if res2.SonarCheckID != expected[1].SonarCheckID {
		t.Errorf("expected %d, got %d", expected[1].SonarCheckID, res2.SonarCheckID)
		return
	}
}

func TestExpectedDNSRecord_RRFailover_UnmarshalJSON(t *testing.T) {
	data := `{"id":31847262,"name":"abc","type":"A","ttl":60,"mode":"roundrobin-failover","region":"default","ipfilter":null,"ipfilterDrop":false,"geoFailover":false,"geoproximity":null,"enabled":true,"value":[{"value":"159.69.18.28","order":1,"sonarCheckId":84874,"enabled":true,"active":false,"failed":true,"status":"DOWN"},{"value":"1.1.1.1","order":2,"sonarCheckId":null,"enabled":true,"active":false,"failed":false,"status":"N\/A"}]}`
	var obj ExpectedDNSRecord
	err := json.Unmarshal([]byte(data), &obj)
	if err != nil {
		t.Error(err)
		return
	}
	if obj.Name != "abc" {
		t.Errorf("expected %q, got %q", "abc", obj.Name)
		return
	}
	if obj.Mode != "roundrobin-failover" {
		t.Errorf("expected %q, got %q", "failover", obj.Mode)
		return
	}
	expected := []*DNSFailoverItemValue{
		{
			Value:        "159.69.18.28",
			Enabled:      true,
			Order:        1,
			SonarCheckID: 84874,
		},
		{
			Value:        "1.1.1.2",
			Enabled:      true,
			Order:        2,
			SonarCheckID: 0,
		},
	}
	res, ok := obj.Value.([]*DNSFailoverItemValue)
	if !ok {
		t.Errorf("unexpected type %T", obj.Value)
		return
	}
	if len(res) != 2 {
		t.Errorf("expected 2, got %d", len(res))
		return
	}
	res1 := res[0]
	if res1.Value != expected[0].Value {
		t.Errorf("expected %q, got %q", expected[0].Value, res1.Value)
		return
	}
	if res1.Enabled != expected[0].Enabled {
		t.Errorf("expected %t, got %t", expected[0].Enabled, res1.Enabled)
		return
	}
	if res1.Order != expected[0].Order {
		t.Errorf("expected %d, got %d", expected[0].Order, res1.Order)
		return
	}
	if res1.SonarCheckID != expected[0].SonarCheckID {
		t.Errorf("expected %d, got %d", expected[0].SonarCheckID, res1.SonarCheckID)
		return
	}
	res2 := res[1]
	if res2.SonarCheckID != expected[1].SonarCheckID {
		t.Errorf("expected %d, got %d", expected[1].SonarCheckID, res2.SonarCheckID)
		return
	}
}

func TestExpectedDNSRecord_Pools_UnmarshalYAML(t *testing.T) {
	data := `
name: abc
type: A
ttl: 60
mode: pools
region: default
enabled: true
value:
  - 1
  - 3
`
	var obj ExpectedDNSRecord
	err := yaml.Unmarshal([]byte(data), &obj)
	if err != nil {
		t.Error(err)
		return
	}
	if obj.Name != "abc" {
		t.Errorf("expected %q, got %q", "abc", obj.Name)
		return
	}
	if obj.Mode != "pools" {
		t.Errorf("expected %q, got %q", "pools", obj.Mode)
		return
	}
	expected := []int{1, 3}

	res, ok := obj.Value.([]int)
	if !ok {
		t.Errorf("unexpected type %T", obj.Value)
		return
	}
	if len(res) != 2 {
		t.Errorf("expected 2, got %d", len(res))
		return
	}
	if slices.Compare(res, expected) != 0 {
		t.Errorf("expected %v, got %v", expected, res)
		return
	}
}

func TestMXRecord(t *testing.T) {
	data := `{"id":12494732,"name":"abc","type":"MX","ttl":86400,"mode":"standard","region":"default","ipfilter":null,"ipfilterDrop":false,"geoFailover":false,"geoproximity":null,"enabled":true,"value":[{"server":"aspmx.l.google.com.","priority":10,"enabled":true}]}`
	var objJ ExpectedDNSRecord
	err := json.Unmarshal([]byte(data), &objJ)
	if err != nil {
		t.Error(err)
		return
	}
	var objY ExpectedDNSRecord
	err = yaml.Unmarshal([]byte(data), &objY)
	if err != nil {
		t.Error(err)
		return
	}

	expectedValue := []*DNSMXStandardItemValue{
		{
			Server:   "aspmx.l.google.com.",
			Priority: 10,
			Enabled:  true,
		},
	}

	Compare := func(obj ExpectedDNSRecord, expectedValue []*DNSMXStandardItemValue, t *testing.T) {
		if objJ.Name != "abc" {
			t.Errorf("expected %q, got %q", "abc", objJ.Name)
			return
		}
		if objJ.Type != "MX" {
			t.Errorf("expected %q, got %q", "MX", objJ.Type)
			return
		}

		res, ok := objJ.Value.([]*DNSMXStandardItemValue)
		if !ok {
			t.Errorf("unexpected type %T", objJ.Value)
			return
		}
		if len(res) != 1 {
			t.Errorf("expected 1, got %d", len(res))
			return
		}
		if res[0].Server != expectedValue[0].Server {
			t.Errorf("expected %q, got %q", expectedValue[0].Server, res[0].Server)
			return
		}
		if res[0].Priority != expectedValue[0].Priority {
			t.Errorf("expected %d, got %d", expectedValue[0].Priority, res[0].Priority)
			return
		}
		if res[0].Enabled != expectedValue[0].Enabled {
			t.Errorf("expected %t, got %t", expectedValue[0].Enabled, res[0].Enabled)
			return
		}
	}

	Compare(objJ, expectedValue, t)
	Compare(objY, expectedValue, t)
}

func TestTXTRecord(t *testing.T) {
	data := `{"id":12494733,"name":"abc","type":"TXT","ttl":300,"mode":"standard","region":"default","ipfilter":null,"ipfilterDrop":false,"geoFailover":false,"geoproximity":null,"enabled":true,"value":[{"value":"\"google-site-verification=fJmN-QfYlsjUbH8uutPamnfirsijC7ynrt2MEGgu3Dc\"","enabled":true},{"value":"\"v=spf1 include:spf.mandrillapp.com ?all\"","enabled":true}],"lastValues":{"standard":[{"value":"\"google-site-verification=fJmN-QfYlsjUbH8uutPamnfirsijC7ynrt2MEGgu3Dc\"","enabled":true},{"value":"\"v=spf1 include:spf.mandrillapp.com ?all\"","enabled":true}]}}`
	var objJ ExpectedDNSRecord
	err := json.Unmarshal([]byte(data), &objJ)
	if err != nil {
		t.Error(err)
		return
	}
	var objY ExpectedDNSRecord
	err = yaml.Unmarshal([]byte(data), &objY)
	if err != nil {
		t.Error(err)
		return
	}

	expectedValue := []*DNSStandardItemValue{
		{
			Value:   "\"google-site-verification=fJmN-QfYlsjUbH8uutPamnfirsijC7ynrt2MEGgu3Dc\"",
			Enabled: true,
		},
		{
			Value:   "\"v=spf1 include:spf.mandrillapp.com ?all\"",
			Enabled: true,
		},
	}

	Compare := func(obj ExpectedDNSRecord, expectedValue []*DNSStandardItemValue, t *testing.T) {
		if objJ.Name != "abc" {
			t.Errorf("expected %q, got %q", "abc", objJ.Name)
			return
		}
		if objJ.Type != "TXT" {
			t.Errorf("expected %q, got %q", "TXT", objJ.Type)
			return
		}

		res, ok := objJ.Value.([]*DNSStandardItemValue)
		if !ok {
			t.Errorf("unexpected type %T", objJ.Value)
			return
		}
		if len(res) != 2 {
			t.Errorf("expected 2, got %d", len(res))
			return
		}
		if res[0].Value != expectedValue[0].Value {
			t.Errorf("expected %q, got %q", expectedValue[0].Value, res[0].Value)
			return
		}
		if res[0].Enabled != expectedValue[0].Enabled {
			t.Errorf("expected %t, got %t", expectedValue[0].Enabled, res[0].Enabled)
			return
		}
		if res[1].Value != expectedValue[1].Value {
			t.Errorf("expected %q, got %q", expectedValue[1].Value, res[1].Value)
			return
		}
		if res[1].Enabled != expectedValue[1].Enabled {
			t.Errorf("expected %t, got %t", expectedValue[1].Enabled, res[1].Enabled)
			return
		}
	}

	Compare(objJ, expectedValue, t)
	Compare(objY, expectedValue, t)
}

func TestHTTPRecord(t *testing.T) {
	dataJ := `{"id":643222,"name":"shawn","type":"HTTP","ttl":1800,"mode":"standard","region":"default","ipfilter":null,"ipfilterDrop":false,"geoFailover":false,"geoproximity":null,"enabled":true,"value":{"hard":false,"redirectType":"frame","title":null,"keywords":null,"description":null,"url":"https:\/\/www.showingweb.com"}}`
	dataY := `
id: 643222
name: shawn
type: HTTP
ttl: 1800
mode: standard
region: default
ipfilter: null
ipfilterDrop: false
geoFailover: false
geoproximity: null
enabled: true
value:
  hard: false
  redirectType: frame
  title: null
  keywords: null
  description: null
  url: https://www.showingweb.com
notes: ""
`
	var objJ ExpectedDNSRecord
	err := json.Unmarshal([]byte(dataJ), &objJ)
	if err != nil {
		t.Error(err)
		return
	}
	var objY ExpectedDNSRecord
	err = yaml.Unmarshal([]byte(dataY), &objY)
	if err != nil {
		t.Error(err)
		return
	}

	expectedValue := &DNSHTTPStandardItemValue{
		Hard:         false,
		RedirectType: "frame",
		Title:        "",
		Keywords:     "",
		Description:  "",
		URL:          "https://www.showingweb.com",
	}

	Compare := func(obj ExpectedDNSRecord, expectedValue *DNSHTTPStandardItemValue, t *testing.T) {
		if objJ.Name != "shawn" {
			t.Errorf("expected %q, got %q", "shawn", objJ.Name)
			return
		}
		if objJ.Type != "HTTP" {
			t.Errorf("expected %q, got %q", "HTTP", objJ.Type)
			return
		}

		_, ok := objJ.Value.(*DNSHTTPStandardItemValue)
		if !ok {
			t.Errorf("unexpected type %T", objJ.Value)
			return
		}
		if reflect.DeepEqual(obj.Value, expectedValue) == false {
			t.Errorf("expected %v, got %v", expectedValue, obj.Value)
			return
		}
	}

	Compare(objJ, expectedValue, t)
	Compare(objY, expectedValue, t)
}

func TestExpectedDNSRecord_CAA_Unmarshal(t *testing.T) {
	dataY := `
name: ""
type: CAA
ttl: 18
mode: standard
region: default
enabled: true
value:
  - tag: issue
    data: letsencrypt.org
    flags: 0
    enabled: true
  - tag: issue
    data: pki.goog; cansignhttpexchanges=yes
    flags: 0
    enabled: true
`
	dataJ := `{"id":58345378,"name":"","type":"CAA","ttl":18,"mode":"standard","region":"default","enabled":true,"value":[{"tag":"issue","data":"letsencrypt.org","flags":0,"enabled":true},{"tag":"issue","data":"pki.goog; cansignhttpexchanges=yes","flags":0,"enabled":true}]}`

	expectedValue := []*DNSCAAStandardItemValue{
		{Tag: "issue", Data: "letsencrypt.org", Flags: 0, Enabled: true},
		{Tag: "issue", Data: "pki.goog; cansignhttpexchanges=yes", Flags: 0, Enabled: true},
	}

	var objY ExpectedDNSRecord
	err := yaml.Unmarshal([]byte(dataY), &objY)
	if err != nil {
		t.Fatal(err)
	}

	var objJ ExpectedDNSRecord
	err = json.Unmarshal([]byte(dataJ), &objJ)
	if err != nil {
		t.Fatal(err)
	}

	if objJ.Type != "CAA" {
		t.Errorf("expected %q, got %q", "CAA", objJ.Type)
	}

	res, ok := objJ.Value.([]*DNSCAAStandardItemValue)
	if !ok {
		t.Fatalf("unexpected type %T", objJ.Value)
	}

	if len(res) != len(expectedValue) {
		t.Errorf("expected %d items, got %d", len(expectedValue), len(res))
	}

	for i, v := range res {
		if !reflect.DeepEqual(v, expectedValue[i]) {
			t.Errorf("item %d: expected %+v, got %+v", i, expectedValue[i], v)
		}
	}
}
