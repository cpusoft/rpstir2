package parsevalidatebase

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	parsevalidatedb "github.com/bgpsecurity/rpstir2/parsevalidate-db"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

func insertMftSyncLogFileModelDb(syncLogFileModel *model.SyncLogFileModel) error {
	start := time.Now()
	belogs.Debug("insertMftSyncLogFileModelDb(): will insert syncLogFileModel:", syncLogFileModel.String())

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertMftSyncLogFileModelDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	// insert new cer
	err = parsevalidatedb.InsertMftDbWithSession(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("insertMftSyncLogFileModelDb(): InsertMftDbWithSession fail, syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel), err)
		return xormdb.RollbackAndLogError(session, "insertMftSyncLogFileModelDb(): InsertMftDbWithSession fail, syncLogFileModel: "+jsonutil.MarshalJson(syncLogFileModel), err)
	}
	belogs.Debug("insertMftSyncLogFileModelDb(): after InsertMftDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	err = parsevalidatedb.UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("insertMftSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel), err)
		return xormdb.RollbackAndLogError(session, "insertMftSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+jsonutil.MarshalJson(syncLogFileModel), err)
	}
	belogs.Debug("insertMftSyncLogFileModelDb(): after UpdateSyncLogFileJsonAllAndStateDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertMftSyncLogFileModelDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("insertMftSyncLogFileModelDb(): ok, syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
	return nil
}

func delMftSyncLogFileModelDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("delMftSyncLogFileModelDb(): will del syncLogFileModel:", syncLogFileModel.String())

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("delMftSyncLogFileModelDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = parsevalidatedb.DelMftByIdDbWithSession(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("delMftSyncLogFileModelDb(): DelMftByIdDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "delMftSyncLogFileModelDb(): DelMftByIdDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}
	belogs.Debug("delMftSyncLogFileModelDb(): after DelMftByIdDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	// only del will upate SyncLogFile's state
	if syncLogFileModel.SyncType == "del" {
		err = parsevalidatedb.UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
		if err != nil {
			belogs.Error("delMftSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
			return xormdb.RollbackAndLogError(session, "delMftSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
		}
		belogs.Debug("delMftSyncLogFileModelDb(): after UpdateSyncLogFileJsonAllAndStateDbWithSession, syncLogFileModel:", syncLogFileModel.String())
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delMftSyncLogFileModelDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delMftSyncLogFileModelDb(): syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
	return nil
}
