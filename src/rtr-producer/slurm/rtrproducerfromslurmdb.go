package slurm

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	rtrcommon "github.com/bgpsecurity/rpstir2/rtr-producer/common"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
)

// tableName: lab_rpki_rtr_full/lab_rpki_rtr_asa_full
func delRtrXFullFromSlurmDb(tableName string) (err error) {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("delRtrXFullFromSlurmDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()

	//  should delete last slurm, then insert new slurm
	sql := `delete from ` + tableName + ` where sourceFrom->'$.source' ='slurm' `
	belogs.Debug("delRtrXFullFromSlurmDb():`delete tableName source=slurm, sql:", sql)
	_, err = session.Exec(sql)
	if err != nil {
		belogs.Error("delRtrXFullFromSlurmDb(): delete tableName source=slurm fail, tableName:", tableName, err)
		return xormdb.RollbackAndLogError(session, "delRtrXFullFromSlurmDb(): delete tableName source=slurm fail", err)
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delRtrXFullFromSlurmDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "delRtrXFullFromSlurmDb(): CommitSession fail: ", err)
	}

	belogs.Info("delRtrXFullFromSlurmDb(): CommitSession ok: time(s):", time.Since(start))
	return nil

}

func insertRtrIncrementalByEffectSlurmDb(newSerialNumberModel *rtrcommon.SerialNumberModel, slurmToRtrFullLogs []model.SlurmToRtrFullLog) (err error) {
	start := time.Now()
	belogs.Debug("insertRtrIncrementalByEffectSlurmDb(): newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertRtrIncrementalByEffectSlurmDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()
	// lab_rpki_rtr_incremental
	var style string
	sql := `insert ignore into lab_rpki_rtr_incremental
				 (serialNumber,style,asn,address,   prefixLength,maxLength, sourceFrom) values
				 (?,?,?,?,  ?,?,?)`
	belogs.Debug("insertRtrIncrementalByEffectSlurmDb():lab_rpki_rtr_incremental, len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))
	for i := range slurmToRtrFullLogs {
		if slurmToRtrFullLogs[i].Style == "prefixAssertions" {
			style = "announce"
		} else if slurmToRtrFullLogs[i].Style == "prefixFilters" {
			style = "withdraw"
		}
		_, err = session.Exec(sql,
			newSerialNumberModel.SerialNumber, style, slurmToRtrFullLogs[i].Asn, slurmToRtrFullLogs[i].Address,
			slurmToRtrFullLogs[i].PrefixLength, slurmToRtrFullLogs[i].MaxLength, slurmToRtrFullLogs[i].SourceFromJson)
		if err != nil {
			belogs.Error("insertRtrIncrementalByEffectSlurmDb():insert into lab_rpki_rtr_incremental fail: new SerialNumber:",
				newSerialNumberModel.SerialNumber, jsonutil.MarshalJson(slurmToRtrFullLogs[i]), err)
			return xormdb.RollbackAndLogError(session, "insertRtrIncrementalByEffectSlurmDb insert into lab_rpki_rtr_incremental fail: ", err)
		}
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrIncrementalByEffectSlurmDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrIncrementalByEffectSlurmDb(): CommitSession fail: ", err)
	}

	belogs.Info("insertRtrIncrementalByEffectSlurmDb(): CommitSession ok: len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), " time(s):", time.Since(start))
	return nil
}

// tableName: lab_rpki_rtr_full/lab_rpki_rtr_asa_full
func updateRtrXFullByNewSerialNumberDb(tableName string, newSerialNumberModel *rtrcommon.SerialNumberModel) (err error) {
	start := time.Now()
	belogs.Debug("updateRtrXFullByNewSerialNumberDb(): tableName:", tableName, "  newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateRtrXFullByNewSerialNumberDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()

	sql := `update ` + tableName + ` set serialNumber= ? `
	_, err = session.Exec(sql, newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("updateRtrXFullByNewSerialNumberDb():update tableName set newSerialNumber fail: tableName:", tableName,
			" new serialNumber:", newSerialNumberModel.SerialNumber, err)
		return xormdb.RollbackAndLogError(session, "updateRtrXFullByNewSerialNumberDb update lab_rpki_rtr_full set newSerialNumber fail: ", err)
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateRtrXFullByNewSerialNumberDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "updateRtrXFullByNewSerialNumberDb(): CommitSession fail: ", err)
	}
	belogs.Info("updateRtrXFullByNewSerialNumberDb(): CommitSession ok:  tableName:", tableName, "   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   time(s):", time.Since(start))
	return nil
}

func insertRtrFullLogFromCurSerialNumberDb(curSerialNumberModel *rtrcommon.SerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel) (err error) {
	start := time.Now()
	belogs.Debug("insertRtrFullLogFromCurSerialNumberDb(): curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertRtrFullLogFromCurSerialNumberDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()

	// lab_rpki_rtr_full_log
	// should ignore last slurm, not insert to new rtr_full_log
	// insert into rtr_full_log and use new slurm update
	sql := `insert into lab_rpki_rtr_full_log (serialNumber,asn,address,prefixLength, maxLength,sourceFrom) 
		select ` + convert.ToString(newSerialNumberModel.SerialNumber) + `,asn,address,prefixLength, maxLength,sourceFrom from lab_rpki_rtr_full_log
		where serialNumber=? and sourceFrom->'$.source' !='slurm' order by id`
	belogs.Debug("insertRtrFullLogFromCurSerialNumberDb():`insert into lab_rpki_rtr_full_log, sql:", sql)
	_, err = session.Exec(sql, curSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("insertRtrFullLogFromCurSerialNumberDb(): insert lab_rpki_rtr_full_log fail, new SerialNumber:", newSerialNumberModel.SerialNumber,
			"  cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return xormdb.RollbackAndLogError(session, "insertRtrFullLogFromCurSerialNumberDb(): insert lab_rpki_rtr_full_log fail: ", err)
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrFullLogFromCurSerialNumberDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrFullLogFromCurSerialNumberDb(): CommitSession fail: ", err)
	}
	belogs.Info("insertRtrFullLogFromCurSerialNumberDb(): CommitSession ok: curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "   time(s):", time.Since(start))
	return nil

}

func insertRtrAsaFullLogFromCurSerialNumberDb(curSerialNumberModel *rtrcommon.SerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel) (err error) {
	start := time.Now()
	belogs.Debug("insertRtrAsaFullLogFromCurSerialNumberDb(): curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertRtrAsaFullLogFromCurSerialNumberDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()
	// lab_rpki_rtr_asa_full_log
	// should ignore last slurm, not insert to new rtr_full_log
	// insert into rtr_full_log and use new slurm update
	sql := `insert into lab_rpki_rtr_asa_full_log (serialNumber,customerAsn,providerAsn,addressFamily,sourceFrom) 
		select ` + convert.ToString(newSerialNumberModel.SerialNumber) + `,customerAsn,providerAsn,addressFamily,sourceFrom from lab_rpki_rtr_asa_full_log
		where serialNumber=? and sourceFrom->'$.source' !='slurm' order by id`
	belogs.Debug("insertRtrAsaFullLogFromCurSerialNumberDb():`insert into lab_rpki_rtr_asa_full_log, sql:", sql)
	_, err = session.Exec(sql, curSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("insertRtrAsaFullLogFromCurSerialNumberDb(): insert lab_rpki_rtr_asa_full_log fail, new SerialNumber:", newSerialNumberModel.SerialNumber,
			"  cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return xormdb.RollbackAndLogError(session, "insertRtrAsaFullLogFromCurSerialNumberDb(): insert lab_rpki_rtr_asa_full_log fail: ", err)
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrAsaFullLogFromCurSerialNumberDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrAsaFullLogFromCurSerialNumberDb(): CommitSession fail: ", err)
	}
	belogs.Info("insertRtrAsaFullLogFromCurSerialNumberDb(): CommitSession ok: curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "   time(s):", time.Since(start))
	return nil

}

// insert SerialNumber globalSerialNumber and subpartSerialNumber
func insertNewSerialNumberDb(newSerialNumberModel *rtrcommon.SerialNumberModel) (err error) {
	//save to lab_rpki_rtr_serial_number, get serialNumber
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertNewSerialNumberDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()

	start := time.Now()
	err = rtrcommon.InsertSerialNumberDb(session, newSerialNumberModel, start)
	if err != nil {
		belogs.Error("insertNewSerialNumberDb():InsertSerialNumberDb fail,newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), err)
		return xormdb.RollbackAndLogError(session, "insertNewSerialNumberDb():InsertSerialNumberDb fail:", err)
	}
	belogs.Debug("insertNewSerialNumberDb():InsertSerialNumberDb, newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertNewSerialNumberDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertNewSerialNumberDb(): CommitSession fail: ", err)
	}

	belogs.Info("insertNewSerialNumberDb():CommitSession ok, new SerialNumber:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))
	return nil
}

func selectSelfNodeDb() (rushNodeModel model.RushNodeModel, has bool, err error) {
	sql := `select id, nodeName, parentNodeId, url, isSelfUrl from lab_rpki_rush_node where isSelfUrl = 'true' `
	has, err = xormdb.XormEngine.SQL(sql).Get(&rushNodeModel)
	if err != nil {
		belogs.Error("selectSelfNodeDb():lab_rpki_rush_node parentNodeId, fail:", err)
		return rushNodeModel, false, err
	}
	belogs.Debug("selectSelfNodeDb():rushNodeModel:", jsonutil.MarshalJson(rushNodeModel), "  has:", has)
	return rushNodeModel, has, nil
}

func insertRtrAsaIncrementalByEffectSlurmDb(newSerialNumberModel *rtrcommon.SerialNumberModel, slurmToRtrFullLogs []model.SlurmToRtrFullLog) (err error) {
	start := time.Now()
	belogs.Debug("insertRtrAsaIncrementalByEffectSlurmDb(): newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertRtrAsaIncrementalByEffectSlurmDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()

	// lab_rpki_rtr_asa_incremental
	var style string
	sql := `insert ignore into lab_rpki_rtr_asa_incremental
				 (serialNumber,style,customerAsn,ProviderAsn,  AddressFamily,sourceFrom) values
				 (?,?,?,?,  ?,?)`
	belogs.Debug("insertRtrAsaIncrementalByEffectSlurmDb():lab_rpki_rtr_asa_incremental, len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))
	for i := range slurmToRtrFullLogs {
		if slurmToRtrFullLogs[i].Style == "aspaAssertions" {
			style = "announce"
		} else if slurmToRtrFullLogs[i].Style == "aspaFilters" {
			style = "withdraw"
		}
		_, err = session.Exec(sql,
			newSerialNumberModel.SerialNumber, style, slurmToRtrFullLogs[i].CustomerAsn, slurmToRtrFullLogs[i].ProviderAsn,
			slurmToRtrFullLogs[i].AddressFamily, slurmToRtrFullLogs[i].SourceFromJson)
		if err != nil {
			belogs.Error("insertRtrAsaIncrementalByEffectSlurmDb():insert into lab_rpki_rtr_asa_incremental fail: new SerialNumber:",
				newSerialNumberModel.SerialNumber, jsonutil.MarshalJson(slurmToRtrFullLogs[i]), err)
			return xormdb.RollbackAndLogError(session, "insertRtrAsaIncrementalByEffectSlurmDb insert into lab_rpki_rtr_asa_incremental fail: ", err)
		}
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrAsaIncrementalByEffectSlurmDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrAsaIncrementalByEffectSlurmDb(): CommitSession fail: ", err)
	}

	belogs.Info("insertRtrAsaIncrementalByEffectSlurmDb(): CommitSession ok: len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), " time(s):", time.Since(start))
	return nil
}

func insertRtrHroaFullLogFromCurSerialNumberDb(curSerialNumberModel *rtrcommon.SerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel) (err error) {
	start := time.Now()
	belogs.Debug("insertRtrHroaFullLogFromCurSerialNumberDb(): curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertRtrHroaFullLogFromCurSerialNumberDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()
	// lab_rpki_rtr_hroa_full_log
	// should ignore last slurm, not insert to new rtr_full_log
	// insert into rtr_full_log and use new slurm update
	sql := `insert into lab_rpki_rtr_hroa_full_log (serialNumber,hroaAsn,subtreeIdentifier,encodedSubtree,afiFlags,sourceFrom) 
		select ` + convert.ToString(newSerialNumberModel.SerialNumber) + `,hroaAsn,subtreeIdentifier,encodedSubtree,afiFlags,sourceFrom from lab_rpki_rtr_hroa_full_log
		where serialNumber=? and sourceFrom->'$.source' !='slurm' order by id`
	belogs.Debug("insertRtrHroaFullLogFromCurSerialNumberDb():`insert into lab_rpki_rtr_hroa_full_log, sql:", sql)
	_, err = session.Exec(sql, curSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("insertRtrHroaFullLogFromCurSerialNumberDb(): insert lab_rpki_rtr_hroa_full_log fail, new SerialNumber:", newSerialNumberModel.SerialNumber,
			"  cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return xormdb.RollbackAndLogError(session, "insertRtrHroaFullLogFromCurSerialNumberDb(): insert lab_rpki_rtr_hroa_full_log fail: ", err)
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrHroaFullLogFromCurSerialNumberDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrHroaFullLogFromCurSerialNumberDb(): CommitSession fail: ", err)
	}
	belogs.Info("insertRtrHroaFullLogFromCurSerialNumberDb(): CommitSession ok: curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "   time(s):", time.Since(start))
	return nil

}

func insertRtrHroaIncrementalByEffectSlurmDb(newSerialNumberModel *rtrcommon.SerialNumberModel, slurmToRtrFullLogs []model.SlurmToRtrFullLog) (err error) {
	start := time.Now()
	belogs.Debug("insertRtrHroaIncrementalByEffectSlurmDb(): newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertRtrHroaIncrementalByEffectSlurmDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()

	// lab_rpki_rtr_hroa_incremental
	var style string
	sql := `insert ignore into lab_rpki_rtr_hroa_incremental
				 (serialNumber,style,hroaAsn,subtreeIdentifier,  encodedSubtree,afiFlags,sourceFrom) values
				 (?,?,?,?,  ?,?,?)`
	belogs.Debug("insertRtrHroaIncrementalByEffectSlurmDb():lab_rpki_rtr_hroa_incremental, len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))
	for i := range slurmToRtrFullLogs {
		if slurmToRtrFullLogs[i].Style == "aspaAssertions" {
			style = "announce"
		} else if slurmToRtrFullLogs[i].Style == "aspaFilters" {
			style = "withdraw"
		}
		_, err = session.Exec(sql,
			newSerialNumberModel.SerialNumber, style, slurmToRtrFullLogs[i].HroaAsn, slurmToRtrFullLogs[i].SubtreeIdentifierBytes,
			slurmToRtrFullLogs[i].EncodedSubtree, slurmToRtrFullLogs[i].AfiFlags, slurmToRtrFullLogs[i].SourceFromJson)
		if err != nil {
			belogs.Error("insertRtrHroaIncrementalByEffectSlurmDb():insert into lab_rpki_rtr_hroa_incremental fail: new SerialNumber:",
				newSerialNumberModel.SerialNumber, jsonutil.MarshalJson(slurmToRtrFullLogs[i]), err)
			return xormdb.RollbackAndLogError(session, "insertRtrHroaIncrementalByEffectSlurmDb insert into lab_rpki_rtr_hroa_incremental fail: ", err)
		}
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrHroaIncrementalByEffectSlurmDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrHroaIncrementalByEffectSlurmDb(): CommitSession fail: ", err)
	}

	belogs.Info("insertRtrHroaIncrementalByEffectSlurmDb(): CommitSession ok: len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), " time(s):", time.Since(start))
	return nil
}

func insertRtrAsraFullLogFromCurSerialNumberDb(curSerialNumberModel *rtrcommon.SerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel) (err error) {
	start := time.Now()
	belogs.Debug("insertRtrAsraFullLogFromCurSerialNumberDb(): curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertRtrAsraFullLogFromCurSerialNumberDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()
	// lab_rpki_rtr_asra_full_log
	// should ignore last slurm, not insert to new rtr_full_log
	// insert into rtr_full_log and use new slurm update
	sql := `insert into lab_rpki_rtr_asra_full_log (serialNumber,customerAsnAsra,addressFamilyAsra,providerAsnAsras,otherNeighborAsnAsras,customerAsnAsras,lateralPeerAsnAsras,hybridAsras,valleyPathAsnAsras,sourceFrom) 
		select ` + convert.ToString(newSerialNumberModel.SerialNumber) + `,customerAsnAsra,addressFamilyAsra,providerAsnAsras,otherNeighborAsnAsras,customerAsnAsras,lateralPeerAsnAsras,hybridAsras,valleyPathAsnAsras,sourceFrom from lab_rpki_rtr_asra_full_log
		where serialNumber=? and sourceFrom->'$.source' !='slurm' order by id`
	belogs.Debug("insertRtrAsraFullLogFromCurSerialNumberDb():`insert into lab_rpki_rtr_asra_full_log, sql:", sql)
	_, err = session.Exec(sql, curSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("insertRtrAsraFullLogFromCurSerialNumberDb(): insert lab_rpki_rtr_asra_full_log fail, new SerialNumber:", newSerialNumberModel.SerialNumber,
			"  cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return xormdb.RollbackAndLogError(session, "insertRtrAsraFullLogFromCurSerialNumberDb(): insert lab_rpki_rtr_asra_full_log fail: ", err)
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrAsraFullLogFromCurSerialNumberDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrAsraFullLogFromCurSerialNumberDb(): CommitSession fail: ", err)
	}
	belogs.Info("insertRtrAsraFullLogFromCurSerialNumberDb(): CommitSession ok: curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), "   time(s):", time.Since(start))
	return nil

}

func insertRtrAsraIncrementalByEffectSlurmDb(newSerialNumberModel *rtrcommon.SerialNumberModel, slurmToRtrFullLogs []model.SlurmToRtrFullLog) (err error) {
	start := time.Now()
	belogs.Debug("insertRtrAsraIncrementalByEffectSlurmDb(): newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("insertRtrAsraIncrementalByEffectSlurmDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()

	// lab_rpki_rtr_asra_incremental
	var style string
	sql := `insert ignore into lab_rpki_rtr_asra_incremental
				 (serialNumber,style,
				 customerAsnAsra,addressFamilyAsra,providerAsnAsras,
				 otherNeighborAsnAsras,customerAsnAsras,lateralPeerAsnAsras,
				 hybridAsras,valleyPathAsnAsras,sourceFrom) values
				 (?,?,
				 ?,?,?,
				 ?,?,?,
				 ?,?,?)`
	belogs.Debug("insertRtrAsraIncrementalByEffectSlurmDb():lab_rpki_rtr_asra_incremental, len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))
	for i := range slurmToRtrFullLogs {
		if slurmToRtrFullLogs[i].Style == "aspaAssertions" {
			style = "announce"
		} else if slurmToRtrFullLogs[i].Style == "aspaFilters" {
			style = "withdraw"
		}
		_, err = session.Exec(sql,
			newSerialNumberModel.SerialNumber, style,
			slurmToRtrFullLogs[i].CustomerAsnAsra, slurmToRtrFullLogs[i].AddressFamilyAsra, xormdb.SqlNullString(slurmToRtrFullLogs[i].ProviderAsnAsrasStr),
			xormdb.SqlNullString(slurmToRtrFullLogs[i].OtherNeighborAsnAsrasStr), xormdb.SqlNullString(slurmToRtrFullLogs[i].CustomerAsnAsrasStr), xormdb.SqlNullString(slurmToRtrFullLogs[i].LateralPeerAsnAsrasStr),
			xormdb.SqlNullString(slurmToRtrFullLogs[i].HybridAsrasStr), xormdb.SqlNullString(slurmToRtrFullLogs[i].ValleyPathAsnAsrasStr), slurmToRtrFullLogs[i].SourceFromJson)
		if err != nil {
			belogs.Error("insertRtrAsraIncrementalByEffectSlurmDb():insert into lab_rpki_rtr_asra_incremental fail: new SerialNumber:",
				newSerialNumberModel.SerialNumber, jsonutil.MarshalJson(slurmToRtrFullLogs[i]), err)
			return xormdb.RollbackAndLogError(session, "insertRtrAsraIncrementalByEffectSlurmDb insert into lab_rpki_rtr_asra_incremental fail: ", err)
		}
	}

	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("insertRtrAsraIncrementalByEffectSlurmDb(): CommitSession fail :", err)
		return xormdb.RollbackAndLogError(session, "insertRtrAsraIncrementalByEffectSlurmDb(): CommitSession fail: ", err)
	}

	belogs.Info("insertRtrAsraIncrementalByEffectSlurmDb(): CommitSession ok: len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs),
		"   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel), " time(s):", time.Since(start))
	return nil
}
