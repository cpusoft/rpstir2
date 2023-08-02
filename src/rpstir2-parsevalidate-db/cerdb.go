package parsevalidatedb

import (
	"errors"
	"strings"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
	"xorm.io/xorm"
)

func AddCerDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("AddCerDb(): will add cer file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("AddCerDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = InsertCerDb(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("AddCerDb(): InsertCerDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "AddCerDb(): InsertCerDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = updateSyncLogFileJsonAllAndStateDb(session, syncLogFileModel)
	if err != nil {
		belogs.Error("AddCerDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "AddCerDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("AddCerDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("AddCerDb(): cer file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func DelCerDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("DelCerDb(): will del cer file, certId:", syncLogFileModel.CertId,
		" filePath:", syncLogFileModel.FilePath, "  fileName:", syncLogFileModel.FileName)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("DelCerDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	if syncLogFileModel.CertId == 0 {
		certId, err := getCertIdByFilePathName(session, "lab_rpki_cer", syncLogFileModel.FilePath, syncLogFileModel.FileName)
		if err != nil {
			belogs.Error("DelCerDb(): getCertIdByFilePathName fail, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName, err)
			return xormdb.RollbackAndLogError(session, "DelCerDb(): getCertIdByFilePathName fail, syncLogFileModel:"+syncLogFileModel.String(), err)
		}
		if certId == 0 {
			belogs.Info("DelCerDb(): file not exist in db, just return, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName)
			return nil
		}
		syncLogFileModel.CertId = certId
		belogs.Debug("DelCerDb(): get certId, certId:", syncLogFileModel.CertId)
	}

	err = DelCerByIdDb(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("DelCerDb(): DelCerByIdDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "DelCerDb(): DelCerByIdDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}
	// only del,will update syncLogFile.
	// when is add/update, will update syncLogFile in AddAsaDb()
	if syncLogFileModel.SyncType == "del" {
		err = updateSyncLogFileJsonAllAndStateDb(session, syncLogFileModel)
		if err != nil {
			belogs.Error("DelCerDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
			return xormdb.RollbackAndLogError(session, "DelCerDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("DelCerDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("DelCerDb(): cer file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func DelCerByIdDb(session *xorm.Session, cerId uint64) (err error) {

	belogs.Debug("DelCerByIdDb():delete lab_rpki_cer by cerId:", cerId)
	// rrdp may have id==0, just return nil
	if cerId <= 0 {
		return nil
	}
	belogs.Info("DelCerByIdDb():delete lab_rpki_cer by cerId, more than 0:", cerId)

	//lab_rpki_cer_sia
	res, err := session.Exec("delete from lab_rpki_cer_sia  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("DelCerByIdDb():delete from lab_rpki_cer_sia failed, cerId:", cerId, "    err:", err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("DelCerByIdDb():delete lab_rpki_cer_sia by cerId:", cerId, "  count:", count)

	//lab_rpki_cer_ipaddress
	res, err = session.Exec("delete from  lab_rpki_cer_ipaddress  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("DelCerByIdDb():delete  from lab_rpki_cer_ipaddress failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelCerByIdDb():delete lab_rpki_cer_ipaddress by cerId:", cerId, "  count:", count)

	//lab_rpki_cer_crldp
	res, err = session.Exec("delete  from lab_rpki_cer_crldp  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("DelCerByIdDb():delete  from lab_rpki_cer_crldp failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelCerByIdDb():delete lab_rpki_cer_crldp by cerId:", cerId, "  count:", count)

	//lab_rpki_cer_asn
	res, err = session.Exec("delete  from lab_rpki_cer_asn  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("DelCerByIdDb():delete  from lab_rpki_cer_asn  failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelCerByIdDb():delete lab_rpki_cer_asn by cerId:", cerId, "  count:", count)

	//lab_rpki_cer_aia
	res, err = session.Exec("delete  from lab_rpki_cer_aia  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("DelCerByIdDb():delete  from lab_rpki_cer_aia  failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelCerByIdDb():delete lab_rpki_cer_aia by cerId:", cerId, "  count:", count)

	//lab_rpki_cer
	res, err = session.Exec("delete  from lab_rpki_cer  where id = ?", cerId)
	if err != nil {
		belogs.Error("DelCerByIdDb():delete  from lab_rpki_cer  failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelCerByIdDb():delete lab_rpki_cer by cerId:", cerId, "  count:", count)

	return nil
}

func InsertCerDb(session *xorm.Session,
	syncLogFileModel *model.SyncLogFileModel, now time.Time) error {

	var cerModel model.CerModel
	json := jsonutil.MarshalJson(syncLogFileModel.CertModel)
	err := jsonutil.UnmarshalJson(json, &cerModel)
	if err != nil {
		belogs.Error("InsertCerDb(): json fail, CertModel to cerModel, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is not cerModel type")
	}
	orginModelJson := jsonutil.MarshalJson(syncLogFileModel.OriginModel)
	notBefore := cerModel.NotBefore
	notAfter := cerModel.NotAfter
	belogs.Debug("InsertCerDb():cerModel filePath,fileName:", cerModel.FilePath, cerModel.FileName, "  orginModel:", orginModelJson,
		"  now ", now, "  notBefore:", notBefore, "  notAfter:", notAfter)

	//lab_rpki_cer
	sqlStr := `INSERT lab_rpki_cer(
	    sn, notBefore,notAfter,subject,
	    issuer,ski,aki,filePath,fileName,
	    fileHash,jsonAll,syncLogId,syncLogFileId,updateTime,
	    state,origin) 	
	    VALUES(?,?,?,?,
	    ?,?,?,?,?,
	    ?,?,?,?,?,
	    ?,?)`
	res, err := session.Exec(sqlStr,
		cerModel.Sn, notBefore, notAfter, cerModel.Subject,
		cerModel.Issuer, xormdb.SqlNullString(cerModel.Ski), xormdb.SqlNullString(cerModel.Aki), cerModel.FilePath, cerModel.FileName,
		cerModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(cerModel)), syncLogFileModel.SyncLogId, syncLogFileModel.Id, now,
		xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)),
		xormdb.SqlNullString(orginModelJson))
	if err != nil {
		belogs.Error("InsertCerDb(): INSERT lab_rpki_cer fail, cerModel:", jsonutil.MarshalJson(cerModel),
			"     syncLogFileModel:", syncLogFileModel.String(), err)
		return err
	}

	cerId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("InsertCerDb(): LastInsertId:", syncLogFileModel.String(), err)
		return err
	}
	belogs.Debug("InsertCerDb():LastInsertId cerId:", cerId)

	//lab_rpki_cer_aia
	belogs.Debug("InsertCerDb(): cerModel.Aia.CaIssuers:", cerModel.AiaModel.CaIssuers)
	if len(cerModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT lab_rpki_cer_aia(cerId, caIssuers) VALUES(?,?)`
		res, err = session.Exec(sqlStr, cerId, cerModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("InsertCerDb(): INSERT lab_rpki_cer_aia Exec:", syncLogFileModel.String(), err)
			return err
		}
	}

	//lab_rpki_cer_asn
	belogs.Debug("InsertCerDb(): cerModel.Asn:", cerModel.AsnModel)
	if len(cerModel.AsnModel.Asns) > 0 {
		sqlAsnStr := `INSERT lab_rpki_cer_asn(cerId, asn) VALUES(?,?)`
		sqlMinMaxStr := `INSERT lab_rpki_cer_asn(cerId, min,max) VALUES(?,?,?)`
		for _, asn := range cerModel.AsnModel.Asns {
			// need  asNum >=0
			if asn.Asn >= 0 {
				res, err = session.Exec(sqlAsnStr, cerId, asn.Asn)
				if err != nil {
					belogs.Error("InsertCerDb(): INSERT sqlAsnStr lab_rpki_cer_asn ,syncLogFileModel err:", syncLogFileModel.String(), err)
					return err
				}
			} else if asn.Max >= 0 && asn.Min >= 0 {
				res, err = session.Exec(sqlMinMaxStr, cerId, asn.Min, asn.Max)
				if err != nil {
					belogs.Error("InsertCerDb(): INSERT sqlMinMaxStr lab_rpki_cer_asn,syncLogFileModel err:", syncLogFileModel.String(), err)
					return err
				}
			} else {
				belogs.Error("InsertCerDb(): INSERT lab_rpki_cer_asn asn/min/max all are zero, syncLogFileModel err:", syncLogFileModel.String())
				return errors.New("insert lab_rpki_cer_asn fail, asn/min/max all are zero")
			}
		}
	}

	//lab_rpki_cer_crldp
	belogs.Debug("InsertCerDb(): cerModel.CRLdp:", cerModel.CrldpModel.Crldps)
	if len(cerModel.CrldpModel.Crldps) > 0 {
		sqlStr = `INSERT lab_rpki_cer_crldp(cerId, crldp) VALUES(?,?)`
		for _, crldp := range cerModel.CrldpModel.Crldps {
			res, err = session.Exec(sqlStr, cerId, crldp)
			if err != nil {
				belogs.Error("InsertCerDb(): INSERT lab_rpki_cer_crldp Exec:", syncLogFileModel.String(), err)
				return err
			}
		}
	}

	//lab_rpki_cer_ipaddress
	belogs.Debug("InsertCerDb(): cerModel.CerIpAddressModel:", cerModel.CerIpAddressModel)
	sqlStr = `INSERT lab_rpki_cer_ipaddress(cerId,addressFamily, addressPrefix,min,max,
	                rangeStart,rangeEnd,addressPrefixRange) 
	                 VALUES(?,?,?,?,?,
	                 ?,?,?)`
	for _, cerIpAddress := range cerModel.CerIpAddressModel.CerIpAddresses {
		res, err = session.Exec(sqlStr,
			cerId, cerIpAddress.AddressFamily, cerIpAddress.AddressPrefix, cerIpAddress.Min, cerIpAddress.Max,
			cerIpAddress.RangeStart, cerIpAddress.RangeEnd, cerIpAddress.AddressPrefixRange)
		if err != nil {
			belogs.Error("InsertCerDb(): INSERT lab_rpki_cer_ipaddress Exec:", syncLogFileModel.String(), err)
			return err
		}
	}

	//lab_rpki_cer_sia
	belogs.Debug("InsertCerDb(): cerModel.Sia:", cerModel.SiaModel)
	if len(cerModel.SiaModel.CaRepository) > 0 ||
		len(cerModel.SiaModel.RpkiManifest) > 0 ||
		len(cerModel.SiaModel.RpkiNotify) > 0 ||
		len(cerModel.SiaModel.SignedObject) > 0 {
		sqlStr = `INSERT lab_rpki_cer_sia(cerId, rpkiManifest,rpkiNotify,caRepository,signedObject) VALUES(?,?,?,?,?)`
		res, err = session.Exec(sqlStr, cerId, cerModel.SiaModel.RpkiManifest,
			cerModel.SiaModel.RpkiNotify, cerModel.SiaModel.CaRepository,
			cerModel.SiaModel.SignedObject)
		if err != nil {
			belogs.Error("InsertCerDb(): INSERT lab_rpki_cer_sia Exec:", syncLogFileModel.String(), err)
			return err
		}
	}
	return nil
}

func getExpireCerDb(now time.Time) (certIdStateModels []model.CertIdStateModel, err error) {

	certIdStateModels = make([]model.CertIdStateModel, 0)
	t := convert.Time2String(now)
	sql := `select id, state as stateStr, c.NotAfter as endTime from  lab_rpki_cer c 
			where timestamp(c.NotAfter) < ? order by id `

	err = xormdb.XormEngine.SQL(sql, t).Find(&certIdStateModels)
	if err != nil {
		belogs.Error("getExpireCerDb(): lab_rpki_cer fail:", t, err)
		return nil, err
	}
	belogs.Info("getExpireCerDb(): now t:", t, "  , len(certIdStateModels):", len(certIdStateModels))
	return certIdStateModels, nil
}

func updateCerStateDb(certIdStateModels []model.CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateCerStateDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	sql := `update lab_rpki_cer c set c.state = ? where id = ? `
	for i := range certIdStateModels {
		belogs.Debug("updateCerStateDb():  certIdStateModels[i]:", certIdStateModels[i].Id, certIdStateModels[i].StateStr)
		_, err := session.Exec(sql, certIdStateModels[i].StateStr, certIdStateModels[i].Id)
		if err != nil {
			belogs.Error("updateCerStateDb(): UPDATE lab_rpki_cer fail :", jsonutil.MarshalJson(certIdStateModels[i]), err)
			return xormdb.RollbackAndLogError(session, "updateCerStateDb(): UPDATE lab_rpki_cer fail : certIdStateModels[i]: "+
				jsonutil.MarshalJson(certIdStateModels[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateCerStateDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("updateCerStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Since(start))

	return nil
}

func UpdateCerByCheckAll(now time.Time) error {
	// check expire
	curCertIdStateModels, err := getExpireCerDb(now)
	if err != nil {
		belogs.Error("UpdateCerByCheckAll(): getExpireCerDb:  err: ", err)
		return err
	}
	belogs.Info("UpdateCerByCheckAll(): len(curCertIdStateModels):", len(curCertIdStateModels))

	newCertIdStateModels := make([]model.CertIdStateModel, 0)
	for i := range curCertIdStateModels {
		// if have this error, ignore
		belogs.Debug("UpdateCerByCheckAll(): old curCertIdStateModels[i]:", jsonutil.MarshalJson(curCertIdStateModels[i]))
		if strings.Contains(curCertIdStateModels[i].StateStr, "NotAfter is earlier than the current time") {
			continue
		}

		// will add error
		stateModel := model.StateModel{}
		jsonutil.UnmarshalJson(curCertIdStateModels[i].StateStr, &stateModel)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NotAfter is earlier than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", notAfter is " + convert.Time2StringZone(curCertIdStateModels[i].EndTime)}
		if conf.Bool("policy::allowStaleCer") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}

		certIdStateModel := model.CertIdStateModel{
			Id:       curCertIdStateModels[i].Id,
			StateStr: jsonutil.MarshalJson(stateModel),
		}
		newCertIdStateModels = append(newCertIdStateModels, certIdStateModel)
		belogs.Debug("UpdateCerByCheckAll(): new certIdStateModel:", jsonutil.MarshalJson(certIdStateModel))
	}

	// update db
	err = updateCerStateDb(newCertIdStateModels)
	if err != nil {
		belogs.Error("UpdateCerByCheckAll(): updateCerStateDb:  err: ", len(newCertIdStateModels), err)
		return err
	}
	belogs.Info("UpdateCerByCheckAll(): ok len(newCertIdStateModels):", len(newCertIdStateModels))
	return nil
}
