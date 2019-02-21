package provider

import (
	"context"
	"strings"
)

type ridKey string

const (
	//CfgFnAPIURL is a config key used as the default URL for resolving the API server - different providers may generate URLs in their own way
	CfgFnAPIURL      = "api-url"
	CfgFnToken       = "token"
	contextRequestID = ridKey("request-id")
)

// ConfigSource abstracts  loading configuration keys from an underlying configuration system such as Viper
type ConfigSource interface {
	GetString(string) string
	GetBool(string) bool
	IsSet(string) bool
}

type mapConfigSource map[string]string

//NewConfigSourceFromMap creates a config source from a map of literal string
func NewConfigSourceFromMap(m map[string]string) ConfigSource {
	return mapConfigSource(m)
}
func (m mapConfigSource) GetString(key string) string {
	return m[key]
}

func (m mapConfigSource) GetBool(key string) bool {
	v := m[key]
	return strings.ToLower(v) == "true"
}

func (m mapConfigSource) IsSet(key string) bool {
	_, ok := m[key]
	return ok
}

// WithRequestID sets the request-id value within the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, contextRequestID, requestID)
}

// GetRequestID returns the request-id from the context
func GetRequestID(ctx context.Context) string {
	requestID, ok := ctx.Value(contextRequestID).(string)
	if ok {
		return requestID
	}

	return ""
}
