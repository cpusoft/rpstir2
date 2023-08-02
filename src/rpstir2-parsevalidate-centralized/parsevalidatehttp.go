package parsevalidatecentralized

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/gin-gonic/gin"
	model "rpstir2-model"
	parsevalidatecore "rpstir2-parsevalidate-core"
)

func ParseValidateStart(c *gin.Context) {
	belogs.Debug("ParseValidateStart(): start: ")

	//check serviceState
	httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
		"/sys/servicestate", `{"operate":"enter","state":"parsevalidate"}`, false)

	go func() {
		nextStep, err := parseValidateStart()
		belogs.Debug("ParseValidateStart():  parseValidateStart end,  nextStep is :", nextStep, err)
		// leave serviceState
		if err != nil {
			// will end this whole sync
			belogs.Error("ParseValidateStart():  parseValidateStart fail", err)
			httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
				"/sys/servicestate", `{"operate":"leave","state":"end"}`, false)
		} else {
			httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
				"/sys/servicestate", `{"operate":"leave","state":"parsevalidate"}`, false)
			// will call chainValidate
			go httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
				"/chainvalidate/start", "", false)
			belogs.Info("ParseValidateStart():  sync.Start end,  nextStep is :", nextStep)
		}

	}()

	ginserver.ResponseOk(c, nil)

}

// upload file to parse;
// only one file
func ParseValidateFile(c *gin.Context) {
	start := time.Now()
	tmpDir, err := ioutil.TempDir("", "ParseValidateFile") // temp dir
	if err != nil {
		belogs.Error("ParseValidateFile(): TempDir fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	defer os.RemoveAll(tmpDir)
	belogs.Debug("ParseValidateFile(): tmpDir:", tmpDir)

	receiveFile, err := ginserver.ReceiveFile(c, tmpDir)
	if err != nil {
		belogs.Error("ParseValidateFile(): ReceiveFile fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("ParseValidateFile(): ReceiveFile, receiveFile:", receiveFile)

	certType, certModel, stateModel, fileHash, err := parsevalidatecore.ParseValidateFile(receiveFile)
	if err != nil {
		belogs.Error("ParseValidateFile(): ParseValidateFile fail: receiveFile:", receiveFile, err, "  time(s):", time.Since(start))
		ginserver.ResponseFail(c, err, "")
		return
	}
	parseCertResponse := model.ParseCertResponse{
		CertType:   certType,
		FileHash:   fileHash,
		CertModel:  certModel,
		StateModel: stateModel,
	}
	belogs.Debug("ParseValidateFile(): parseCertResponse:", jsonutil.MarshalJson(parseCertResponse), "  time(s):", time.Since(start))
	ginserver.ResponseOk(c, parseCertResponse)
}

// upload file to parse;
// only one file
func ParseFile(c *gin.Context) {
	start := time.Now()
	tmpDir, err := ioutil.TempDir("", "ParseFile") // temp dir
	if err != nil {
		belogs.Error("ParseFile(): TempDir fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	defer os.RemoveAll(tmpDir)
	belogs.Debug("ParseFile(): tmpDir:", tmpDir)

	receiveFile, err := ginserver.ReceiveFile(c, tmpDir)
	if err != nil {
		belogs.Error("ParseFile(): ReceiveFile fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("ParseFile(): ReceiveFile, receiveFile:", receiveFile)

	certModel, err := parsevalidatecore.ParseFile(receiveFile)
	if err != nil {
		belogs.Error("ParseFile(): ParseFile: err:", receiveFile, err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Info("ParseFile(): ok, certModel:", jsonutil.MarshallJsonIndent(certModel),
		"  time(s):", time.Since(start))
	ginserver.ResponseOk(c, certModel)
}

// upload file to parse to get ca repo
func ParseFileSimple(c *gin.Context) {
	start := time.Now()
	tmpDir, err := ioutil.TempDir("", "ParseFileSimple") // temp dir
	if err != nil {
		belogs.Error("ParseFileSimple(): TempDir fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	defer os.RemoveAll(tmpDir)
	belogs.Debug("ParseFileSimple(): tmpDir:", tmpDir)

	receiveFile, err := ginserver.ReceiveFile(c, tmpDir)
	if err != nil {
		belogs.Error("ParseFileSimple(): ReceiveFile fail:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("ParseFileSimple(): ReceiveFile, receiveFile:", receiveFile)

	parseCerSimple, err := parsevalidatecore.ParseFileSimple(receiveFile)
	if err != nil {
		belogs.Error("ParseFileSimple(): ParseFileSimple: err:", err)
		ginserver.ResponseFail(c, err, "")
		return
	}
	belogs.Debug("ParseFileSimple():ok, parseCerSimple:", jsonutil.MarshalJson(parseCerSimple))
	belogs.Info("ParseFileSimple():ok, parseCerSimple.RpkiNotify:", parseCerSimple.RpkiNotify,
		" parseCerSimple.CaRepository:", parseCerSimple.CaRepository, "   time(s):", time.Since(start))

	ginserver.ResponseOk(c, parseCerSimple)
}
