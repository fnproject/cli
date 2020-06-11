package adapter

type OCIClient struct {
	ociFn      *OCIFnClient
	ociApp     *OCIAppClient
	ociTrigger *OCITriggerClient
}

func (oci *OCIClient) GetFnsClient() FnClient {
	return oci.ociFn
}

func (oci *OCIClient) GetAppsClient() AppClient {
	return oci.ociApp
}

func (oci *OCIClient) GetTriggersClient() TriggerClient {
	return oci.ociTrigger
}
