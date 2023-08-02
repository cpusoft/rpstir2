package parsevalidatecore

import (
	"errors"
	"strings"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/jsonutil"
	model "rpstir2-model"
)

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
func ParseValidateFile(certFile string) (certType string, certModel interface{},
	stateModel model.StateModel, fileHash string, err error) {
	belogs.Debug("ParseValidateFile(): parsevalidate start:", certFile)
	fileHash, err = hashutil.Sha256File(certFile)
	if err != nil {
		belogs.Error("ParseValidateFile(): Sha256File fail, certFile:", certFile, err)
		return "", nil, stateModel, "", err
	}
	if strings.HasSuffix(certFile, ".cer") {
		cerModel, stateModel, err := ParseValidateCer(certFile)
		belogs.Debug("ParseValidateFile(): after ParseValidateCer():certFile:", certFile,
			"  stateModel:", jsonutil.MarshalJson(stateModel), "  err:", err)
		return "cer", cerModel, stateModel, fileHash, err
	} else if strings.HasSuffix(certFile, ".crl") {
		crlModel, stateModel, err := ParseValidateCrl(certFile)
		belogs.Debug("ParseValidateFile(): after ParseValidateCrl(): certFile:", certFile,
			"  stateModel:", jsonutil.MarshalJson(stateModel), "  err:", err)
		return "crl", crlModel, stateModel, fileHash, err
	} else if strings.HasSuffix(certFile, ".mft") {
		mftModel, stateModel, err := ParseValidateMft(certFile)
		belogs.Debug("ParseValidateFile(): after ParseValidateMft():certFile:", certFile,
			"  stateModel:", jsonutil.MarshalJson(stateModel), "  err:", err)
		return "mft", mftModel, stateModel, fileHash, err
	} else if strings.HasSuffix(certFile, ".roa") {
		roaModel, stateModel, err := ParseValidateRoa(certFile)
		belogs.Debug("ParseValidateFile():after ParseValidateRoa(): certFile:", certFile,
			"  stateModel:", jsonutil.MarshalJson(stateModel), "  err:", err)
		return "roa", roaModel, stateModel, fileHash, err
	} else if strings.HasSuffix(certFile, ".sig") {
		sigModel, stateModel, err := ParseValidateSig(certFile)
		belogs.Debug("ParseValidateFile():after ParseValidateSig(): certFile:", certFile,
			"  stateModel:", jsonutil.MarshalJson(stateModel), "  err:", err)
		return "sig", sigModel, stateModel, fileHash, err
	} else if strings.HasSuffix(certFile, ".asa") {
		asaModel, stateModel, err := ParseValidateAsa(certFile)
		belogs.Debug("ParseValidateFile():after ParseValidateAsa(): certFile:", certFile,
			"  stateModel:", jsonutil.MarshalJson(stateModel), "  err:", err)
		return "asa", asaModel, stateModel, fileHash, err
	} else if strings.HasSuffix(certFile, ".gbr") {
		belogs.Info("ParseValidateFile():not support .gbr file: certFile:", certFile)
		return "", nil, stateModel, "", errors.New("unknown gbr file type")
	} else {
		return "", nil, stateModel, "", errors.New("unknown file type")
	}
}

func ParseFile(certFile string) (certModel interface{}, err error) {
	belogs.Debug("ParseFile(): parsevalidate start:", certFile)
	if strings.HasSuffix(certFile, ".cer") {
		cerModel, _, err := ParseValidateCer(certFile)
		if err != nil {
			belogs.Error("ParseFile(): ParseValidateCer:", certFile, "  err:", err)
			return nil, err
		}
		cerModel.FilePath = ""
		belogs.Debug("ParseFile(): certFile,cerModel:", certFile, cerModel)
		return cerModel, nil

	} else if strings.HasSuffix(certFile, ".crl") {
		crlModel, _, err := ParseValidateCrl(certFile)
		if err != nil {
			belogs.Error("ParseFile(): ParseValidateCrl:", certFile, "  err:", err)
			return nil, err
		}
		crlModel.FilePath = ""
		belogs.Debug("ParseFile(): certFile, crlModel:", certFile, crlModel)
		return crlModel, nil

	} else if strings.HasSuffix(certFile, ".mft") {
		mftModel, _, err := ParseValidateMft(certFile)
		if err != nil {
			belogs.Error("ParseFile(): ParseValidateMft:", certFile, "  err:", err)
			return nil, err
		}
		mftModel.FilePath = ""
		belogs.Debug("ParseFile(): certFile, mftModel:", certFile, mftModel)
		return mftModel, nil

	} else if strings.HasSuffix(certFile, ".roa") {
		roaModel, _, err := ParseValidateRoa(certFile)
		if err != nil {
			belogs.Error("ParseFile(): ParseValidateRoa:", certFile, "  err:", err)
			return nil, err
		}
		roaModel.FilePath = ""
		belogs.Debug("ParseFile(): certFile, roaModel:", certFile, roaModel)
		return roaModel, nil

	} else if strings.HasSuffix(certFile, ".sig") {
		sigModel, _, err := ParseValidateSig(certFile)
		if err != nil {
			belogs.Error("ParseFile(): ParseValidateSig:", certFile, "  err:", err)
			return nil, err
		}
		sigModel.FilePath = ""
		belogs.Debug("ParseFile(): certFile, sigModel:", certFile, sigModel)
		return sigModel, nil

	} else if strings.HasSuffix(certFile, ".asa") {
		asaModel, _, err := ParseValidateAsa(certFile)
		if err != nil {
			belogs.Error("ParseFile(): ParseValidateAsa:", certFile, "  err:", err)
			return nil, err
		}
		asaModel.FilePath = ""
		belogs.Debug("ParseFile(): certFile, asaModel:", certFile, asaModel)
		return asaModel, nil

	} else {
		return nil, errors.New("unknown file type")
	}
}

// only parse cer to get ca repository/rpkiNotify, raw subjct public key info
func ParseFileSimple(certFile string) (parseCerSimple model.ParseCerSimple, err error) {
	belogs.Info("parseCerSimple(): certFile:", certFile)
	if strings.HasSuffix(certFile, ".cer") {
		return ParseCerSimpleModel(certFile)
	}
	return parseCerSimple, errors.New("unknown file type")
}
