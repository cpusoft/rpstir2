package parsevalidatebase

import (
	"sync"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	parsevalidatecore "github.com/bgpsecurity/rpstir2/parsevalidate-core"
	parsevalidatedb "github.com/bgpsecurity/rpstir2/parsevalidate-db"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"golang.org/x/sync/errgroup"
)

// ParseValidateStart: start
func parseValidateStart() (nextStep string, err error) {

	start := time.Now()
	belogs.Info("parseValidateStart(): start")
	// save starttime to lab_rpki_sync_log
	syncLogId, err := parsevalidatedb.UpdateSyncLogParseValidateStartDb("parsevalidating")
	if err != nil {
		belogs.Error("parseValidateStart():updateRsyncLogParseValidateStartDb fail:", err)
		return "", err
	}
	belogs.Debug("parseValidateStart():updateRsyncLogParseValidateStartDb, syncLogId:", syncLogId)

	// get all need rsyncLogFile
	syncLogFileModels, err := getSyncLogFileModelsBySyncLogIdDb(syncLogId)
	if err != nil {
		belogs.Error("parseValidateStart():getSyncLogFileModelsBySyncLogIdDb fail:", syncLogId, err)
		return "", err
	}
	belogs.Debug("parseValidateStart(): getSyncLogFileModelsBySyncLogIdDb, syncLogId:", syncLogId, "  len(syncLogFileModels):", len(syncLogFileModels))

	//process "del" and "update" rsyncLogFile
	err = delSyncLogFileModels(syncLogId, syncLogFileModels)
	if err != nil {
		belogs.Error("parseValidateStart(): delSyncLogFileModels fail:", err)
		return "", err
	}
	belogs.Debug("parseValidateStart(): after delSyncLogFileModels, syncLogId:", syncLogId, "  len(syncLogFileModels):", len(syncLogFileModels))

	// process "add" and "update" rsyncLogFile
	err = parseValidateAndInsertSyncLogFileModels(syncLogId, syncLogFileModels)
	if err != nil {
		belogs.Error("parseValidateStart():insertCertByAddAndUpdate fail:", err)
		return "", err
	}

	// will check all certs, not only this rsyncLogFiles : expire
	err = updateCertByCheckAll()
	if err != nil {
		belogs.Error("parseValidateStart():updateCertByCheckAll fail:", err)
		return "", err
	}

	// save to db
	err = parsevalidatedb.UpdateSyncLogParseValidateStateEndDb(syncLogId, "parsevalidated", make([]string, 0))
	if err != nil {
		belogs.Debug("parseValidateStart(): UpdateRsyncLogAndCert fail: ", err)
		return "", err
	}

	belogs.Info("parseValidateStart(): end, will call chainvalidate,  time(s):", time.Since(start))
	return "chainvalidate", nil
}

func delSyncLogFileModels(syncLogId uint64, syncLogFileModels []*model.SyncLogFileModel) (err error) {
	start := time.Now()

	belogs.Debug("delSyncLogFileModels():syncLogId:", syncLogId, " len(syncLogFileModels):", len(syncLogFileModels))

	var delWg sync.WaitGroup
	delConcurrentCountCh := make(chan int, conf.Int("parse::delConcurrentCount"))
	for i := range syncLogFileModels {
		delWg.Add(1)
		delConcurrentCountCh <- 1
		go func(i int) {
			defer func() {
				delWg.Done()
				<-delConcurrentCountCh
			}()
			belogs.Debug("delSyncLogFileModels(): for i:", i, "  certId:", syncLogFileModels[i].CertId, "  syncLogId:", syncLogId)
			if syncLogFileModels[i].CertId > 0 {
				fileType := syncLogFileModels[i].FileType
				belogs.Debug("delSyncLogFileModels(): fileType:", fileType, "  i:", i)
				switch fileType {
				case "asa":
					delAsaSyncLogFileModelDb(syncLogFileModels[i])
				case "cer":
					delCerSyncLogFileModelDb(syncLogFileModels[i])
				case "crl":
					delCrlSyncLogFileModelDb(syncLogFileModels[i])
				case "mft":
					delMftSyncLogFileModelDb(syncLogFileModels[i])
				case "roa":
					delRoaSyncLogFileModelDb(syncLogFileModels[i])
				}
			}
		}(i)
	}
	delWg.Wait()
	close(delConcurrentCountCh)
	belogs.Info("delSyncLogFileModels(): end, syncLogId:", syncLogId, " len(syncLogFileModels):", len(syncLogFileModels), " time(s):", time.Since(start))
	return nil

}

