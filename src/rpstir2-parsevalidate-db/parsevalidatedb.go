package parsevalidatedb

import (
	"errors"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
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
	belogs.Debug("ReceiveDistributedCountResult(): saveToSyncRrdpNotifyDb, resultType is snapshotcout or deltacount,",
		"  distributedResult:", distributedResult.String(), "  time(s):", time.Since(start))
	return nil
}

func ReceiveDistributedPublishWithdrawResults(distributedResults []model.DistributedResult) (err error) {
	start := time.Now()
	belogs.Debug("ReceiveDistributedPublishWithdrawResults(): len(distributedResults):", len(distributedResults))

	for i := range distributedResults {
		distributedResult := distributedResults[i]
		syncLogFileId, err := saveToSyncLogFileDb(&distributedResult)
		if err != nil {
			belogs.Error("ReceiveDistributedPublishWithdrawResults(): saveToSyncLogFileDb fail, resultType is publish or withdraw,",
				"  distributedResult:", distributedResult.String(), err, "  time(s):", time.Since(start))
			continue
		}
		belogs.Debug("ReceiveDistributedPublishWithdrawResults(): saveToSyncLogFileDb, resultType is publish or withdraw,",
			"  distributedResult:", distributedResult.String(), "  time(s):", time.Since(start))

		err = saveToCertDb(syncLogFileId, &distributedResult)
		if err != nil {
			belogs.Error("ReceiveDistributedPublishWithdrawResults(): saveToCertDb fail, resultType is publish or withdraw,",
				"  distributedResult:", distributedResult.String(), err, "  time(s):", time.Since(start))
			continue
		}
		belogs.Debug("ReceiveDistributedPublishWithdrawResults(): saveToCertDb, resultType is publish or withdraw, ",
			"  distributedResult:", distributedResult.String(), "  time(s):", time.Since(start))
	}

	belogs.Debug("ReceiveDistributedPublishWithdrawResults():ok, distributedResults:", len(distributedResults), "  time(s):", time.Since(start))
	return nil
}

