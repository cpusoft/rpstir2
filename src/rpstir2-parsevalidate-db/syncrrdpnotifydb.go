package parsevalidatedb

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
)

func saveToSyncRrdpNotifyDb(distributedResult *model.DistributedResult) error {
	start := time.Now()
	belogs.Debug("saveToSyncRrdpNotifyDb(): distributedResult:", distributedResult.String())

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("saveToSyncRrdpNotifyDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	resultType := distributedResult.ResultType
	sql := `update lab_rpki_sync_rrdp_notify set curSerial = ? , downloadTime = ? where notifyUrl = ? and ( curSerial is null or curSerial < ? )`
	var curSerial uint64
	var downloadTime time.Time
	var notifyUrl string

	if resultType == model.DISTRIBUTED_RESULT_TYPE_SNAPSHOTCOUNT {
		curSerial = distributedResult.DistributedSnapshotCountResult.Serial
		notifyUrl = distributedResult.DistributedSnapshotCountResult.NotifyUrl
		downloadTime = distributedResult.DistributedSnapshotCountResult.SyncTime
	} else if resultType == model.DISTRIBUTED_RESULT_TYPE_DELTACOUNT {
		curSerial = distributedResult.DistributedDeltaCountResult.Serial
		notifyUrl = distributedResult.DistributedDeltaCountResult.NotifyUrl
		downloadTime = distributedResult.DistributedDeltaCountResult.SyncTime
	}
	belogs.Debug("saveToSyncRrdpNotifyDb(): curSerial:", curSerial, "  notifyUrl:", notifyUrl, "  downloadTime:", downloadTime)
	affected, err := session.Exec(sql, curSerial, downloadTime, notifyUrl, curSerial)
	if err != nil {
		belogs.Error("saveToSyncRrdpNotifyDb(): update lab_rpki_sync_rrdp_notify fail, curSerial:",
			curSerial, "  notifyUrl:", notifyUrl, "  downloadTime:", downloadTime, err)
		return xormdb.RollbackAndLogError(session, "saveToSyncRrdpNotifyDb(): update lab_rpki_sync_rrdp_notify fail:"+
			notifyUrl, err)
	}
	rows, err := affected.RowsAffected()
	belogs.Debug("saveToSyncRrdpNotifyDb(): update lab_rpki_sync_rrdp_notify, notifyUrl:", notifyUrl, "  affected rows:", rows)
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("saveToSyncRrdpNotifyDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("saveToSyncRrdpNotifyDb(): ok, notifyUrl:", notifyUrl, "  curSerial:", curSerial, "  time(s):", time.Since(start))
	return nil
}
