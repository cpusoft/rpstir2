package rtrserver

import (
	"errors"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
)

func getSessionIdDb() (sessionId uint16, err error) {
	// lab_rpki_rtr_session, get sessionId
	sql := `select max(sessionId) as sessionId from lab_rpki_rtr_session `
	has, err := xormdb.XormEngine.SQL(sql).Get(&sessionId)
	if err != nil {
		belogs.Error("getSessionIdDb():select max(sessionId) lab_rpki_rtr_session fail:", err)
		return sessionId, err
	}
	if !has {
		belogs.Error("getSessionIdDb():select max(sessionId) lab_rpki_rtr_session have no sessionId:", has)
		return sessionId, errors.New("select max(sessionId) lab_rpki_rtr_session have no sessionId")
	}
	belogs.Debug("getSessionIdDb():select max(sessionId) lab_rpki_rtr_session, sessionId :", sessionId)
	return sessionId, nil
}

func getSpanSerialNumbersDb(clientSerialNumber uint32) (serialNumbers []uint32, err error) {
	serialNumbers = make([]uint32, 0)
	err = xormdb.XormEngine.Table("lab_rpki_rtr_serial_number").Cols("serialNumber").Where("serialNumber > ?", clientSerialNumber).Find(&serialNumbers)
	if err != nil {
		belogs.Error("getSpanSerialNumbersDb():get serialNumbers fail, clientSerialNumber: ", clientSerialNumber, err)
		return serialNumbers, err
	}
	belogs.Debug("getSpanSerialNumbersDb(): clientSerialNumber : ", clientSerialNumber, "   serialNumbers:", serialNumbers)
	return serialNumbers, nil
}

