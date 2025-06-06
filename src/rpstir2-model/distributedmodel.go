package model

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/rrdputil"
	"github.com/cpusoft/goutil/urlutil"
)

type DistributedSnapshotModel struct {
	Id          int    `json:"id"  xorm:"id int"`
	NotifyUrl   string `json:"notifyUrl" xorm:"notifyUrl varchar(512)"`
	SnapshotUrl string `json:"snapshotUrl" xorm:"snapshotUrl varchar(512)"`
	MaxSerial   uint64 `json:"maxSerial" xorm:"maxSerial int"`

	CenterNodeUrl string `json:"centerNodeUrl"`
	Index         uint64 `json:"index"`
	SyncLogId     uint64 `json:"syncLogId"`
}

// result
type DistributedSnapshotCountResult struct {
	NotifyUrl   string `json:"notifyUrl"`
	SnapshotUrl string `json:"snapshotUrl"`
	SessionId   string `json:"sessionId"`
	SyncLogId   uint64 `json:"syncLogId"`

	Serial       uint64    `json:"serial"`
	PublishCount uint64    `json:"publishCount"`
	SyncTime     time.Time `json:"syncTime"`

	// if not "", means has error
	ErrMsg string `json:"errMsg"`
}

func (c DistributedSnapshotCountResult) String() string {
	return jsonutil.MarshalJson(c)
}

func NewDistributedSnapshotCountResult(distributedSnapshotModel *DistributedSnapshotModel,
	sessionId string, publishCount uint64, syncTime time.Time) *DistributedSnapshotCountResult {

	return &DistributedSnapshotCountResult{
		NotifyUrl:    distributedSnapshotModel.NotifyUrl,
		SnapshotUrl:  distributedSnapshotModel.SnapshotUrl,
		SessionId:    sessionId,
		SyncLogId:    distributedSnapshotModel.SyncLogId,
		Serial:       distributedSnapshotModel.MaxSerial,
		PublishCount: publishCount,
		SyncTime:     syncTime,
	}
}

func NewDistributedSnapshotCountErrorResult(distributedSnapshotModel *DistributedSnapshotModel,
	errMsg string, syncTime time.Time) *DistributedSnapshotCountResult {

	return &DistributedSnapshotCountResult{
		NotifyUrl:   distributedSnapshotModel.NotifyUrl,
		SnapshotUrl: distributedSnapshotModel.SnapshotUrl,
		SyncLogId:   distributedSnapshotModel.SyncLogId,
		Serial:      distributedSnapshotModel.MaxSerial,
		SyncTime:    syncTime,
		ErrMsg:      errMsg,
	}
}

type DistributedDeltaModel struct {
	Id        int    `json:"id"  xorm:"id int"`
	NotifyUrl string `json:"notifyUrl" xorm:"notifyUrl varchar(512)"`
	DeltaUrl  string `json:"deltaUrl" xorm:"deltaUrl varchar(512)"`
	Serial    uint64 `json:"serial" xorm:"serial int"`

	CenterNodeUrl string `json:"centerNodeUrl"`
	Index         uint64 `json:"index"`
	SyncLogId     uint64 `json:"syncLogId"`
}

type DistributedDeltaCountResults struct {
	DistributedDeltaCountResults []DistributedDeltaCountResult `json:"distributedDeltaCountResults"`
}

func (c DistributedDeltaCountResults) String() string {
	return jsonutil.MarshalJson(c)
}

type DistributedDeltaCountResult struct {
	NotifyUrl string `json:"notifyUrl"`
	DeltaUrl  string `json:"deltaUrl"`
	SessionId string `json:"sessionId"`
	SyncLogId uint64 `json:"syncLogId"`

	Serial        uint64    `json:"serial"`
	PublishCount  uint64    `json:"publishCount"`
	WithdrawCount uint64    `json:"withdrawCount"`
	SyncTime      time.Time `json:"syncTime"`

	// if not "", means has error
	ErrMsg string `json:"errMsg"`
}

func (c DistributedDeltaCountResult) String() string {
	return jsonutil.MarshalJson(c)
}
func NewDistributedDeltaCountResult(distributedDeltaModel *DistributedDeltaModel,
	sessionId string, serial, publishCount, withdrawCount uint64, syncTime time.Time) *DistributedDeltaCountResult {
	return &DistributedDeltaCountResult{
		NotifyUrl:     distributedDeltaModel.NotifyUrl,
		DeltaUrl:      distributedDeltaModel.DeltaUrl,
		SessionId:     sessionId,
		SyncLogId:     distributedDeltaModel.SyncLogId,
		Serial:        serial,
		PublishCount:  publishCount,
		WithdrawCount: withdrawCount,
		SyncTime:      syncTime,
	}
}

