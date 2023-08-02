package parsevalidatedb

import (
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
	"xorm.io/xorm"
)

// need endCh, when error
func GetSyncLogFileModelBySyncLogIdDb(labRpkiSyncLogId uint64, syncLogFileModelCh chan *model.SyncLogFileModel,
	parseConcurrentCh chan int, endCh chan bool, wg *sync.WaitGroup) (err error) {
	start := time.Now()

	belogs.Debug("GetSyncLogFileModelBySyncLogIdDb(): labRpkiSyncLogId:", labRpkiSyncLogId)
	syncLogFileModel := new(model.SyncLogFileModel)
	sql := `select s.id,s.syncLogId,s.filePath,s.fileName, s.fileType, s.syncType, 
				cast(CONCAT(IFNULL(c.id,''),IFNULL(m.id,''),IFNULL(l.id,''),IFNULL(r.id,''),IFNULL(a.id,'')) as unsigned int) as certId from lab_rpki_sync_log_file s 
			left join lab_rpki_cer c on c.filePath = s.filePath and c.fileName = s.fileName  
			left join lab_rpki_mft m on m.filePath = s.filePath and m.fileName = s.fileName  
			left join lab_rpki_crl l on l.filePath = s.filePath and l.fileName = s.fileName  
			left join lab_rpki_roa r on r.filePath = s.filePath and r.fileName = s.fileName 
			left join lab_rpki_asa a on a.filePath = s.filePath and a.fileName = s.fileName 
			where s.state->>'$.updateCertTable'='notYet' and s.syncLogId=? order by s.id `
	rows, err := xormdb.XormEngine.SQL(sql, labRpkiSyncLogId).Rows(syncLogFileModel)
	if err != nil {
		belogs.Error("GetSyncLogFileModelBySyncLogIdDb(): select from rpki_*** fail:", err)
		return err
	}
	belogs.Debug("GetSyncLogFileModelBySyncLogIdDb(): will call rows.Next(), time(s):", time.Since(start))

	defer rows.Close()
	var index uint64
	for rows.Next() {
		// control parse speed
		parseConcurrentCh <- 1
		// get new *syncLogFileModel every Scan
		syncLogFileModel = new(model.SyncLogFileModel)
		err = rows.Scan(syncLogFileModel)
		if err != nil {
			belogs.Error("GetSyncLogFileModelBySyncLogIdDb(): Scan fail:", err)
			continue
		}
		syncLogFileModel.Index = index
		belogs.Debug("GetSyncLogFileModelBySyncLogIdDb(): Scan, wg.Add() id:", syncLogFileModel.Id, " index:", index,
			"  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
			"  , time(s):", time.Since(start))

		wg.Add(1)
		syncLogFileModelCh <- syncLogFileModel
		index++
	}

	belogs.Info("GetSyncLogFileModelBySyncLogIdDb(): get all syncLogFileModel,labRpkiSyncLogId:", labRpkiSyncLogId, "   count:", index, "  time(s):", time.Since(start))
	return nil
}

func updateSyncLogFileJsonAllAndStateDb(session *xorm.Session, syncLogFileModel *model.SyncLogFileModel) error {
	belogs.Debug("updateSyncLogFileJsonAllAndStateDb(): syncLogFileModel:", syncLogFileModel.String())
	sqlStr := `update lab_rpki_sync_log_file f set 	
	  f.state=json_replace(f.state,'$.updateCertTable','finished','$.rtr',?) ,
	  f.jsonAll=?  where f.id=?`
	rtrState := "notNeed"
	jsonAll := ""
	if (syncLogFileModel.FileType == "roa" || syncLogFileModel.FileType == "asa") &&
		syncLogFileModel.SyncType != "del" {
		rtrState = "notYet"
	}

	//when del or update(before del), syncLogFileModel.CertModel is nil
	if syncLogFileModel.CertModel == nil {
		belogs.Debug("updateSyncLogFileJsonAllAndStateDb(): del or update, CertModel is nil, syncLogFileModel:",
			syncLogFileModel.String())
	} else {
		// when add or update(after del), syncLogFileModel.CertModel is not nil
		jsonAll = jsonutil.MarshalJson(syncLogFileModel.CertModel)
	}
	belogs.Debug("updateSyncLogFileJsonAllAndStateDb(): id:", syncLogFileModel.Id,
		"  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  len(jsonAll):", len(jsonAll))

	_, err := session.Exec(sqlStr, rtrState, xormdb.SqlNullString(jsonAll), syncLogFileModel.Id)
	if err != nil {
		belogs.Error("updateSyncLogFileJsonAllAndStateDb(): updateSyncLogFileJsonAllAndState fail:",
			"   id:", syncLogFileModel.Id,
			"   file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
			"   rtrState:", rtrState, "  jsonAll:", jsonAll, err)
		return err
	}
	belogs.Debug("updateSyncLogFileJsonAllAndStateDb(): update lab_rpki_sync_log_file, id:", syncLogFileModel.Id,
		"   file:", syncLogFileModel.FilePath, syncLogFileModel.FileName)
	return nil
}

