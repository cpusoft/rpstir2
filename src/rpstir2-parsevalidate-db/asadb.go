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

func AddAsaDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("AddAsaDb(): will add asa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("AddAsaDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = InsertAsaDb(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("AddAsaDb(): InsertAsaDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "AddAsaDb(): InsertAsaDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = updateSyncLogFileJsonAllAndStateDb(session, syncLogFileModel)
	if err != nil {
		belogs.Error("AddAsaDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "AddAsaDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("AddAsaDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("AddAsaDb(): asa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func DelAsaDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("DelAsaDb(): will del asa file, certId:", syncLogFileModel.CertId,
		" filePath:", syncLogFileModel.FilePath, "  fileName:", syncLogFileModel.FileName)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("DelAsaDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	if syncLogFileModel.CertId == 0 {
		certId, err := getCertIdByFilePathName(session, "lab_rpki_asa", syncLogFileModel.FilePath, syncLogFileModel.FileName)
		if err != nil {
			belogs.Error("DelAsaDb(): getCertIdByFilePathName fail, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName, err)
			return xormdb.RollbackAndLogError(session, "DelAsaDb(): getCertIdByFilePathName fail, syncLogFileModel:"+syncLogFileModel.String(), err)
		}
		if certId == 0 {
			belogs.Info("DelAsaDb(): file not exist in db, just return, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName)
			return nil
		}
		syncLogFileModel.CertId = certId
		belogs.Debug("DelAsaDb(): get certId, certId:", syncLogFileModel.CertId)
	}

	err = DelAsaByIdDb(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("DelAsaDb(): DelAsaByIdDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "DelAsaDb(): DelAsaByIdDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}
	// only del,will update syncLogFile.
	// when is add/update, will update syncLogFile in AddAsaDb()
	if syncLogFileModel.SyncType == "del" {
		err = updateSyncLogFileJsonAllAndStateDb(session, syncLogFileModel)
		if err != nil {
			belogs.Error("DelAsaDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
			return xormdb.RollbackAndLogError(session, "DelAsaDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("DelAsaDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("DelAsaDb(): asa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func DelAsaByIdDb(session *xorm.Session, asaId uint64) (err error) {

	belogs.Debug("DelAsaByIdDb():delete lab_rpki_asa by asaId:", asaId)

	// rrdp may have id==0, just return nil
	if asaId <= 0 {
		return nil
	}
	belogs.Info("DelAsaByIdDb():delete lab_rpki_asa by asaId, more than 0:", asaId)

	//lab_rpki_asa_provider_asn
	res, err := session.Exec("delete from lab_rpki_asa_provider_asn  where asaId = ?", asaId)
	if err != nil {
		belogs.Error("DelAsaByIdDb():delete  from lab_rpki_asa_provider_asn fail: asaId: ", asaId, err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("DelAsaByIdDb():delete lab_rpki_asa_provider_asn by asaId:", asaId, "  count:", count)

	//lab_rpki_asa_customer_asn
	res, err = session.Exec("delete from lab_rpki_asa_customer_asn  where asaId = ?", asaId)
	if err != nil {
		belogs.Error("DelAsaByIdDb():delete  from lab_rpki_asa_customer_asn fail: asaId: ", asaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelAsaByIdDb():delete lab_rpki_asa_customer_asn by asaId:", asaId, "  count:", count)

	//lab_rpki_asa_aia
	res, err = session.Exec("delete from  lab_rpki_asa_aia  where asaId = ?", asaId)
	if err != nil {
		belogs.Error("DelAsaByIdDb():delete  from lab_rpki_asa_aia fail: asaId: ", asaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelAsaByIdDb():delete lab_rpki_asa_aia by asaId:", asaId, "  count:", count)

	//lab_rpki_asa_sia
	res, err = session.Exec("delete from  lab_rpki_asa_sia  where asaId = ?", asaId)
	if err != nil {
		belogs.Error("DelAsaByIdDb():delete  from lab_rpki_asa_sia fail: asaId: ", asaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelAsaByIdDb():delete lab_rpki_asa_sia by asaId:", asaId, "  count:", count)

	//lab_rpki_asa
	res, err = session.Exec("delete from  lab_rpki_asa  where id = ?", asaId)
	if err != nil {
		belogs.Error("DelAsaByIdDb():delete  from lab_rpki_asa fail: asaId: ", asaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelAsaByIdDb():delete lab_rpki_asa by asaId:", asaId, "  count:", count)

	return nil
}

func InsertAsaDb(session *xorm.Session,
	syncLogFileModel *model.SyncLogFileModel, now time.Time) error {

	var asaModel model.AsaModel
	json := jsonutil.MarshalJson(syncLogFileModel.CertModel)
	err := jsonutil.UnmarshalJson(json, &asaModel)
	if err != nil {
		belogs.Error("InsertAsaDb(): json fail, CertModel to asaModel, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is not asaModel type")
	}

	orginModelJson := jsonutil.MarshalJson(syncLogFileModel.OriginModel)
	belogs.Debug("InsertAsaDb():asaModel filePath,fileName:", asaModel.FilePath, asaModel.FileName, "  orginModel:", orginModelJson, "  now ", now)

	//lab_rpki_asa
	sqlStr := `INSERT lab_rpki_asa(
	                ski, aki, filePath,fileName, 
	                fileHash,jsonAll,syncLogId, syncLogFileId, updateTime,
	                state,origin)
					VALUES(?,?,?,?,
					?,?,?,?,?,
					?,?)`
	res, err := session.Exec(sqlStr,
		xormdb.SqlNullString(asaModel.Ski), xormdb.SqlNullString(asaModel.Aki), asaModel.FilePath, asaModel.FileName,
		asaModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(asaModel)), syncLogFileModel.SyncLogId, syncLogFileModel.Id, now,
		xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)),
		xormdb.SqlNullString(orginModelJson))
	if err != nil {
		belogs.Error("InsertAsaDb(): INSERT lab_rpki_asa Exec :", syncLogFileModel.String(), err)
		return err
	}

	asaId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("InsertAsaDb(): LastInsertId asaId:", syncLogFileModel.String(), err)
		return err
	}

	//lab_rpki_asa_aia
	belogs.Debug("InsertAsaDb(): asaModel.Aia.CaIssuers:", asaModel.AiaModel.CaIssuers)
	if len(asaModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT lab_rpki_asa_aia(asaId, caIssuers)
				VALUES(?,?)`
		res, err = session.Exec(sqlStr, asaId, asaModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("InsertAsaDb(): INSERT lab_rpki_asa_aia Exec:", syncLogFileModel.String(), err)
			return err
		}
	}

	//lab_rpki_asa_sia
	belogs.Debug("InsertAsaDb(): asaModel.Sia:", asaModel.SiaModel)
	if len(asaModel.SiaModel.CaRepository) > 0 ||
		len(asaModel.SiaModel.RpkiManifest) > 0 ||
		len(asaModel.SiaModel.RpkiNotify) > 0 ||
		len(asaModel.SiaModel.SignedObject) > 0 {
		sqlStr = `INSERT lab_rpki_asa_sia(asaId, rpkiManifest,rpkiNotify,caRepository,signedObject)
				VALUES(?,?,?,?,?)`
		res, err = session.Exec(sqlStr, asaId, asaModel.SiaModel.RpkiManifest,
			asaModel.SiaModel.RpkiNotify, asaModel.SiaModel.CaRepository,
			asaModel.SiaModel.SignedObject)
		if err != nil {
			belogs.Error("InsertAsaDb(): INSERT lab_rpki_asa_sia Exec:", syncLogFileModel.String(), err)
			return err
		}
	}

	//lab_rpki_asa_customer_asn
	belogs.Debug("InsertAsaDb(): asaModel.CustomerAsns:", jsonutil.MarshalJson(asaModel.CustomerAsns))
	customerSqlStr := `INSERT lab_rpki_asa_customer_asn(asaId, customerAsn)
						VALUES(?,?)`
	providerSqlStr := `INSERT lab_rpki_asa_provider_asn(asaId,customerAsnId, providerAsn,providerOrder) 
						VALUES(?,?,?,?)`
	if asaModel.CustomerAsns != nil && len(asaModel.CustomerAsns) > 0 {
		for _, customerAsn := range asaModel.CustomerAsns {
			res, err = session.Exec(customerSqlStr, asaId, customerAsn.CustomerAsn)
			if err != nil {
				belogs.Error("InsertAsaDb(): INSERT lab_rpki_asa_customer_asn Exec:", syncLogFileModel.String(), err)
				return err
			}
			customerAsnId, err := res.LastInsertId()
			if err != nil {
				belogs.Error("InsertAsaDb(): LastInsertId customerAsnId:", syncLogFileModel.String(), err)
				return err
			}
			//lab_rpki_asa_provider_asn
			belogs.Debug("InsertAsaDb(): customerAsn.ProviderAsns:", syncLogFileModel.String())

			for i, providerAsn := range customerAsn.ProviderAsns {
				res, err = session.Exec(providerSqlStr,
					asaId, customerAsnId, providerAsn, i)
				if err != nil {
					belogs.Error("InsertAsaDb(): INSERT lab_rpki_asa_provider_asn Exec:", syncLogFileModel.String(), err)
					return err
				}
			}
		}
	}
	return nil
}

func getExpireAsaDb(now time.Time) (certIdStateModels []model.CertIdStateModel, err error) {

	certIdStateModels = make([]model.CertIdStateModel, 0)
	t := now.Local().Format("2006-01-02T15:04:05-0700")
	sql := `select id, state as stateStr,str_to_date( SUBSTRING_INDEX(c.jsonAll->>'$.eeCertModel.notAfter','+',1),'%Y-%m-%dT%H:%i:%S')  as endTime  from  lab_rpki_asa c 
			where c.jsonAll->>'$.eeCertModel.notAfter' < ? order by id `

	err = xormdb.XormEngine.SQL(sql, t).Find(&certIdStateModels)
	if err != nil {
		belogs.Error("getExpireAsaDb(): lab_rpki_asa fail:", t, err)
		return nil, err
	}
	belogs.Info("getExpireAsaDb(): now t:", t, "  , len(certIdStateModels):", len(certIdStateModels))
	return certIdStateModels, nil
}

func updateAsaStateDb(certIdStateModels []model.CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateAsaStateDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	sql := `update lab_rpki_asa c set c.state = ? where id = ? `
	for i := range certIdStateModels {
		belogs.Debug("updateAsaStateDb():  certIdStateModels[i]:", certIdStateModels[i].Id, certIdStateModels[i].StateStr)
		_, err := session.Exec(sql, certIdStateModels[i].StateStr, certIdStateModels[i].Id)
		if err != nil {
			belogs.Error("updateAsaStateDb(): UPDATE lab_rpki_asa fail :", jsonutil.MarshalJson(certIdStateModels[i]), err)
			return xormdb.RollbackAndLogError(session, "updateAsaStateDb(): UPDATE lab_rpki_asa fail : certIdStateModels[i]: "+
				jsonutil.MarshalJson(certIdStateModels[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateAsaStateDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("updateAsaStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Since(start))

	return nil
}

func UpdateAsaByCheckAll(now time.Time) error {
	// check expire
	curCertIdStateModels, err := getExpireAsaDb(now)
	if err != nil {
		belogs.Error("UpdateAsaByCheckAll(): getExpireAsaDb:  err: ", err)
		return err
	}
	belogs.Info("UpdateAsaByCheckAll(): len(curCertIdStateModels):", len(curCertIdStateModels))

	newCertIdStateModels := make([]model.CertIdStateModel, 0)
	for i := range curCertIdStateModels {
		// if have this error, ignore
		belogs.Debug("UpdateAsaByCheckAll(): old curCertIdStateModels[i]:", jsonutil.MarshalJson(curCertIdStateModels[i]))
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
		belogs.Debug("UpdateAsaByCheckAll(): new certIdStateModel:", jsonutil.MarshalJson(certIdStateModel))
	}

	// update db
	err = updateAsaStateDb(newCertIdStateModels)
	if err != nil {
		belogs.Error("UpdateAsaByCheckAll(): updateAsaStateDb:  err: ", len(newCertIdStateModels), err)
		return err
	}
	belogs.Info("UpdateAsaByCheckAll(): ok len(newCertIdStateModels):", len(newCertIdStateModels))
	return nil

}