func NewDistributedDeltaCountErrorResult(distributedDeltaModel *DistributedDeltaModel,
	errMsg string, syncTime time.Time) *DistributedDeltaCountResult {
	return &DistributedDeltaCountResult{
		NotifyUrl: distributedDeltaModel.NotifyUrl,
		DeltaUrl:  distributedDeltaModel.DeltaUrl,
		SyncLogId: distributedDeltaModel.SyncLogId,
		SyncTime:  syncTime,
		ErrMsg:    errMsg,
	}
}

// publish result (include from snapshot or delta)
type DistributedPublishWithdrawResult struct {
	//
	NotifyUrl string `json:"notifyUrl"`
	SessionId string `json:"sessionId"`
	SyncLogId uint64 `json:"syncLogId"`

	//
	SnapshotOrDeltaUrl   string `json:"snapshotOrDeltaUrl"` // snapshot or delta
	IsSnapshot           bool   `json:"isSnapshot"`
	PublishOrWithdrawUrl string `json:"publishOrWithdrawUrl"` // publish or withdraw
	IsPublish            bool   `json:"isPublish"`
	Serial               uint64 `json:"serial"`
	Index                uint64 `json:"index"`
	CenterNodeUrl        string `json:"centerNodeUrl"`
	//
	FilePathName string      `json:"filePathName"`
	FileHash     string      `json:"fileHash"`
	Base64       string      `json:"base64"`
	CertModel    interface{} `json:"certModel"`
	StateModel   StateModel  `json:"stateModel"`
	OriginModel  OriginModel `json:"originModel"`
	SyncTime     time.Time   `json:"syncTime"`

	// if not "", means has error
	ErrMsg string `json:"errMsg"`
}

func (c DistributedPublishWithdrawResult) String() string {
	m := make(map[string]interface{})
	m["notifyUrl"] = c.NotifyUrl
	m["sessionId"] = c.SessionId
	m["syncLogId"] = c.SyncLogId

	m["snapshotOrDeltaUrl"] = c.SnapshotOrDeltaUrl
	m["isSnapshot"] = c.IsSnapshot
	m["publishOrWithdrawUrl"] = c.PublishOrWithdrawUrl
	m["isPublish"] = c.IsPublish
	m["serial"] = c.Serial
	m["index"] = c.Index
	m["centerNodeUrl"] = c.CenterNodeUrl

	m["filePathName"] = c.FilePathName
	m["syncTime"] = c.SyncTime
	m["errMsg"] = c.ErrMsg
	return jsonutil.MarshalJson(m)
}
func (c DistributedPublishWithdrawResult) SimpleString() string {
	m := make(map[string]interface{})
	m["notifyUrl"] = c.NotifyUrl
	m["syncLogId"] = c.SyncLogId

	m["snapshotOrDeltaUrl"] = c.SnapshotOrDeltaUrl
	m["publishOrWithdrawUrl"] = c.PublishOrWithdrawUrl
	m["serial"] = c.Serial

	m["filePathName"] = c.FilePathName
	m["syncTime"] = c.SyncTime
	m["errMsg"] = c.ErrMsg
	return jsonutil.MarshalJson(m)
}

