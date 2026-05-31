package dto

type Record struct {
	FieldType string `json:"fieldType"`
	Id        int    `json:"id"`
	SubDomain string `json:"subDomain"`
	Target    string `json:"target"`
	Ttl       int    `json:"ttl"`
	Zone      string `json:"zone"`
}
