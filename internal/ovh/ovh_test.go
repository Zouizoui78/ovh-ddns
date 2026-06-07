package ovh

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"testing"

	"github.com/Zouizoui78/ovh-ddns/internal/model"
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

	f.responses[url] = reqBody.(model.Record)
	return nil
}

func TestGetDomainsIps(t *testing.T) {
	client := &fakeClient{
		responses: map[string]any{
			"/domain/zone/example.com/record?fieldType=A": []int{1},
			"/domain/zone/example.com/record/1": model.Record{
				Target:     net.ParseIP("1.2.3.4"),
				RecordType: model.RecordTypeA,
			},
			"/domain/zone/example.com/record?fieldType=AAAA": []int{2},
			"/domain/zone/example.com/record/2": model.Record{
				Target:     net.ParseIP("2001:db8::1"),
				RecordType: model.RecordTypeAAAA,
			},
			"/domain/zone/example.net/record?fieldType=A": []int{3},
			"/domain/zone/example.net/record/3": model.Record{
				Target:     net.ParseIP("5.6.7.8"),
				RecordType: model.RecordTypeA,
			},
			"/domain/zone/example.net/record?fieldType=AAAA": []int{4},
			"/domain/zone/example.net/record/4": model.Record{
				Target:     net.ParseIP("2001:db8::2"),
				RecordType: model.RecordTypeAAAA,
			},
		},
	}

	ovh := NewFromClient(client)
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

func TestGetDomainsIpsReturnsErrorWhenRecordMissing(t *testing.T) {
	client := &fakeClient{
		responses: map[string]any{
			"/domain/zone/example.com/record?fieldType=A":    []int{},
			"/domain/zone/example.com/record?fieldType=AAAA": []int{2},
			"/domain/zone/example.com/record/2": model.Record{
				Target:     net.ParseIP("2001:db8::1"),
				RecordType: model.RecordTypeAAAA,
			},
		},
	}

	ovh := NewFromClient(client)
	ips, err := ovh.GetDnsZones(context.Background(), []string{"example.com"})
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	v4 := ips["example.com"].A.Target
	if v4 != nil {
		t.Fatalf("expected no A record, got %s", v4)
	}
}
