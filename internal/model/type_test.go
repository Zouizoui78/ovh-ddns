package model

import (
	"testing"

	"github.com/Zouizoui78/ovh-ddns/internal/utils"
)

func TestRecordA(t *testing.T) {
	r := RecordTypeFromString("A")
	if r != RecordTypeA {
		t.Errorf("expected RecordTypeA")
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name string
		in   RecordType
		out  string
	}{
		{
			name: "A record",
			in:   RecordTypeA,
			out:  "A",
		},
		{
			name: "AAAA record",
			in:   RecordTypeAAAA,
			out:  "AAAA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.in.String()
			if actual != tt.out {
				t.Errorf("expected %v, got %v", tt.out, actual)
			}
		})
	}
}

func TestFromString(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		out   RecordType
		panic bool
	}{
		{
			name: "A record",
			in:   "A",
			out:  RecordTypeA,
		},
		{
			name: "AAAA record",
			in:   "AAAA",
			out:  RecordTypeAAAA,
		},
		{
			name:  "invalid record",
			in:    "unknown record type",
			panic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panic {
				utils.AssertPanic(t, func() { RecordTypeFromString(tt.in) })
				return
			}

			actual := RecordTypeFromString(tt.in)
			if actual != tt.out {
				t.Errorf("expected %v, got %v", tt.out, actual)
			}
		})
	}
}
