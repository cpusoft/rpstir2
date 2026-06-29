package model

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

type TestConnectModel struct {
	RepoUrl     string `json:"repoUrl"`
	PreceptNode string `json:"preceptNode"`
}
type DownloadConnectModel struct {
	RrdpNotifyUrl string `json:"rrdpNotifyUrl"`
	PreceptNode   string `json:"preceptNode"`
}
type RepoState struct {
	State        string        `json:"state"` // valid/invalid
	DurationTime time.Duration `json:"durationTime"`
	RepoUrl      string        `json:"repoUrl"`
	PreceptNode  string        `json:"preceptNode"`
	ErrMsg       string        `json:"errMsg"` // error
}

func NewRepoState(state string, durationTime time.Duration,
	repoUrl, preceptNode string, err error) *RepoState {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	return &RepoState{
		State:        state,
		RepoUrl:      repoUrl,
		PreceptNode:  preceptNode,
		DurationTime: durationTime,
		ErrMsg:       errMsg,
	}
}

type RepoStateList struct {
	RepoStates []*RepoState `json:"repoStates"`
}

func NewRepoStateList() *RepoStateList {
	c := &RepoStateList{}
	c.RepoStates = make([]*RepoState, 0)
	return c
}
func (c *RepoStateList) Add(repoState *RepoState) {
	c.RepoStates = append(c.RepoStates, repoState)
}

type RepoStateListMap struct {
	RepoStatess map[string]*RepoStateList `json:"repoStatess"`
}

func NewRepoStateListMap() RepoStateListMap {
	c := RepoStateListMap{}
	c.RepoStatess = make(map[string]*RepoStateList)
	return c
}

func (c *RepoStateListMap) Adds(repoStates []*RepoState) {
	belogs.Debug("RepoStateListMap.Adds(): len(repoStates):", len(repoStates))
	for i := range repoStates {
		receptNode := repoStates[i].PreceptNode
		belogs.Debug("RepoStateListMap.Adds(): repoState:", jsonutil.MarshalJson(repoStates[i]))
		if receptNode == "localhost" {
			continue
		}
		rss, ok := c.RepoStatess[receptNode]
		belogs.Debug("RepoStateListMap.Adds(): get rss from RepoStatess by receptNode, ",
			"   receptNode:", receptNode, "  rss:", jsonutil.MarshalJson(rss), "  ok:", ok)
		if ok {
			rss.Add(repoStates[i])
		} else {
			rss = NewRepoStateList()
			rss.Add(repoStates[i])
		}
		c.RepoStatess[receptNode] = rss
	}
	belogs.Debug("RepoStateListMap.Adds(): len(RepoStatess):", len(c.RepoStatess))

}
