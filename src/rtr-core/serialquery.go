package rtrcore

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/guregu/null"
)

type RtrSerialQueryModel struct {
	ProtocolVersion uint8  `json:"protocolVersion"`
	PduType         uint8  `json:"pduType"`
	SessionId       uint16 `json:"sessionId"`
	Length          uint32 `json:"length"`
	SerialNumber    uint32 `json:"serialNumber"`
}

func NewRtrSerialQueryModel(protocolVersion uint8, sessionId uint16,
	serialNumber uint32) *RtrSerialQueryModel {
	return &RtrSerialQueryModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_SERIAL_QUERY,
		SessionId:       sessionId,
		Length:          12,
		SerialNumber:    serialNumber,
	}
}

func (p *RtrSerialQueryModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.SessionId)
	binary.Write(wr, binary.BigEndian, p.Length)
	binary.Write(wr, binary.BigEndian, p.SerialNumber)
	return wr.Bytes()
}
func (p *RtrSerialQueryModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrSerialQueryModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}
func (p *RtrSerialQueryModel) GetPduType() uint8 {
	return p.PduType
}
func ParseToSerialQuery(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	var sessionId uint16
	var serialNumber uint32
	var length uint32

	// get sessionId
	err = binary.Read(buf, binary.BigEndian, &sessionId)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get sessionId fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get SessionId")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 12 {
		belogs.Error("ParseToSerialQuery():PDU_TYPE_SERIAL_QUERY,  length must be 12, buf:", buf, "  length:", length)
		rtrError := NewRtrError(
			errors.New("pduType is SERIAL QUERY, length must be 12"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	// get serialNumber
	err = binary.Read(buf, binary.BigEndian, &serialNumber)
	if err != nil {
		belogs.Error("ParseToSerialQuery(): PDU_TYPE_SERIAL_QUERY get serialNumber fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get serialNumber")
		return rtrPduModel, rtrError
	}
	sq := NewRtrSerialQueryModel(protocolVersion, sessionId, serialNumber)
	belogs.Debug("ParseToSerialQuery():get PDU_TYPE_SERIAL_QUERY, buf:", buf, "  sq:", jsonutil.MarshalJson(sq))
	return sq, nil
}

// when len(rtrIncrementals)==0, just return endofdata, it is not an error
func AssembleSerialResponses(rtrIncrementals []model.LabRpkiRtrIncremental,
	rtrAsaIncrementals []model.LabRpkiRtrAsaIncremental,
	rtrHroaIncrementals []model.LabRpkiRtrHroaIncremental,
	rtrAsraIncrementals []model.LabRpkiRtrAsraIncremental,
	protocolVersion uint8, sessionId uint16, serialNumber uint32) (rtrPduModels []RtrPduModel, err error) {
	start := time.Now()
	belogs.Info("AssembleSerialResponses(): len(rtrIncrementals):", len(rtrIncrementals),
		"   len(rtrAsaIncrementals):", len(rtrAsaIncrementals),
		"   len(rtrHroaIncrementals):", len(rtrHroaIncrementals),
		"   len(rtrAsraIncrementals):", len(rtrAsraIncrementals),
		"   protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)
	if protocolVersion != PDU_PROTOCOL_VERSION_0 &&
		protocolVersion != PDU_PROTOCOL_VERSION_1 &&
		protocolVersion != PDU_PROTOCOL_VERSION_2 &&
		protocolVersion != PDU_PROTOCOL_VERSION_3 {
		belogs.Error("AssembleSerialResponses(): protocolVersion is error, protocolVersion:", protocolVersion)
		return nil, errors.New("protocolVersion is error")
	}

	rtrPduModels = make([]RtrPduModel, 0)
	dataAvailable := false
	supportHroa := conf.String("rtr::supportHroa") == "true"
	supportAsra := conf.String("rtr::supportAsra") == "true"

	// start response
	cacheResponseModel := NewRtrCacheResponseModel(protocolVersion, sessionId)
	rtrPduModels = append(rtrPduModels, cacheResponseModel)
	prefixAsaVersion := protocolVersion
	hroaAsraVersion := protocolVersion
	if (supportHroa || supportAsra) && protocolVersion >= PDU_PROTOCOL_VERSION_3 {
		prefixAsaVersion = PDU_PROTOCOL_VERSION_2
	}
	belogs.Debug("AssembleSerialResponses(): cacheResponseModel:", jsonutil.MarshalJson(cacheResponseModel),
		"   supportHroa:", supportHroa, "   supportAsra:", supportAsra,
		"   protocolVersion:", protocolVersion,
		"   prefixAsaVersion:", prefixAsaVersion,
		"   hroaAsraVersion:", hroaAsraVersion)

	//rtr incr from roa rtr
	if protocolVersion == PDU_PROTOCOL_VERSION_0 ||
		protocolVersion == PDU_PROTOCOL_VERSION_1 ||
		protocolVersion == PDU_PROTOCOL_VERSION_2 ||
		protocolVersion == PDU_PROTOCOL_VERSION_3 {
		if len(rtrIncrementals) > 0 {
			belogs.Debug("AssembleSerialResponses(): will get rtrIncrementals, len(rtrIncrementals):", len(rtrIncrementals),
				"   prefixAsaVersion:", prefixAsaVersion,
				"   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// rtr incr to response
			rtrIncrementalPduModels, err := convertRtrIncrementalsToRtrPduModels(rtrIncrementals, prefixAsaVersion)
			if err != nil {
				belogs.Error("AssembleSerialResponses(): convertRtrIncrementalsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrIncrementalPduModels...)
			dataAvailable = true
			belogs.Info("AssembleSerialResponses(): get rtrIncrementalPduModels, len(rtrIncrementalPduModels):", len(rtrIncrementalPduModels),
				"  time(s):", time.Since(start))
		}
	}

	if protocolVersion == PDU_PROTOCOL_VERSION_2 ||
		protocolVersion == PDU_PROTOCOL_VERSION_3 {
		//rtr incr from asa rtr
		if len(rtrAsaIncrementals) > 0 {
			belogs.Debug("AssembleSerialResponses():  will get rtrAsaIncrementals,  len(rtrAsaIncrementals):", len(rtrAsaIncrementals),
				"   prefixAsaVersion:", prefixAsaVersion,
				"   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// rtr asa incr to response
			rtrAsaIncrementalPduModels, err := convertRtrAsaIncrementalsToRtrPduModels(rtrAsaIncrementals, prefixAsaVersion)
			if err != nil {
				belogs.Error("AssembleSerialResponses(): convertRtrAsaIncrementalsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrAsaIncrementalPduModels...)
			dataAvailable = true
			belogs.Debug("AssembleSerialResponses():  get rtrAsaIncrementalPduModels, len(rtrAsaIncrementalPduModels):", len(rtrAsaIncrementalPduModels))
		}

	}
	if supportHroa && protocolVersion == PDU_PROTOCOL_VERSION_3 {
		//rtr incremental from hroa rtr
		if len(rtrHroaIncrementals) > 0 {
			belogs.Debug("AssembleSerialResponses(): will get rtrHroaIncrementals, len(rtrHroaIncrementals):", len(rtrHroaIncrementals),
				"   hroaAsraVersion:", hroaAsraVersion,
				"   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// rtr hroa incremental to response
			rtrHroaIncrementalPduModels, err := convertRtrHroaIncrementalsToRtrPduModels(rtrHroaIncrementals, hroaAsraVersion)
			if err != nil {
				belogs.Error("AssembleSerialResponses(): convertRtrHroaIncrementalsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrHroaIncrementalPduModels...)
			dataAvailable = true
			belogs.Info("AssembleSerialResponses(): get rtrHroaIncrementalPduModels, len(rtrHroaIncrementalPduModels):", len(rtrHroaIncrementalPduModels),
				"  time(s):", time.Since(start))
		}
	}

	if supportAsra && protocolVersion == PDU_PROTOCOL_VERSION_3 {
		//rtr incremental from asra rtr
		if len(rtrAsraIncrementals) > 0 {
			belogs.Debug("AssembleSerialResponses(): will get rtrAsraIncrementals, len(rtrAsraIncrementals):", len(rtrAsraIncrementals),
				"   hroaAsraVersion:", hroaAsraVersion,
				"   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// rtr asra incremental to response
			rtrAsraIncrementalPduModels, err := convertRtrAsraIncrementalsToRtrPduModels(rtrAsraIncrementals, hroaAsraVersion)
			if err != nil {
				belogs.Error("AssembleSerialResponses(): convertRtrAsraIncrementalsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrAsraIncrementalPduModels...)
			dataAvailable = true
			belogs.Info("AssembleSerialResponses(): get rtrAsraIncrementalPduModels, len(rtrAsraIncrementalPduModels):", len(rtrAsraIncrementalPduModels),
				"  time(s):", time.Since(start))
		}
	}
	if !dataAvailable {
		errorReportModel := NewRtrErrorReportModel(protocolVersion, PDU_TYPE_ERROR_CODE_NO_DATA_AVAILABLE, nil, nil)
		rtrPduModels = append(rtrPduModels, errorReportModel)
		belogs.Info("AssembleSerialResponses(): there is no rtr this time,  will send errorReport with not_data_available, ",
			"  receive protocolVersion:", protocolVersion, "   sessionId:", sessionId,
			"  serialNumber:", serialNumber, "  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels), "  time(s):", time.Since(start))
	} else {
		// end response
		endOfDataModel := AssembleEndOfDataResponse(protocolVersion, sessionId, serialNumber)
		rtrPduModels = append(rtrPduModels, endOfDataModel)
		belogs.Debug("AssembleSerialResponses(): endOfDataModel:", jsonutil.MarshalJson(endOfDataModel))

		belogs.Info("AssembleSerialResponses(): will send Cache Response of incrtmental rtr,",
			"   supportHroa:", supportHroa, "   supportAsra:", supportAsra,
			"   protocolVersion:", protocolVersion,
			"   prefixAsaVersion:", prefixAsaVersion,
			"   hroaAsraVersion:", hroaAsraVersion,
			"   sessionId:", sessionId, "  serialNumber:", serialNumber,
			"   len(rtrIncrementals):", len(rtrIncrementals),
			"   len(rtrAsaIncrementals):", len(rtrAsaIncrementals),
			"   len(rtrHroaIncrementals):", len(rtrHroaIncrementals),
			"   len(rtrAsraIncrementals):", len(rtrAsraIncrementals),
			"   len(rtrPduModels):", len(rtrPduModels), "  time(s):", time.Since(start))
	}
	belogs.Debug("AssembleSerialResponses(): ok, rtrPduModels:", jsonutil.MarshalJson(rtrPduModels), "  time(s):", time.Since(start))
	return rtrPduModels, nil

}

func convertRtrIncrementalsToRtrPduModels(rtrIncrementals []model.LabRpkiRtrIncremental,
	protocolVersion uint8) (rtrPduModels []RtrPduModel, err error) {
	rtrPduModels = make([]RtrPduModel, 0)
	for i, _ := range rtrIncrementals {
		rtrPduModel, err := convertRtrIncrementalToRtrPduModel(&rtrIncrementals[i], protocolVersion)
		if err != nil {
			belogs.Error("convertRtrIncrementalsToRtrPduModels(): convertRtrIncrementalToRtrPduModel fail: ", err)
			return nil, err
		}
		rtrPduModels = append(rtrPduModels, rtrPduModel)
	}
	belogs.Debug("convertRtrIncrementalsToRtrPduModels(): len(rtrIncrementals):", len(rtrIncrementals), " len(rtrPduModels):", len(rtrPduModels))
	return rtrPduModels, nil
}

func convertRtrIncrementalToRtrPduModel(rtrIncremental *model.LabRpkiRtrIncremental,
	protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {

	ipHex, ipType, err := iputil.AddressToRtrFormatByte(rtrIncremental.Address)
	if ipType == iputil.Ipv4Type {
		ipv4 := [4]byte{0x00}
		copy(ipv4[:], ipHex[:])
		rtrIpv4PrefixModel := NewRtrIpv4PrefixModel(protocolVersion, getModelFlagsFromStyle(rtrIncremental.Style),
			uint8(rtrIncremental.PrefixLength), uint8(rtrIncremental.MaxLength), ipv4, uint32(rtrIncremental.Asn))
		return rtrIpv4PrefixModel, nil
	} else if ipType == iputil.Ipv6Type {
		ipv6 := [16]byte{0x00}
		copy(ipv6[:], ipHex[:])
		rtrIpv6PrefixModel := NewRtrIpv6PrefixModel(protocolVersion, getModelFlagsFromStyle(rtrIncremental.Style),
			uint8(rtrIncremental.PrefixLength), uint8(rtrIncremental.MaxLength), ipv6, uint32(rtrIncremental.Asn))
		return rtrIpv6PrefixModel, nil
	}
	return rtrPduModel, errors.New("convert to rtr format, error ipType")
}

func convertRtrAsaIncrementalsToRtrPduModels(rtrAsaIncrementals []model.LabRpkiRtrAsaIncremental,
	protocolVersion uint8) (rtrAsaPduModels []RtrPduModel, err error) {
	belogs.Debug("convertRtrAsaIncrementalsToRtrPduModels(): len(rtrAsaIncrementals):", len(rtrAsaIncrementals), "  protocolVersion:", protocolVersion)

	start := time.Now()
	sameCustomerAsnAfi := make(map[string]*RtrAsaModel)
	rtrAsaPduModels = make([]RtrPduModel, 0)
	for i := range rtrAsaIncrementals {
		rtrPduModel := NewRtrAsaModelFromDb(protocolVersion, getModelFlagsFromStyle(rtrAsaIncrementals[i].Style),
			rtrAsaIncrementals[i].AddressFamily, uint32(rtrAsaIncrementals[i].CustomerAsn))
		key := rtrPduModel.GetKey()
		belogs.Debug("convertRtrAsaIncrementalsToRtrPduModels(): will add key:", key)
		if v, ok := sameCustomerAsnAfi[key]; ok {
			v.AddProviderAsn(uint32(rtrAsaIncrementals[i].ProviderAsn))
			sameCustomerAsnAfi[key] = v
		} else {
			rtrPduModel.AddProviderAsn(uint32(rtrAsaIncrementals[i].ProviderAsn))
			sameCustomerAsnAfi[key] = rtrPduModel
		}
	}
	for _, v := range sameCustomerAsnAfi {
		rtrAsaPduModels = append(rtrAsaPduModels, v)
		belogs.Debug("convertRtrAsaIncrementalsToRtrPduModels(): v: ", jsonutil.MarshalJson(v))
	}
	belogs.Info("convertRtrAsaIncrementalsToRtrPduModels(): len(rtrAsaIncrementals):", len(rtrAsaIncrementals),
		" len(rtrAsaPduModels):", len(rtrAsaPduModels), "  time(s):", time.Since(start))

	return rtrAsaPduModels, nil
}

func convertRtrHroaIncrementalToRtrPduModel(rtrHroaIncremental *model.LabRpkiRtrHroaIncremental, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	belogs.Debug("convertRtrHroaIncrementalToRtrPduModel(): rtrHroaIncremental:", jsonutil.MarshalJson(rtrHroaIncremental))
	afiFlags := rtrHroaIncremental.AfiFlags
	encodedSubTree := [4]byte{0x00}
	b, _ := convert.IntToBytes(rtrHroaIncremental.EncodedSubtree.ValueOrZero())
	copy(encodedSubTree[:], b[4:]) // only 4 bytes
	belogs.Debug("convertRtrHroaIncrementalToRtrPduModel(): b:", b, "   encodedSubTree:", encodedSubTree)
	if afiFlags == iputil.Ipv4Type {
		subtreeIdentifier := [4]byte{0x00}
		copy(subtreeIdentifier[:], rtrHroaIncremental.SubtreeIdentifierBytes[12:])
		rtrIpv4HroaModel := NewRtrIpv4HroaModel(protocolVersion,
			subtreeIdentifier, encodedSubTree,
			uint32(rtrHroaIncremental.HroaAsn.ValueOrZero()))
		belogs.Debug("convertRtrHroaIncrementalToRtrPduModel(): rtrIpv4HroaModel:", jsonutil.MarshalJson(rtrIpv4HroaModel))
		return rtrIpv4HroaModel, nil
	} else if afiFlags == iputil.Ipv6Type {
		rtrIpv6HroaModel := NewRtrIpv6HroaModel(protocolVersion,
			rtrHroaIncremental.SubtreeIdentifierBytes, encodedSubTree,
			uint32(rtrHroaIncremental.HroaAsn.ValueOrZero()))
		belogs.Debug("convertRtrHroaIncrementalToRtrPduModel(): rtrIpv6HroaModel:", jsonutil.MarshalJson(rtrIpv6HroaModel))
		return rtrIpv6HroaModel, nil
	}
	return rtrPduModel, errors.New("convert to rtr format, error afiFlags")
}

func convertRtrHroaIncrementalsToRtrPduModels(rtrHroaIncrementals []model.LabRpkiRtrHroaIncremental,
	protocolVersion uint8) (rtrPduModels []RtrPduModel, err error) {
	rtrPduModels = make([]RtrPduModel, 0)
	for i := range rtrHroaIncrementals {
		rtrPduModel, err := convertRtrHroaIncrementalToRtrPduModel(&rtrHroaIncrementals[i], protocolVersion)
		if err != nil {
			belogs.Error("convertRtrHroaIncrementalsToRtrPduModels(): convertRtrHroaIncrementalToRtrPduModel fail: ", err)
			return nil, err
		}
		rtrPduModels = append(rtrPduModels, rtrPduModel)
	}
	belogs.Debug("convertRtrHroaIncrementalsToRtrPduModels(): len(rtrHroaIncrementals):", len(rtrHroaIncrementals),
		" len(rtrPduModels):", len(rtrPduModels))
	return rtrPduModels, nil
}

func convertRtrAsraIncrementalToRtrPduModel(rtrAsraIncremental *model.LabRpkiRtrAsraIncremental, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	belogs.Debug("convertRtrAsraIncrementalToRtrPduModel(): rtrAsraIncremental:", jsonutil.MarshalJson(rtrAsraIncremental))

	customerAsnAsra := uint32(rtrAsraIncremental.CustomerAsnAsra.ValueOrZero())
	providerAsnAsras := make([]uint32, 0)
	err = jsonutil.UnmarshalJson(rtrAsraIncremental.ProviderAsnAsrasStr, &providerAsnAsras)
	if err != nil {
		belogs.Error("convertRtrAsraIncrementalToRtrPduModel(): get providerAsnAsras fail, rtrAsraIncremental.ProviderAsnAsrasStr:",
			rtrAsraIncremental.ProviderAsnAsrasStr, err)
		return rtrPduModel, err
	}
	rtrAsraModel := NewRtrAsraModel(protocolVersion, PDU_FLAG_ANNOUNCE, null.IntFrom(0),
		customerAsnAsra, providerAsnAsras)

	if len(rtrAsraIncremental.OtherNeighborAsnAsrasStr) > 0 {
		asNumber := make([]uint32, 0)
		err = jsonutil.UnmarshalJson(rtrAsraIncremental.OtherNeighborAsnAsrasStr, &asNumber)
		if err != nil {
			belogs.Error("convertRtrAsraIncrementalToRtrPduModel(): get OtherNeighborAsnAsrasStr fail, rtrAsraIncremental.OtherNeighborAsnAsrasStr:",
				rtrAsraIncremental.OtherNeighborAsnAsrasStr, err)
			return rtrPduModel, err
		}
		rtrAsraModel.AddAsNumbers(PDU_TYPE_ASRA_TYPE_ON, asNumber)
	}
	if len(rtrAsraIncremental.CustomerAsnAsrasStr) > 0 {
		asNumber := make([]uint32, 0)
		err = jsonutil.UnmarshalJson(rtrAsraIncremental.CustomerAsnAsrasStr, &asNumber)
		if err != nil {
			belogs.Error("convertRtrAsraIncrementalToRtrPduModel(): get CustomerAsnAsrasStr fail, rtrAsraIncremental.CustomerAsnAsrasStr:",
				rtrAsraIncremental.CustomerAsnAsrasStr, err)
			return rtrPduModel, err
		}
		rtrAsraModel.AddAsNumbers(PDU_TYPE_ASRA_TYPE_C, asNumber)
	}
	if len(rtrAsraIncremental.LateralPeerAsnAsrasStr) > 0 {
		asNumber := make([]uint32, 0)
		err = jsonutil.UnmarshalJson(rtrAsraIncremental.LateralPeerAsnAsrasStr, &asNumber)
		if err != nil {
			belogs.Error("convertRtrAsraIncrementalToRtrPduModel(): get LateralPeerAsnAsrasStr fail, rtrAsraIncremental.LateralPeerAsnAsrasStr:",
				rtrAsraIncremental.LateralPeerAsnAsrasStr, err)
			return rtrPduModel, err
		}
		rtrAsraModel.AddAsNumbers(PDU_TYPE_ASRA_TYPE_L, asNumber)
	}
	return rtrAsraModel, nil
}
func convertRtrAsraIncrementalsToRtrPduModels(rtrAsraIncrementals []model.LabRpkiRtrAsraIncremental,
	protocolVersion uint8) (rtrPduModels []RtrPduModel, err error) {
	rtrPduModels = make([]RtrPduModel, 0)
	for i := range rtrAsraIncrementals {
		rtrPduModel, err := convertRtrAsraIncrementalToRtrPduModel(&rtrAsraIncrementals[i], protocolVersion)
		if err != nil {
			belogs.Error("convertRtrAsraIncrementalsToRtrPduModels(): convertRtrAsraIncrementalToRtrPduModel fail: ", err)
			return nil, err
		}
		rtrPduModels = append(rtrPduModels, rtrPduModel)
	}
	belogs.Debug("convertRtrAsraIncrementalsToRtrPduModels(): len(rtrAsraIncrementals):", len(rtrAsraIncrementals),
		" len(rtrPduModels):", len(rtrPduModels))
	return rtrPduModels, nil
}
