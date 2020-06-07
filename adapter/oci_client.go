package adapter

type ociClient struct{
	ociFn 		*ociFnClient
	ociApp 		*ociAppClient
	ociTrigger 	*ociTriggerClient
}

func (oci *ociClient) getFnClient() FnClient {
	return oci.ociFn
}

func (oci *ociClient) getAppClient() AppClient {
	return oci.ociApp
}

func (oci *ociClient) getTriggerClient() TriggerClient {
	return oci.ociTrigger
}
