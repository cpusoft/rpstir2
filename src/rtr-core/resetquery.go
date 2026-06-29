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

type RtrResetQueryModel struct {
	ProtocolVersion uint8  `json:"protocolVersion"`
	PduType         uint8  `json:"pduType"`
	Zero            uint16 `json:"zero"`
	Length          uint32 `json:"length"`
}

func NewRtrResetQueryModel(protocolVersion uint8) *RtrResetQueryModel {
	return &RtrResetQueryModel{
		ProtocolVersion: protocolVersion,
		PduType:         PDU_TYPE_RESET_QUERY,
		Zero:            0,
		Length:          8,
	}
}

func (p *RtrResetQueryModel) Bytes() []byte {
	wr := bytes.NewBuffer([]byte{})
	binary.Write(wr, binary.BigEndian, p.ProtocolVersion)
	binary.Write(wr, binary.BigEndian, p.PduType)
	binary.Write(wr, binary.BigEndian, p.Zero)
	binary.Write(wr, binary.BigEndian, p.Length)
	return wr.Bytes()
}
func (p *RtrResetQueryModel) PrintBytes() string {
	return convert.PrintBytes(p.Bytes(), 8)
}
func (p *RtrResetQueryModel) GetProtocolVersion() uint8 {
	return p.ProtocolVersion
}
func (p *RtrResetQueryModel) GetPduType() uint8 {
	return p.PduType
}

