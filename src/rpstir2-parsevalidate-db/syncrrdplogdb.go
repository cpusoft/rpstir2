package parsevalidatedb

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
)

func saveToSyncRrdpLogDb(distributedResult *model.DistributedResult) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("saveToSyncRrdpLogDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	resultType := distributedResult.ResultType
	belogs.Debug("saveToSyncRrdpLogDb(): resultType:", resultType)
	var syncLogId, lastSerial, curSerial uint64
	var notifyUrl, sessionId, rrdpType, snapshotOrDeltaUrl string
	var rrdpTime time.Time
	if resultType == model.DISTRIBUTED_RESULT_TYPE_SNAPSHOTCOUNT {
		syncLogId = distributedResult.DistributedSnapshotCountResult.SyncLogId
		lastSerial = 0
		curSerial = distributedResult.DistributedSnapshotCountResult.Serial
		notifyUrl = distributedResult.DistributedSnapshotCountResult.NotifyUrl
		sessionId = distributedResult.DistributedSnapshotCountResult.SessionId
		rrdpType = "snapshot"
		snapshotOrDeltaUrl = distributedResult.DistributedSnapshotCountResult.SnapshotUrl
		rrdpTime = distributedResult.DistributedSnapshotCountResult.SyncTime
	} else if resultType == model.DISTRIBUTED_RESULT_TYPE_DELTACOUNT {
		syncLogId = distributedResult.DistributedDeltaCountResult.SyncLogId
		lastSerial = 0
		curSerial = distributedResult.DistributedDeltaCountResult.Serial
		notifyUrl = distributedResult.DistributedDeltaCountResult.NotifyUrl
		sessionId = distributedResult.DistributedDeltaCountResult.SessionId
		rrdpType = "delta"
		snapshotOrDeltaUrl = distributedResult.DistributedDeltaCountResult.DeltaUrl
		rrdpTime = distributedResult.DistributedDeltaCountResult.SyncTime
	}
	belogs.Debug("saveToSyncRrdpLogDb(): syncLogId:", syncLogId, "  curSerial:", curSerial, "  notifyUrl:", notifyUrl,
		"   sessionId:", sessionId, "  rrdpType:", rrdpType, "  snapshotOrDeltaUrl:", snapshotOrDeltaUrl, "   rrdpTime:", rrdpTime)
	sqlStr := `INSERT lab_rpki_sync_rrdp_log(syncLogId,  notifyUrl,  sessionId,  
		lastSerial,	  curSerial,  
		rrdpTime,  rrdpType, snapshotOrDeltaUrl)
		VALUES(?,?,?,   ?,?,    ?,?,?)`
	_, err = session.Exec(sqlStr, syncLogId, notifyUrl, sessionId,
		xormdb.SqlNullInt(int64(lastSerial)), curSerial,
		rrdpTime, rrdpType, snapshotOrDeltaUrl)
	if err != nil {
		belogs.Error("saveToSyncRrdpLogDb(): INSERT lab_rpki_sync_rrdp_log fail, syncLogId:",
			syncLogId, "  distributedResult:", distributedResult.String(), err)
		return xormdb.RollbackAndLogError(session, "saveToSyncRrdpLogDb(): INSERT lab_rpki_sync_rrdp_log fail: "+
			snapshotOrDeltaUrl, err)
	}
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("saveToSyncRrdpLogDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("saveToSyncRrdpLogDb(): ok ,syncLogId:", syncLogId, "  curSerial:", curSerial, "  notifyUrl:", notifyUrl,
		"   sessionId:", sessionId, "  rrdpType:", rrdpType, "  snapshotOrDeltaUrl:", snapshotOrDeltaUrl, "  time(s):", time.Since(start))
	return nil
}