func saveToSyncLogFileDb(distributedResult *model.DistributedResult) (syncLogFileId uint64, err error) {

	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("saveToSyncLogFileDb(): NewSession fail:", err)
		return 0, err
	}
	defer session.Close()

	syncLogId := distributedResult.DistributedPublishWithdrawResult.SyncLogId
	filePathName := distributedResult.DistributedPublishWithdrawResult.FilePathName
	filePath, fileName := osutil.Split(filePathName)
	fileType := osutil.ExtNoDot(fileName)
	sourceUrl := distributedResult.DistributedPublishWithdrawResult.SnapshotOrDeltaUrl
	syncTime := distributedResult.DistributedPublishWithdrawResult.SyncTime
	syncStyle := "rrdp"
	fileHash := distributedResult.DistributedPublishWithdrawResult.FileHash
	isSnapshot := distributedResult.DistributedPublishWithdrawResult.IsSnapshot
	isPublish := distributedResult.DistributedPublishWithdrawResult.IsPublish
	var syncType, jsonAll string
	if isPublish {
		syncType = "add"
		jsonAll = jsonutil.MarshalJson(distributedResult.DistributedPublishWithdrawResult.CertModel)
	} else {
		syncType = "del"
		jsonAll = "" // when withdraw, jsonAll is empty
	}
	rtr := "notNeed"
	if isPublish &&
		(fileType == "roa" || fileType == "asa") {
		rtr = "notYet"
	}
	syncLogFileState := model.LabRpkiSyncLogFileState{
		Sync:            "finished",
		UpdateCertTable: "finished",
		Rtr:             rtr,
	}
	state := jsonutil.MarshalJson(syncLogFileState)
	belogs.Debug("saveToSyncLogFileDb(): syncLogId:", syncLogId, "  filePathName:", filePathName,
		"  isSnapshot:", isSnapshot, " isPublish", isPublish, "  sourceUrl:", sourceUrl,
		"  syncType:", syncType, "  syncStyle:", syncStyle, "  fileHash:", fileHash, "  len(jsonAll):", len(jsonAll))

	//lab_rpki_sync_log_file
	sqlStr := `INSERT lab_rpki_sync_log_file
			   (syncLogId,fileType,syncTime,
			   filePath,fileName,sourceUrl,
			   syncType,syncStyle,state,
			   fileHash,jsonAll)
		 VALUES(?,?,?,
		 ?,?,?,
		 ?,?,?,
		 ?,?)`
	res, err := session.Exec(sqlStr,
		syncLogId, fileType, syncTime,
		filePath, fileName, sourceUrl,
		syncType, syncStyle, state,
		xormdb.SqlNullString(fileHash),
		xormdb.SqlNullString(jsonAll))
	if err != nil {
		belogs.Error("saveToSyncLogFileDb():INSERT lab_rpki_sync_log_file fail, distributedResult:", distributedResult.String(), err)
		return 0, xormdb.RollbackAndLogError(session, "INSERT lab_rpki_sync_log_file fail", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		belogs.Error("saveToSyncLogFileDb(): LastInsertId id fail,distributedResult:", distributedResult.String(), err)
		return 0, err
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session, "saveToSyncLogFileDb(): CommitSession fail:", err)
	}
	belogs.Debug("saveToSyncLogFileDb(): ok, syncLogId:", syncLogId, "  syncLogFileId:", id, "  filePathName:", filePathName, "  sourceUrl:", sourceUrl,
		"  syncType:", syncType, "  syncStyle:", syncStyle, "  fileHash:", fileHash, "  time(s):", time.Since(start))

	return uint64(id), nil
}
