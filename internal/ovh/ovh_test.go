package ovh

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"path"
	"strconv"
	"testing"

	"github.com/Zouizoui78/ovh-ddns/internal/ovh/dto"
)

type fakeClient struct {
	responses map[string]any
}

func (f *fakeClient) GetWithContext(_ context.Context, url string, resType any) error {
	value, ok := f.responses[url]
	if !ok {
		return fmt.Errorf("unexpected url: %s", url)
	}

	if err, ok := value.(error); ok {
		return err
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return json.Unmarshal(payload, resType)
}

func (f *fakeClient) PostWithContext(ctx context.Context, url string, reqBody, resType any) error {
	if _, ok := f.responses[url]; ok {
		return fmt.Errorf("record already present")
	}

	d := reqBody.(dto.RecordPost)
	res, ok := resType.(*dto.Record)
	if !ok {
		panic("expected resType to be a *dto.Record")
	}

	*res = dto.Record{
		Id:        len(f.responses),
		FieldType: d.FieldType,
		SubDomain: d.SubDomain,
		Target:    d.Target,
		Ttl:       d.Ttl,
	}
	f.responses[path.Join(url, strconv.FormatInt(int64(res.Id), 10))] = res
	return nil
}

func (f *fakeClient) PutWithContext(ctx context.Context, url string, reqBody, resType any) error {
	value, ok := f.responses[url]
	if !ok {
		return fmt.Errorf("record not found")
	}

	oldValue, ok := value.(dto.Record)
	if !ok {
		return fmt.Errorf("expected Record dto")
	}

	newValue, ok := value.(dto.RecordPut)
	if !ok {
		return fmt.Errorf("expected RecordPut dto")
	}

	oldValue.SubDomain = newValue.SubDomain
	oldValue.Target = newValue.Target
	oldValue.Ttl = newValue.Ttl

	f.responses[url] = oldValue

	return nil
}

func TestGetDnsZones(t *testing.T) {
	client := &fakeClient{
		responses: map[string]any{
			"/domain/zone/example.com/record?fieldType=A":    []int{1},
			"/domain/zone/example.com/record/1":              newARecordDto("1.2.3.4"),
			"/domain/zone/example.com/record?fieldType=AAAA": []int{2},
			"/domain/zone/example.com/record/2":              newAAAARecordDto("2001:db8::1"),
			"/domain/zone/example.net/record?fieldType=A":    []int{3},
			"/domain/zone/example.net/record/3":              newARecordDto("5.6.7.8"),
			"/domain/zone/example.net/record?fieldType=AAAA": []int{4},
			"/domain/zone/example.net/record/4":              newAAAARecordDto("2001:db8::2"),
		},
	}

	ovh := NewFromClient(client, false)
	got, err := ovh.GetDnsZones(context.Background(), []string{"example.com", "example.net"})
	if err != nil {
		t.Fatalf("GetDomainsIps returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(got))
	}

	v4 := got["example.com"].A.Target
	v6 := got["example.com"].AAAA.Target
	if !v4.Equal(net.ParseIP("1.2.3.4")) {
		t.Fatalf("unexpected ipv4 for example.com: %v", v4)
	}
	if !v6.Equal(net.ParseIP("2001:db8::1")) {
		t.Fatalf("unexpected ipv6 for example.com: %v", v6)
	}

	v4 = got["example.net"].A.Target
	v6 = got["example.net"].AAAA.Target
	if !v4.Equal(net.ParseIP("5.6.7.8")) {
		t.Fatalf("unexpected ipv4 for example.net: %v", v4)
	}
	if !v6.Equal(net.ParseIP("2001:db8::2")) {
		t.Fatalf("unexpected ipv6 for example.net: %v", v6)
	}
}

func TestGetDnsZonesReturnNilAddrWhenNoRecord(t *testing.T) {
	client := &fakeClient{
		responses: map[string]any{
			"/domain/zone/example.com/record?fieldType=A":    []int{},
			"/domain/zone/example.com/record?fieldType=AAAA": []int{2},
			"/domain/zone/example.com/record/2":              newAAAARecordDto("2001:db8::1"),
		},
	}

	ovh := NewFromClient(client, false)
	zones, err := ovh.GetDnsZones(context.Background(), []string{"example.com"})
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	if zones["example.com"].A != nil {
		t.Fatalf("expected no A record")
	}
}

func TestPostRecord(t *testing.T) {
	ovh := NewFromClient(&fakeClient{
		responses: map[string]any{},
	}, false)

	r, err := ovh.PostARecord(context.Background(), "example.com", net.ParseIP("1.2.3.4"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if r.Id != 0 {
		t.Errorf("expected id 0, got %v", r.Id)
	}

	r, err = ovh.PostAAAARecord(context.Background(), "example.com", net.ParseIP("2001:db8::1"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if r.Id != 1 {
		t.Errorf("expected id 1, got %v", r.Id)
	}
}

// Helpers

func newARecordDto(target string) dto.Record {
	return dto.Record{
		Target:    net.ParseIP(target),
		FieldType: "A",
	}
}

func newAAAARecordDto(target string) dto.Record {
	return dto.Record{
		Target:    net.ParseIP(target),
		FieldType: "AAAA",
	}
}
