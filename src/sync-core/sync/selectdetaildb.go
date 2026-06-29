package sync

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
)

func InsertSelectDetailDb(logID string, index int, startTime time.Time, detail string) (syncLogId uint64, err error) {
	belogs.Debug("InsertSelectDetailDb():logID:", logID, "   index:", index, "  detail:", detail)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("InsertSyncLogStartDb(): NewSession fail :", err)
		return 0, err
	}
	defer session.Close()

	//lab_rpki_distributed_select_detail
	sqlStr := `INSERT lab_rpki_distributed_select_detail(logId, idx, startTime, selectTime, detail)
					VALUES(?,?,?,?,?)`
	res, err := session.Exec(sqlStr, logID, index, startTime, time.Now(), detail)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session,
			"InsertSelectDetailDb(): INSERT sync_log fail:"+logID+","+","+detail, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session,
			"InsertSelectDetailDb(): LastInsertId fail:"+logID+","+detail+",", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return 0, xormdb.RollbackAndLogError(session,
			"InsertSelectDetailDb(): CommitSession fail:"+logID+","+detail+",", err)

	}

	belogs.Debug("InsertSelectDetailDb():new syncLogId:", id)
	return uint64(id), nil
}
