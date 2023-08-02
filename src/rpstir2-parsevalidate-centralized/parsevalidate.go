package parsevalidatecentralized

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/osutil"
	"golang.org/x/sync/errgroup"
	model "rpstir2-model"
	parsevalidatecore "rpstir2-parsevalidate-core"
	parsevalidatedb "rpstir2-parsevalidate-db"
)

// ParseValidateStart: start
func parseValidateStart() (nextStep string, err error) {

	start := time.Now()
	belogs.Info("parseValidateStart(): start")
	// save starttime to lab_rpki_sync_log
	labRpkiSyncLogId, err := parsevalidatedb.UpdateSyncLogParseValidateStartDb("parsevalidating")
	if err != nil {
		belogs.Error("parseValidateStart():updateRsyncLogParseValidateStartDb fail:", err)
		return "", err
	}
	belogs.Debug("parseValidateStart():updateRsyncLogParseValidateStartDb, labRpkiSyncLogId:", labRpkiSyncLogId)

	// get all need rsyncLogFile
	syncLogFileModels, err := getSyncLogFileModelsBySyncLogIdDb(labRpkiSyncLogId)
	if err != nil {
		belogs.Error("parseValidateStart():getSyncLogFileModelsBySyncLogIdDb fail:", labRpkiSyncLogId, err)
		return "", err
	}
	belogs.Debug("parseValidateStart(): getSyncLogFileModelsBySyncLogIdDb, syncLogFileModels.SyncLogId:", labRpkiSyncLogId, syncLogFileModels.SyncLogId)

	//process "del" and "update" rsyncLogFile
	err = delCertByDelAndUpdate(syncLogFileModels)
	if err != nil {
		belogs.Error("parseValidateStart():delCertByDelAndUpdate fail:", err)
		return "", err
	}
	belogs.Debug("parseValidateStart(): after delCertByDelAndUpdate, syncLogFileModels.SyncLogId:", syncLogFileModels.SyncLogId)

	// process "add" and "update" rsyncLogFile
	err = insertCertByAddAndUpdate(syncLogFileModels)
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
	err = parsevalidatedb.UpdateSyncLogParseValidateStateEndDb(labRpkiSyncLogId, "parsevalidated", make([]string, 0))
	if err != nil {
		belogs.Debug("parseValidateStart(): UpdateRsyncLogAndCert fail: ", err)
		return "", err
	}

	belogs.Info("parseValidateStart(): end, will call chainvalidate,  time(s):", time.Since(start))
	return "chainvalidate", nil
}

// get del;
// get update, because "update" should del first
func delCertByDelAndUpdate(syncLogFileModels *SyncLogFileModels) (err error) {
	start := time.Now()

	belogs.Debug("delCertByDelAndUpdate(): syncLogFileModels.SyncLogId.:", syncLogFileModels.SyncLogId)

	var wg sync.WaitGroup

	// get "del" and "update" cer synclog files to del
	belogs.Debug("delCertByDelAndUpdate(): len(syncLogFileModels.DelCerSyncLogFileModels):", len(syncLogFileModels.DelCerSyncLogFileModels),
		"       len(syncLogFileModels.UpdateCerSyncLogFileModels):", len(syncLogFileModels.UpdateCerSyncLogFileModels))
	if len(syncLogFileModels.DelCerSyncLogFileModels) > 0 || len(syncLogFileModels.UpdateCerSyncLogFileModels) > 0 {
		wg.Add(1)
		go delCersDb(syncLogFileModels.DelCerSyncLogFileModels, syncLogFileModels.UpdateCerSyncLogFileModels, &wg)
	}

	// get "del" and "update" crl synclog files to del
	belogs.Debug("delCertByDelAndUpdate(): len(syncLogFileModels.DelCrlSyncLogFileModels):", len(syncLogFileModels.DelCrlSyncLogFileModels),
		"       len(syncLogFileModels.UpdateCrlSyncLogFileModels):", len(syncLogFileModels.UpdateCrlSyncLogFileModels))
	if len(syncLogFileModels.DelCrlSyncLogFileModels) > 0 || len(syncLogFileModels.UpdateCrlSyncLogFileModels) > 0 {
		wg.Add(1)
		go delCrlsDb(syncLogFileModels.DelCrlSyncLogFileModels, syncLogFileModels.UpdateCrlSyncLogFileModels, &wg)
	}

	// get "del" and "update" mft synclog files to del
	belogs.Debug("delCertByDelAndUpdate(): len(syncLogFileModels.DelMftSyncLogFileModels):", len(syncLogFileModels.DelMftSyncLogFileModels),
		"       len(syncLogFileModels.UpdateMftSyncLogFileModels):", len(syncLogFileModels.UpdateMftSyncLogFileModels))
	if len(syncLogFileModels.DelMftSyncLogFileModels) > 0 || len(syncLogFileModels.UpdateMftSyncLogFileModels) > 0 {
		wg.Add(1)
		go delMftsDb(syncLogFileModels.DelMftSyncLogFileModels, syncLogFileModels.UpdateMftSyncLogFileModels, &wg)
	}

	// get "del" and "update" roa synclog files to del
	belogs.Debug("delCertByDelAndUpdate(): len(syncLogFileModels.DelRoaSyncLogFileModels):", len(syncLogFileModels.DelRoaSyncLogFileModels),
		"       len(syncLogFileModels.UpdateRoaSyncLogFileModels):", len(syncLogFileModels.UpdateRoaSyncLogFileModels))
	if len(syncLogFileModels.DelRoaSyncLogFileModels) > 0 || len(syncLogFileModels.UpdateRoaSyncLogFileModels) > 0 {
		wg.Add(1)
		go delRoasDb(syncLogFileModels.DelRoaSyncLogFileModels, syncLogFileModels.UpdateRoaSyncLogFileModels, &wg)
	}

	// get "del" and "update" asa synclog files to del
	belogs.Debug("delCertByDelAndUpdate(): len(syncLogFileModels.DelAsaSyncLogFileModels):", len(syncLogFileModels.DelAsaSyncLogFileModels),
		"       len(syncLogFileModels.UpdateAsaSyncLogFileModels):", len(syncLogFileModels.UpdateAsaSyncLogFileModels))
	if len(syncLogFileModels.DelAsaSyncLogFileModels) > 0 || len(syncLogFileModels.UpdateAsaSyncLogFileModels) > 0 {
		wg.Add(1)
		go delAsasDb(syncLogFileModels.DelAsaSyncLogFileModels, syncLogFileModels.UpdateAsaSyncLogFileModels, &wg)
	}

	wg.Wait()
	belogs.Info("delCertByDelAndUpdate(): end,  time(s):", time.Since(start))
	return nil

}

