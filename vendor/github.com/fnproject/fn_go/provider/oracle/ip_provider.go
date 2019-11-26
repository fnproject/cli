package oracle

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/common/auth"

	"github.com/fnproject/fn_go/provider"
)

func NewIPProvider(configSource provider.ConfigSource, passphraseSource provider.PassPhraseSource) (provider.Provider, error) {
	ip, err := auth.InstancePrincipalConfigurationProvider()
	if err != nil {
		return nil, err
	}

	cfgApiUrl := configSource.GetString(provider.CfgFnAPIURL)
	if cfgApiUrl == "" {
		region, err := ip.Region()
		if err != nil {
			return nil, err
		}
		// Construct the API endpoint from the "nearby" endpoint
		cfgApiUrl = fmt.Sprintf(FunctionsAPIURLTmpl, region)
	}
	apiUrl, err := provider.CanonicalFnAPIUrl(cfgApiUrl)
	if err != nil {
		return nil, err
	}

	compartmentID := configSource.GetString(CfgCompartmentID)
	if compartmentID == "" {
		// Get the local compartment ID from the metadata endpoint
		resp, err := http.DefaultClient.Get(CompartmentMetadata)
		if err != nil {
			return nil, fmt.Errorf("problem fetching compartment Id from metadata endpoint %s", err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("problem fetching compartment Id from metadata endpoint %s", err)
		}
		compartmentID = string(body)
	}
	return &OracleProvider{
		FnApiUrl:      apiUrl,
		Signer:        common.DefaultRequestSigner(ip),
		Interceptor:   nil,
		DisableCerts:  configSource.GetBool(CfgDisableCerts),
		CompartmentID: compartmentID,
	}, nil
}
