package oracle

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/oracle/oci-go-sdk/v65/functions"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"

	"github.com/fnproject/fn_go/provider"
)

func NewIPProvider(configSource provider.ConfigSource, passphraseSource provider.PassPhraseSource) (provider.Provider, error) {
	// Set OCI SDK to use IMDS to fetch region info (second-level domain etc.)
	common.EnableInstanceMetadataServiceLookup()

	configProvider, err := auth.InstancePrincipalConfigurationProvider()
	if err != nil {
		return nil, err
	}

	compartmentID := configSource.GetString(CfgCompartmentID)
	if compartmentID == "" {
		// Get the local compartment ID from the metadata endpoint
		req, err := http.NewRequest("GET", CompartmentMetadata, nil)
		if err != nil {
			return nil, fmt.Errorf("problem fetching compartment OCID from metadata endpoint %s", err)
		}

		// IMDS v2 requires authorisation header for any request
		req.Header.Add("Authorization", "Bearer Oracle")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("problem fetching compartment OCID from metadata endpoint %s", err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("problem fetching compartment OCID from metadata endpoint %s", err)
		}
		compartmentID = string(body)
	}

	ociClient, err := functions.NewFunctionsManagementClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, err
	}

	ociClient.UserAgent = fmt.Sprintf("%s %s", userAgentPrefixIp, ociClient.UserAgent)

	disableCerts := configSource.GetBool(CfgDisableCerts)
	if disableCerts {
		c := ociClient.HTTPClient.(*http.Client)
		c.Transport = InsecureRoundTripper(c.Transport)
	}

	// If we have an explicit api-url configured then use that, otherwise let OCI client compute the url from the standard
	// production url template and the configured region from environment.
	cfgApiUrl := configSource.GetString(provider.CfgFnAPIURL)
	var apiUrl *url.URL
	if cfgApiUrl != "" {
		apiUrl, err = provider.CanonicalFnAPIUrl(cfgApiUrl)
		if err != nil {
			return nil, err
		}
		ociClient.Host = apiUrl.String()
	} else {
		// Even if URL is computed by OCI SDK itself, we still populate FnApiUrl in the Provider for compatibility's sake
		apiUrl, err = provider.CanonicalFnAPIUrl(ociClient.Host)
		if err != nil {
			return nil, err
		}
	}

	return &OracleProvider{
		FnApiUrl:              apiUrl,
		Signer:                common.DefaultRequestSigner(configProvider),
		Interceptor:           nil,
		DisableCerts:          disableCerts,
		CompartmentID:         compartmentID,
		ImageCompartmentID:    configSource.GetString(CfgImageCompartmentID),
		ConfigurationProvider: configProvider,
		ociClient:             ociClient,
	}, nil
}