func getRtrXIncrementalDb(clientSerialNumber uint32) (
	rtrIncrementals []model.LabRpkiRtrIncremental,
	rtrAsaIncrementals []model.LabRpkiRtrAsaIncremental,
	rtrHroaIncrementals []model.LabRpkiRtrHroaIncremental,
	rtrAsraIncrementals []model.LabRpkiRtrAsraIncremental,
	sessionId uint16, serialNumber uint32, err error) {

	start := time.Now()
	rtrIncrementals = make([]model.LabRpkiRtrIncremental, 0)
	err = xormdb.XormEngine.Where("serialNumber > ?", clientSerialNumber).Find(&rtrIncrementals)
	if err != nil {
		belogs.Error("getRtrXIncrementalDb():get rtrIncrementals fail:  clientSerialNumber is ", clientSerialNumber, err)
		return nil, nil, nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrXIncrementalDb():select lab_rpki_rtr_incremental, clientSerialNumber:", clientSerialNumber,
		"  len(rtrIncrementals):", len(rtrIncrementals))

	rtrAsaIncrementals = make([]model.LabRpkiRtrAsaIncremental, 0)
	err = xormdb.XormEngine.Where("serialNumber > ?", clientSerialNumber).Find(&rtrAsaIncrementals)
	if err != nil {
		belogs.Error("getRtrXIncrementalDb():get rtrAsaIncrementals fail:  clientSerialNumber is ", clientSerialNumber, err)
		return nil, nil, nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrXIncrementalDb():select lab_rpki_rtr_asa_incremental, clientSerialNumber:", clientSerialNumber,
		" len(rtrAsaIncrementals):", len(rtrAsaIncrementals))

	rtrHroaIncrementals = make([]model.LabRpkiRtrHroaIncremental, 0)
	err = xormdb.XormEngine.SQL(`select id, serialNumber, style,hroaAsn, subtreeIdentifier, encodedSubtree, afiFlags 
			from lab_rpki_rtr_hroa_incremental  
			where serialNumber > ? order by id`, clientSerialNumber).Find(&rtrHroaIncrementals)
	if err != nil {
		belogs.Error("getRtrXIncrementalDb():get rtrHroaIncrementals fail:  clientSerialNumber is ", clientSerialNumber, err)
		return nil, nil, nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrXIncrementalDb():select lab_rpki_rtr_hroa_incremental, clientSerialNumber:", clientSerialNumber,
		"  len(rtrHroaIncrementals):", len(rtrHroaIncrementals))

	rtrAsraIncrementals = make([]model.LabRpkiRtrAsraIncremental, 0)
	err = xormdb.XormEngine.SQL(`select id, serialNumber,style, customerAsnAsra, addressFamilyAsra, providerAsnAsras as providerAsnAsrasStr, otherNeighborAsnAsras as otherNeighborAsnAsrasStr, customerAsnAsras as customerAsnAsrasStr, lateralPeerAsnAsras as lateralPeerAsnAsrasStr, hybridAsras as hybridAsrasStr, valleyPathAsnAsras as valleyPathAsnAsrasStr 
			from lab_rpki_rtr_asra_incremental
			where serialNumber > ? order by id`, clientSerialNumber).Find(&rtrAsraIncrementals)
	if err != nil {
		belogs.Error("getRtrXIncrementalDb():get rtrAsraIncrementals fail:  clientSerialNumber is ", clientSerialNumber, err)
		return nil, nil, nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrXIncrementalDb():select lab_rpki_rtr_asra_incremental, clientSerialNumber:", clientSerialNumber,
		"  len(rtrAsraIncrementals):", len(rtrAsraIncrementals))

	sessionId, err = getSessionIdDb()
	if err != nil {
		belogs.Error("getRtrXIncrementalDb():getSessionIdDb fail:", err)
		return nil, nil, nil, nil, sessionId, serialNumber, err
	}

	// lab_rpki_rtr_serial_number, get serialNumber
	serialNumber, err = getMaxSerialNumberDb()
	if err != nil {
		belogs.Error("getRtrXIncrementalDb():getMaxSerialNumberDb fail:", err)
		return nil, nil, nil, nil, sessionId, serialNumber, err
	}

	belogs.Info("getRtrXIncrementalDb():len(rtrIncrementals):", len(rtrIncrementals),
		"   len(rtrAsaIncrementals):", len(rtrAsaIncrementals),
		"   len(rtrHroaIncrementals):", len(rtrHroaIncrementals),
		"   len(rtrAsraIncrementals):", len(rtrAsraIncrementals),
		"   sessionId:", sessionId, "  serialNumber:", serialNumber,
		"   clientSerialNumber:", clientSerialNumber, "  time(s):", time.Since(start))
	return rtrIncrementals, rtrAsaIncrementals, rtrHroaIncrementals, rtrAsraIncrementals, sessionId, serialNumber, nil
}

