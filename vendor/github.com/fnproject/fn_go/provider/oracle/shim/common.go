package shim

import "context"

const annotationCompartmentId = "oracle.com/oci/compartmentId"

// OCI update config is wholesale replacement of the map. Here we do the FnV2 server-side merge on the client instead.
// Based on https://github.com/fnproject/fn/blob/d55e01ab7d565e9796748f2f40662e94394aff07/api/models/fn.go#L274-L285
func mergeConfig(oldConfig map[string]string, changeConfig map[string]string) map[string]string {
	if changeConfig != nil {
		if oldConfig == nil {
			oldConfig = make(map[string]string)
		}
		for k, v := range changeConfig {
			if v == "" {
				delete(oldConfig, k)
			} else {
				oldConfig[k] = v
			}
		}
	}
	return oldConfig
}

// Helper func to convert nil context to context.Background
func ctxOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
