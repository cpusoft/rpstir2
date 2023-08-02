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
func addCersDb(syncLogFileModels []model.SyncLogFileModel) error {
	belogs.Info("addCersDb(): will insert len(syncLogFileModels):", len(syncLogFileModels))
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("addCersDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	start := time.Now()

	belogs.Debug("addCersDb(): len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		// insert new cer
		err = parsevalidatedb.InsertCerDb(session, &syncLogFileModels[i], start)
		if err != nil {
			belogs.Error("addCersDb(): InsertCerDb fail:", jsonutil.MarshalJson(&syncLogFileModels[i]), err)
			return xormdb.RollbackAndLogError(session, "addCersDb(): InsertCerDb fail: "+jsonutil.MarshalJson(&syncLogFileModels[i]), err)
		}
	}

	err = updateSyncLogFilesJsonAllAndStateDb(session, syncLogFileModels)
	if err != nil {
		belogs.Error("addCersDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "addCersDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("addCersDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("addCersDb(): len(syncLogFileModels):", len(syncLogFileModels), "  time(s):", time.Since(start))
	return nil

}

// del
func delCersDb(delSyncLogFileModels []model.SyncLogFileModel, updateSyncLogFileModels []model.SyncLogFileModel, wg *sync.WaitGroup) (err error) {
	defer func() {
		wg.Done()
	}()

	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("delCersDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	syncLogFileModels := append(delSyncLogFileModels, updateSyncLogFileModels...)
	belogs.Info("delCersDb(): will del len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = parsevalidatedb.DelCerByIdDb(session, syncLogFileModels[i].CertId)
		if err != nil {
			belogs.Error("delCersDb(): DelCerByIdDb fail, cerId:", syncLogFileModels[i].CertId, err)
			return xormdb.RollbackAndLogError(session, "delCersDb(): DelCerByIdDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	// only update delSyncLogFileModels
	err = updateSyncLogFilesJsonAllAndStateDb(session, delSyncLogFileModels)
	if err != nil {
		belogs.Error("delCersDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "delCersDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delCersDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delCersDb(): len(cers):", len(syncLogFileModels), "  time(s):", time.Since(start))
	return nil
}
