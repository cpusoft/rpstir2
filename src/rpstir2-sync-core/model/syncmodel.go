package model

import (
	"time"

	model "rpstir2-model"
)

type SyncState struct {
	SyncStyle string `json:"syncStyle"`

	StartTime time.Time `json:"startTime,omitempty"`
	EndTime   time.Time `json:"endTime,omitempty"`

	SyncUrls   []string         `json:"syncUrls"`
	SyncResult model.SyncResult `json:"syncResult"`
}