// insertCertByAddAndUpdate :  use update, because "update" should add
func insertCertByAddAndUpdate(syncLogFileModels *SyncLogFileModels) (err error) {

	start := time.Now()
	belogs.Debug("insertCertByAddAndUpdate(): syncLogFileModels.SyncLogId:", syncLogFileModels.SyncLogId)

	var wg sync.WaitGroup

	// add/update crl
	belogs.Info("insertCertByAddAndUpdate():len(syncLogFileModels.UpdateCerSyncLogFileModels):", len(syncLogFileModels.UpdateCerSyncLogFileModels))
	if len(syncLogFileModels.UpdateCerSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddCerts(syncLogFileModels.UpdateCerSyncLogFileModels, "cer", &wg)
	}

	// add/update crl
	belogs.Info("insertCertByAddAndUpdate():len(syncLogFileModels.UpdateCrlSyncLogFileModels):", len(syncLogFileModels.UpdateCrlSyncLogFileModels))
	if len(syncLogFileModels.UpdateCrlSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddCerts(syncLogFileModels.UpdateCrlSyncLogFileModels, "crl", &wg)
	}

	// add/update mft
	belogs.Info("insertCertByAddAndUpdate():len(syncLogFileModels.UpdateMftSyncLogFileModels):", len(syncLogFileModels.UpdateMftSyncLogFileModels))
	if len(syncLogFileModels.UpdateMftSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddCerts(syncLogFileModels.UpdateMftSyncLogFileModels, "mft", &wg)
	}

	// add/update roa
	belogs.Info("insertCertByAddAndUpdate():len(syncLogFileModels.UpdateRoaSyncLogFileModels):", len(syncLogFileModels.UpdateRoaSyncLogFileModels))
	if len(syncLogFileModels.UpdateRoaSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddCerts(syncLogFileModels.UpdateRoaSyncLogFileModels, "roa", &wg)
	}

	// add/update asa
	belogs.Info("insertCertByAddAndUpdate():len(syncLogFileModels.UpdateAsaSyncLogFileModels):", len(syncLogFileModels.UpdateAsaSyncLogFileModels))
	if len(syncLogFileModels.UpdateAsaSyncLogFileModels) > 0 {
		wg.Add(1)
		go parseValidateAndAddCerts(syncLogFileModels.UpdateAsaSyncLogFileModels, "asa", &wg)
	}

	wg.Wait()
	belogs.Info("insertCertByAddAndUpdate(): end,  time(s):", time.Since(start))
	return nil
}

