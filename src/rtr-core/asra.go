package rtrcore

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/guregu/null"
)

type RtrAsraModel struct {
	ProtocolVersion         uint8              `json:"protocolVersion"`
	PduType                 uint8              `json:"pduType"`
	Zero                    uint16             `json:"zero"`
	Length                  uint32             `json:"length"`
	Flags                   uint8              `json:"flags"`
	AfiFlags                uint8              `json:"afiFlags"`
	ProviderAsCount         uint16             `json:"providerAsCount"`
	CustomerAsn             uint32             `json:"customerAsn"`
	ProviderAsns            []uint32           `json:"providerAsns"`
	AsraType                uint8              `json:"asraType"` // ON, C, L, P, V
	AsCountOrHybridTlvCount uint16             `json:"asCountOrHybridTlvCount"`
	RrtrAsraAsNumber        *RrtrAsraAsNumber  `json:"rtrAsraAsNumber"`
	RrtrAsraHybridTlv       *RrtrAsraHybridTlv `json:"rtrAsraHybridTlv"`
}
type RrtrAsraAsNumber struct {
	AsNumber []uint32 `json:"asNumber"`
}
type RrtrAsraHybridTlv struct {
	NeighborAsn               uint32                      `json:"neighborAsn"`
	RelationCount             uint16                      `json:"relationCount"`
	RrtrAsraHybridTagRelation []RrtrAsraHybridTagRelation `json:"rrtrAsraHybridTagRelation"`
}
type RrtrAsraHybridTagRelation struct {
	Tag           uint32 `json:"tag"`
	RelationCount uint8  `json:"relationCount"`
}

func NewRtrAsraModel(protocolVersion uint8,
	flags uint8, addressFamily null.Int, // afiFlags uint8,
	customerAsn uint32, providerAsns []uint32) *RtrAsraModel {
	var afiFlags uint8
	if addressFamily.Valid && addressFamily.ValueOrZero() > 0 {
		afiFlags = uint8(addressFamily.ValueOrZero()) - 1 // addressFamily:ipv4 is 1, ipv6 is 2; afiFlags: ipv4 is 0, ipv6 is 1
	} else {
		afiFlags = 0
	}
	length := 4 + 4 + 4 + 4 + 4*len(providerAsns)
	return &RtrAsraModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_ASRA,
		Zero:            0,
		Length:          uint32(length),
		Flags:           flags,
		AfiFlags:        afiFlags,
		ProviderAsCount: uint16(len(providerAsns)),
		CustomerAsn:     customerAsn,
		ProviderAsns:    providerAsns,
	}
}

func (p *RtrAsraModel) AddAsNumbers(asraType uint8, asNumber []uint32) {
	p.AsraType = asraType
	p.AsCountOrHybridTlvCount = uint16(len(asNumber))
	p.RrtrAsraAsNumber = &RrtrAsraAsNumber{
		AsNumber: asNumber,
	}
	p.Length = p.Length + 1 + 2 + 4*uint32(len(asNumber))
}

// TODO AddHybridTlvs
func (p *RtrAsraModel) AddHybridTlvs(asraType uint8, hybridTlvCount uint16) {
	p.AsraType = asraType
	p.AsCountOrHybridTlvCount = hybridTlvCount

}

func (p *RtrAsraModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.Zero)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.Flags)
	binary.Write(wr, binary.BigEndian, p.AfiFlags)
	binary.Write(wr, binary.BigEndian, p.ProviderAsCount)
	binary.Write(wr, binary.BigEndian, p.CustomerAsn)
	binary.Write(wr, binary.BigEndian, p.ProviderAsns)
	if p.AsraType <= 4 {
		binary.Write(wr, binary.BigEndian, p.AsraType)
		binary.Write(wr, binary.BigEndian, p.AsCountOrHybridTlvCount)
		binary.Write(wr, binary.BigEndian, p.RrtrAsraAsNumber.AsNumber)
	}
	return wr.Bytes()
}
func (p *RtrAsraModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrAsraModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrAsraModel) GetPduType() uint8 {
	return p.PduType
}