func ParseToResetQuery(buf *bytes.Reader, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	var zero16 uint16
	var length uint32

	// get zero16
	err = binary.Read(buf, binary.BigEndian, &zero16)
	if err != nil {
		belogs.Error("ParseToResetQuery(): PDU_TYPE_RESET_QUERY get zero fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get zero")
		return rtrPduModel, rtrError
	}

	// get length
	err = binary.Read(buf, binary.BigEndian, &length)
	if err != nil {
		belogs.Error("ParseToResetQuery(): PDU_TYPE_RESET_QUERY get length fail, buf:", buf, err)
		rtrError := NewRtrError(
			err,
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}
	if length != 8 {
		belogs.Error("ParseToResetQuery():PDU_TYPE_RESET_QUERY, length must be 8, buf:", buf, "  length:", length)
		rtrError := NewRtrError(
			errors.New("pduType is RESET QUERY, length must be 8"),
			true, protocolVersion, PDU_TYPE_ERROR_CODE_CORRUPT_DATA,
			buf, "Fail to get length")
		return rtrPduModel, rtrError
	}

	rq := NewRtrResetQueryModel(protocolVersion)
	belogs.Debug("ParseToResetQuery():get PDU_TYPE_RESET_QUERY, buf:", buf, "   rq:", jsonutil.MarshalJson(rq))
	return rq, nil
}

// when len(rtrFull)==0, it is an error with no_data_available
func AssembleResetResponses(rtrFulls []model.LabRpkiRtrFull,
	rtrAsaFulls []model.LabRpkiRtrAsaFull,
	rtrHroaFulls []model.LabRpkiRtrHroaFull,
	rtrAsraFulls []model.LabRpkiRtrAsraFull,
	protocolVersion uint8, sessionId uint16, serialNumber uint32) (rtrPduModels []RtrPduModel, err error) {
	start := time.Now()
	belogs.Info("AssembleResetResponses(): len(rtrFulls):", len(rtrFulls),
		"   len(rtrAsaFulls):", len(rtrAsaFulls),
		"   len(rtrHroaFulls):", len(rtrHroaFulls),
		"   len(rtrAsraFulls):", len(rtrAsraFulls),
		"   protocolVersion:", protocolVersion, "   sessionId:", sessionId, "   serialNumber:", serialNumber)
	if protocolVersion != PDU_PROTOCOL_VERSION_0 &&
		protocolVersion != PDU_PROTOCOL_VERSION_1 &&
		protocolVersion != PDU_PROTOCOL_VERSION_2 &&
		protocolVersion != PDU_PROTOCOL_VERSION_3 {
		belogs.Error("AssembleResetResponses(): protocolVersion is error, protocolVersion:", protocolVersion)
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
	belogs.Debug("AssembleResetResponses(): cacheResponseModel:", jsonutil.MarshalJson(cacheResponseModel),
		"   supportHroa:", supportHroa, "   supportAsra:", supportAsra,
		"   protocolVersion:", protocolVersion,
		"   prefixAsaVersion:", prefixAsaVersion,
		"   hroaAsraVersion:", hroaAsraVersion)
	// rtr full from roa rtr
	if protocolVersion == PDU_PROTOCOL_VERSION_0 ||
		protocolVersion == PDU_PROTOCOL_VERSION_1 ||
		protocolVersion == PDU_PROTOCOL_VERSION_2 ||
		protocolVersion == PDU_PROTOCOL_VERSION_3 {
		if len(rtrFulls) > 0 {

			belogs.Debug("AssembleResetResponses(): will get rtrFullPduModels, len(rtrFulls)>0, len(rtrFulls):", len(rtrFulls),
				"   prefixAsaVersion:", prefixAsaVersion,
				"   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// rtr full to response
			rtrFullPduModels, err := convertRtrFullsToRtrPduModels(rtrFulls, prefixAsaVersion)
			if err != nil {
				belogs.Error("AssembleResetResponses(): convertRtrIncrementalsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrFullPduModels...)
			dataAvailable = true
			belogs.Info("AssembleResetResponses(): get rtrFullPduModels, len(rtrFullPduModels):", len(rtrFullPduModels),
				"  time(s):", time.Since(start))
		}
	}
	if protocolVersion == PDU_PROTOCOL_VERSION_2 ||
		protocolVersion == PDU_PROTOCOL_VERSION_3 {
		//rtr full from asa rtr
		if len(rtrAsaFulls) > 0 {
			belogs.Debug("AssembleResetResponses(): will get rtrAsaFullPduModels, len(rtrAsaFulls):", len(rtrAsaFulls),
				"   prefixAsaVersion:", prefixAsaVersion,
				"   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// rtr asa full to response
			rtrAsaFullPduModels, err := convertRtrAsaFullsToRtrPduModels(rtrAsaFulls, prefixAsaVersion)
			if err != nil {
				belogs.Error("AssembleResetResponses(): convertRtrAsaFullsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrAsaFullPduModels...)
			dataAvailable = true
			belogs.Info("AssembleResetResponses(): get rtrAsaFullPduModels, len(rtrAsaFullPduModels):", len(rtrAsaFullPduModels),
				"  time(s):", time.Since(start))
		}
	}

	if supportHroa && protocolVersion == PDU_PROTOCOL_VERSION_3 {
		//rtr full from hroa rtr
		if len(rtrHroaFulls) > 0 {
			belogs.Debug("AssembleResetResponses(): will get rtrHroaFullPduModels, len(rtrHroaFulls):", len(rtrHroaFulls),
				"   hroaAsraVersion:", hroaAsraVersion,
				"   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// rtr hroa full to response
			rtrHroaFullPduModels, err := convertRtrHroaFullsToRtrPduModels(rtrHroaFulls, hroaAsraVersion)
			if err != nil {
				belogs.Error("AssembleResetResponses(): convertRtrHroaFullsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrHroaFullPduModels...)
			dataAvailable = true
			belogs.Info("AssembleResetResponses(): get rtrHroaFullPduModels, len(rtrHroaFullPduModels):", len(rtrHroaFullPduModels),
				"  time(s):", time.Since(start))
		}
	}
	if supportAsra && protocolVersion == PDU_PROTOCOL_VERSION_3 {
		//rtr full from asra rtr
		if len(rtrAsraFulls) > 0 {
			belogs.Debug("AssembleResetResponses(): will get rtrHroaFullPduModels, len(rtrAsraFulls):", len(rtrAsraFulls),
				"   hroaAsraVersion:", hroaAsraVersion,
				"   sessionId:", sessionId, "   serialNumber:", serialNumber)

			// rtr hroa full to response
			rtrAsraFullPduModels, err := convertRtrAsraFullsToRtrPduModels(rtrAsraFulls, hroaAsraVersion)
			if err != nil {
				belogs.Error("AssembleResetResponses(): convertRtrAsraFullsToRtrPduModels fail: ", err)
				return nil, err
			}
			rtrPduModels = append(rtrPduModels, rtrAsraFullPduModels...)
			dataAvailable = true
			belogs.Info("AssembleResetResponses(): get rtrAsraFullPduModels, len(rtrAsraFullPduModels):", len(rtrAsraFullPduModels),
				"  time(s):", time.Since(start))
		}
	}

	if !dataAvailable {
		errorReportModel := NewRtrErrorReportModel(protocolVersion, PDU_TYPE_ERROR_CODE_NO_DATA_AVAILABLE, nil, nil)
		rtrPduModels = append(rtrPduModels, errorReportModel)
		belogs.Info("AssembleResetResponses(): there is no rtr this time,  will send errorReport with not_data_available, ",
			"  receive protocolVersion:", protocolVersion, "   sessionId:", sessionId,
			"  serialNumber:", serialNumber, "  rtrPduModels:", jsonutil.MarshalJson(rtrPduModels), "  time(s):", time.Since(start))
	} else {
		// end response
		endOfDataModel := AssembleEndOfDataResponse(protocolVersion, sessionId, serialNumber)
		rtrPduModels = append(rtrPduModels, endOfDataModel)
		belogs.Debug("AssembleResetResponses(): endOfDataModel:", jsonutil.MarshalJson(endOfDataModel))

		belogs.Info("AssembleResetResponses(): will send Cache Response of all rtr,",
			"   supportHroa:", supportHroa, "   supportAsra:", supportAsra,
			"   protocolVersion:", protocolVersion,
			"   prefixAsaVersion:", prefixAsaVersion,
			"   hroaAsraVersion:", hroaAsraVersion,
			"   sessionId:", sessionId, "  serialNumber:", serialNumber,
			"   len(rtrFulls):", len(rtrFulls),
			"   len(rtrAsaFulls):", len(rtrAsaFulls),
			"   len(rtrHroaFulls):", len(rtrHroaFulls),
			"   len(rtrAsraFulls):", len(rtrAsraFulls),
			"   len(rtrPduModels):", len(rtrPduModels), "  time(s):", time.Since(start))
	}
	belogs.Debug("AssembleResetResponses(): ok, rtrPduModels:", jsonutil.MarshalJson(rtrPduModels), "  time(s):", time.Since(start))
	return rtrPduModels, nil

}

func convertRtrFullToRtrPduModel(rtrFull *model.LabRpkiRtrFull, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {

	ipHex, ipType, err := iputil.AddressToRtrFormatByte(rtrFull.Address)
	if ipType == iputil.Ipv4Type {
		ipv4 := [4]byte{0x00}
		copy(ipv4[:], ipHex[:])
		rtrIpv4PrefixModel := NewRtrIpv4PrefixModel(protocolVersion, PDU_FLAG_ANNOUNCE, uint8(rtrFull.PrefixLength),
			uint8(rtrFull.MaxLength), ipv4, uint32(rtrFull.Asn))
		return rtrIpv4PrefixModel, nil
	} else if ipType == iputil.Ipv6Type {
		ipv6 := [16]byte{0x00}
		copy(ipv6[:], ipHex[:])
		rtrIpv6PrefixModel := NewRtrIpv6PrefixModel(protocolVersion, PDU_FLAG_ANNOUNCE, uint8(rtrFull.PrefixLength),
			uint8(rtrFull.MaxLength), ipv6, uint32(rtrFull.Asn))
		return rtrIpv6PrefixModel, nil
	}
	return rtrPduModel, errors.New("convert to rtr format, error ipType")
}

func convertRtrFullsToRtrPduModels(rtrFulls []model.LabRpkiRtrFull,
	protocolVersion uint8) (rtrPduModels []RtrPduModel, err error) {
	rtrPduModels = make([]RtrPduModel, 0)
	for i := range rtrFulls {
		rtrPduModel, err := convertRtrFullToRtrPduModel(&rtrFulls[i], protocolVersion)
		if err != nil {
			belogs.Error("convertRtrFullsToRtrPduModels(): convertRtrFullToRtrPduModel fail: ", err)
			return nil, err
		}
		rtrPduModels = append(rtrPduModels, rtrPduModel)
	}
	belogs.Debug("convertRtrFullsToRtrPduModels(): len(rtrFulls):", len(rtrFulls), " len(rtrPduModels):", len(rtrPduModels))
	return rtrPduModels, nil
}

func convertRtrAsaFullsToRtrPduModels(rtrAsaFulls []model.LabRpkiRtrAsaFull,
	protocolVersion uint8) (rtrAsaPduModels []RtrPduModel, err error) {
	belogs.Debug("convertRtrAsaFullsToRtrPduModels(): len(rtrAsaFulls):", len(rtrAsaFulls), "  protocolVersion:", protocolVersion)

	start := time.Now()
	sameCustomerAsnAfi := make(map[string]*RtrAsaModel)
	rtrAsaPduModels = make([]RtrPduModel, 0)
	for i := range rtrAsaFulls {
		rtrPduModel := NewRtrAsaModelFromDb(protocolVersion, PDU_FLAG_ANNOUNCE,
			rtrAsaFulls[i].AddressFamily, uint32(rtrAsaFulls[i].CustomerAsn))
		key := rtrPduModel.GetKey()
		belogs.Debug("convertRtrAsaFullsToRtrPduModels(): will add key:", key)
		if v, ok := sameCustomerAsnAfi[key]; ok {
			v.AddProviderAsn(uint32(rtrAsaFulls[i].ProviderAsn))
			sameCustomerAsnAfi[key] = v
		} else {
			rtrPduModel.AddProviderAsn(uint32(rtrAsaFulls[i].ProviderAsn))
			sameCustomerAsnAfi[key] = rtrPduModel
		}
	}
	for _, v := range sameCustomerAsnAfi {
		rtrAsaPduModels = append(rtrAsaPduModels, v)
		belogs.Debug("convertRtrAsaFullsToRtrPduModels(): v: ", jsonutil.MarshalJson(v))
	}
	belogs.Info("convertRtrAsaFullsToRtrPduModels(): len(rtrAsaFulls):", len(rtrAsaFulls),
		" len(rtrAsaPduModels):", len(rtrAsaPduModels), "  time(s):", time.Since(start))
	return rtrAsaPduModels, nil
}

func convertRtrHroaFullToRtrPduModel(rtrHroaFull *model.LabRpkiRtrHroaFull, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	belogs.Debug("convertRtrHroaFullToRtrPduModel(): rtrHroaFull:", jsonutil.MarshalJson(rtrHroaFull))
	afiFlags := rtrHroaFull.AfiFlags
	encodedSubTree := [4]byte{0x00}
	b, _ := convert.IntToBytes(rtrHroaFull.EncodedSubtree.ValueOrZero())
	copy(encodedSubTree[:], b[4:]) // only 4 bytes
	belogs.Debug("convertRtrHroaFullToRtrPduModel(): b:", b, "   encodedSubTree:", encodedSubTree)
	if afiFlags == iputil.Ipv4Type {
		subtreeIdentifier := [4]byte{0x00}
		copy(subtreeIdentifier[:], rtrHroaFull.SubtreeIdentifierBytes[12:])
		rtrIpv4HroaModel := NewRtrIpv4HroaModel(protocolVersion,
			subtreeIdentifier, encodedSubTree,
			uint32(rtrHroaFull.HroaAsn.ValueOrZero()))
		belogs.Debug("convertRtrHroaFullToRtrPduModel(): rtrIpv4HroaModel:", jsonutil.MarshalJson(rtrIpv4HroaModel))
		return rtrIpv4HroaModel, nil
	} else if afiFlags == iputil.Ipv6Type {
		rtrIpv6HroaModel := NewRtrIpv6HroaModel(protocolVersion,
			rtrHroaFull.SubtreeIdentifierBytes, encodedSubTree,
			uint32(rtrHroaFull.HroaAsn.ValueOrZero()))
		belogs.Debug("convertRtrHroaFullToRtrPduModel(): rtrIpv6HroaModel:", jsonutil.MarshalJson(rtrIpv6HroaModel))
		return rtrIpv6HroaModel, nil
	}
	return rtrPduModel, errors.New("convert to rtr format, error afiFlags")
}

func convertRtrHroaFullsToRtrPduModels(rtrHroaFulls []model.LabRpkiRtrHroaFull,
	protocolVersion uint8) (rtrPduModels []RtrPduModel, err error) {
	rtrPduModels = make([]RtrPduModel, 0)
	for i := range rtrHroaFulls {
		rtrPduModel, err := convertRtrHroaFullToRtrPduModel(&rtrHroaFulls[i], protocolVersion)
		if err != nil {
			belogs.Error("convertRtrHroaFullsToRtrPduModels(): convertRtrHroaFullToRtrPduModel fail: ", err)
			return nil, err
		}
		rtrPduModels = append(rtrPduModels, rtrPduModel)
	}
	belogs.Debug("convertRtrHroaFullsToRtrPduModels(): len(rtrHroaFulls):", len(rtrHroaFulls),
		" len(rtrPduModels):", len(rtrPduModels))
	return rtrPduModels, nil
}

func convertRtrAsraFullToRtrPduModel(rtrAsraFull *model.LabRpkiRtrAsraFull, protocolVersion uint8) (rtrPduModel RtrPduModel, err error) {
	belogs.Debug("convertRtrAsraFullToRtrPduModel(): rtrAsraFull:", jsonutil.MarshalJson(rtrAsraFull))

	customerAsnAsra := uint32(rtrAsraFull.CustomerAsnAsra.ValueOrZero())
	providerAsnAsras := make([]uint32, 0)
	err = jsonutil.UnmarshalJson(rtrAsraFull.ProviderAsnAsrasStr, &providerAsnAsras)
	if err != nil {
		belogs.Error("convertRtrAsraFullToRtrPduModel(): get providerAsnAsras fail, rtrAsraFull.ProviderAsnAsrasStr:",
			rtrAsraFull.ProviderAsnAsrasStr, err)
		return rtrPduModel, err
	}
	rtrAsraModel := NewRtrAsraModel(protocolVersion, PDU_FLAG_ANNOUNCE, null.IntFrom(0),
		customerAsnAsra, providerAsnAsras)

	if len(rtrAsraFull.OtherNeighborAsnAsrasStr) > 0 {
		asNumber := make([]uint32, 0)
		err = jsonutil.UnmarshalJson(rtrAsraFull.OtherNeighborAsnAsrasStr, &asNumber)
		if err != nil {
			belogs.Error("convertRtrAsraFullToRtrPduModel(): get OtherNeighborAsnAsrasStr fail, rtrAsraFull.OtherNeighborAsnAsrasStr:",
				rtrAsraFull.OtherNeighborAsnAsrasStr, err)
			return rtrPduModel, err
		}
		rtrAsraModel.AddAsNumbers(PDU_TYPE_ASRA_TYPE_ON, asNumber)
	}
	if len(rtrAsraFull.CustomerAsnAsrasStr) > 0 {
		asNumber := make([]uint32, 0)
		err = jsonutil.UnmarshalJson(rtrAsraFull.CustomerAsnAsrasStr, &asNumber)
		if err != nil {
			belogs.Error("convertRtrAsraFullToRtrPduModel(): get CustomerAsnAsrasStr fail, rtrAsraFull.CustomerAsnAsrasStr:",
				rtrAsraFull.CustomerAsnAsrasStr, err)
			return rtrPduModel, err
		}
		rtrAsraModel.AddAsNumbers(PDU_TYPE_ASRA_TYPE_C, asNumber)
	}
	if len(rtrAsraFull.LateralPeerAsnAsrasStr) > 0 {
		asNumber := make([]uint32, 0)
		err = jsonutil.UnmarshalJson(rtrAsraFull.LateralPeerAsnAsrasStr, &asNumber)
		if err != nil {
			belogs.Error("convertRtrAsraFullToRtrPduModel(): get LateralPeerAsnAsrasStr fail, rtrAsraFull.LateralPeerAsnAsrasStr:",
				rtrAsraFull.LateralPeerAsnAsrasStr, err)
			return rtrPduModel, err
		}
		rtrAsraModel.AddAsNumbers(PDU_TYPE_ASRA_TYPE_L, asNumber)
	}
	return rtrAsraModel, nil
}

func convertRtrAsraFullsToRtrPduModels(rtrAsraFulls []model.LabRpkiRtrAsraFull,
	protocolVersion uint8) (rtrPduModels []RtrPduModel, err error) {
	rtrPduModels = make([]RtrPduModel, 0)
	for i := range rtrAsraFulls {
		rtrPduModel, err := convertRtrAsraFullToRtrPduModel(&rtrAsraFulls[i], protocolVersion)
		if err != nil {
			belogs.Error("convertRtrAsraFullsToRtrPduModels(): convertRtrAsraFullToRtrPduModel fail: ", err)
			return nil, err
		}
		rtrPduModels = append(rtrPduModels, rtrPduModel)
	}
	belogs.Debug("convertRtrAsraFullsToRtrPduModels(): len(rtrAsraFulls):", len(rtrAsraFulls),
		" len(rtrPduModels):", len(rtrPduModels))
	return rtrPduModels, nil
}
