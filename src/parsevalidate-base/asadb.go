package parsevalidatebase

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	parsevalidatedb "github.com/bgpsecurity/rpstir2/parsevalidate-db"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

func insertAsaSyncLogFileModelDb(syncLogFileModel *model.SyncLogFileModel) error {
	start := time.Now()
	belogs.Debug("insertAsaSyncLogFileModelDb(): will insert syncLogFileModel:", syncLogFileModel.String())

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertAsaSyncLogFileModelDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	// insert new cer
	err = parsevalidatedb.InsertAsaDbWithSession(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("insertAsaSyncLogFileModelDb(): InsertAsaDbWithSession fail, syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel), err)
		return xormdb.RollbackAndLogError(session, "insertAsaSyncLogFileModelDb(): InsertAsaDbWithSession fail, syncLogFileModel: "+jsonutil.MarshalJson(syncLogFileModel), err)
	}
	belogs.Debug("insertAsaSyncLogFileModelDb(): after InsertAsaDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	err = parsevalidatedb.UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("insertAsaSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel), err)
		return xormdb.RollbackAndLogError(session, "insertAsaSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+jsonutil.MarshalJson(syncLogFileModel), err)
	}
	belogs.Debug("insertAsaSyncLogFileModelDb(): after UpdateSyncLogFileJsonAllAndStateDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertAsaSyncLogFileModelDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("insertAsaSyncLogFileModelDb(): ok, syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
	return nil
}

func delAsaSyncLogFileModelDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("delAsaSyncLogFileModelDb(): will del syncLogFileModel:", syncLogFileModel.String())

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("delAsaSyncLogFileModelDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = parsevalidatedb.DelAsaByIdDbWithSession(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("delAsaSyncLogFileModelDb(): DelAsaByIdDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "delAsaSyncLogFileModelDb(): DelAsaByIdDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}
	belogs.Debug("delAsaSyncLogFileModelDb(): after DelAsaByIdDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	// only del will upate SyncLogFile's state
	if syncLogFileModel.SyncType == "del" {
		err = parsevalidatedb.UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
		if err != nil {
			belogs.Error("delAsaSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
			return xormdb.RollbackAndLogError(session, "delAsaSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
		}
		belogs.Debug("delAsaSyncLogFileModelDb(): after UpdateSyncLogFileJsonAllAndStateDbWithSession, syncLogFileModel:", syncLogFileModel.String())
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delAsaSyncLogFileModelDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delAsaSyncLogFileModelDb(): syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
	return nil
}
