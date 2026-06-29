package parsevalidatebase

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	parsevalidatedb "github.com/bgpsecurity/rpstir2/parsevalidate-db"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

func insertRoaSyncLogFileModelDb(syncLogFileModel *model.SyncLogFileModel) error {
	start := time.Now()
	belogs.Debug("insertRoaSyncLogFileModelDb(): will insert syncLogFileModel:", syncLogFileModel.String())

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertRoaSyncLogFileModelDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	// insert new cer
	err = parsevalidatedb.InsertRoaDbWithSession(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("insertRoaSyncLogFileModelDb(): InsertRoaDbWithSession fail, syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel), err)
		return xormdb.RollbackAndLogError(session, "insertRoaSyncLogFileModelDb(): InsertRoaDbWithSession fail, syncLogFileModel: "+jsonutil.MarshalJson(syncLogFileModel), err)
	}
	belogs.Debug("insertRoaSyncLogFileModelDb(): after InsertRoaDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	err = parsevalidatedb.UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("insertRoaSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel), err)
		return xormdb.RollbackAndLogError(session, "insertRoaSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+jsonutil.MarshalJson(syncLogFileModel), err)
	}
	belogs.Debug("insertRoaSyncLogFileModelDb(): after UpdateSyncLogFileJsonAllAndStateDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRoaSyncLogFileModelDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("insertRoaSyncLogFileModelDb(): ok, syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
	return nil
}

func delRoaSyncLogFileModelDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("delRoaSyncLogFileModelDb(): will del syncLogFileModel:", syncLogFileModel.String())

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("delRoaSyncLogFileModelDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = parsevalidatedb.DelRoaByIdDbWithSession(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("delRoaSyncLogFileModelDb(): DelRoaByIdDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "delRoaSyncLogFileModelDb(): DelRoaByIdDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}
	belogs.Debug("delRoaSyncLogFileModelDb(): after DelRoaByIdDbWithSession, syncLogFileModel:", syncLogFileModel.String())

	// only del will upate SyncLogFile's state
	if syncLogFileModel.SyncType == "del" {
		err = parsevalidatedb.UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
		if err != nil {
			belogs.Error("delRoaSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
			return xormdb.RollbackAndLogError(session, "delRoaSyncLogFileModelDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
		}
		belogs.Debug("delRoaSyncLogFileModelDb(): after UpdateSyncLogFileJsonAllAndStateDbWithSession, syncLogFileModel:", syncLogFileModel.String())
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delRoaSyncLogFileModelDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delRoaSyncLogFileModelDb(): syncLogFileModel:", syncLogFileModel.String(), "  time(s):", time.Since(start))
	return nil
}
