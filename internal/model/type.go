package model

import "fmt"

type RecordType int

const (
	RecordTypeA RecordType = iota
	RecordTypeAAAA
)

func (r RecordType) String() string {
	switch r {
	case RecordTypeA:
		return "A"
	case RecordTypeAAAA:
		return "AAAA"
	default:
		panic("unhandled record type")
	}
}

func RecordTypeFromString(r string) RecordType {
	switch r {
	case "A":
		return RecordTypeA
	case "AAAA":
		return RecordTypeAAAA
	default:
		panic(fmt.Errorf("unhandled record type: %s", r))
	}
}
