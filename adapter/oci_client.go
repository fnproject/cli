package adapter

type OCIClient struct {
	ociFn      *OCIFnClient
	ociApp     *OCIAppClient
	ociTrigger *OCITriggerClient
}

func (oci *OCIClient) FnClient() FnClient {
	return oci.ociFn
}

func (oci *OCIClient) AppClient() AppClient {
	return oci.ociApp
}

func (oci *OCIClient) TriggerClient() TriggerClient {
	return oci.ociTrigger
}
