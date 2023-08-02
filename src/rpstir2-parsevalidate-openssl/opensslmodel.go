package openssl

//  1.2.840.113549.1.9.16.1.49
type AsProviderAttestation struct {
	Version      Version `json:"version" asn1:"class:2,tag:0"` //default 0
	CustomerAsn  int     `json:"customerAsn"`
	ProviderAsns []int   `json:"ProviderAsns"`
}
type Version struct {
	Version int `json:"version"`
}
