package dto

import (
	"net"

	"github.com/Zouizoui78/ovh-ddns/internal/model"
)

type Record struct {
	Id        int    `json:"id"`
	FieldType string `json:"fieldType"`
	SubDomain string `json:"subDomain,omitempty"`
	Target    net.IP `json:"target"`
	Ttl       int    `json:"ttl,omitempty"`
	Zone      string `json:"zone"`
}

func (r Record) ToModel() model.Record {
	return model.Record{
		Id:         r.Id,
		RecordType: model.RecordTypeFromString(r.FieldType),
		SubDomain:  r.SubDomain,
		Target:     r.Target,
		Ttl:        r.Ttl,
		Zone:       r.Zone,
	}
}
