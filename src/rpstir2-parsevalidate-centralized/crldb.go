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
func addCrlsDb(syncLogFileModels []model.SyncLogFileModel) error {
	belogs.Info("addCrlsDb(): will insert len(syncLogFileModels):", len(syncLogFileModels))
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("addCrlsDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	start := time.Now()

	// add
	belogs.Debug("addCrlsDb(): len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = parsevalidatedb.InsertCrlDb(session, &syncLogFileModels[i], start)
		if err != nil {
			belogs.Error("addCrlsDb(): InsertCrlDb fail:", jsonutil.MarshalJson(syncLogFileModels[i]), err)
			return xormdb.RollbackAndLogError(session, "addCrlsDb(): InsertCrlDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	err = updateSyncLogFilesJsonAllAndStateDb(session, syncLogFileModels)
	if err != nil {
		belogs.Error("addCrlsDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "addCrlsDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("addCrlsDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("addCrlsDb(): len(syncLogFileModels):", len(syncLogFileModels), "  time(s):", time.Since(start))
	return nil

}

// del
func delCrlsDb(delSyncLogFileModels []model.SyncLogFileModel, updateSyncLogFileModels []model.SyncLogFileModel, wg *sync.WaitGroup) (err error) {
	defer func() {
		wg.Done()
	}()

	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("delCrlsDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	syncLogFileModels := append(delSyncLogFileModels, updateSyncLogFileModels...)
	belogs.Info("delCrlsDb(): will del len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = parsevalidatedb.DelCrlByIdDb(session, syncLogFileModels[i].CertId)
		if err != nil {
			belogs.Error("delCrlsDb(): DelCrlByIdDb fail, cerId:", syncLogFileModels[i].CertId, err)
			return xormdb.RollbackAndLogError(session, "delCrlsDb(): DelCrlByIdDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	// only update delSyncLogFileModels
	err = updateSyncLogFilesJsonAllAndStateDb(session, delSyncLogFileModels)
	if err != nil {
		belogs.Error("delCrlsDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "delCrlsDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delCrlsDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delCrlsDb(): len(crls):", len(syncLogFileModels), "  time(s):", time.Since(start))
	return nil
}
