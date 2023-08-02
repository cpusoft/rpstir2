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
func addMftsDb(syncLogFileModels []model.SyncLogFileModel) error {
	belogs.Info("addMftsDb(): will insert len(syncLogFileModels):", len(syncLogFileModels))
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("addMftsDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	start := time.Now()

	belogs.Debug("addMftsDb(): len(syncLogFileModels):", len(syncLogFileModels))
	// insert new mft
	for i := range syncLogFileModels {
		err = parsevalidatedb.InsertMftDb(session, &syncLogFileModels[i], start)
		if err != nil {
			belogs.Error("addMftsDb(): InsertMftDb fail:", jsonutil.MarshalJson(syncLogFileModels[i]), err)
			return xormdb.RollbackAndLogError(session, "addMftsDb(): InsertMftDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	err = updateSyncLogFilesJsonAllAndStateDb(session, syncLogFileModels)
	if err != nil {
		belogs.Error("addMftsDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "addMftsDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("addMftsDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("addMftsDb(): len(syncLogFileModels):", len(syncLogFileModels), "  time(s):", time.Since(start))
	return nil
}

// del
func delMftsDb(delSyncLogFileModels []model.SyncLogFileModel, updateSyncLogFileModels []model.SyncLogFileModel, wg *sync.WaitGroup) (err error) {
	defer func() {
		wg.Done()
	}()

	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("delMftsDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	syncLogFileModels := append(delSyncLogFileModels, updateSyncLogFileModels...)
	belogs.Info("delMftsDb(): will del len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = parsevalidatedb.DelMftByIdDb(session, syncLogFileModels[i].CertId)
		if err != nil {
			belogs.Error("delMftsDb(): DelMftByIdDb fail, cerId:", syncLogFileModels[i].CertId, err)
			return xormdb.RollbackAndLogError(session, "delMftsDb(): DelMftByIdDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	// only update delSyncLogFileModels
	err = updateSyncLogFilesJsonAllAndStateDb(session, delSyncLogFileModels)
	if err != nil {
		belogs.Error("delMftsDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "delMftsDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delMftsDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delMftsDb(): len(mfts):", len(syncLogFileModels), "  time(s):", time.Since(start))
	return nil
}