func parseValidateAndAddCerts(syncLogFileModels []model.SyncLogFileModel, fileType string, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	start := time.Now()

	// parsevalidate
	belogs.Debug("parseValidateAndAddCerts(): len(syncLogFileModels):", len(syncLogFileModels), "  fileType:", fileType)
	var parseValidateWg sync.WaitGroup
	parseValidateCh := make(chan int, conf.Int("parse::parseConcurrentCount"))
	for i := range syncLogFileModels {
		parseValidateWg.Add(1)
		parseValidateCh <- 1
		go parseValidateCert(&syncLogFileModels[i], &parseValidateWg, parseValidateCh)
	}
	parseValidateWg.Wait()
	close(parseValidateCh)

	belogs.Info("parseValidateAndAddCerts():end parseValidate, len(syncLogFileModels):", len(syncLogFileModels), "  fileType:", fileType, "  fileType:", fileType,
		"  time(s):", time.Since(start), ", and will save to db")

	// add to db
	switch fileType {
	case "cer":
		addCersDb(syncLogFileModels)
	case "crl":
		addCrlsDb(syncLogFileModels)
	case "mft":
		addMftsDb(syncLogFileModels)
	case "roa":
		addRoasDb(syncLogFileModels)
	case "asa":
		addAsasDb(syncLogFileModels)
	}
	belogs.Info("parseValidateAndAddCerts():end add***Db(), len(syncLogFileModels):", len(syncLogFileModels), "  fileType:", fileType, "  time(s):", time.Since(start))
}

func parseValidateCert(syncLogFileModel *model.SyncLogFileModel,
	wg *sync.WaitGroup, parseValidateCh chan int) (parseFailFile string, err error) {
	defer func() {
		wg.Done()
		<-parseValidateCh
	}()

	start := time.Now()
	belogs.Debug("parseValidateCert(): syncLogFileModel :", syncLogFileModel.String())
	file := osutil.JoinPathFile(syncLogFileModel.FilePath, syncLogFileModel.FileName)
	belogs.Debug("parseValidateCert(): file :", file)
	_, certModel, stateModel, err := parseValidateFile(file)
	if err != nil {
		belogs.Error("parseValidateCert(): parseValidateFile fail: ", file, err)
		return file, err
	}
	syncLogFileModel.CertModel = certModel
	syncLogFileModel.StateModel = stateModel
	belogs.Debug("parseValidateCert(): parseValidateFile file :", file,
		"   syncType:", syncLogFileModel.SyncType, "  time(s):", time.Since(start))

	return "", nil

}

/*
MFT: Manifests for the Resource Public Key Infrastructure (RPKI)
https://datatracker.ietf.org/doc/rfc6486/?include_text=1

ROA: A Profile for Route Origin Authorizations (ROAs)
https://datatracker.ietf.org/doc/rfc6482/?include_text=1

CRL: Internet X.509 Public Key Infrastructure Certificate and Certificate Revocation List (CRL) Profile
https://datatracker.ietf.org/doc/rfc5280/?include_text=1

EE: Signed Object Template for the Resource Public Key Infrastructure (RPKI)
https://datatracker.ietf.org/doc/rfc6488/?include_text=1

CER: IP/AS:  X.509 Extensions for IP Addresses and AS Identifiers
https://datatracker.ietf.org/doc/rfc3779/?include_text=1

CER: A Profile for X.509 PKIX Resource Certificates
https://datatracker.ietf.org/doc/rfc6487/?include_text=1



A Profile for X.509 PKIX Resource Certificates
https://datatracker.ietf.org/doc/rfc6487/?include_text=1


A Profile for Route Origin Authorizations (ROAs)
https://datatracker.ietf.org/doc/rfc6482/?include_text=1

Signed Object Template for the Resource Public Key Infrastructure (RPKI)
https://datatracker.ietf.org/doc/rfc6488/?include_text=1

X.509 Extensions for IP Addresses and AS Identifiers
https://datatracker.ietf.org/doc/rfc3779/?include_text=1


Internet X.509 Public Key Infrastructure Certificate and Certificate Revocation List (CRL) Profile
https://datatracker.ietf.org/doc/rfc5280/?include_text=1
*/
// upload file to parse
func parseValidateFile(certFile string) (certType string, certModel interface{}, stateModel model.StateModel, err error) {
	belogs.Debug("parseValidateFile(): parsevalidate start:", certFile)

	if strings.HasSuffix(certFile, ".cer") {
		cerModel, stateModel, err := parsevalidatecore.ParseValidateCer(certFile)
		belogs.Debug("parseValidateFile():  after ParseValidateCer():certFile, stateModel:", certFile, stateModel, "  err:", err)
		return "cer", cerModel, stateModel, err
	} else if strings.HasSuffix(certFile, ".crl") {
		crlModel, stateModel, err := parsevalidatecore.ParseValidateCrl(certFile)
		belogs.Debug("parseValidateFile(): after ParseValidateCrl(): certFile,stateModel:", certFile, stateModel, "  err:", err)
		return "crl", crlModel, stateModel, err
	} else if strings.HasSuffix(certFile, ".mft") {
		mftModel, stateModel, err := parsevalidatecore.ParseValidateMft(certFile)
		belogs.Debug("parseValidateFile(): after ParseValidateMft():certFile,stateModel:", certFile, stateModel, "  err:", err)
		return "mft", mftModel, stateModel, err
	} else if strings.HasSuffix(certFile, ".roa") {
		roaModel, stateModel, err := parsevalidatecore.ParseValidateRoa(certFile)
		belogs.Debug("parseValidateFile():after ParseValidateRoa(): certFile,stateModel:", certFile, stateModel, "  err:", err)
		return "roa", roaModel, stateModel, err
	} else if strings.HasSuffix(certFile, ".sig") {
		sigModel, stateModel, err := parsevalidatecore.ParseValidateSig(certFile)
		belogs.Debug("parseValidateFile():after ParseValidateSig(): certFile,stateModel:", certFile, stateModel, "  err:", err)
		return "sig", sigModel, stateModel, err
	} else if strings.HasSuffix(certFile, ".asa") {
		asaModel, stateModel, err := parsevalidatecore.ParseValidateAsa(certFile)
		belogs.Debug("parseValidateFile():after ParseValidateAsa(): certFile,stateModel:", certFile, stateModel, "  err:", err)
		return "asa", asaModel, stateModel, err
	} else {
		return "", nil, stateModel, errors.New("unknown file type")
	}
}

