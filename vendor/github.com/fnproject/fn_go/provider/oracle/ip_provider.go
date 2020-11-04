package oracle

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/oracle/oci-go-sdk/v27/common"
	"github.com/oracle/oci-go-sdk/v27/common/auth"

	"github.com/fnproject/fn_go/provider"
)

func NewIPProvider(configSource provider.ConfigSource, passphraseSource provider.PassPhraseSource) (provider.Provider, error) {
	// Set OCI SDK to use IMDS to fetch region info (second-level domain etc.)
	common.EnableInstanceMetadataServiceLookup()

	ip, err := auth.InstancePrincipalConfigurationProvider()
	if err != nil {
		return nil, err
	}

	// If we have an explicit api-url configured then use that, otherwise compute the url from the standard
	// production url template and the configured region from environment.
	cfgApiUrl := configSource.GetString(provider.CfgFnAPIURL)
	if cfgApiUrl == "" {
		region, err := ip.Region()
		if err != nil {
			return nil, err
		}
		domain, err := GetRealmDomain()
		if err != nil {
			return nil, err
		}
		cfgApiUrl = fmt.Sprintf(FunctionsAPIURLTmpl, region, domain)
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
			return nil, fmt.Errorf("problem fetching compartment OCID from metadata endpoint %s", err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("problem fetching compartment OCID from metadata endpoint %s", err)
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
