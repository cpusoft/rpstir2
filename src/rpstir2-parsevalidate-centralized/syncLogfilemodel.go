package parsevalidatecentralized

import (
	model "rpstir2-model"
)

type SyncLogFileModels struct {
	SyncLogId                  uint64                   `json:"syncLogId"`
	UpdateCerSyncLogFileModels []model.SyncLogFileModel `json:"updateCerSyncLogFileModels"`
	DelCerSyncLogFileModels    []model.SyncLogFileModel `json:"delCerSyncLogFileModels"`

	UpdateMftSyncLogFileModels []model.SyncLogFileModel `json:"updateMftSyncLogFileModels"`
	DelMftSyncLogFileModels    []model.SyncLogFileModel `json:"delMftSyncLogFileModels"`

	UpdateCrlSyncLogFileModels []model.SyncLogFileModel `json:"updateCrlSyncLogFileModels"`
	DelCrlSyncLogFileModels    []model.SyncLogFileModel `json:"delCrlSyncLogFileModels"`

	UpdateRoaSyncLogFileModels []model.SyncLogFileModel `json:"updateRoaSyncLogFileModels"`
	DelRoaSyncLogFileModels    []model.SyncLogFileModel `json:"delRoaSyncLogFileModels"`

	UpdateAsaSyncLogFileModels []model.SyncLogFileModel `json:"updateAsaSyncLogFileModels"`
	DelAsaSyncLogFileModels    []model.SyncLogFileModel `json:"delAsaSyncLogFileModels"`
}

func NewSyncLogFileModels(syncLogId uint64, dbSyncLogFileModels []model.SyncLogFileModel) *SyncLogFileModels {
	syncLogFileModels := &SyncLogFileModels{}
	syncLogFileModels.SyncLogId = syncLogId

	syncLogFileModels.UpdateCerSyncLogFileModels = make([]model.SyncLogFileModel, 0)
	syncLogFileModels.DelCerSyncLogFileModels = make([]model.SyncLogFileModel, 0)

	syncLogFileModels.UpdateMftSyncLogFileModels = make([]model.SyncLogFileModel, 0)
	syncLogFileModels.DelMftSyncLogFileModels = make([]model.SyncLogFileModel, 0)

	syncLogFileModels.UpdateCrlSyncLogFileModels = make([]model.SyncLogFileModel, 0)
	syncLogFileModels.DelCrlSyncLogFileModels = make([]model.SyncLogFileModel, 0)

	syncLogFileModels.UpdateRoaSyncLogFileModels = make([]model.SyncLogFileModel, 0)
	syncLogFileModels.DelRoaSyncLogFileModels = make([]model.SyncLogFileModel, 0)

	syncLogFileModels.UpdateAsaSyncLogFileModels = make([]model.SyncLogFileModel, 0)
	syncLogFileModels.DelAsaSyncLogFileModels = make([]model.SyncLogFileModel, 0)

	for i := range dbSyncLogFileModels {
		if dbSyncLogFileModels[i].FileType == "cer" {
			if dbSyncLogFileModels[i].SyncType == "add" || dbSyncLogFileModels[i].SyncType == "update" {
				syncLogFileModels.UpdateCerSyncLogFileModels = append(syncLogFileModels.UpdateCerSyncLogFileModels, dbSyncLogFileModels[i])
			} else if dbSyncLogFileModels[i].SyncType == "del" {
				syncLogFileModels.DelCerSyncLogFileModels = append(syncLogFileModels.DelCerSyncLogFileModels, dbSyncLogFileModels[i])
			}
		} else if dbSyncLogFileModels[i].FileType == "mft" {
			if dbSyncLogFileModels[i].SyncType == "add" || dbSyncLogFileModels[i].SyncType == "update" {
				syncLogFileModels.UpdateMftSyncLogFileModels = append(syncLogFileModels.UpdateMftSyncLogFileModels, dbSyncLogFileModels[i])
			} else if dbSyncLogFileModels[i].SyncType == "del" {
				syncLogFileModels.DelMftSyncLogFileModels = append(syncLogFileModels.DelMftSyncLogFileModels, dbSyncLogFileModels[i])
			}
		} else if dbSyncLogFileModels[i].FileType == "crl" {
			if dbSyncLogFileModels[i].SyncType == "add" || dbSyncLogFileModels[i].SyncType == "update" {
				syncLogFileModels.UpdateCrlSyncLogFileModels = append(syncLogFileModels.UpdateCrlSyncLogFileModels, dbSyncLogFileModels[i])
			} else if dbSyncLogFileModels[i].SyncType == "del" {
				syncLogFileModels.DelCrlSyncLogFileModels = append(syncLogFileModels.DelCrlSyncLogFileModels, dbSyncLogFileModels[i])
			}
		} else if dbSyncLogFileModels[i].FileType == "roa" {
			if dbSyncLogFileModels[i].SyncType == "add" || dbSyncLogFileModels[i].SyncType == "update" {
				syncLogFileModels.UpdateRoaSyncLogFileModels = append(syncLogFileModels.UpdateRoaSyncLogFileModels, dbSyncLogFileModels[i])
			} else if dbSyncLogFileModels[i].SyncType == "del" {
				syncLogFileModels.DelRoaSyncLogFileModels = append(syncLogFileModels.DelRoaSyncLogFileModels, dbSyncLogFileModels[i])
			}
		} else if dbSyncLogFileModels[i].FileType == "asa" {
			if dbSyncLogFileModels[i].SyncType == "add" || dbSyncLogFileModels[i].SyncType == "update" {
				syncLogFileModels.UpdateAsaSyncLogFileModels = append(syncLogFileModels.UpdateAsaSyncLogFileModels, dbSyncLogFileModels[i])
			} else if dbSyncLogFileModels[i].SyncType == "del" {
				syncLogFileModels.DelAsaSyncLogFileModels = append(syncLogFileModels.DelAsaSyncLogFileModels, dbSyncLogFileModels[i])
			}
		}
	}
	return syncLogFileModels
}
