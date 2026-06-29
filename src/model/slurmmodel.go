package model

import (
	"math/big"

	"github.com/guregu/null"
)

//////////////////////////////////////
//  SLURM
//////////////////

// filter
// FormatPrefix, MaxPrefixLength and PrefixLength are not in json
type PrefixFilters struct {
	Asn             null.Int `json:"asn"`
	Prefix          string   `json:"prefix"`
	FormatPrefix    string   `json:"-"`
	PrefixLength    uint64   `json:"-"`
	MaxPrefixLength uint64   `json:"maxPrefixLength"`
	Comment         string   `json:"comment"`
}

// set asn==-1 means asn is empty
type BgpsecFilters struct {
	Asn     null.Int `json:"asn"`
	SKI     string   `json:"SKI"`
	Comment string   `json:"comment"`
}

type AspaFilters struct {
	CustomerAsn  null.Int       `json:"customerAsid"`
	ProviderAsns []ProviderAsns `json:"providers"`
	Comment      string         `json:"comment"`
}

type ProviderAsns struct {
	ProviderAsn   null.Int `json:"providerAsid"`
	AddressFamily string   `json:"afiLimit"` //IPv4 IPv6
}

type HroaFilters struct {
	HroaAsn                null.Int `json:"asn"`
	SubtreeIdentifier      big.Int  `json:"subtree_identifier"`
	SubtreeIdentifierBytes [16]byte `json:"subtree_identifier_bytes"`
	EncodedSubtree         null.Int `json:"encoded_subtree"`
	AfiFlags               uint64   `json:"afiFlags"`
	Comment                string   `json:"comment"`
}

type AsraFilters struct {
	CustomerAsn       null.Int   `json:"customer_asid"`
	AddressFamily     string     `json:"afi"` //IPv4 IPv6
	ProviderAsns      []null.Int `json:"provider_set"`
	OtherNeighborAsns []null.Int `json:"other_neighbor_set"`
	CustomerAsns      []null.Int `json:"customer_set"`
	LateralPeerAsns   []null.Int `json:"lateral_peer_set"`
	Hybrids           []Hybrid   `json:"hybrid_set"`
	ValleyPathAsns    []null.Int `json:"valley_path_set"`
	Comment           string     `json:"comment"`
}

const (
	SLURM_PROVIDER_ASNS_ADDRESS_FAMILY_IPV4 = "IPv4"
	SLURM_PROVIDER_ASNS_ADDRESS_FAMILY_IPV6 = "IPv6"
)

type ValidationOutputFilters struct {
	PrefixFilters []PrefixFilters `json:"prefixFilters"`
	BgpsecFilters []BgpsecFilters `json:"bgpsecFilters"`
	AspaFilters   []AspaFilters   `json:"aspaFilters"`
	HroaFilters   []HroaFilters   `json:"hroaFilters"`
	AsraFilters   []AsraFilters   `json:"asraFilters"`
}

// assertion
// set !asn.Valid means asn is empty
type PrefixAssertions struct {
	Asn             null.Int `json:"asn"`
	Prefix          string   `json:"prefix"`
	MaxPrefixLength uint64   `json:"maxPrefixLength"`
	Comment         string   `json:"comment"`
}

// set asn==-1 means asn is empty
type BgpsecAssertions struct {
	Asn             null.Int `json:"asn"`
	Comment         string   `json:"comment"`
	SKI             string   `json:"SKI"`
	RouterPublicKey string   `json:"RouterPublicKey"`
}

type AspaAssertions struct {
	CustomerAsn  null.Int       `json:"customerAsid"`
	ProviderAsns []ProviderAsns `json:"providers"`
	Comment      string         `json:"comment"`
}
type HroaAssertions struct {
	HroaAsn                null.Int `json:"asn"`
	SubtreeIdentifier      big.Int  `json:"subtree_identifier"`
	SubtreeIdentifierBytes [16]byte `json:"subtree_identifier_bytes"`
	EncodedSubtree         null.Int `json:"encoded_subtree"`
	AfiFlags               uint64   `json:"afiFlags"`
	Comment                string   `json:"comment"`
}

type AsraAssertions struct {
	CustomerAsn       null.Int     `json:"customer_asid"`
	AddressFamily     string       `json:"afi"` //IPv4 IPv6
	ProviderAsns      []null.Int   `json:"provider_set"`
	OtherNeighborAsns []null.Int   `json:"other_neighbor_set"`
	CustomerAsns      []null.Int   `json:"customer_set"`
	LateralPeerAsns   []null.Int   `json:"lateral_peer_set"`
	Hybrids           []Hybrid     `json:"hybrid_set"`
	ValleyPathAsns    [][]null.Int `json:"valley_path_set"`
	Comment           string       `json:"comment"`
}

type Hybrid struct {
	NeighborAsn  null.Int   `json:"neighbor_asid"`
	ProviderAsns []null.Int `json:"provider"`
	CustomerAsns []null.Int `json:"customer"`
}

type LocallyAddedAssertions struct {
	PrefixAssertions []PrefixAssertions `json:"prefixAssertions"`
	BgpsecAssertions []BgpsecAssertions `json:"bgpsecAssertions"`
	AspaAssertions   []AspaAssertions   `json:"aspaAssertions"`
	HroaAssertions   []HroaAssertions   `json:"hroaAssertions"`
	AsraAssertions   []AsraAssertions   `json:"asraAssertions"`
}

type Slurm struct {
	SlurmVersion            int                     `json:"slurmVersion"`
	ValidationOutputFilters ValidationOutputFilters `json:"validationOutputFilters"`
	LocallyAddedAssertions  LocallyAddedAssertions  `json:"locallyAddedAssertions"`
}
