package ociparity

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Spec struct {
	Root map[string]interface{}
}

func LoadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	normalized := normalizeYAML(raw)
	root, ok := normalized.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("spec root is not an object")
	}
	return &Spec{Root: root}, nil
}

func normalizeYAML(v interface{}) interface{} {
	switch t := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(t))
		for k, v2 := range t {
			m[fmt.Sprint(k)] = normalizeYAML(v2)
		}
		return m
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, item := range t {
			out[i] = normalizeYAML(item)
		}
		return out
	default:
		return v
	}
}

func (s *Spec) ResolveRef(ref string) (map[string]interface{}, error) {
	if !strings.HasPrefix(ref, "#/") {
		return nil, fmt.Errorf("unsupported ref %q", ref)
	}
	parts := strings.Split(strings.TrimPrefix(ref, "#/"), "/")
	var cur interface{} = s.Root
	for _, part := range parts {
		obj, ok := cur.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid ref path %q", ref)
		}
		cur, ok = obj[part]
		if !ok {
			return nil, fmt.Errorf("ref not found %q", ref)
		}
	}
	obj, ok := cur.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("ref %q does not point to an object", ref)
	}
	return obj, nil
}

func (s *Spec) Schema(name string) (map[string]interface{}, error) {
	if components, ok := s.Root["components"].(map[string]interface{}); ok {
		if schemas, ok := components["schemas"].(map[string]interface{}); ok {
			if entry, ok := schemas[name].(map[string]interface{}); ok {
				return entry, nil
			}
		}
	}
	if definitions, ok := s.Root["definitions"].(map[string]interface{}); ok {
		if entry, ok := definitions[name].(map[string]interface{}); ok {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("schema %q not found in components.schemas or definitions", name)
}

func ParseObjectOrRef(s *Spec, node map[string]interface{}) (map[string]interface{}, error) {
	if ref, ok := node["$ref"].(string); ok {
		return s.ResolveRef(ref)
	}
	return node, nil
}

func marshalNormalized(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
