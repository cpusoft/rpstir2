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

type RtrAsaModel struct {
	ProtocolVersion uint8    `json:"protocolVersion"`
	PduType         uint8    `json:"pduType"`
	Zero0           uint16   `json:"zero0"`
	Length          uint32   `json:"length"`
	Flags           uint8    `json:"flags"`
	AfiFlags        uint8    `json:"afiFlags"`
	ProviderAsCount uint16   `json:"providerAsCount"`
	CustomerAsn     uint32   `json:"customerAsn"`
	ProviderAsns    []uint32 `json:"providerAsns"`
}

func NewRtrAsaModelFromDb(protocolVersion uint8, flags uint8, addressFamily null.Int, // afiFlags uint8,
	customerAsn uint32) *RtrAsaModel {
	length := 16 // header+flags+afi+providerAsCount+CustomerAsn, will increase when providerAsn is added
	var afiFlags uint8
	if addressFamily.Valid && addressFamily.ValueOrZero() > 0 {
		afiFlags = uint8(addressFamily.ValueOrZero()) - 1 // addressFamily:ipv4 is 1, ipv6 is 2; afiFlags: ipv4 is 0, ipv6 is 1
	} else {
		afiFlags = 0
	}
	return &RtrAsaModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_ASA,
		Zero0:           0,
		Length:          uint32(length),
		Flags:           flags,
		AfiFlags:        afiFlags,
		ProviderAsCount: 0,
		CustomerAsn:     customerAsn,
		ProviderAsns:    make([]uint32, 0),
	}
}
func NewRtrAsaModelFromParse(protocolVersion uint8, flags uint8, afiFlags uint8,
	customerAsn uint32, providerAsns []uint32) *RtrAsaModel {
	length := 16 + len(providerAsns)*4

	return &RtrAsaModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_ASA,
		Zero0:           0,
		Length:          uint32(length),
		Flags:           flags,
		AfiFlags:        afiFlags,
		ProviderAsCount: uint16(len(providerAsns)),
		CustomerAsn:     customerAsn,
		ProviderAsns:    providerAsns,
	}
}

func (p *RtrAsaModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.Zero0)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.Flags)
	binary.Write(wr, binary.BigEndian, p.AfiFlags)
	binary.Write(wr, binary.BigEndian, p.ProviderAsCount)
	binary.Write(wr, binary.BigEndian, p.CustomerAsn)
	if len(p.ProviderAsns) > 0 {
		for i := range p.ProviderAsns {
			binary.Write(wr, binary.BigEndian, p.ProviderAsns[i])
		}
	}
	return wr.Bytes()
}
func (p *RtrAsaModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrAsaModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrAsaModel) GetPduType() uint8 {
	return p.PduType
}
func (p *RtrAsaModel) AddProviderAsn(providerAsn uint32) {
	if p.ProviderAsns == nil {
		p.ProviderAsns = make([]uint32, 0)
	}
	p.ProviderAsns = append(p.ProviderAsns, providerAsn)
	p.Length += 4
	p.ProviderAsCount++
	belogs.Debug("AddProviderAsn(): providerAsn:", providerAsn, "  length:", p.Length,
		"  len(providerAsns):", len(p.ProviderAsns), "   ProviderAsCount:", p.ProviderAsCount)
}

func (p *RtrAsaModel) GetKey() string {
	return convert.ToString(p.Flags) + "_" + convert.ToString(p.AfiFlags) +
		"_" + convert.ToString(p.CustomerAsn)
}

func ParseToAsa(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	/*
		ProtocolVersion uint8    `json:"protocolVersion"`
		PduType         uint8    `json:"pduType"`
		Zero0           uint16   `json:"zero0"`
		Length          uint32   `json:"length"`
		Flags           uint8    `json:"flags"`
		Zero1           uint8    `json:"zero1"`
		ProviderAsCount uint16   `json:"providerAsCount"`
		CustomerAsn     uint32   `json:"customerAsn"`
		ProviderAsns    []uint32 `json:"providerAsns"`
	*/

	var zero0 uint16
	var length uint32
	var flags uint8
	var afiFlags uint8
	var providerAsCount uint16
	var customerAsn uint32
	var providerAsns []uint32

	// get zero0
	err = binary.Read(buf, binary.BigEndian, &zero0)
	if err != nil {
		belogs.Error("ParseToAsa(): PDU_TYPE_ASA get zero0 fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero0")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToAsa(): PDU_TYPE_ASA get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length < 16 {
		belogs.Error("ParseToAsa():PDU_TYPE_ASA, length must be more than 16, buf:", buf, length)
		rtrError := NewRtrError(
			errors.New("pduType is ASA, length must be more than 16"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError

	}

	// get flags
	err = binary.Read(buf, binary.BigEndian, &flags)
	if err != nil {
		belogs.Error("ParseToAsa(): PDU_TYPE_ASA get flags fail, buf:", buf, err)
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
		belogs.Error("ParseToAsa():PDU_TYPE_ASA, flags is only use bits, buf:", buf, "  flags:", flags)
		rtrError := NewRtrError(
			errors.New("pduType is IPV4 PREFIX, flags is only use bits"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get flags")
		return rtrPduModel, rtrError
	}

	// get afiFlags
	err = binary.Read(buf, binary.BigEndian, &afiFlags)
	if err != nil {
		belogs.Error("ParseToAsa(): PDU_TYPE_ASA get afiFlags fail:  buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get afiFlags")
		return rtrPduModel, rtrError
	}

	// get providerAsCount
	err = binary.Read(buf, binary.BigEndian, &providerAsCount)
	if err != nil {
		belogs.Error("ParseToAsa(): PDU_TYPE_ASA get providerAsCount fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get providerAsCount")
		return rtrPduModel, rtrError
	}

	// get customerAsn
	err = binary.Read(buf, binary.BigEndian, &customerAsn)
	if err != nil {
		belogs.Error("ParseToAsa(): PDU_TYPE_ASA get customerAsn fail, buf:", buf, err)
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
			belogs.Error("ParseToAsa(): PDU_TYPE_ASA get providerAsn fail, buf:", buf, err)
			rtrError := NewRtrError(
				err,
				true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
				buf, "Fail to get providerAsn")
			return rtrPduModel, rtrError
		}
		providerAsns = append(providerAsns, providerAsn)
	}
	sq := NewRtrAsaModelFromParse(protocolVersion, flags, afiFlags,
		customerAsn, providerAsns)

	belogs.Debug("ParseToAsa():get PDU_TYPE_ASA, buf:", buf, jsonutil.MarshalJson(sq))
	return sq, nil
}
