package openssl

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	model "rpstir2-model"
)

func convertAsProviderAttestationToCustomerAsns(asProviderAttestation AsProviderAttestation) (customerAsns []model.CustomerAsn, err error) {
	belogs.Debug("convertAsProviderAttestationToCustomerAsns(): asProviderAttestation:", jsonutil.MarshalJson(asProviderAttestation))

	customerAsns = make([]model.CustomerAsn, 0)
	customerAsn := model.CustomerAsn{}
	customerAsn.Version = uint64(asProviderAttestation.Version.Version)
	customerAsn.CustomerAsn = uint64(asProviderAttestation.CustomerAsn)
	providerAsns := make([]uint64, 0)
	for i := range asProviderAttestation.ProviderAsns {
		providerAsns = append(providerAsns, uint64(asProviderAttestation.ProviderAsns[i]))
	}
	customerAsn.ProviderAsns = providerAsns
	customerAsns = append(customerAsns, customerAsn)
	belogs.Debug("convertAsProviderAttestationToCustomerAsns(): customerAsns:", jsonutil.MarshalJson(customerAsns))

	return customerAsns, nil
}
