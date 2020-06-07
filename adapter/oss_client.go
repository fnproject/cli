package adapter

type ossClient struct {
	ossFn 		*ossFnClient
	ossApp 		*ossAppClient
	ossTrigger 	*ossTriggerClient
}

func (oss ossClient) getFnClient() FnClient {
	return oss.ossFn
}

func (oss ossClient) getAppClient() AppClient {
	return oss.ossApp
}

func (oss ossClient) getTriggerClient() TriggerClient {
	return oss.ossTrigger
}