package parsevalidatebase

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	parsevalidatedb "github.com/bgpsecurity/rpstir2/parsevalidate-db"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

func insertCerSyncLogFileModelDb(syncLogFileModel *model.SyncLogFileModel) error {
	start := time.Now()
	belogs.Debug("insertCerSyncLogFileModelDb(): will insert syncLogFileModel:", syncLogFileModel.String())

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertCerSyncLogFileModelDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	// insert new cer
	err = parsevalidatedb.InsertCerDbWithSession(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("insertCerSyncLogFileModelDb(): InsertCerDbWithSession fail, syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel), err)
		return xormdb.RollbackAndLogError(session, "insertCerSyncLogFileModelDb(): InsertCerDbWithSession fail, syncLogFileModel: "+jsonutil.MarshalJson(syncLogFileModel), err)
	}
	belogs.Debug("insertCerSyncLogFileModelDb(): after InsertCerDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	err = parsevalidatedb.UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("insertCerSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel), err)
		return xormdb.RollbackAndLogError(session, "insertCerSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+jsonutil.MarshalJson(syncLogFileModel), err)
	}
	belogs.Debug("insertCerSyncLogFileModelDb(): after UpdateSyncLogFileJsonAllAndStateDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertCerSyncLogFileModelDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("insertCerSyncLogFileModelDb(): ok, syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
	return nil

}

func delCerSyncLogFileModelDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("delCerSyncLogFileModelDb(): will del syncLogFileModel:", syncLogFileModel.String())

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("delCerSyncLogFileModelDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = parsevalidatedb.DelCerByIdDbWithSession(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("delCerSyncLogFileModelDb(): DelCerByIdDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "delCerSyncLogFileModelDb(): DelCerByIdDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}
	belogs.Debug("delCerSyncLogFileModelDb(): after DelCerByIdDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	// only del will upate SyncLogFile's state
	if syncLogFileModel.SyncType == "del" {
		err = parsevalidatedb.UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
		if err != nil {
			belogs.Error("delCerSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
			return xormdb.RollbackAndLogError(session, "delCerSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
		}
		belogs.Debug("delCerSyncLogFileModelDb(): after UpdateSyncLogFileJsonAllAndStateDbWithSession, syncLogFileModel:", syncLogFileModel.String())
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delCerSyncLogFileModelDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delCerSyncLogFileModelDb(): syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
	return nil
}