func ParseToAsra(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	/*
		ProtocolVersion         uint8              `json:"protocolVersion"`
			PduType                 uint8              `json:"pduType"`
			Zero                    uint16             `json:"zero"`
			Length                  uint32             `json:"length"`
			Flags                   uint8              `json:"flags"`
			AfiFlags                uint8              `json:"afiFlags"`
			ProviderAsCount         uint16             `json:"providerAsCount"`
			CustomerAsn             uint32             `json:"customerAsn"`
			ProviderAsns            []uint32           `json:"providerAsns"`
			AsraType                uint8              `json:"asraType"` // ON, C, L, P, V
			AsCountOrHybridTlvCount uint16             `json:"asCountOrHybridTlvCount"`
			RrtrAsraAsNumber        *RrtrAsraAsNumber  `json:"rtrAsraAsNumber"`
			RrtrAsraHybridTlv       *RrtrAsraHybridTlv `json:"rtrAsraHybridTlv"`
	*/

	var zero uint16
	var length uint32
	var flags uint8
	var afiFlags uint8
	var providerAsCount uint16
	var customerAsn uint32
	var providerAsns []uint32
	var asraType uint8
	var asCountOrHybridTlvCount uint16
	var asNumber []uint32

	// get zero
	err = binary.Read(buf, binary.BigEndian, &zero)
	if err != nil {
		belogs.Error("ParseToAsra(): PDU_TYPE_ASRA get zero fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToAsra(): PDU_TYPE_ASRA get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length < 16 {
		belogs.Error("ParseToAsra(): PDU_TYPE_ASRA, length must be more than 16, buf:", buf, length)
		rtrError := NewRtrError(
			errors.New("pduType is ASRA, length must be more than 16"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError

	}

	// get flags
	err = binary.Read(buf, binary.BigEndian, &flags)
	if err != nil {
		belogs.Error("ParseToAsra(): PDU_TYPE_ASRA get flags fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}
	/*
		Bit     Bit Name
		----    -------------------
		0      AFI (IPv4 == 0, IPv6 == 1)
		1      Announce == 1, Delete == 0
		2-7    Reserved, must be zero
	*/
	if flags != 0 && flags != 1 && flags != 2 && flags != 3 {
		belogs.Error("ParseToAsra(): PDU_TYPE_ASRA, flags is only use bits, buf:", buf, "  flags:", flags)
		rtrError := NewRtrError(
			errors.New("pduType is IPV4 PREFIX, flags is only use bits"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}

	// get afiFlags
	err = binary.Read(buf, binary.BigEndian, &afiFlags)
	if err != nil {
		belogs.Error("ParseToAsra(): PDU_TYPE_ASRA get afiFlags fail:  buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get afiFlags")
		return rtrPduModel, rtrError
	}

	// get providerAsCount
	err = binary.Read(buf, binary.BigEndian, &providerAsCount)
	if err != nil {
		belogs.Error("ParseToAsra(): PDU_TYPE_ASRA get providerAsCount fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get providerAsCount")
		return rtrPduModel, rtrError
	}

	// get customerAsn
	err = binary.Read(buf, binary.BigEndian, &customerAsn)
	if err != nil {
		belogs.Error("ParseToAsra(): PDU_TYPE_ASRA get customerAsn fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get customerAsn")
		return rtrPduModel, rtrError
	}
	providerAsns = make([]uint32, 0)
	for i := uint16(0); i < providerAsCount; i++ {
		var providerAsn uint32
		err = binary.Read(buf, binary.BigEndian, &providerAsn)
		if err != nil {
			belogs.Error("ParseToAsra(): PDU_TYPE_ASRA get providerAsn fail, buf:", buf, err)
			rtrError := NewRtrError(
				err,
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get providerAsn")
			return rtrPduModel, rtrError
		}
		providerAsns = append(providerAsns, providerAsn)
	}
	sq := NewRtrAsraModel(protocolVersion, flags, null.IntFrom(0),
		customerAsn, providerAsns)

	// get AsraType
	err = binary.Read(buf, binary.BigEndian, &asraType)
	if err != nil {
		belogs.Error("ParseToAsra(): PDU_TYPE_ASRA get asraType fail:  buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get asraType")
		return rtrPduModel, rtrError
	}
	// get asCountOrHybridTlvCount
	err = binary.Read(buf, binary.BigEndian, &asCountOrHybridTlvCount)
	if err != nil {
		belogs.Error("ParseToAsra(): PDU_TYPE_ASRA get asCountOrHybridTlvCount fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get asCountOrHybridTlvCount")
		return rtrPduModel, rtrError
	}
	asNumber = make([]uint32, 0)
	for i := uint16(0); i < asCountOrHybridTlvCount; i++ {
		var asn uint32
		err = binary.Read(buf, binary.BigEndian, &asn)
		if err != nil {
			belogs.Error("ParseToAsra(): PDU_TYPE_ASRA get asNumber fail, buf:", buf, err)
			rtrError := NewRtrError(
				err,
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get asNumber")
			return rtrPduModel, rtrError
		}
		asNumber = append(asNumber, asn)
	}
	sq.AddAsNumbers(asraType, asNumber)
	belogs.Debug("ParseToAsra(): PDU_TYPE_ASRA, buf:", buf, jsonutil.MarshalJson(sq))
	return sq, nil
}
