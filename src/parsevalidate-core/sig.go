package parsevalidatecore

import (
	"errors"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/fileutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"

	//asn1 "github.com/bgpsecurity/rpstir2/parsevalidate-asn1"
	openssl "github.com/bgpsecurity/rpstir2/parsevalidate-openssl"
)

func parseValidateSig(fileModel *model.FileModel) (sigModel model.SigModel, stateModel model.StateModel, err error) {
	start := time.Now()
	belogs.Debug("parseValidateSig(): fileModel:", jsonutil.MarshalJson(fileModel))
	stateModel = model.NewStateModel()
	err = parseSigModel(fileModel, &sigModel, &stateModel)
	if err != nil {
		belogs.Error("parseValidateSig():parseSigModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return sigModel, stateModel, err
	}
	belogs.Debug("parseValidateSig(): sigModel:", sigModel.String(), "  time(s):", time.Since(start))

	err = validateSigModel(&sigModel, &stateModel)
	if err != nil {
		belogs.Error("parseValidateSig():validateSigModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return sigModel, stateModel, err
	}
	if len(stateModel.Errors) > 0 || len(stateModel.Warnings) > 0 {
		belogs.Info("parseValidateSig():stateModel have errors or warnings, fileModel:", jsonutil.MarshalJson(fileModel), "     stateModel:", jsonutil.MarshalJson(stateModel))
	}
	stateModel.JudgeState()
	belogs.Debug("parseValidateSig(): sigModel.FilePath, sigModel.FileName, sigModel.Ski, sigModel.Aki:",
		sigModel.FilePath, sigModel.FileName, sigModel.Ski, sigModel.Aki, "  time(s):", time.Since(start))
	return sigModel, stateModel, nil
}

// Try to store the error in statemode instead of returning err
func parseSig(fileModel *model.FileModel) (sigModel model.SigModel, err error) {
	belogs.Debug("parseSig(): fileModel:", jsonutil.MarshalJson(fileModel))
	stateModel := model.NewStateModel()
	err = parseSigModel(fileModel, &sigModel, &stateModel)
	if err != nil {
		belogs.Error("parseSig():parseSigModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		return sigModel, nil
	}
	belogs.Debug("parseSig(): sigModel:", sigModel.String())
	return sigModel, nil
}

// some parse may return err, will stop
func parseSigModel(fileModel *model.FileModel, sigModel *model.SigModel, stateModel *model.StateModel) error {
	belogs.Debug("parseSigModel(): fileModel:", jsonutil.MarshalJson(fileModel))

	fileLength, err := fileutil.GetFileLength(fileModel.TempFilePathName)
	if err != nil {
		belogs.Error("parseSigModel(): GetFileLength fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "Fail to open file",
			Detail: err.Error()}
		stateModel.AddError(&stateMsg)
		return err
	}
	belogs.Debug("parseSigModel(): fileLength:", fileLength, " tempFile:", fileModel.TempFilePathName)

	if fileLength == 0 {
		belogs.Error("parseSigModel(): GetFileLength fail, fileLenght is empty, fileModel:", jsonutil.MarshalJson(fileModel))
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "File is empty",
			Detail: ""}
		stateModel.AddError(&stateMsg)
		return errors.New("File " + fileModel.FilePathName + " is empty")
	}

	sigModel.FilePath, sigModel.FileName = osutil.GetFilePathAndFileName(fileModel.FilePathName)
	sigModel.Version = 0 //default
	/*
		//TODO
		err = asn1.ParseSigModelByAsn1(fileModel, sigModel, stateModel)
		if err != nil {
			belogs.Error("parseSigModel(): ParseSigModelByAsn1 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			// have set stateMsg in ParseSigModelByAsn1 .
			return err
		}
	*/
	err = openssl.ParseSigModelByOpenssl(fileModel, sigModel, stateModel)
	if err != nil {
		belogs.Error("parseSigModel(): ParseSigModelByOpenssl fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
		// have set stateMsg in ParseSigModelByOpenssl .
		return err
	}
	/*
		//https://blog.csdn.net/Zhymax/article/details/7683925
		//openssl asn1parse -in -ard.sig -inform DER
		results, err := opensslutil.GetResultsByOpensslAns1(fileModel.TempFilePathName)
		if err != nil {
			belogs.Error("parseSigModel(): GetResultsByOpensslAns1 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file by openssl",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
		belogs.Debug("parseSigModel(): len(results):", len(results))

		//get file hash
		sigModel.FileHash, err = hashutil.Sha256File(fileModel.TempFilePathName)
		if err != nil {
			belogs.Error("parseSigModel(): Sha256File fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to read file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}

		// get sig hex
		// first HEX DUMP

			   39:d=4  hl=2 l=  11 prim: OBJECT            :1.2.840.113549.1.9.16.1.26
			   52:d=4  hl=2 l=inf  cons: cont [ 0 ]
			   54:d=5  hl=2 l=inf  cons: OCTET STRING
			   56:d=6  hl=3 l= 137 prim: OCTET STRING      [HEX DUMP]:308186020200CA180F323031383036323831373030
			32345A180F32303138303632393138303032345A060960864801650304020130533051162C36353736393433633735383262
			3164656266666261303564363235343034323462633765626363352E63726C032100154269177B0346014642A367DA415F32
			C2BFE7C4EAD8AED59ACCF8F20220F89C


		err = openssl.ParseSigModelByOpensslResults(results, sigModel)
		if err != nil {
			belogs.Error("parseSigModel():ParseSigModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		sigModel.EContentType, err = openssl.ParseSigEContentTypeByOpensslResults(results)
		if err != nil {
			belogs.Error("parseSigModel():ParseSigEContentTypeByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		sigModel.SignerInfoModel, err = openssl.ParseSignerInfoModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseSigModel():ParseSignerInfoModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		// get cer info in mft
		eeCerFile, fileByte, start, end, err := openssl.ParseByOpensslAns1ToX509(fileModel.TempFilePathName, results)
		if err != nil {
			belogs.Error("parseSigModel():ParseByOpensslAns1ToX509 fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse ee certificate by openssl",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
		defer osutil.CloseAndRemoveFile(eeCerFile)
		belogs.Debug("parseSigModel():ParseByOpensslAns1ToX509:", eeCerFile, fileByte, start, end)

		results, err = opensslutil.GetResultsByOpensslX509(eeCerFile.Name())
		if err != nil {
			belogs.Error("parseSigModel(): GetResultsByOpensslX509 fail, eeCerFile:", eeCerFile.Name(), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse ee certificate by openssl",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
			return err
		}
		belogs.Debug("parseSigModel(): len(results):", len(results))

		sigModel.Aki, sigModel.Ski, err = openssl.ParseAkiSkiByOpensslResults(results)
		if err != nil {
			belogs.Error("parseMftByOpenssl(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		// AIA, no SIA
		sigModel.AiaModel, _, err = openssl.ParseAiaModelSiaModelByOpensslResults(results)
		if err != nil {
			belogs.Error("parseSigModel(): ParseAiaModelSiaModelByOpensslResults fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}

		sigModel.EeCertModel, err = ParseEeCertModel(eeCerFile.Name(), fileByte, start, end)
		if err != nil {
			belogs.Error("parseSigModel(): ParseEeCertModel fail, fileModel:", jsonutil.MarshalJson(fileModel), err)
			stateMsg := model.StateMsg{Stage: "parsevalidate",
				Fail:   "Fail to parse file",
				Detail: err.Error()}
			stateModel.AddError(&stateMsg)
		}
	*/
	belogs.Debug("parseSigModel(): sigModel:", sigModel.String())
	return nil
}

func validateSigModel(sigModel *model.SigModel, stateModel *model.StateModel) (err error) {
	return
}
