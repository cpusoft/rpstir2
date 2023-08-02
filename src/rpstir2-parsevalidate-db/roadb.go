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

func AddRoaDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("AddRoaDb(): will add roa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("AddRoaDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = InsertRoaDb(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("AddRoaDb(): InsertRoaDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "AddRoaDb(): InsertRoaDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = updateSyncLogFileJsonAllAndStateDb(session, syncLogFileModel)
	if err != nil {
		belogs.Error("AddRoaDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "AddRoaDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("AddRoaDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("AddRoaDb(): roa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func DelRoaDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("DelRoaDb(): will del roa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("DelRoaDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	if syncLogFileModel.CertId == 0 {
		certId, err := getCertIdByFilePathName(session, "lab_rpki_roa", syncLogFileModel.FilePath, syncLogFileModel.FileName)
		if err != nil {
			belogs.Error("DelRoaDb(): getCertIdByFilePathName fail, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName, err)
			return xormdb.RollbackAndLogError(session, "DelRoaDb(): getCertIdByFilePathName fail, syncLogFileModel:"+syncLogFileModel.String(), err)
		}
		if certId == 0 {
			belogs.Info("DelRoaDb(): file not exist in db, just return, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName)
			return nil
		}
		syncLogFileModel.CertId = certId
		belogs.Debug("DelRoaDb(): get certId, certId:", syncLogFileModel.CertId)
	}

	err = DelRoaByIdDb(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("DelRoaDb(): DelRoaByIdDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "DelRoaDb(): DelRoaByIdDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}
	// only del,will update syncLogFile.
	// when is add/update, will update syncLogFile in AddAsaDb()
	if syncLogFileModel.SyncType == "del" {
		err = updateSyncLogFileJsonAllAndStateDb(session, syncLogFileModel)
		if err != nil {
			belogs.Error("DelRoaDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
			return xormdb.RollbackAndLogError(session, "DelRoaDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
		}
	}
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("DelRoaDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("DelRoaDb(): roa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func DelRoaByIdDb(session *xorm.Session, roaId uint64) (err error) {

	belogs.Debug("DelRoaByIdDb():delete lab_rpki_roa by roaId:", roaId)
	// rrdp may have id==0, just return nil
	if roaId <= 0 {
		return nil
	}
	belogs.Info("DelRoaByIdDb():delete lab_rpki_roa by roaId, more than 0:", roaId)

	//lab_rpki_roa_ipaddress
	res, err := session.Exec("delete from lab_rpki_roa_ipaddress  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("DelRoaByIdDb():delete  from lab_rpki_roa_ipaddress fail: roaId: ", roaId, err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("DelRoaByIdDb():delete lab_rpki_roa_ipaddress by roaId:", roaId, "  count:", count)

	//lab_rpki_roa_ee_ipaddress
	res, err = session.Exec("delete from lab_rpki_roa_ee_ipaddress  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("DelRoaByIdDb():delete  from lab_rpki_roa_ee_ipaddress fail: roaId: ", roaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelRoaByIdDb():delete lab_rpki_roa_ee_ipaddress by roaId:", roaId, "  count:", count)

	//lab_rpki_roa_sia
	res, err = session.Exec("delete from  lab_rpki_roa_sia  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("DelRoaByIdDb():delete  from lab_rpki_roa_sia fail: roaId: ", roaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelRoaByIdDb():delete lab_rpki_roa_sia by roaId:", roaId, "  count:", count)

	//lab_rpki_roa_sia
	res, err = session.Exec("delete from  lab_rpki_roa_aia  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("DelRoaByIdDb():delete  from lab_rpki_roa_aia fail: roaId: ", roaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelRoaByIdDb():delete lab_rpki_roa_aia by roaId:", roaId, "  count:", count)

	//lab_rpki_roa
	res, err = session.Exec("delete from  lab_rpki_roa  where id = ?", roaId)
	if err != nil {
		belogs.Error("DelRoaByIdDb():delete  from lab_rpki_roa fail: roaId: ", roaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelRoaByIdDb():delete lab_rpki_roa by roaId:", roaId, "  count:", count)

	return nil
}

func InsertRoaDb(session *xorm.Session,
	syncLogFileModel *model.SyncLogFileModel, now time.Time) error {

	var roaModel model.RoaModel
	json := jsonutil.MarshalJson(syncLogFileModel.CertModel)
	err := jsonutil.UnmarshalJson(json, &roaModel)
	if err != nil {
		belogs.Error("InsertRoaDb(): json fail, CertModel to RoaModel, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is not roaModel type")
	}
	originModelJson := jsonutil.MarshalJson(syncLogFileModel.OriginModel)
	belogs.Debug("InsertRoaDb(): roaModel filePath,fileName:", roaModel.FilePath, roaModel.FileName, " originModel:", originModelJson, " now ", now)

	//lab_rpki_roa
	sqlStr := `INSERT lab_rpki_roa(
	                asn,  ski, aki, filePath,fileName, 
	                fileHash,jsonAll,syncLogId, syncLogFileId, updateTime,
	                state,origin)
					VALUES(?,?,?,?,?,
					?,?,?,?,?,
					?,?)`
	res, err := session.Exec(sqlStr,
		roaModel.Asn, xormdb.SqlNullString(roaModel.Ski), xormdb.SqlNullString(roaModel.Aki), roaModel.FilePath, roaModel.FileName,
		roaModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(roaModel)), syncLogFileModel.SyncLogId, syncLogFileModel.Id, now,
		xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)),
		xormdb.SqlNullString(originModelJson))
	if err != nil {
		belogs.Error("InsertRoaDb(): INSERT lab_rpki_roa Exec :", syncLogFileModel.String(), err)
		return err
	}

	roaId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("InsertRoaDb(): LastInsertId :", syncLogFileModel.String(), err)
		return err
	}

	//lab_rpki_roa_aia
	belogs.Debug("InsertRoaDb(): roaModel.Aia.CaIssuers:", roaModel.AiaModel.CaIssuers)
	if len(roaModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT lab_rpki_roa_aia(roaId, caIssuers)
				VALUES(?,?)`
		res, err = session.Exec(sqlStr, roaId, roaModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("InsertRoaDb(): INSERT lab_rpki_roa_aia Exec :", syncLogFileModel.String(), err)
			return err
		}
	}

	//lab_rpki_roa_sia
	belogs.Debug("InsertRoaDb(): roaModel.Sia:", roaModel.SiaModel)
	if len(roaModel.SiaModel.CaRepository) > 0 ||
		len(roaModel.SiaModel.RpkiManifest) > 0 ||
		len(roaModel.SiaModel.RpkiNotify) > 0 ||
		len(roaModel.SiaModel.SignedObject) > 0 {
		sqlStr = `INSERT lab_rpki_roa_sia(roaId, rpkiManifest,rpkiNotify,caRepository,signedObject)
				VALUES(?,?,?,?,?)`
		res, err = session.Exec(sqlStr, roaId, roaModel.SiaModel.RpkiManifest,
			roaModel.SiaModel.RpkiNotify, roaModel.SiaModel.CaRepository,
			roaModel.SiaModel.SignedObject)
		if err != nil {
			belogs.Error("InsertRoaDb(): INSERT lab_rpki_roa_sia Exec :", syncLogFileModel.String(), err)
			return err
		}
	}

	//lab_rpki_roa_ipaddress
	belogs.Debug("InsertRoaDb(): roaModel.IPAddrBlocks:", jsonutil.MarshalJson(roaModel.RoaIpAddressModels))
	if roaModel.RoaIpAddressModels != nil && len(roaModel.RoaIpAddressModels) > 0 {
		sqlStr = `INSERT lab_rpki_roa_ipaddress(roaId, addressFamily,addressPrefix,maxLength, rangeStart, rangeEnd,addressPrefixRange )
						VALUES(?,?,?,?,?,?,?)`
		for _, roaIpAddressModel := range roaModel.RoaIpAddressModels {
			res, err = session.Exec(sqlStr, roaId, roaIpAddressModel.AddressFamily,
				roaIpAddressModel.AddressPrefix, roaIpAddressModel.MaxLength,
				roaIpAddressModel.RangeStart, roaIpAddressModel.RangeEnd, roaIpAddressModel.AddressPrefixRange)
			if err != nil {
				belogs.Error("InsertRoaDb(): INSERT lab_rpki_roa_ipaddress Exec :", syncLogFileModel.String(), err)
				return err
			}

		}
	}

	//lab_rpki_roa_ee_ipaddress
	belogs.Debug("InsertRoaDb(): roaModel.CerIpAddressModel:", roaModel.EeCertModel.CerIpAddressModel)
	sqlStr = `INSERT lab_rpki_roa_ee_ipaddress(roaId,addressFamily, addressPrefix,min,max,
	                rangeStart,rangeEnd,addressPrefixRange) 
	                 VALUES(?,?,?,?,?,
	                 ?,?,?)`
	for _, cerIpAddress := range roaModel.EeCertModel.CerIpAddressModel.CerIpAddresses {
		res, err = session.Exec(sqlStr,
			roaId, cerIpAddress.AddressFamily, cerIpAddress.AddressPrefix, cerIpAddress.Min, cerIpAddress.Max,
			cerIpAddress.RangeStart, cerIpAddress.RangeEnd, cerIpAddress.AddressPrefixRange)
		if err != nil {
			belogs.Error("InsertRoaDb(): INSERT lab_rpki_roa_ee_ipaddress Exec:", syncLogFileModel.String(), err)
			return err
		}
	}
	return nil
}

func getExpireRoaDb(now time.Time) (certIdStateModels []model.CertIdStateModel, err error) {

	certIdStateModels = make([]model.CertIdStateModel, 0)
	t := now.Local().Format("2006-01-02T15:04:05-0700")
	sql := `select id, state as stateStr,str_to_date( SUBSTRING_INDEX(c.jsonAll->>'$.eeCertModel.notAfter','+',1),'%Y-%m-%dT%H:%i:%S')  as endTime  from  lab_rpki_roa c 
			where c.jsonAll->>'$.eeCertModel.notAfter' < ? order by id `

	err = xormdb.XormEngine.SQL(sql, t).Find(&certIdStateModels)
	if err != nil {
		belogs.Error("getExpireRoaDb(): lab_rpki_roa fail:", t, err)
		return nil, err
	}
	belogs.Info("getExpireRoaDb(): now t:", t, "  , len(certIdStateModels):", len(certIdStateModels))
	return certIdStateModels, nil
}

func updateRoaStateDb(certIdStateModels []model.CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateRoaStateDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	sql := `update lab_rpki_roa c set c.state = ? where id = ? `
	for i := range certIdStateModels {
		belogs.Debug("updateRoaStateDb():  certIdStateModels[i]:", certIdStateModels[i].Id, certIdStateModels[i].StateStr)
		_, err := session.Exec(sql, certIdStateModels[i].StateStr, certIdStateModels[i].Id)
		if err != nil {
			belogs.Error("updateRoaStateDb(): UPDATE lab_rpki_roa fail :", jsonutil.MarshalJson(certIdStateModels[i]), err)
			return xormdb.RollbackAndLogError(session, "updateRoaStateDb(): UPDATE lab_rpki_roa fail : certIdStateModels[i]: "+
				jsonutil.MarshalJson(certIdStateModels[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateRoaStateDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("updateRoaStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Since(start))

	return nil
}

func UpdateRoaByCheckAll(now time.Time) error {
	// check expire
	curCertIdStateModels, err := getExpireRoaDb(now)
	if err != nil {
		belogs.Error("UpdateRoaByCheckAll(): getExpireRoaDb:  err: ", err)
		return err
	}
	belogs.Info("UpdateRoaByCheckAll(): len(curCertIdStateModels):", len(curCertIdStateModels))

	newCertIdStateModels := make([]model.CertIdStateModel, 0)
	for i := range curCertIdStateModels {
		// if have this error, ignore
		belogs.Debug("UpdateRoaByCheckAll(): old curCertIdStateModels[i]:", jsonutil.MarshalJson(curCertIdStateModels[i]))
		if strings.Contains(curCertIdStateModels[i].StateStr, "NotAfter of EE is earlier than the current time") {
			continue
		}

		// will add error
		stateModel := model.StateModel{}
		jsonutil.UnmarshalJson(curCertIdStateModels[i].StateStr, &stateModel)

		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NotAfter of EE is earlier than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", notAfter is " + convert.Time2StringZone(curCertIdStateModels[i].EndTime)}
		if conf.Bool("policy::allowStaleEe") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}

		certIdStateModel := model.CertIdStateModel{
			Id:       curCertIdStateModels[i].Id,
			StateStr: jsonutil.MarshalJson(stateModel),
		}
		newCertIdStateModels = append(newCertIdStateModels, certIdStateModel)
		belogs.Debug("UpdateRoaByCheckAll(): new certIdStateModel:", jsonutil.MarshalJson(certIdStateModel))
	}

	// update db
	err = updateRoaStateDb(newCertIdStateModels)
	if err != nil {
		belogs.Error("UpdateRoaByCheckAll(): updateRoaStateDb:  err: ", len(newCertIdStateModels), err)
		return err
	}
	belogs.Info("UpdateRoaByCheckAll(): ok len(newCertIdStateModels):", len(newCertIdStateModels))
	return nil

}
