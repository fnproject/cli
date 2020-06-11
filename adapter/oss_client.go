package adapter

type OSSClient struct {
	ossFn      *OSSFnClient
	ossApp     *OSSAppClient
	ossTrigger *OSSTriggerClient
}

func (oss *OSSClient) GetFnsClient() FnClient {
	return oss.ossFn
}

func (oss *OSSClient) GetAppsClient() AppClient {
	return oss.ossApp
}

func (oss *OSSClient) GetTriggersClient() TriggerClient {
	return oss.ossTrigger
}