func parseFile(certFile string) (certModel interface{}, err error) {
	belogs.Debug("parseFile(): parsevalidate start:", certFile)
	if strings.HasSuffix(certFile, ".cer") {
		cerModel, _, err := parsevalidatecore.ParseValidateCer(certFile)
		if err != nil {
			belogs.Error("parseFile(): ParseValidateCer:", certFile, "  err:", err)
			return nil, err
		}
		cerModel.FilePath = ""
		belogs.Debug("parseFile(): certFile,cerModel:", certFile, cerModel)
		return cerModel, nil

	} else if strings.HasSuffix(certFile, ".crl") {
		crlModel, _, err := parsevalidatecore.ParseValidateCrl(certFile)
		if err != nil {
			belogs.Error("parseFile(): ParseValidateCrl:", certFile, "  err:", err)
			return nil, err
		}
		crlModel.FilePath = ""
		belogs.Debug("parseFile(): certFile, crlModel:", certFile, crlModel)
		return crlModel, nil

	} else if strings.HasSuffix(certFile, ".mft") {
		mftModel, _, err := parsevalidatecore.ParseValidateMft(certFile)
		if err != nil {
			belogs.Error("parseFile(): ParseValidateMft:", certFile, "  err:", err)
			return nil, err
		}
		mftModel.FilePath = ""
		belogs.Debug("parseFile(): certFile, mftModel:", certFile, mftModel)
		return mftModel, nil

	} else if strings.HasSuffix(certFile, ".roa") {
		roaModel, _, err := parsevalidatecore.ParseValidateRoa(certFile)
		if err != nil {
			belogs.Error("parseFile(): ParseValidateRoa:", certFile, "  err:", err)
			return nil, err
		}
		roaModel.FilePath = ""
		belogs.Debug("parseFile(): certFile, roaModel:", certFile, roaModel)
		return roaModel, nil

	} else if strings.HasSuffix(certFile, ".sig") {
		sigModel, _, err := parsevalidatecore.ParseValidateSig(certFile)
		if err != nil {
			belogs.Error("parseFile(): ParseValidateSig:", certFile, "  err:", err)
			return nil, err
		}
		sigModel.FilePath = ""
		belogs.Debug("parseFile(): certFile, sigModel:", certFile, sigModel)
		return sigModel, nil

	} else if strings.HasSuffix(certFile, ".asa") {
		asaModel, _, err := parsevalidatecore.ParseValidateAsa(certFile)
		if err != nil {
			belogs.Error("parseFile(): ParseValidateAsa:", certFile, "  err:", err)
			return nil, err
		}
		asaModel.FilePath = ""
		belogs.Debug("parseFile(): certFile, asaModel:", certFile, asaModel)
		return asaModel, nil

	} else {
		return nil, errors.New("unknown file type")
	}
}

// only parse cer to get ca repository/rpkiNotify, raw subjct public key info
func parseFileSimple(certFile string) (parseCerSimple model.ParseCerSimple, err error) {
	belogs.Info("parseCerSimple(): certFile:", certFile)
	if strings.HasSuffix(certFile, ".cer") {
		return parsevalidatecore.ParseCerSimpleModel(certFile)
	}
	return parseCerSimple, errors.New("unknown file type")
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
