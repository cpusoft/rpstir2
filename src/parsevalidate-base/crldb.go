package parsevalidatebase

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	parsevalidatedb "github.com/bgpsecurity/rpstir2/parsevalidate-db"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

func insertCrlSyncLogFileModelDb(syncLogFileModel *model.SyncLogFileModel) error {
	start := time.Now()
	belogs.Debug("insertCrlSyncLogFileModelDb(): will insert syncLogFileModel:", syncLogFileModel.String())

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertCrlSyncLogFileModelDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	// insert new cer
	err = parsevalidatedb.InsertCrlDbWithSession(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("insertCrlSyncLogFileModelDb(): InsertCrlDbWithSession fail, syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel), err)
		return xormdb.RollbackAndLogError(session, "insertCrlSyncLogFileModelDb(): InsertCrlDbWithSession fail, syncLogFileModel: "+jsonutil.MarshalJson(syncLogFileModel), err)
	}
	belogs.Debug("insertCrlSyncLogFileModelDb(): after InsertCrlDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	err = parsevalidatedb.UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("insertCrlSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel), err)
		return xormdb.RollbackAndLogError(session, "insertCrlSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+jsonutil.MarshalJson(syncLogFileModel), err)
	}
	belogs.Debug("insertCrlSyncLogFileModelDb(): after UpdateSyncLogFileJsonAllAndStateDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertCrlSyncLogFileModelDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("insertCrlSyncLogFileModelDb(): ok, syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
	return nil
}

func delCrlSyncLogFileModelDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("delCrlSyncLogFileModelDb(): will del syncLogFileModel:", syncLogFileModel.String())

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("delCrlSyncLogFileModelDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = parsevalidatedb.DelCrlByIdDbWithSession(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("delCrlSyncLogFileModelDb(): DelCrlByIdDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "delCrlSyncLogFileModelDb(): DelCrlByIdDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}
	belogs.Debug("delCrlSyncLogFileModelDb(): after DelCrlByIdDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	// only del will upate SyncLogFile's state
	if syncLogFileModel.SyncType == "del" {
		err = parsevalidatedb.UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
		if err != nil {
			belogs.Error("delCrlSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
			return xormdb.RollbackAndLogError(session, "delCrlSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
		}
		belogs.Debug("delCrlSyncLogFileModelDb(): after UpdateSyncLogFileJsonAllAndStateDbWithSession, syncLogFileModel:", syncLogFileModel.String())
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delCrlSyncLogFileModelDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delCrlSyncLogFileModelDb(): syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
	return nil
}
