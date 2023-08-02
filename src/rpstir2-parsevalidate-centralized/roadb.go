package parsevalidatecentralized

import (
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
	parsevalidatedb "rpstir2-parsevalidate-db"
)

// add
func addRoasDb(syncLogFileModels []model.SyncLogFileModel) error {
	belogs.Info("addRoasDb(): will insert len(syncLogFileModels):", len(syncLogFileModels))
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("addRoasDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	start := time.Now()

	belogs.Debug("addRoasDb(): len(syncLogFileModels):", len(syncLogFileModels))
	// insert new roa
	for i := range syncLogFileModels {
		err = parsevalidatedb.InsertRoaDb(session, &syncLogFileModels[i], start)
		if err != nil {
			belogs.Error("addRoasDb(): InsertRoaDb fail:", jsonutil.MarshalJson(syncLogFileModels[i]), err)
			return xormdb.RollbackAndLogError(session, "addRoasDb(): InsertRoaDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	err = updateSyncLogFilesJsonAllAndStateDb(session, syncLogFileModels)
	if err != nil {
		belogs.Error("addRoasDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "addRoasDb(): updateSyncLogFilesJsonAllAndStateDb fail ", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("addRoasDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("addRoasDb(): len(syncLogFileModels):", len(syncLogFileModels), "  time(s):", time.Since(start))
	return nil
}

// del
func delRoasDb(delSyncLogFileModels []model.SyncLogFileModel, updateSyncLogFileModels []model.SyncLogFileModel, wg *sync.WaitGroup) (err error) {
	defer func() {
		wg.Done()
	}()
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("delRoasDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	syncLogFileModels := append(delSyncLogFileModels, updateSyncLogFileModels...)
	belogs.Info("delRoasDb(): will del len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = parsevalidatedb.DelRoaByIdDb(session, syncLogFileModels[i].CertId)
		if err != nil {
			belogs.Error("delRoasDb(): DelRoaByIdDb fail, cerId:", syncLogFileModels[i].CertId, err)
			return xormdb.RollbackAndLogError(session, "delRoasDb(): DelRoaByIdDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	// only update delSyncLogFileModels
	err = updateSyncLogFilesJsonAllAndStateDb(session, delSyncLogFileModels)
	if err != nil {
		belogs.Error("delRoasDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "delRoasDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delRoasDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delRoasDb(): len(roas), ", len(syncLogFileModels), "  time(s):", time.Since(start))
	return nil
}