func saveToCertDb(syncLogFileId uint64, distributedResult *model.DistributedResult) error {
	start := time.Now()
	belogs.Debug("saveToCertDb(): syncLogFileId:", syncLogFileId, "  distributedResult:", distributedResult.String())
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("saveToCertDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	syncLogFileModel := new(model.SyncLogFileModel)
	syncLogFileModel.Id = syncLogFileId
	filePathName := distributedResult.DistributedPublishWithdrawResult.FilePathName
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
	belogs.Debug("saveToCertDb(): syncLogFileModel:", syncLogFileModel.String(), "  selectForUpdateWaitSec:", selectForUpdateWaitSec)

	switch syncLogFileModel.FileType {
	case "cer":
		belogs.Debug("saveToCertDb(): fileType is cer, syncLogFileModel:", syncLogFileModel.String())
		// select * from cer for update wait 10
		err = selectByFilePathNameForUpdate("lab_rpki_cer", syncLogFileModel.FilePath, syncLogFileModel.FileName, selectForUpdateWaitSec)
		if err != nil {
			belogs.Debug("saveToCertDb(): selectByFilePathNameForUpdate lab_rpki_cer fail,syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}

		err = DelCerDb(syncLogFileModel)
		if err != nil {
			belogs.Debug("saveToCertDb(): DelCerDb fail,syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}
		if syncLogFileModel.SyncType == "add" || syncLogFileModel.SyncType == "update" {
			err = AddCerDb(syncLogFileModel)
			if err != nil {
				belogs.Debug("saveToCertDb(): AddCerDb fail,syncLogFileModel:", syncLogFileModel.String(), err)
				return err
			}
		}
	case "crl":
		belogs.Debug("saveToCertDb(): fileType is crl, syncLogFileModel:", syncLogFileModel.String())
		err = selectByFilePathNameForUpdate("lab_rpki_crl", syncLogFileModel.FilePath, syncLogFileModel.FileName, selectForUpdateWaitSec)
		if err != nil {
			belogs.Debug("saveToCertDb(): selectByFilePathNameForUpdate lab_rpki_crl fail,syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}

		err = DelCrlDb(syncLogFileModel)
		if err != nil {
			belogs.Debug("saveToCertDb(): DelCrlDb fail,syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}
		if syncLogFileModel.SyncType == "add" || syncLogFileModel.SyncType == "update" {
			err = AddCrlDb(syncLogFileModel)
			if err != nil {
				belogs.Debug("saveToCertDb(): AddCrlDb fail,syncLogFileModel:", syncLogFileModel.String(), err)
				return err
			}
		}
	case "roa":
		belogs.Debug("saveToCertDb(): fileType is roa, syncLogFileModel:", syncLogFileModel.String())
		err = selectByFilePathNameForUpdate("lab_rpki_roa", syncLogFileModel.FilePath, syncLogFileModel.FileName, selectForUpdateWaitSec)
		if err != nil {
			belogs.Debug("saveToCertDb(): selectByFilePathNameForUpdate lab_rpki_roa fail,syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}

		err = DelRoaDb(syncLogFileModel)
		if err != nil {
			belogs.Debug("saveToCertDb(): DelRoaDb fail,syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}
		if syncLogFileModel.SyncType == "add" || syncLogFileModel.SyncType == "update" {
			err = AddRoaDb(syncLogFileModel)
			if err != nil {
				belogs.Debug("saveToCertDb(): AddRoaDb fail,syncLogFileModel:", syncLogFileModel.String(), err)
				return err
			}
		}
	case "mft":
		belogs.Debug("saveToCertDb(): fileType is mft, syncLogFileModel:", syncLogFileModel.String())
		err = selectByFilePathNameForUpdate("lab_rpki_mft", syncLogFileModel.FilePath, syncLogFileModel.FileName, selectForUpdateWaitSec)
		if err != nil {
			belogs.Debug("saveToCertDb(): selectByFilePathNameForUpdate lab_rpki_mft fail,syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}

		err = DelMftDb(syncLogFileModel)
		if err != nil {
			belogs.Debug("saveToCertDb(): DelMftDb fail,syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}
		if syncLogFileModel.SyncType == "add" || syncLogFileModel.SyncType == "update" {
			err = AddMftDb(syncLogFileModel)
			if err != nil {
				belogs.Debug("saveToCertDb(): AddMftDb fail,syncLogFileModel:", syncLogFileModel.String(), err)
				return err
			}
		}
	case "asa":
		belogs.Debug("saveToCertDb(): fileType is asa, syncLogFileModel:", syncLogFileModel.String())
		err = selectByFilePathNameForUpdate("lab_rpki_asa", syncLogFileModel.FilePath, syncLogFileModel.FileName, selectForUpdateWaitSec)
		if err != nil {
			belogs.Debug("saveToCertDb(): selectByFilePathNameForUpdate lab_rpki_asa fail,syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}

		err = DelAsaDb(syncLogFileModel)
		if err != nil {
			belogs.Debug("saveToCertDb(): DelAsaDb fail,syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}
		if syncLogFileModel.SyncType == "add" || syncLogFileModel.SyncType == "update" {
			err = AddAsaDb(syncLogFileModel)
			if err != nil {
				belogs.Debug("saveToCertDb(): AddAsaDb fail,syncLogFileModel:", syncLogFileModel.String(), err)
				return err
			}
		}
	}
	belogs.Debug("saveToCertDb(): ok, syncLogFileModel:", syncLogFileModel.String(), " time(s):", time.Since(start))
	return nil
}

func getCertIdByFilePathName(session *xorm.Session, tableName, filePath, fileName string) (uint64, error) {
	belogs.Debug("getCertIdByFilePathName(): tableName:", tableName, "  filePath:", filePath, "  fileName:", fileName)
	var id int
	sql := `select id from ` + tableName + ` where filepath=? and filename=?`
	has, err := session.SQL(sql, filePath, fileName).Get(&id)
	if err != nil {
		belogs.Error("getCertIdByFilePathName(): get id failed, tableName:", tableName, " filePath:", filePath, "  fileName:", fileName, "    err:", err)
		return 0, err
	}
	if !has {
		belogs.Debug("getCertIdByFilePathName(): not found from tableName:", tableName, " filePath:", filePath, "  fileName:", fileName)
		return 0, nil

	}
	belogs.Debug("getCertIdByFilePathName(): get CertId, filePath:", filePath, "  fileName:", fileName, " id:", id)
	return uint64(id), nil
}

func selectByFilePathNameForUpdate(tableName, filePath, fileName, selectForUpdateWaitSec string) error {
	belogs.Debug("selectByFilePathNameForUpdate(): tableName:", tableName, "  filePath:", filePath,
		"  fileName:", fileName, "  selectForUpdateWaitSec:", selectForUpdateWaitSec)
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("selectByFilePathNameForUpdate(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	var id int
	start := time.Now()
	sql := `select id from ` + tableName + ` where filepath=? and filename=? for update wait ` + selectForUpdateWaitSec
	_, err = session.SQL(sql, filePath, fileName).Get(&id)
	if err != nil {
		belogs.Error("selectByFilePathNameForUpdate(): selectforupdate failed, tableName:", tableName, " filePath:", filePath, "  fileName:", fileName,
			"    err:", err, " time(s):", time.Since(start))
		return err
	}
	belogs.Debug("selectByFilePathNameForUpdate(): pass selectforupdate, filePath:", filePath, "  fileName:", fileName,
		" id:", id, " time(s):", time.Since(start))
	return nil
}
