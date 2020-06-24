package adapter

type OSSClient struct {
	ossFn      *OSSFnClient
	ossApp     *OSSAppClient
	ossTrigger *OSSTriggerClient
}

func (oss *OSSClient) FnClient() FnClient {
	return oss.ossFn
}

func (oss *OSSClient) AppClient() AppClient {
	return oss.ossApp
}

func (oss *OSSClient) TriggerClient() TriggerClient {
	return oss.ossTrigger
}