func parseValidateAndInsertSyncLogFileModels(syncLogId uint64, syncLogFileModels []*model.SyncLogFileModel) (err error) {

	start := time.Now()
	belogs.Debug("parseValidateAndInsertSyncLogFileModels(): syncLogId:", syncLogId, "   len(syncLogFileModels):", len(syncLogFileModels))

	var parseValidateAndInsertWg sync.WaitGroup
	parseValidateAndInsertCh := make(chan int, conf.Int("parse::parseInsertConcurrentCount"))
	for i := range syncLogFileModels {
		if syncLogFileModels[i].SyncType == "del" {
			continue
		}
		parseValidateAndInsertWg.Add(1)
		parseValidateAndInsertCh <- 1
		go func(i int) {
			defer func() {
				parseValidateAndInsertWg.Done()
				<-parseValidateAndInsertCh
			}()

			file := osutil.JoinPathFile(syncLogFileModels[i].FilePath, syncLogFileModels[i].FileName)
			belogs.Debug("parseValidateSyncLogFileModel(): file :", file)
			_, certModel, stateModel, originModel, _, err := parsevalidatecore.ParseValidateFile(file)
			if err != nil {
				belogs.Error("parseValidateSyncLogFileModel(): parsevalidatecore.ParseValidateFile fail:", file, err)
				return
			}

			syncLogFileModels[i].CertModel = certModel
			syncLogFileModels[i].StateModel = stateModel
			syncLogFileModels[i].OriginModel = originModel
			belogs.Debug("parseValidateAndInsertSyncLogFileModels(): after parseValidateSyncLogFileModel, file:", syncLogFileModels[i].FilePath, syncLogFileModels[i].FileName, "  i:", i,
				"  syncLogFileModel.CertModel:", jsonutil.MarshalJson(syncLogFileModels[i].CertModel),
				"  syncLogFileModel.StateModel:", jsonutil.MarshalJson(syncLogFileModels[i].StateModel),
				"  syncLogFileModel.OriginModel:", jsonutil.MarshalJson(syncLogFileModels[i].OriginModel))
			fileType := syncLogFileModels[i].FileType
			switch fileType {
			case "cer":
				insertCerSyncLogFileModelDb(syncLogFileModels[i])
			case "crl":
				insertCrlSyncLogFileModelDb(syncLogFileModels[i])
			case "mft":
				insertMftSyncLogFileModelDb(syncLogFileModels[i])
			case "roa":
				insertRoaSyncLogFileModelDb(syncLogFileModels[i])
			case "asa":
				insertAsaSyncLogFileModelDb(syncLogFileModels[i])
			}
		}(i)
	}

	parseValidateAndInsertWg.Wait()
	close(parseValidateAndInsertCh)
	belogs.Info("parseValidateAndInsertSyncLogFileModels(): end, syncLogId:", syncLogId, "   len(syncLogFileModels):", len(syncLogFileModels), " time(s):", time.Since(start))
	return nil
}

func updateCertByCheckAll() (err error) {

	start := time.Now()
	belogs.Info("updateCertByCheckAll():start:")

	var g errgroup.Group
	g.Go(func() error {
		er := parsevalidatedb.UpdateCerByCheckAll(start)
		if er != nil {
			belogs.Error("updateCertByCheckAll(): UpdateCerByCheckAll:  err:", er)
		}
		return er
	})

	g.Go(func() error {
		er := parsevalidatedb.UpdateCrlByCheckAll(start)
		if er != nil {
			belogs.Error("updateCertByCheckAll(): UpdateCrlByCheckAll:  err:", er)
		}
		return er
	})

	g.Go(func() error {
		er := parsevalidatedb.UpdateMftByCheckAll(start)
		if er != nil {
			belogs.Error("updateCertByCheckAll(): UpdateMftByCheckAll:  err:", er)
		}
		return er
	})

	g.Go(func() error {
		er := parsevalidatedb.UpdateRoaByCheckAll(start)
		if er != nil {
			belogs.Error("updateCertByCheckAll(): UpdateRoaByCheckAll:  err:", er)
		}
		return er
	})
	g.Go(func() error {
		er := parsevalidatedb.UpdateAsaByCheckAll(start)
		if er != nil {
			belogs.Error("updateCertByCheckAll(): UpdateAsaByCheckAll:  err:", er)
		}
		return er
	})

	if err := g.Wait(); err != nil {
		belogs.Error("updateCertByCheckAll(): fail, err:", err, "   time(s):", time.Since(start))
		return err
	}
	belogs.Info("updateCertByCheckAll(): ok,   time(s):", time.Since(start))
	return nil
}
