package parsevalidatecore

import (
	"fmt"
	"testing"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/jsonutil"
)

func TestParseValidateSig(t *testing.T) {

	file := `../../../go-study/src/asn1sig2/checklist.sig`
	fileModel := &model.FileModel{
		FilePathName:     file,
		TempFilePathName: file,
	}
	sigModel, stateModel, err := parseValidateSig(fileModel)
	fmt.Println(jsonutil.MarshalJson(sigModel), jsonutil.MarshalJson(stateModel), err)

}
