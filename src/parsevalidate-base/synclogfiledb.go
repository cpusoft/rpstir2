package parsevalidatebase

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
)

func getSyncLogFileModelsBySyncLogIdDb(syncLogId uint64) (syncLogFileModels []*model.SyncLogFileModel, err error) {
	start := time.Now()

	belogs.Debug("getSyncLogFileModelsBySyncLogIdDb(): syncLogId:", syncLogId)
	syncLogFileModels = make([]*model.SyncLogFileModel, 0)
	sql := `select s.id,s.syncLogId,s.filePath,s.fileName, s.fileType, s.syncType, 
				cast(CONCAT(IFNULL(c.id,''),IFNULL(m.id,''),IFNULL(l.id,''),IFNULL(r.id,''),IFNULL(a.id,'')) as unsigned int) as certId from lab_rpki_sync_log_file s 
			left join lab_rpki_cer c on c.filePath = s.filePath and c.fileName = s.fileName  
			left join lab_rpki_mft m on m.filePath = s.filePath and m.fileName = s.fileName  
			left join lab_rpki_crl l on l.filePath = s.filePath and l.fileName = s.fileName  
			left join lab_rpki_roa r on r.filePath = s.filePath and r.fileName = s.fileName 
			left join lab_rpki_asa a on a.filePath = s.filePath and a.fileName = s.fileName 
			where s.state->>'$.updateCertTable'='notYet' and s.syncLogId=? order by s.id `
	err = xormdb.XormEngine.SQL(sql, syncLogId).Find(&syncLogFileModels)
	if err != nil {
		belogs.Error("getSyncLogFileModelsBySyncLogIdDb(): Find fail:", err)
		return nil, err
	}
	belogs.Info("getSyncLogFileModelsBySyncLogIdDb(): len(syncLogFileModels):", len(syncLogFileModels), " time(s):", time.Since(start))
	//syncLogFileModels = NewSyncLogFileModels(syncLogId, dbSyncLogFileModels)
	//belogs.Info("getSyncLogFileModelsBySyncLogIdDb(): end, len(dbSyncLogFileModels):", len(dbSyncLogFileModels), " time(s):", time.Since(start))
	return syncLogFileModels, nil
}

/*
func updateSyncLogFilesJsonAllAndStateDb(session *xorm.Session, syncLogFileModels []model.SyncLogFileModel) error {
	start := time.Now()
	belogs.Debug("updateSyncLogFilesJsonAllAndStateDb(): len(syncLogFileModels):", len(syncLogFileModels))
	sqlStr := `update lab_rpki_sync_log_file f set
	  f.state=json_replace(f.state,'$.updateCertTable','finished','$.rtr',?) ,
	  f.jsonAll=?  where f.id=?`
	for i := range syncLogFileModels {
		rtrState := "notNeed"
		jsonAll := ""
		if (syncLogFileModels[i].FileType == "roa" || syncLogFileModels[i].FileType == "asa") &&
			syncLogFileModels[i].SyncType != "del" {
			rtrState = "notYet"
		}

		//when del or update(before del), syncLogFileModels[i].CertModel is nil
		if syncLogFileModels[i].CertModel == nil {
			belogs.Debug("updateSyncLogFilesJsonAllAndStateDb(): del or update, CertModel is nil:",
				jsonutil.MarshalJson(syncLogFileModels[i]))
		} else {
			jsonAll = jsonutil.MarshalJson(syncLogFileModels[i].CertModel)
			belogs.Debug("updateSyncLogFilesJsonAllAndStateDb(): jsonAll:", jsonAll, " syncLogFileModels[i]:", syncLogFileModels[i].String())
		}

		_, err := session.Exec(sqlStr, rtrState, xormdb.SqlNullString(jsonAll), syncLogFileModels[i].Id)
		if err != nil {
			belogs.Error("updateSyncLogFilesJsonAllAndStateDb(): updateSyncLogFileJsonAllAndState fail:",
				jsonutil.MarshalJson(syncLogFileModels[i]),
				"   syncLogFileId:", syncLogFileModels[i].Id, err)
			return err
		}
	}
	belogs.Debug("updateSyncLogFilesJsonAllAndStateDb(): ok len(syncLogFileModels):", len(syncLogFileModels), "  time(s):", time.Since(start))
	return nil
}

func updateSyncLogFileJsonAllAndStateDb(session *xorm.Session, syncLogFileModel *model.SyncLogFileModel) error {
	start := time.Now()
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

	//when del or update(before del), syncLogFileModels[i].CertModel is nil
	if syncLogFileModel.CertModel == nil {
		belogs.Debug("updateSyncLogFileJsonAllAndStateDb(): del or update, CertModel is nil:",
			jsonutil.MarshalJson(syncLogFileModel))
	} else {
		jsonAll = jsonutil.MarshalJson(syncLogFileModel.CertModel)
		belogs.Debug("updateSyncLogFileJsonAllAndStateDb(): jsonAll:", jsonAll, " syncLogFileModel:", syncLogFileModel.String())
	}

	_, err := session.Exec(sqlStr, rtrState, xormdb.SqlNullString(jsonAll), syncLogFileModel.Id)
	if err != nil {
		belogs.Error("updateSyncLogFileJsonAllAndStateDb(): updateSyncLogFileJsonAllAndState fail:",
			jsonutil.MarshalJson(syncLogFileModel),
			"   syncLogFileId:", syncLogFileModel.Id, err)
		return err
	}

	belogs.Debug("updateSyncLogFileJsonAllAndStateDb(): ok syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
	return nil
}
*/