func getRtrXFullDb() (rtrFulls []model.LabRpkiRtrFull,
	rtrAsaFulls []model.LabRpkiRtrAsaFull,
	rtrHroaFulls []model.LabRpkiRtrHroaFull,
	rtrAsraFulls []model.LabRpkiRtrAsraFull,
	sessionId uint16, serialNumber uint32, err error) {
	start := time.Now()
	/*
		sql := `select id, serialNumber, asn,address, prefixLength,maxLength
		from lab_rpki_rtr_full order by id`
	*/
	rtrFulls = make([]model.LabRpkiRtrFull, 0)
	err = xormdb.XormEngine.Table("lab_rpki_rtr_full").Cols("id, serialNumber, asn,address, prefixLength,maxLength").
		OrderBy("id").Find(&rtrFulls)
	if err != nil {
		belogs.Error("getRtrXFullDb():select  lab_rpki_rtr_full fail:", err)
		return nil, nil, nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrXFullDb():select lab_rpki_rtr_full, len :", len(rtrFulls))

	rtrAsaFulls = make([]model.LabRpkiRtrAsaFull, 0)
	err = xormdb.XormEngine.Table("lab_rpki_rtr_asa_full").Cols("id, serialNumber, customerAsn, providerAsn, addressFamily").
		OrderBy("id").Find(&rtrAsaFulls)
	if err != nil {
		belogs.Error("getRtrXFullDb():select  lab_rpki_rtr_asa_full fail:", err)
		return nil, nil, nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrXFullDb():select lab_rpki_rtr_asa_full, len :", len(rtrAsaFulls))

	rtrHroaFulls = make([]model.LabRpkiRtrHroaFull, 0)
	err = xormdb.XormEngine.Table("lab_rpki_rtr_hroa_full").Cols("id, serialNumber, hroaAsn, subtreeIdentifier, encodedSubtree, afiFlags").
		OrderBy("id").Find(&rtrHroaFulls)
	if err != nil {
		belogs.Error("getRtrXFullDb():select  lab_rpki_rtr_hroa_full fail:", err)
		return nil, nil, nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrXFullDb():select lab_rpki_rtr_hroa_full, len :", len(rtrHroaFulls))

	rtrAsraFulls = make([]model.LabRpkiRtrAsraFull, 0)
	err = xormdb.XormEngine.Table("lab_rpki_rtr_asra_full").Cols("id, serialNumber, customerAsnAsra, addressFamilyAsra, providerAsnAsras as providerAsnAsrasStr, otherNeighborAsnAsras as otherNeighborAsnAsrasStr, customerAsnAsras as customerAsnAsrasStr, lateralPeerAsnAsras as lateralPeerAsnAsrasStr, hybridAsras as hybridAsrasStr, valleyPathAsnAsras as valleyPathAsnAsrasStr").
		OrderBy("id").Find(&rtrAsraFulls)
	if err != nil {
		belogs.Error("getRtrXFullDb():select  lab_rpki_rtr_asra_full fail:", err)
		return nil, nil, nil, nil, sessionId, serialNumber, err
	}
	belogs.Debug("getRtrXFullDb():select lab_rpki_rtr_asra_full, len :", len(rtrHroaFulls))

	// lab_rpki_rtr_serial_number, get serialNumber
	serialNumber, err = getMaxSerialNumberDb()
	if err != nil {
		belogs.Error("getRtrXFullDb():getMaxSerialNumberDb fail:", err)
		return nil, nil, nil, nil, sessionId, serialNumber, err
	}

	sessionId, err = getSessionIdDb()
	if err != nil {
		belogs.Error("getRtrXFullDb():getSessionIdDb fail:", err)
		return nil, nil, nil, nil, sessionId, serialNumber, err
	}
	belogs.Info("getRtrXFullDb():len(rtrFulls):", len(rtrFulls),
		"   len(rtrAsaFulls):", len(rtrAsaFulls),
		"   len(rtrHroaFulls):", len(rtrHroaFulls),
		"   len(rtrAsraFulls):", len(rtrAsraFulls),
		"   sessionId:", sessionId, "  serialNumber:", serialNumber,
		"   time(s):", time.Since(start))
	return rtrFulls, rtrAsaFulls, rtrHroaFulls, rtrAsraFulls, sessionId, serialNumber, nil
}
func getSessionIdAndSerialNumberDb() (sessionId uint16, serialNumber uint32, err error) {

	// lab_rpki_rtr_serial_number, get serialNumber
	serialNumber, err = getMaxSerialNumberDb()
	if err != nil {
		belogs.Error("getSessionIdAndSerialNumberDb():getMaxSerialNumberDb fail:", err)
		return sessionId, serialNumber, err
	}

	sessionId, err = getSessionIdDb()
	if err != nil {
		belogs.Error("getSessionIdAndSerialNumberDb():getSessionIdDb fail:", err)
		return sessionId, serialNumber, err
	}
	belogs.Debug("getSessionIdAndSerialNumberDb(): serialNumber:", serialNumber, "   sessionId:", sessionId)
	return sessionId, serialNumber, nil
}
