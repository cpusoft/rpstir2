package rtrcore

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
)

type RtrIpv6HroaModel struct {
	ProtocolVersion   uint8    `json:"protocolVersion"`
	PduType           uint8    `json:"pduType"`
	Zero              uint16   `json:"zero"`
	Length            uint32   `json:"length"` //32
	SubtreeIdentifier [16]byte `json:"subtreeIdentifier"`
	EncodedSubtree    [4]byte  `json:"encodedSubtree"`
	Asn               uint32   `json:"asn"`
}

func NewRtrIpv6HroaModel(protocolVersion uint8,
	subtreeIdentifier [16]byte, encodedSubtree [4]byte, asn uint32) *RtrIpv6HroaModel {
	return &RtrIpv6HroaModel{
		ProtocolVersion:   protocolVersion,
		PduType:           PDU_TYPE_IPV6_HROA,
		Zero:              0,
		Length:            32,
		SubtreeIdentifier: subtreeIdentifier,
		EncodedSubtree:    encodedSubtree,
		Asn:               asn,
	}
}

func (p *RtrIpv6HroaModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.Zero)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.SubtreeIdentifier)
	binary.Write(wr, binary.BigEndian, p.EncodedSubtree)
	binary.Write(wr, binary.BigEndian, p.Asn)
	return wr.Bytes()
}
func (p *RtrIpv6HroaModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrIpv6HroaModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}

func (p *RtrIpv6HroaModel) GetPduType() uint8 {
	return p.PduType
}

func ParseToIpv6Hroa(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	/*
		ProtocolVersion   uint8   `json:"protocolVersion"`
		PduType           uint8   `json:"pduType"`
		Zero              uint16  `json:"zero"`
		Length            uint32  `json:"length"` // 32
		SubtreeIdentifier [16]byte `json:"subtreeIdentifier"`
		EncodedSubtree    [4]byte `json:"encodedSubtree"`
		Asn               uint32  `json:"asn"`
	*/
	var zero uint16
	var length uint32
	var subtreeIdentifier [16]byte
	var encodedSubtree [4]byte
	var asn uint32

	// get zero0
	err = binary.Read(buf, binary.BigEndian, &zero)
	if err != nil {
		belogs.Error("ParseToIpv6Hroa(): PDU_TYPE_IPV6_HROA get zero fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToIpv6Hroa(): PDU_TYPE_IPV6_HROA get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 32 {
		belogs.Error("ParseToIpv6Hroa():PDU_TYPE_IPV6_HROA, length must be 20, buf:", buf, "   length:", length)
		rtrError := NewRtrError(
			errors.New("pduType is IPV4 HROA, length must be 32"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError

	}

	// get subtreeIdentifier
	err = binary.Read(buf, binary.BigEndian, &subtreeIdentifier)
	if err != nil {
		belogs.Error("ParseToIpv6Hroa(): PDU_TYPE_IPV6_HROA get subtreeIdentifier fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get subtreeIdentifier")
		return rtrPduModel, rtrError
	}

	// get encodedSubtree
	err = binary.Read(buf, binary.BigEndian, &encodedSubtree)
	if err != nil {
		belogs.Error("ParseToIpv6Hroa(): PDU_TYPE_IPV6_HROA get encodedSubtree fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get encodedSubtree")
		return rtrPduModel, rtrError
	}

	// get asn
	err = binary.Read(buf, binary.BigEndian, &asn)
	if err != nil {
		belogs.Error("ParseToIpv6Hroa(): PDU_TYPE_IPV6_HROA get asn fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get asn")
		return rtrPduModel, rtrError
	}

	sq := NewRtrIpv6HroaModel(protocolVersion, subtreeIdentifier, encodedSubtree, asn)

	belogs.Debug("ParseToIpv6Hroa():get PDU_TYPE_IPV6_HROA, buf:", buf, "  sq:", jsonutil.MarshalJson(sq))
	return sq, nil
}
