package model

import (
	"errors"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
)

type CompileParam struct {
	Rp          []string `json:"rp"`
	Vc          []string `json:"vc"`
	Distributed []string `json:"distributed"`
}

func NewCompileParam(compileParamStr string) (*CompileParam, error) {
	if len(compileParamStr) == 0 {
		belogs.Error("NewCompileParam():compileParam is empty")
		return nil, errors.New("compileParam is empty")
	}
	compileParam := &CompileParam{}
	err := jsonutil.UnmarshalJson(compileParamStr, compileParam)
	if err != nil {
		belogs.Error("NewCompileParam(): UnmarshalJson fail, compileParamStr:", compileParamStr)
		return nil, err
	}
	belogs.Debug("NewCompileParam():compileParam:", jsonutil.MarshalJson(compileParam))
	return compileParam, nil
}

func FoundProgram(compileParam *CompileParam, programName string) bool {
	if programName == "rp" {
		return len(compileParam.Rp) > 0
	} else if programName == "vc" {
		return len(compileParam.Vc) > 0
	}
	return false
}

func FoundProgramModel(compileParam *CompileParam, programName, programModelDetail string) bool {
	if programName == "rp" {
		for _, d := range compileParam.Rp {
			if d == programModelDetail {
				return true
			}
		}
		return false
	} else if programName == "vc" {
		for _, d := range compileParam.Vc {
			if d == programModelDetail {
				return true
			}
		}
		return false
	}
	return false
}
