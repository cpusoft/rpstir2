package hroa

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	rtrcommon "github.com/bgpsecurity/rpstir2/rtr-producer/common"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

func getRtrHroaFullFromRtrFullLogDb(serialNumber uint64) (rtrHroaFulls map[string]model.LabRpkiRtrHroaFull, err error) {
	start := time.Now()
	belogs.Debug("getRtrHroaFullFromRtrFullLogDb():serialNumber:", serialNumber)
	rtrHroaFs := make([]model.LabRpkiRtrHroaFull, 0)
	sql :=
		`select serialNumber,hroaAsn,subtreeIdentifier,encodedSubtree,afiFlags,sourceFrom 
	    from lab_rpki_rtr_hroa_full_log 
	    where serialNumber = ? 
		order by id `
	err = xormdb.XormEngine.SQL(sql, serialNumber).Find(&rtrHroaFs)
	if err != nil {
		belogs.Error("getRtrHroaFullFromRtrFullLogDb(): get rtr_hroa_full_log fail: serialNumber: ", serialNumber, err)
		return nil, err
	}
	if len(rtrHroaFs) == 0 {
		belogs.Debug("getRtrHroaFullFromRtrFullLogDb(): len(rtrHroaFs)==0: serialNumber", serialNumber)
		return make(map[string]model.LabRpkiRtrHroaFull, 0), nil
	}
	belogs.Debug("getRtrHroaFullFromRtrFullLogDb():model.LabRpkiRtrHroaFull, serialNumber, len(rtrHroaFs) : ", serialNumber, len(rtrHroaFs))

	rtrHroaFulls = make(map[string]model.LabRpkiRtrHroaFull, len(rtrHroaFs)+50)
	for i := range rtrHroaFs {
		key := convert.ToString(rtrHroaFs[i].HroaAsn.Value) + "_" +
			rtrHroaFs[i].SubtreeIdentifier.String() + "_" +
			convert.ToString(rtrHroaFs[i].EncodedSubtree.Value)
		rtrHroaFulls[key] = rtrHroaFs[i]
	}
	belogs.Info("getRtrHroaFullFromRtrFullLogDb():map LabRpkiRtrHroaFull, serialNumber:",
		serialNumber, "  , len(rtrHroaFs):", len(rtrHroaFs), "   time(s):", time.Since(start))
	return rtrHroaFulls, nil

}

func updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb(newSerialNumberModel *rtrcommon.SerialNumberModel,
	rtrHroaIncrementals []model.LabRpkiRtrHroaIncremental) (err error) {
	start := time.Now()
	belogs.Debug("updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb(): newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrHroaIncrementals):", len(rtrHroaIncrementals))

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()

	// serialnumber/rtrhroafull/rtrhroaincr should in one session
	// insert new serial number
	err = rtrcommon.InsertSerialNumberDb(session, newSerialNumberModel, start)
	if err != nil {
		belogs.Error("updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb():InsertSerialNumberDb fail,newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), err)
		return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb():InsertSerialNumberDb fail:", err)
	}
	belogs.Debug("updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb():InsertSerialNumberDb, newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))

	// delete and insert into lab_rpki_rtr_hroa_full
	sql := `delete from lab_rpki_rtr_hroa_full`
	_, err = session.Exec(sql)
	if err != nil {
		belogs.Error("updateRtrFullAndIncrementalAndRsyncLogRtrStateEndDb():delete lab_rpki_rtr_hroa_full fail:", err)
		return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb():delete lab_rpki_rtr_hroa_full fail:", err)
	}
	belogs.Debug("updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb():delete lab_rpki_rtr_hroa_full, time(s):", time.Since(start))

	// insert rtr_hroa_full from rtr_full_hroa_log
	sql = `
	insert ignore into lab_rpki_rtr_hroa_full 
		  (serialNumber, hroaAsn, subtreeIdentifier, 
		   encodedSubtree,afiFlags,sourceFrom ) 
	select serialNumber, hroaAsn, subtreeIdentifier, 
	       encodedSubtree, afiFlags,sourceFrom 
	from lab_rpki_rtr_hroa_full_log where serialNumber=? order by id`
	_, err = session.Exec(sql, newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb():insert into lab_rpki_rtr_hroa_full from lab_rpki_rtr_hroa_full_log fail: newSerialNumber:",
			jsonutil.MarshalJson(newSerialNumberModel), err)
		return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb():insert into lab_rpki_rtr_hroa_full from lab_rpki_rtr_hroa_full_log fail: ", err)
	}
	belogs.Debug("updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb():insert into lab_rpki_rtr_hroa_full from lab_rpki_rtr_hroa_full_log , time(s):", time.Since(start))

	// insert into lab_rpki_rtr_hroa_incremental
	sql = `insert ignore into lab_rpki_rtr_hroa_incremental
		(serialNumber,style,hroaAsn,subtreeIdentifier,   encodedSubtree,afiFlags,sourceFrom) values
		(?,?,?,?,  ?,?,?)`
	for i := range rtrHroaIncrementals {
		_, err = session.Exec(sql,
			newSerialNumberModel.SerialNumber, rtrHroaIncrementals[i].Style, rtrHroaIncrementals[i].HroaAsn, rtrHroaIncrementals[i].SubtreeIdentifier,
			rtrHroaIncrementals[i].EncodedSubtree, rtrHroaIncrementals[i].AfiFlags, rtrHroaIncrementals[i].SourceFrom)
		if err != nil {
			belogs.Error("updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb():insert into lab_rpki_rtr_hroa_incremental fail: newSerialNumber:",
				jsonutil.MarshalJson(newSerialNumberModel), jsonutil.MarshalJson(rtrHroaIncrementals[i]), err)
			return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb():insert into lab_rpki_rtr_hroa_incremental fail: ", err)
		}
	}
	belogs.Debug("updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb():insert into lab_rpki_rtr_hroa_incremental, time(s):", time.Since(start))

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb(): CommitSession fail: ", err)
	}

	belogs.Info("updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb(): CommitSession ok: newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrHroaIncrementals):", len(rtrHroaIncrementals), "   time(s):", time.Since(start))
	return nil
}
