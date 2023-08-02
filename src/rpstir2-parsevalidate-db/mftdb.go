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

func AddMftDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("AddMftDb(): will add mft file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("AddMftDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = InsertMftDb(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("AddMftDb(): InsertMftDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "AddMftDb(): InsertMftDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = updateSyncLogFileJsonAllAndStateDb(session, syncLogFileModel)
	if err != nil {
		belogs.Error("AddMftDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "AddMftDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("AddMftDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("AddMftDb(): mft file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func DelMftDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("DelMftDb(): will del mft file:", syncLogFileModel.FilePath, syncLogFileModel.FileName)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("DelMftDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	if syncLogFileModel.CertId == 0 {
		certId, err := getCertIdByFilePathName(session, "lab_rpki_mft", syncLogFileModel.FilePath, syncLogFileModel.FileName)
		if err != nil {
			belogs.Error("DelMftDb(): getCertIdByFilePathName fail, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName, err)
			return xormdb.RollbackAndLogError(session, "DelMftDb(): getCertIdByFilePathName fail, syncLogFileModel:"+syncLogFileModel.String(), err)
		}
		if certId == 0 {
			belogs.Info("DelMftDb(): file not exist in db, just return, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName)
			return nil
		}
		syncLogFileModel.CertId = certId
		belogs.Debug("DelMftDb(): get certId, certId:", syncLogFileModel.CertId)
	}

	err = DelMftByIdDb(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("DelMftDb(): DelMftByIdDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "DelMftDb(): DelMftByIdDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}
	// only del,will update syncLogFile.
	// when is add/update, will update syncLogFile in AddAsaDb()
	if syncLogFileModel.SyncType == "del" {
		err = updateSyncLogFileJsonAllAndStateDb(session, syncLogFileModel)
		if err != nil {
			belogs.Error("DelMftDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:", syncLogFileModel.String(), err)
			return xormdb.RollbackAndLogError(session, "DelMftDb(): updateSyncLogFileJsonAllAndStateDb fail, syncLogFileModel:"+syncLogFileModel.String(), err)
		}
	}
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("DelMftDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("DelMftDb(): mft file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func DelMftByIdDb(session *xorm.Session, mftId uint64) (err error) {
	belogs.Debug("DelMftByIdDb():delete lab_rpki_mft by mftId:", mftId)

	// rrdp may have id==0, just return nil
	if mftId <= 0 {
		return nil
	}
	belogs.Info("DelMftByIdDb():delete lab_rpki_mft by mftId, more than 0:", mftId)

	//lab_rpki_mft_file_hash
	res, err := session.Exec("delete from lab_rpki_mft_file_hash  where mftId = ?", mftId)
	if err != nil {
		belogs.Error("DelMftByIdDb():delete  from lab_rpki_mft_file_hash fail: mftId: ", mftId, err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("DelMftByIdDb():delete lab_rpki_mft_file_hash by mftId:", mftId, "  count:", count)

	//lab_rpki_mft_sia
	res, err = session.Exec("delete from  lab_rpki_mft_sia  where mftId = ?", mftId)
	if err != nil {
		belogs.Error("DelMftByIdDb():delete  from lab_rpki_mft_sia fail:mftId: ", mftId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelMftByIdDb():delete lab_rpki_mft_sia by mftId:", mftId, "  count:", count)

	//lab_rpki_mft_aia
	res, err = session.Exec("delete from  lab_rpki_mft_aia  where mftId = ?", mftId)
	if err != nil {
		belogs.Error("DelMftByIdDb():delete  from lab_rpki_mft_aia fail:mftId: ", mftId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelMftByIdDb():delete lab_rpki_mft_aia by mftId:", mftId, "  count:", count)

	//lab_rpki_mft
	res, err = session.Exec("delete from  lab_rpki_mft  where id = ?", mftId)
	if err != nil {
		belogs.Error("DelMftByIdDb():delete  from lab_rpki_mft fail:mftId: ", mftId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelMftByIdDb():delete lab_rpki_mft by mftId:", mftId, "  count:", count)

	return nil

}

func InsertMftDb(session *xorm.Session,
	syncLogFileModel *model.SyncLogFileModel, now time.Time) error {

	var mftModel model.MftModel
	json := jsonutil.MarshalJson(syncLogFileModel.CertModel)
	err := jsonutil.UnmarshalJson(json, &mftModel)
	if err != nil {
		belogs.Error("InsertMftDb(): json fail, CertModel to mftModel, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is not mftModel type")
	}
	originModelJson := jsonutil.MarshalJson(syncLogFileModel.OriginModel)
	thisUpdate := mftModel.ThisUpdate
	nextUpdate := mftModel.NextUpdate
	belogs.Debug("InsertMftDb(): mftModel filePath,fileName:", mftModel.FilePath, mftModel.FileName, " originModel:", originModelJson, " now ", now, "  thisUpdate:", thisUpdate, "  nextUpdate:", nextUpdate)

	//lab_rpki_manifest
	sqlStr := `INSERT lab_rpki_mft(
	           mftNumber, thisUpdate, nextUpdate, ski, aki, 
	           filePath,fileName,fileHash, jsonAll,syncLogId, 
	           syncLogFileId, updateTime,state,origin)
				VALUES(
				?,?,?,?,?,
				?,?,?,?,?,
				?,?,?,?)`
	res, err := session.Exec(sqlStr,
		mftModel.MftNumber, thisUpdate, nextUpdate, xormdb.SqlNullString(mftModel.Ski), xormdb.SqlNullString(mftModel.Aki),
		mftModel.FilePath, mftModel.FileName, mftModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(mftModel)), syncLogFileModel.SyncLogId,
		syncLogFileModel.Id, now,
		xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)),
		xormdb.SqlNullString(originModelJson))
	if err != nil {
		belogs.Error("InsertMftDb(): INSERT lab_rpki_mft Exec :", syncLogFileModel.String(), err)
		return err
	}

	mftId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("InsertMftDb(): LastInsertId :", syncLogFileModel.String(), err)
		return err
	}

	//lab_rpki_mft_aia
	belogs.Debug("InsertMftDb(): mftModel.Aia.CaIssuers:", mftModel.AiaModel.CaIssuers)
	if len(mftModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT lab_rpki_mft_aia(mftId, caIssuers) 
			VALUES(?,?)`
		res, err = session.Exec(sqlStr, mftId, mftModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("InsertMftDb(): INSERT lab_rpki_mft_aia Exec :", syncLogFileModel.String(), err)
			return err
		}
	}

	//lab_rpki_mft_sia
	belogs.Debug("InsertMftDb(): mftModel.Sia:", mftModel.SiaModel)
	if len(mftModel.SiaModel.CaRepository) > 0 ||
		len(mftModel.SiaModel.RpkiManifest) > 0 ||
		len(mftModel.SiaModel.RpkiNotify) > 0 ||
		len(mftModel.SiaModel.SignedObject) > 0 {
		sqlStr = `INSERT lab_rpki_mft_sia(mftId, rpkiManifest,rpkiNotify,caRepository,signedObject) 
			VALUES(?,?,?,?,?)`
		res, err = session.Exec(sqlStr, mftId, mftModel.SiaModel.RpkiManifest,
			mftModel.SiaModel.RpkiNotify, mftModel.SiaModel.CaRepository,
			mftModel.SiaModel.SignedObject)
		if err != nil {
			belogs.Error("InsertMftDb(): INSERT lab_rpki_mft_sia Exec :", syncLogFileModel.String(), err)
			return err
		}
	}

	//lab_rpki_mft_fileAndHashs
	belogs.Debug("InsertMftDb(): len(mftModel.FileHashModels):", len(mftModel.FileHashModels))
	if mftModel.FileHashModels != nil && len(mftModel.FileHashModels) > 0 {
		sqlStr = `INSERT lab_rpki_mft_file_hash(mftId, file,hash) VALUES(?,?,?)`
		for _, fileHashModel := range mftModel.FileHashModels {
			res, err = session.Exec(sqlStr, mftId, fileHashModel.File, fileHashModel.Hash)
			if err != nil {
				belogs.Error("InsertMftDb(): INSERT lab_rpki_mft_file_hash Exec :", syncLogFileModel.String(), err)
				return err
			}
		}
	}
	return nil
}

func getExpireMftDb(now time.Time) (certIdStateModels []model.CertIdStateModel, err error) {

	certIdStateModels = make([]model.CertIdStateModel, 0)
	t := convert.Time2String(now)
	sql := `select id, state as stateStr, c.nextUpdate  as endTime from  lab_rpki_mft c 
			where timestamp(c.nextUpdate) < ? order by id `

	err = xormdb.XormEngine.SQL(sql, t).Find(&certIdStateModels)
	if err != nil {
		belogs.Error("getExpireMftDb(): lab_rpki_mft fail:", t, err)
		return nil, err
	}
	belogs.Info("getExpireMftDb(): now t:", t, "  , len(certIdStateModels):", len(certIdStateModels))
	return certIdStateModels, nil
}

func updateMftStateDb(certIdStateModels []model.CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateMftStateDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	sql := `update lab_rpki_mft c set c.state = ? where id = ? `
	for i := range certIdStateModels {
		belogs.Debug("updateMftStateDb():  certIdStateModels[i]:", certIdStateModels[i].Id, certIdStateModels[i].StateStr)
		_, err := session.Exec(sql, certIdStateModels[i].StateStr, certIdStateModels[i].Id)
		if err != nil {
			belogs.Error("updateMftStateDb(): UPDATE lab_rpki_mft fail :", jsonutil.MarshalJson(certIdStateModels[i]), err)
			return xormdb.RollbackAndLogError(session, "updateMftStateDb(): UPDATE lab_rpki_mft fail : certIdStateModels[i]: "+
				jsonutil.MarshalJson(certIdStateModels[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateMftStateDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("updateMftStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Since(start))

	return nil
}

func UpdateMftByCheckAll(now time.Time) error {
	// check expire
	curCertIdStateModels, err := getExpireMftDb(now)
	if err != nil {
		belogs.Error("UpdateMftByCheckAll(): getExpireMftDb:  err: ", err)
		return err
	}
	belogs.Info("UpdateMftByCheckAll(): len(curCertIdStateModels):", len(curCertIdStateModels))

	newCertIdStateModels := make([]model.CertIdStateModel, 0)
	for i := range curCertIdStateModels {
		// if have this error, ignore
		belogs.Debug("UpdateMftByCheckAll(): old curCertIdStateModels[i]:", jsonutil.MarshalJson(curCertIdStateModels[i]))
		if strings.Contains(curCertIdStateModels[i].StateStr, "NextUpdate is earlier than the current time") {
			continue
		}

		// will add error
		stateModel := model.StateModel{}
		jsonutil.UnmarshalJson(curCertIdStateModels[i].StateStr, &stateModel)

		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NextUpdate is earlier than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", nextUpdate is " + convert.Time2StringZone(curCertIdStateModels[i].EndTime)}
		if conf.Bool("policy::allowStaleMft") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}

		certIdStateModel := model.CertIdStateModel{
			Id:       curCertIdStateModels[i].Id,
			StateStr: jsonutil.MarshalJson(stateModel),
		}
		newCertIdStateModels = append(newCertIdStateModels, certIdStateModel)
		belogs.Debug("UpdateMftByCheckAll(): new certIdStateModel:", jsonutil.MarshalJson(certIdStateModel))
	}

	// update db
	err = updateMftStateDb(newCertIdStateModels)
	if err != nil {
		belogs.Error("UpdateMftByCheckAll(): updateMftStateDb:  err: ", len(newCertIdStateModels), err)
		return err
	}
	belogs.Info("UpdateMftByCheckAll(): ok len(newCertIdStateModels):", len(newCertIdStateModels))
	return nil

}
