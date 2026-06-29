package rtrserver

import (
	"time"

	rtrcore "github.com/bgpsecurity/rpstir2/rtr-core"
	"github.com/cpusoft/goutil/belogs"
)

func processResetQuery(rtrPduModel rtrcore.RtrPduModel) (resetResponses []rtrcore.RtrPduModel, err error) {
	start := time.Now()
	rtrFulls, rtrAsaFulls, rtrHroaFulls, rtrAsarFulls, sessionId, serialNumber, err := getRtrXFullDb()
	if err != nil {
		belogs.Error("processResetQuery(): GetRtrFullAndSerialNumAndSessionId fail: ", err)
		return resetResponses, err
	}
	belogs.Debug("processResetQuery(): len(rtrFulls):", len(rtrFulls),
		"  len(rtrAsaFulls):", len(rtrAsaFulls),
		"  len(rtrHroaFulls):", len(rtrHroaFulls),
		"  len(rtrAsarFulls):", len(rtrAsarFulls),
		"  sessionId:", sessionId, "  serialNumber: ", serialNumber, "  time(s):", time.Since(start))
	rtrPduModels, err := rtrcore.AssembleResetResponses(rtrFulls,
		rtrAsaFulls, rtrHroaFulls, rtrAsarFulls,
		rtrPduModel.GetProtocolVersion(), sessionId, serialNumber)
	if err != nil {
		belogs.Error("processResetQuery(): AssembleResetResponses fail: ", err)
		return resetResponses, err
	}
	return rtrPduModels, nil
}
