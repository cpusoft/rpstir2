package parsevalidatedb

import (
	"errors"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

func ReceiveDistributedCountResult(distributedResult *model.DistributedResult) (err error) {
	start := time.Now()
	belogs.Debug("ReceiveDistributedCountResult(): distributedResult:", distributedResult.String())
	resultType := distributedResult.ResultType
	var errMsg string
	if resultType == model.DISTRIBUTED_RESULT_TYPE_SNAPSHOTCOUNT {
		errMsg = distributedResult.DistributedSnapshotCountResult.ErrMsg
	} else if resultType == model.DISTRIBUTED_RESULT_TYPE_DELTACOUNT {
		errMsg = distributedResult.DistributedDeltaCountResult.ErrMsg
	} else {
		belogs.Error("ReceiveDistributedCountResult(): resultType is unsupported, distributedResult:",
			distributedResult.String(), err, "  time(s):", time.Since(start))
		return errors.New("resultType is unsupported")
	}
	belogs.Debug("ReceiveDistributedCountResult(): resultType:", resultType, "   errMsg:", errMsg,
		" distributedResult:", distributedResult.String())

	belogs.Debug("ReceiveDistributedCountResult(): resultType is snapshotcout or deltacount, resultType:", resultType,
		"   errMsg:", errMsg, "   distributedResult:", distributedResult.String())
	err = saveToSyncRrdpLogDb(distributedResult)
	if err != nil {
		belogs.Error("ReceiveDistributedCountResult(): saveToSyncRrdpLogDb fail, resultType is snapshotcout or deltacount, ",
			"  distributedResult:", distributedResult.String(), err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Debug("ReceiveDistributedCountResult(): saveToSyncRrdpLogDb, resultType is snapshotcout or deltacount,",
		"  distributedResult:", distributedResult.String(), "  time(s):", time.Since(start))

	err = saveToSyncRrdpNotifyDb(distributedResult)
	if err != nil {
		belogs.Error("ReceiveDistributedCountResult(): saveToSyncRrdpNotifyDb fail, resultType is snapshotcout or deltacount, ",
			"  distributedResult:", distributedResult.String(), err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Info("ReceiveDistributedCountResult(): ok, resultType is snapshotcout or deltacount,",
		"  distributedResult:", distributedResult.String(), "  time(s):", time.Since(start))
	return nil
}

func ReceiveDistributedResults(distributedResults []model.DistributedResult) (err error) {
	start := time.Now()
	belogs.Debug("ReceiveDistributedResults(): len(distributedResults):", len(distributedResults))

	for i := range distributedResults {
		startOne := time.Now()
		distributedResult := distributedResults[i]
		err = ReceiveDistributedResult(&distributedResult)
		if err != nil {
			belogs.Error("ReceiveDistributedResults(): ReceiveDistributedResult fail, ",
				"  distributedResult:", distributedResult.String(), err, "  time(s):", time.Since(start))
			continue
		}
		belogs.Debug("ReceiveDistributedResults(): ReceiveDistributedResult ok, ",
			"  distributedResult:", distributedResult.String(), "  time(s):", time.Since(startOne))
	}

	belogs.Info("ReceiveDistributedResults(): ok, len(distributedResults):", len(distributedResults), "  time(s):", time.Since(start))
	return nil
}

func ReceiveDistributedResult(distributedResult *model.DistributedResult) (err error) {
	start := time.Now()
	belogs.Debug("ReceiveDistributedResults(): distributedResult:", distributedResult.String())
	syncLogFileId, err := saveToSyncLogFileDb(distributedResult)
	if err != nil {
		belogs.Error("ReceiveDistributedResults(): saveToSyncLogFileDb fail, resultType is publish or withdraw,",
			"  distributedResult:", distributedResult.String(), err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Debug("ReceiveDistributedResults(): saveToSyncLogFileDb, resultType is publish or withdraw,",
		"  distributedResult:", distributedResult.String(), "  time(s):", time.Since(start))

	err = saveToCertDb(syncLogFileId, distributedResult)
	if err != nil {
		belogs.Error("ReceiveDistributedResults(): saveToCertDb fail, resultType is publish or withdraw,",
			"  distributedResult:", distributedResult.String(), err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Debug("ReceiveDistributedResults(): saveToCertDb, resultType is publish or withdraw, ",
		"  distributedResult:", distributedResult.String(), "  time(s):", time.Since(start))
	return nil
}

func saveToCertDb(syncLogFileId uint64, distributedResult *model.DistributedResult) error {
	start := time.Now()
	belogs.Debug("saveToCertDb(): syncLogFileId:", syncLogFileId, "  distributedResult:", distributedResult.String())

	syncLogFileModel := new(model.SyncLogFileModel)
	syncLogFileModel.Id = syncLogFileId
	filePathName := distributedResult.DistributedPublishWithdrawResult.CenterFilePathName
	syncLogFileModel.FilePath, syncLogFileModel.FileName = osutil.Split(filePathName)
	syncLogFileModel.FileType = osutil.ExtNoDot(syncLogFileModel.FileName)
	syncLogFileModel.SyncLogId = distributedResult.DistributedPublishWithdrawResult.SyncLogId
	syncLogFileModel.CertModel = distributedResult.DistributedPublishWithdrawResult.CertModel
	syncLogFileModel.StateModel = distributedResult.DistributedPublishWithdrawResult.StateModel
	syncLogFileModel.OriginModel = distributedResult.DistributedPublishWithdrawResult.OriginModel
	if distributedResult.DistributedPublishWithdrawResult.IsPublish {
		syncLogFileModel.SyncType = "add"
	} else {
		syncLogFileModel.SyncType = "del"
	}
	selectForUpdateWaitSec := conf.String("parse::selectForUpdateWaitSec")
	belogs.Info("saveToCertDb(): syncLogFileModel:", syncLogFileModel.String(), "  selectForUpdateWaitSec:", selectForUpdateWaitSec)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("saveToCertDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	switch syncLogFileModel.FileType {
	case "cer":
		belogs.Debug("saveToCertDb(): fileType is cer, syncLogFileModel:", syncLogFileModel.String())
		certIds, err := selectByFilePathNameDbWithSession(session, "lab_rpki_cer", syncLogFileModel.FilePath, syncLogFileModel.FileName, selectForUpdateWaitSec)
		if err != nil {
			belogs.Error("saveToCertDb(): selectByFilePathNameDbWithSession lab_rpki_cer fail, syncLogFileModel:", syncLogFileModel.String(), err)
			return xormdb.RollbackAndLogError(session, "saveToCertDb(): selectByFilePathNameDbWithSession lab_rpki_cer fail: "+syncLogFileModel.String(), err)
		}
		belogs.Info("saveToCertDb(): selectByFilePathNameDbWithSession from lab_rpki_cer, certIds:", certIds,
			" syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		if len(certIds) > 0 {
			for _, certId := range certIds {
				everyDelTime := time.Now()
				syncLogFileModel.CertId = uint64(certId)
				err = DelCerDbWithSession(session, syncLogFileModel)
				if err != nil {
					belogs.Error("saveToCertDb(): DelCerDbWithSession fail,syncLogFileModel:", syncLogFileModel.String(), err)
					return xormdb.RollbackAndLogError(session, "saveToCertDb(): DelCerDbWithSession fail: "+syncLogFileModel.String(), err)
				}
				belogs.Debug("saveToCertDb(): DelCerDbWithSession, syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(everyDelTime))
			}
			belogs.Info("saveToCertDb(): DelCerDbWithSession, certIds:", certIds, "  syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		}

		if syncLogFileModel.SyncType == "add" || syncLogFileModel.SyncType == "update" {
			err = InsertCerDbWithSession(session, syncLogFileModel, start)
			if err != nil {
				belogs.Error("saveToCertDb(): InsertCerDbWithSession fail,syncLogFileModel:", syncLogFileModel.String(), err)
				return xormdb.RollbackAndLogError(session, "saveToCertDb(): InsertCerDbWithSession fail: "+syncLogFileModel.String(), err)
			}
			belogs.Info("saveToCertDb(): InsertCerDbWithSession, syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		}
	case "crl":
		belogs.Debug("saveToCertDb(): fileType is crl, syncLogFileModel:", syncLogFileModel.String())
		certIds, err := selectByFilePathNameDbWithSession(session, "lab_rpki_crl", syncLogFileModel.FilePath, syncLogFileModel.FileName, selectForUpdateWaitSec)
		if err != nil {
			belogs.Error("saveToCertDb(): selectByFilePathNameDbWithSession lab_rpki_crl fail,syncLogFileModel:", syncLogFileModel.String(), err)
			return xormdb.RollbackAndLogError(session, "saveToCertDb(): selectByFilePathNameDbWithSession lab_rpki_crl fail: "+syncLogFileModel.String(), err)
		}
		belogs.Info("saveToCertDb(): selectByFilePathNameDbWithSession from lab_rpki_crl, certIds:", certIds,
			" syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		if len(certIds) > 0 {
			for _, certId := range certIds {
				everyDelTime := time.Now()
				syncLogFileModel.CertId = uint64(certId)
				err = DelCrlDbWithSession(session, syncLogFileModel)
				if err != nil {
					belogs.Error("saveToCertDb(): DelCrlDbWithSession fail,syncLogFileModel:", syncLogFileModel.String(), err)
					return xormdb.RollbackAndLogError(session, "saveToCertDb(): DelCrlDbWithSession fail: "+syncLogFileModel.String(), err)
				}
				belogs.Debug("saveToCertDb(): DelCrlDbWithSession, certId:", certId, "  syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(everyDelTime))
			}
			belogs.Info("saveToCertDb(): DelCrlDbWithSession, certIds:", certIds, "  syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		}
		if syncLogFileModel.SyncType == "add" || syncLogFileModel.SyncType == "update" {
			err = InsertCrlDbWithSession(session, syncLogFileModel, start)
			if err != nil {
				belogs.Error("saveToCertDb(): InsertCrlDbWithSession fail,syncLogFileModel:", syncLogFileModel.String(), err)
				return xormdb.RollbackAndLogError(session, "saveToCertDb(): InsertCrlDbWithSession fail: "+syncLogFileModel.String(), err)
			}
			belogs.Info("saveToCertDb(): InsertCrlDbWithSession, syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		}
	case "roa":
		belogs.Debug("saveToCertDb(): fileType is roa, syncLogFileModel:", syncLogFileModel.String())
		certIds, err := selectByFilePathNameDbWithSession(session, "lab_rpki_roa", syncLogFileModel.FilePath, syncLogFileModel.FileName, selectForUpdateWaitSec)
		if err != nil {
			belogs.Error("saveToCertDb(): selectByFilePathNameDbWithSession lab_rpki_roa fail,syncLogFileModel:", syncLogFileModel.String(), err)
			return xormdb.RollbackAndLogError(session, "saveToCertDb(): selectByFilePathNameDbWithSession lab_rpki_roa fail: "+syncLogFileModel.String(), err)
		}
		belogs.Info("saveToCertDb(): selectByFilePathNameDbWithSession from lab_rpki_roa, certIds:", certIds,
			" syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		if len(certIds) > 0 {
			for _, certId := range certIds {
				everyDelTime := time.Now()
				syncLogFileModel.CertId = uint64(certId)
				err = DelRoaDbWithSession(session, syncLogFileModel)
				if err != nil {
					belogs.Error("saveToCertDb(): DelRoaDbWithSession fail,syncLogFileModel:", syncLogFileModel.String(), err)
					return xormdb.RollbackAndLogError(session, "saveToCertDb(): DelRoaDbWithSession fail: "+syncLogFileModel.String(), err)
				}
				belogs.Debug("saveToCertDb(): DelRoaDbWithSession, certId:", certId, "  syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(everyDelTime))
			}
			belogs.Info("saveToCertDb(): DelRoaDbWithSession, certIds:", certIds, "  syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		}
		if syncLogFileModel.SyncType == "add" || syncLogFileModel.SyncType == "update" {
			err = InsertRoaDbWithSession(session, syncLogFileModel, start)
			if err != nil {
				belogs.Error("saveToCertDb(): InsertRoaDbWithSession fail,syncLogFileModel:", syncLogFileModel.String(), err)
				return xormdb.RollbackAndLogError(session, "saveToCertDb(): InsertRoaDbWithSession fail: "+syncLogFileModel.String(), err)
			}
			belogs.Info("saveToCertDb(): InsertRoaDbWithSession, syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		}
	case "mft":
		belogs.Debug("saveToCertDb(): fileType is mft, syncLogFileModel:", syncLogFileModel.String())
		certIds, err := selectByFilePathNameDbWithSession(session, "lab_rpki_mft", syncLogFileModel.FilePath, syncLogFileModel.FileName, selectForUpdateWaitSec)
		if err != nil {
			belogs.Error("saveToCertDb(): selectByFilePathNameDbWithSession lab_rpki_mft fail,syncLogFileModel:", syncLogFileModel.String(), err)
			return xormdb.RollbackAndLogError(session, "saveToCertDb(): selectByFilePathNameDbWithSession lab_rpki_mft fail: "+syncLogFileModel.String(), err)
		}
		belogs.Info("saveToCertDb(): selectByFilePathNameDbWithSession from lab_rpki_mft, certIds:", certIds,
			" syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		if len(certIds) > 0 {
			for _, certId := range certIds {
				everyDelTime := time.Now()
				syncLogFileModel.CertId = uint64(certId)
				err = DelMftDbWithSession(session, syncLogFileModel)
				if err != nil {
					belogs.Error("saveToCertDb(): DelMftDbWithSession fail,syncLogFileModel:", syncLogFileModel.String(), err)
					return xormdb.RollbackAndLogError(session, "saveToCertDb(): DelMftDbWithSession fail: "+syncLogFileModel.String(), err)
				}
				belogs.Debug("saveToCertDb(): DelMftDbWithSession, syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(everyDelTime))
			}
			belogs.Info("saveToCertDb(): DelMftDbWithSession, certIds:", certIds, "  syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		}
		if syncLogFileModel.SyncType == "add" || syncLogFileModel.SyncType == "update" {
			err = InsertMftDbWithSession(session, syncLogFileModel, start)
			if err != nil {
				belogs.Error("saveToCertDb(): InsertMftDbWithSession fail,syncLogFileModel:", syncLogFileModel.String(), err)
				return xormdb.RollbackAndLogError(session, "saveToCertDb(): InsertMftDbWithSession fail: "+syncLogFileModel.String(), err)
			}
			belogs.Info("saveToCertDb(): InsertMftDbWithSession, syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		}
	case "asa":
		belogs.Debug("saveToCertDb(): fileType is asa, syncLogFileModel:", syncLogFileModel.String())
		certIds, err := selectByFilePathNameDbWithSession(session, "lab_rpki_asa", syncLogFileModel.FilePath, syncLogFileModel.FileName, selectForUpdateWaitSec)
		if err != nil {
			belogs.Error("saveToCertDb(): selectByFilePathNameDbWithSession lab_rpki_asa fail,syncLogFileModel:", syncLogFileModel.String(), err)
			return xormdb.RollbackAndLogError(session, "saveToCertDb(): selectByFilePathNameDbWithSession lab_rpki_asa fail: "+syncLogFileModel.String(), err)
		}
		belogs.Info("saveToCertDb(): selectByFilePathNameDbWithSession from lab_rpki_asa, certIds:", certIds,
			" syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		if len(certIds) > 0 {
			for _, certId := range certIds {
				everyDelTime := time.Now()
				syncLogFileModel.CertId = uint64(certId)
				err = DelAsaDbWithSession(session, syncLogFileModel)
				if err != nil {
					belogs.Error("saveToCertDb(): DelAsaDbWithSession fail,syncLogFileModel:", syncLogFileModel.String(), err)
					return xormdb.RollbackAndLogError(session, "saveToCertDb(): DelAsaDbWithSession fail: "+syncLogFileModel.String(), err)
				}
				belogs.Debug("saveToCertDb(): DelAsaDbWithSession, syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(everyDelTime))
			}
			belogs.Info("saveToCertDb(): DelAsaDbWithSession, certIds:", certIds, "  syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		}
		if syncLogFileModel.SyncType == "add" || syncLogFileModel.SyncType == "update" {
			err = InsertAsaDbWithSession(session, syncLogFileModel, start)
			if err != nil {
				belogs.Error("saveToCertDb(): InsertAsaDbWithSession fail,syncLogFileModel:", syncLogFileModel.String(), err)
				return xormdb.RollbackAndLogError(session, "saveToCertDb(): InsertAsaDbWithSession fail: "+syncLogFileModel.String(), err)
			}
			belogs.Info("saveToCertDb(): InsertAsaDbWithSession, syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("saveToCertDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("saveToCertDb(): ok, syncLogFileModel:", syncLogFileModel.String(), " time(s):", time.Since(start))
	return nil
}

func getCertIdByFilePathNameWithSession(session *xorm.Session, tableName, filePath, fileName string) (uint64, error) {
	belogs.Debug("getCertIdByFilePathNameWithSession(): tableName:", tableName, "  filePath:", filePath, "  fileName:", fileName)
	var id int
	sql := `select id from ` + tableName + ` where filepath=? and filename=?`
	has, err := session.SQL(sql, filePath, fileName).Get(&id)
	if err != nil {
		belogs.Error("getCertIdByFilePathNameWithSession(): get id failed, tableName:", tableName, " filePath:", filePath, "  fileName:", fileName, "    err:", err)
		return 0, err
	}
	if !has {
		belogs.Debug("getCertIdByFilePathNameWithSession(): not found from tableName:", tableName, " filePath:", filePath, "  fileName:", fileName)
		return 0, nil

	}
	belogs.Debug("getCertIdByFilePathNameWithSession(): get CertId, filePath:", filePath, "  fileName:", fileName, " id:", id)
	return uint64(id), nil
}

func selectByFilePathNameDbWithSession(session *xorm.Session, tableName, filePath, fileName, selectForUpdateWaitSec string) ([]int, error) {
	belogs.Debug("selectByFilePathNameDbWithSession(): tableName:", tableName, "  filePath:", filePath,
		"  fileName:", fileName, "  selectForUpdateWaitSec:", selectForUpdateWaitSec)

	start := time.Now()

	ids := make([]int, 0)
	sql := `select id from ` + tableName + ` where filepath=? and filename=? `
	belogs.Debug("selectByFilePathNameDbWithSession():select id from tableName, sql:", sql)
	err := session.SQL(sql, filePath, fileName).Find(&ids)
	if err != nil {
		belogs.Error("selectByFilePathNameDbWithSession(): select id failed, tableName:", tableName, " filePath:", filePath, "  fileName:", fileName,
			"    err:", err, " time(s):", time.Since(start))
		return nil, err
	}
	belogs.Debug("selectByFilePathNameDbWithSession(): pass select id, filePath:", filePath, "  fileName:", fileName,
		" ids:", ids, " time(s):", time.Since(start))
	return ids, nil
}
