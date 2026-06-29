package hroa

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	rtrcommon "github.com/bgpsecurity/rpstir2/rtr-producer/common"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

func RtrUpdateByHroaFromSync(curSerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel) (err error) {
	start := time.Now()
	belogs.Info("RtrUpdateByHroaFromSync():start, curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel),
		"    newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	// get all slurm
	slurmToRtrFullLogs, err := rtrcommon.GetAllSlurmsDb("hroa")
	if err != nil {
		belogs.Error("RtrUpdateByHroaFromSync(): GetAllSlurmsDb fail:", err)
		return err
	}
	belogs.Info("RtrUpdateByHroaFromSync(): len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs), "  time(s):", time.Since(start))

	//when both  len are 0, return nil
	if len(slurmToRtrFullLogs) == 0 {
		belogs.Info("RtrUpdateByHroaFromSync():hroa slurm are empty")
		return nil
	}

	_, err = rtrcommon.UpdateRtrHroaFullOrFullLogFromSlurmDb("lab_rpki_rtr_hroa_full_log", newSerialNumberModel.SerialNumber, slurmToRtrFullLogs, false)
	if err != nil {
		belogs.Error("RtrUpdateByHroaFromSync(): UpdateRtrHroaFullOrFullLogFromSlurmDb lab_rpki_rtr_hroa_full_log, fail:", err)
		return err
	}
	belogs.Info("RtrUpdateByHroaFromSync():UpdateRtrHroaFullOrFullLogFromSlurmDb new serialNumber:", newSerialNumberModel.SerialNumber,
		"  len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs), "  time(s):", time.Since(start))

	// get incrementals from curRtrFullLog and newRtrFullLog different
	rtrHroaIncrementals, err := getRtrHroaIncrementals(curSerialNumberModel, newSerialNumberModel)
	if err != nil {
		belogs.Error("RtrUpdateByHroaFromSync():getRtrHroaIncrementals fail: curSerialNumberModel:", curSerialNumberModel,
			"   newSerialNumber:", newSerialNumberModel, err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Info("RtrUpdateByHroaFromSync():getRtrHroaIncrementals, len(rtrHroaIncrementals)", len(rtrHroaIncrementals),
		"  curSerialNumberModel:", curSerialNumberModel, "   newSerialNumber:", newSerialNumberModel, "  time(s):", time.Since(start))

	err = updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb(newSerialNumberModel, rtrHroaIncrementals)
	if err != nil {
		belogs.Error("RtrUpdateByHroaFromSync():updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb: fail: newSerialNumber:",
			jsonutil.MarshalJson(newSerialNumberModel), "   len(rtrHroaIncrementals):", len(rtrHroaIncrementals), err, "  time(s):", time.Since(start))
		return err
	}
	belogs.Info("RtrUpdateByHroaFromSync(): updateSerialNumberAndRtrHroaFullAndRtrHroaIncrementalDb,  newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel),
		"   len(rtrHroaIncrementals):", len(rtrHroaIncrementals), "  time(s):", time.Since(start))
	return nil
}

func getRtrHroaIncrementals(curSerialNumberModel, newSerialNumberModel *rtrcommon.SerialNumberModel) (rtrHroaIncrementals []model.LabRpkiRtrHroaIncremental, err error) {
	start := time.Now()
	belogs.Debug("getRtrHroaIncrementals(): curSerialNumberModel:", jsonutil.MarshalJson(curSerialNumberModel), "   newSerialNumberModel:", jsonutil.MarshalJson(newSerialNumberModel))

	// get cur rtrFull
	rtrHroaFullCurs, err := getRtrHroaFullFromRtrFullLogDb(curSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrHroaIncrementals():getRtrHroaFullFromRtrFullLogDb rtrHroaFullCurs fail: cur SerialNumber:", curSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrHroaIncrementals(): getRtrHroaFullFromRtrFullLogDb len(rtrHroaFullCurs):", len(rtrHroaFullCurs),
		" cur serialNumber:", curSerialNumberModel.SerialNumber, "  time(s):", time.Since(start))

	// get latest rtrFull
	rtrHroaFullNews, err := getRtrHroaFullFromRtrFullLogDb(newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrHroaIncrementals():getRtrHroaFullFromRtrFullLogDb rtrHroaFullNews fail: new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrHroaIncrementals(): getRtrHroaFullFromRtrFullLogDb, len(rtrHroaFullNews):", len(rtrHroaFullNews),
		"  new SerialNumber:", newSerialNumberModel.SerialNumber, "  time(s):", time.Since(start))

	// get rtr incrementals
	rtrHroaIncrementals, err = diffRtrHroaFullToRtrHroaIncremental(rtrHroaFullCurs, rtrHroaFullNews, newSerialNumberModel.SerialNumber)
	if err != nil {
		belogs.Error("getRtrHroaIncrementals():GetRtrFull rtrFullLast fail: new SerialNumber:", newSerialNumberModel.SerialNumber, err)
		return nil, err
	}
	belogs.Info("getRtrHroaIncrementals():diffRtrHroaFullToRtrHroaIncremental, len(rtrHroaIncrementals)", len(rtrHroaIncrementals),
		" new  SerialNumber:", newSerialNumberModel.SerialNumber, "  time(s):", time.Since(start))
	return rtrHroaIncrementals, nil
}

func diffRtrHroaFullToRtrHroaIncremental(rtrHroaFullCurs, rtrHroaFullNews map[string]model.LabRpkiRtrHroaFull,
	newSerialNumber uint64) (rtrHroaIncrementals []model.LabRpkiRtrHroaIncremental, err error) {
	belogs.Debug("diffRtrHroaFullToRtrHroaIncremental(): len(rtrHroaFullsCurs):", len(rtrHroaFullCurs),
		"   len(rtrHroaFullNews):", len(rtrHroaFullNews), "   newSerialNumber:", newSerialNumber)

	rtrHroaIncrementals = make([]model.LabRpkiRtrHroaIncremental, 0, len(rtrHroaFullCurs))
	for keyNew, valueNew := range rtrHroaFullNews {
		// new exist in cur, then del in cur
		if _, ok := rtrHroaFullCurs[keyNew]; ok {
			belogs.Debug("diffRtrHroaFullToRtrHroaIncremental(): keyNew found in rtrHroaFullCurs:", keyNew,
				"  will del in rtrHroaFullCurs:", jsonutil.MarshalJson(rtrHroaFullCurs[keyNew]),
				"  and will ignore in rtrHroaFullNews:", jsonutil.MarshalJson(valueNew))
			delete(rtrHroaFullCurs, keyNew)
		} else {
			// new is not exist in cur, then this is announce
			rtrHroaIncremental := model.LabRpkiRtrHroaIncremental{
				Style:             "announce",
				HroaAsn:           valueNew.HroaAsn,
				SubtreeIdentifier: valueNew.SubtreeIdentifier,
				EncodedSubtree:    valueNew.EncodedSubtree,
				AfiFlags:          valueNew.AfiFlags,
				SerialNumber:      uint64(newSerialNumber),
				SourceFrom:        valueNew.SourceFrom,
			}
			belogs.Debug("diffRtrHroaFullToRtrHroaIncremental():keyNew not found in rtrHroaFullCurs, valueNew:", jsonutil.MarshalJson(valueNew),
				"   will set as announce incremental:", jsonutil.MarshalJson(rtrHroaIncremental))
			rtrHroaIncrementals = append(rtrHroaIncrementals, rtrHroaIncremental)
		}
	}
	belogs.Debug("diffRtrHroaFullToRtrHroaIncremental(): after announce, remain will as withdraw len(rtrHroaFullCurs):",
		len(rtrHroaFullCurs))
	// remain in cur, is not show in new, so this is withdraw
	for _, valueCur := range rtrHroaFullCurs {
		rtrHroaIncremental := model.LabRpkiRtrHroaIncremental{
			Style:             "withdraw",
			HroaAsn:           valueCur.HroaAsn,
			SubtreeIdentifier: valueCur.SubtreeIdentifier,
			EncodedSubtree:    valueCur.EncodedSubtree,
			AfiFlags:          valueCur.AfiFlags,
			SerialNumber:      uint64(newSerialNumber),
			SourceFrom:        valueCur.SourceFrom,
		}
		belogs.Debug("diffRtrHroaFullToRtrHroaIncremental(): withdraw incremental:",
			jsonutil.MarshalJson(rtrHroaIncremental))
		rtrHroaIncrementals = append(rtrHroaIncrementals, rtrHroaIncremental)
	}
	belogs.Debug("diffRtrHroaFullToRtrHroaIncremental(): newSerialNumber, len(rtrHroaIncrementals):", newSerialNumber, len(rtrHroaIncrementals))
	return rtrHroaIncrementals, nil
}
