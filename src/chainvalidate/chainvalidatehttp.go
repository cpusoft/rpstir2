package chainvalidate

import (
	"github.com/cpusoft/goutil/belogs"
	conf "github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/ginserver"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/gin-gonic/gin"
)

// ChainValidatePartialStart  开始证书链验证接口（支持局部验证）
//
//	@param c 服务端上下文信息
func ChainValidatePartialStart(c *gin.Context) {
	belogs.Info("ChainValidatePartialStart(): use partial chainvalidate")
	chainValidateHttpStart(c, false)
}

// ChainValidateStart 开始证书链验证接口
//
//	@param c 服务端上下文信息
func ChainValidateStart(c *gin.Context) {
	belogs.Info("ChainValidateStart(): use global chainvalidate")
	chainValidateHttpStart(c, true)
}

// chainValidateHttpStart 实际执行证书链验证
//
//	@param c 服务端上下文信息
//	@param isGlobal 是否支持证书链局部验证
func chainValidateHttpStart(c *gin.Context, isGlobal bool) {
	rpstir2RpUrl := "https://" + conf.String("rpstir2-rp::serverHost") + ":" + conf.String("rpstir2-rp::serverHttpsPort")
	rpstir2VcUrl := "https://" + conf.String("rpstir2-vc::serverHost") + ":" + conf.String("rpstir2-vc::serverHttpsPort")
	belogs.Info("chainValidateHttpStart(): start,  rpstir2RpUrl:", rpstir2RpUrl, "   rpstir2VcUrl:", rpstir2VcUrl, "  isGlobal:", isGlobal, "  usecache:", UseCache)

	//check serviceState
	_, _, err := httpclient.Post(rpstir2RpUrl+"/sys/servicestate", `{"operate":"enter","state":"chainvalidate"}`, false)
	if err != nil {
		belogs.Error("chainValidateHttpStart(): servicestate enter chainvalidate fail, rpstir2RpUrl:", rpstir2RpUrl, err)
	}
	go func() {
		nextStep, err := chainValidateStart(isGlobal)
		belogs.Debug("chainValidateHttpStart(): chainValidateStart end,  nextStep is:", nextStep, err)
		// leave serviceState
		if err != nil {
			// will end this whole sync
			belogs.Error("chainValidateHttpStart(): chainValidateStart fail", err)
			_, _, err = httpclient.Post(rpstir2RpUrl+"/sys/servicestate", `{"operate":"leave","state":"end"}`, false)
			if err != nil {
				belogs.Error("chainValidateHttpStart(): chainValidateStart fail, servicestate leave end fail, rpstir2RpUrl:", rpstir2RpUrl, err)
			}
		} else {
			// leave serviceState
			_, _, err = httpclient.Post(rpstir2RpUrl+"/sys/servicestate", `{"operate":"leave","state":"chainvalidate"}`, false)
			if err != nil {
				belogs.Error("chainValidateHttpStart(): chainValidateStart ok, servicestate leave chainvalidate fail, rpstir2RpUrl:", rpstir2RpUrl, err)
			}
			// rtr producer
			go func() {
				_, _, err = httpclient.Post(rpstir2VcUrl+"/rtrproducer/updatefromsync", `{"lastStep":"chainValidateStart"}`, false)
				if err != nil {
					//  there is may be no vc. so, just leave
					belogs.Info("chainValidateHttpStart(): post to vc's /rtrproducer/ no response, there is may be no vc. ", err)
					_, _, err = httpclient.Post(rpstir2RpUrl+"/sys/servicestate", `{"operate":"leave","state":"end"}`, false)
					if err != nil {
						belogs.Error("chainValidateHttpStart(): after updatefromsync fail, servicestate leave end fail, rpstir2RpUrl:", rpstir2RpUrl, err)
					}
				}
			}()

			// call statistics
			go httpclient.Post(rpstir2RpUrl+"/statistic/start", "", false)

			// call roacompete
			go httpclient.Post(rpstir2RpUrl+"/roacompete/start", "", false)

			// call roahistory
			//go httpclient.Post(rpstir2RpUrl+"/roahistory/start", "", false)
			belogs.Info("ChainValidateStart(): end,  nextStep is :", nextStep)
		}
	}()
	ginserver.ResponseOk(c, nil)
}
