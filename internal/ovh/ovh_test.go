package ovh

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
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

func TestGetDomainsIps(t *testing.T) {
	client := &fakeClient{
		responses: map[string]any{
			"/domain/zone/example.com/record?fieldType=A":    []int{1},
			"/domain/zone/example.com/record/1":              dto.Record{Target: "1.2.3.4"},
			"/domain/zone/example.com/record?fieldType=AAAA": []int{2},
			"/domain/zone/example.com/record/2":              dto.Record{Target: "2001:db8::1"},
			"/domain/zone/example.net/record?fieldType=A":    []int{3},
			"/domain/zone/example.net/record/3":              dto.Record{Target: "5.6.7.8"},
			"/domain/zone/example.net/record?fieldType=AAAA": []int{4},
			"/domain/zone/example.net/record/4":              dto.Record{Target: "2001:db8::2"},
		},
	}

	ovh := NewFromClient(client)
	got, err := ovh.GetDomainsIps(context.Background(), []string{"example.com", "example.net"})
	if err != nil {
		t.Fatalf("GetDomainsIps returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(got))
	}

	if !got["example.com"].V4.Equal(net.ParseIP("1.2.3.4")) {
		t.Fatalf("unexpected ipv4 for example.com: %v", got["example.com"].V4)
	}
	if !got["example.com"].V6.Equal(net.ParseIP("2001:db8::1")) {
		t.Fatalf("unexpected ipv6 for example.com: %v", got["example.com"].V6)
	}

	if !got["example.net"].V4.Equal(net.ParseIP("5.6.7.8")) {
		t.Fatalf("unexpected ipv4 for example.net: %v", got["example.net"].V4)
	}
	if !got["example.net"].V6.Equal(net.ParseIP("2001:db8::2")) {
		t.Fatalf("unexpected ipv6 for example.net: %v", got["example.net"].V6)
	}
}

func TestGetDomainsIpsReturnsErrorWhenRecordMissing(t *testing.T) {
	client := &fakeClient{
		responses: map[string]any{
			"/domain/zone/example.com/record?fieldType=A":    []int{},
			"/domain/zone/example.com/record?fieldType=AAAA": []int{2},
			"/domain/zone/example.com/record/2":              dto.Record{Target: "2001:db8::1"},
		},
	}

	ovh := NewFromClient(client)
	ips, err := ovh.GetDomainsIps(context.Background(), []string{"example.com"})
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	v4 := ips["example.com"].V4
	if v4 != nil {
		t.Fatalf("expected no A record, got %s", v4)
	}
}
