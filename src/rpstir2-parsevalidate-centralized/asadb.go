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
func addAsasDb(syncLogFileModels []model.SyncLogFileModel) error {
	belogs.Info("addAsasDb(): will insert len(syncLogFileModels):", len(syncLogFileModels))
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("addAsasDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	start := time.Now()

	belogs.Debug("addAsasDb(): len(syncLogFileModels):", len(syncLogFileModels))
	// insert new asa
	for i := range syncLogFileModels {
		err = parsevalidatedb.InsertAsaDb(session, &syncLogFileModels[i], start)
		if err != nil {
			belogs.Error("addAsasDb(): InsertAsaDb fail:", jsonutil.MarshalJson(syncLogFileModels[i]), err)
			return xormdb.RollbackAndLogError(session, "addAsasDb(): InsertAsaDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	err = updateSyncLogFilesJsonAllAndStateDb(session, syncLogFileModels)
	if err != nil {
		belogs.Error("addAsasDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "addAsasDb(): updateSyncLogFilesJsonAllAndStateDb fail ", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("addAsasDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("addAsasDb(): len(syncLogFileModels):", len(syncLogFileModels), "  time(s):", time.Since(start))
	return nil
}

// del
func delAsasDb(delSyncLogFileModels []model.SyncLogFileModel, updateSyncLogFileModels []model.SyncLogFileModel, wg *sync.WaitGroup) (err error) {
	defer func() {
		wg.Done()
	}()
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("delAsasDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	syncLogFileModels := append(delSyncLogFileModels, updateSyncLogFileModels...)
	belogs.Info("delAsasDb(): will del len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = parsevalidatedb.DelAsaByIdDb(session, syncLogFileModels[i].CertId)
		if err != nil {
			belogs.Error("delAsasDb(): DelAsaByIdDb fail, cerId:", syncLogFileModels[i].CertId, err)
			return xormdb.RollbackAndLogError(session, "delAsasDb(): DelAsaByIdDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	// only update delSyncLogFileModels
	err = updateSyncLogFilesJsonAllAndStateDb(session, delSyncLogFileModels)
	if err != nil {
		belogs.Error("delAsasDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "delAsasDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delAsasDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delAsasDb(): len(asas), ", len(syncLogFileModels), "  time(s):", time.Since(start))
	return nil
}