func NewDistributedPublishWithdrawResultFromSnapshot(distributedSnapshotModel *DistributedSnapshotModel,
	snapshotPublish *rrdputil.SnapshotPublish,
	sessionId string, index uint64) (distributedPublishWithdrawResult *DistributedPublishWithdrawResult) {
	c := &DistributedPublishWithdrawResult{
		NotifyUrl: distributedSnapshotModel.NotifyUrl,
		SessionId: sessionId,
		SyncLogId: distributedSnapshotModel.SyncLogId,

		SnapshotOrDeltaUrl:   distributedSnapshotModel.SnapshotUrl,
		IsSnapshot:           true,
		PublishOrWithdrawUrl: snapshotPublish.Uri,
		IsPublish:            true,
		Serial:               distributedSnapshotModel.MaxSerial,
		Index:                index,
		CenterNodeUrl:        distributedSnapshotModel.CenterNodeUrl,
		Base64:               snapshotPublish.Base64,
		OriginModel:          *JudgeOrigin(snapshotPublish.Uri, distributedSnapshotModel.NotifyUrl),
	}
	destPath := conf.String("rrdp::destPath") + "/"
	c.FilePathName, _ = urlutil.JoinPrefixPathAndUrlFileName(destPath, snapshotPublish.Uri)
	belogs.Debug("NewDistributedPublishWithdrawResultFromSnapshot(): c:", c.String())
	return c
}
func NewDistributedPublishWithdrawResultFromDelta(distributedDeltaModel *DistributedDeltaModel,
	deltaPublish *rrdputil.DeltaPublish, deltaWithdraw *rrdputil.DeltaWithdraw,
	sessionId string, index uint64) (distributedPublishWithdrawResult *DistributedPublishWithdrawResult) {
	c := &DistributedPublishWithdrawResult{
		NotifyUrl: distributedDeltaModel.NotifyUrl,
		SessionId: sessionId,
		SyncLogId: distributedDeltaModel.SyncLogId,

		SnapshotOrDeltaUrl: distributedDeltaModel.DeltaUrl,
		IsSnapshot:         false,
		Serial:             distributedDeltaModel.Serial,
		Index:              index,
		CenterNodeUrl:      distributedDeltaModel.CenterNodeUrl,
	}

	if deltaPublish != nil {
		c.PublishOrWithdrawUrl = deltaPublish.Uri
		c.IsPublish = true
		c.Base64 = deltaPublish.Base64
		c.OriginModel = *JudgeOrigin(deltaPublish.Uri, distributedDeltaModel.NotifyUrl)
	} else if deltaWithdraw != nil {
		c.PublishOrWithdrawUrl = deltaWithdraw.Uri
		c.IsPublish = false
		c.FileHash = deltaWithdraw.Hash
		c.OriginModel = *JudgeOrigin(deltaWithdraw.Uri, distributedDeltaModel.NotifyUrl)
	}
	destPath := conf.String("rrdp::destPath") + "/"
	c.FilePathName, _ = urlutil.JoinPrefixPathAndUrlFileName(destPath, c.PublishOrWithdrawUrl)

	belogs.Debug("NewDistributedPublishWithdrawResultFromDelta(): c:", c.String())
	return c
}

func (c *DistributedPublishWithdrawResult) SetParseValidateResult(
	fileHash string, certModel interface{}, stateModel StateModel, syncTime time.Time) {
	c.FileHash = fileHash
	c.CertModel = certModel
	c.StateModel = stateModel
	c.SyncTime = syncTime
}

func (c *DistributedPublishWithdrawResult) SetErrMsg(errMsg string, syncTime time.Time) {
	c.ErrMsg = errMsg
	c.SyncTime = syncTime
}

const (
	DISTRIBUTED_RESULT_TYPE_SNAPSHOTCOUNT   = "snapshotCount"
	DISTRIBUTED_RESULT_TYPE_DELTACOUNT      = "deltaCount"
	DISTRIBUTED_RESULT_TYPE_PUBLISHWITHDRAW = "publishwithdraw"
)

type DistributedResult struct {
	ResultType string `json:"resultType"` //snapshotCount/deltaCount/publish/withdraws

	DistributedSnapshotCountResult   *DistributedSnapshotCountResult   `json:"distributedSnapshotCountResult"`
	DistributedDeltaCountResult      *DistributedDeltaCountResult      `json:"distributedDeltaCountResult"`
	DistributedPublishWithdrawResult *DistributedPublishWithdrawResult `json:"distributedPublishWithdrawResult"`
}

func (c DistributedResult) String() string {
	m := make(map[string]string)
	m["resultType"] = c.ResultType
	if c.ResultType == DISTRIBUTED_RESULT_TYPE_SNAPSHOTCOUNT {
		m["distributedSnapshotCountResult"] = c.DistributedSnapshotCountResult.String()
	} else if c.ResultType == DISTRIBUTED_RESULT_TYPE_DELTACOUNT {
		m["distributedDeltaCountResult"] = c.DistributedDeltaCountResult.String()
	} else if c.ResultType == DISTRIBUTED_RESULT_TYPE_PUBLISHWITHDRAW {
		m["distributedPublishWithdrawResult"] = c.DistributedPublishWithdrawResult.String()
	}
	return jsonutil.MarshalJson(m)
}
