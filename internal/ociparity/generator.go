package ociparity

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type GeneratedField struct {
	FlagName    string
	Property    string
	Description string
	Kind        string
}

type AppGenerationModel struct {
	CreateUpdateFields []GeneratedField
	ListFields         []GeneratedField
}

type FnGenerationModel struct {
	CreateUpdateFields []GeneratedField
	ListFields         []GeneratedField
}

func BuildAppModel(spec *Spec) (*AppGenerationModel, error) {
	createSchema, err := spec.Schema("CreateApplicationDetails")
	if err != nil {
		return nil, err
	}
	updateSchema, err := spec.Schema("UpdateApplicationDetails")
	if err != nil {
		return nil, err
	}
	createSchema, err = ParseObjectOrRef(spec, createSchema)
	if err != nil {
		return nil, err
	}
	updateSchema, err = ParseObjectOrRef(spec, updateSchema)
	if err != nil {
		return nil, err
	}
	fields := map[string]GeneratedField{}
	collect := func(schema map[string]interface{}) {
		props, _ := schema["properties"].(map[string]interface{})
		for _, propName := range []string{"traceConfig", "networkSecurityGroupIds", "imagePolicyConfig", "securityAttributes"} {
			if propNode, ok := props[propName].(map[string]interface{}); ok {
				kind := "string"
				switch propName {
				case "traceConfig", "imagePolicyConfig", "securityAttributes":
					kind = "json"
				case "networkSecurityGroupIds":
					kind = "string-slice"
				}
				flagName := toFlagName(propName)
				fields[propName] = GeneratedField{
					FlagName:    flagName,
					Property:    propName,
					Description: normalizeDescription(propNode["description"]),
					Kind:        kind,
				}
			}
		}
	}
	collect(createSchema)
	collect(updateSchema)
	ordered := make([]GeneratedField, 0, len(fields))
	for _, f := range fields {
		ordered = append(ordered, f)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].FlagName < ordered[j].FlagName })
	listFields := []GeneratedField{
		{FlagName: "display-name", Property: "displayName", Description: "Filter applications by exact display name", Kind: "string"},
		{FlagName: "id", Property: "id", Description: "Filter applications by OCID", Kind: "string"},
		{FlagName: "lifecycle-state", Property: "lifecycleState", Description: "Filter applications by lifecycle state", Kind: "string"},
		{FlagName: "sort-by", Property: "sortBy", Description: "Sort applications by a supported field", Kind: "string"},
		{FlagName: "sort-order", Property: "sortOrder", Description: "Sort order for list results", Kind: "string"},
	}
	return &AppGenerationModel{CreateUpdateFields: ordered, ListFields: listFields}, nil
}

func toFlagName(property string) string {
	var out []rune
	for i, r := range property {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out = append(out, '-')
		}
		if r >= 'A' && r <= 'Z' {
			r = r + ('a' - 'A')
		}
		out = append(out, r)
	}
	return string(out)
}

func GenerateAppFiles(specPath string) (map[string]string, error) {
	spec, err := LoadSpec(specPath)
	if err != nil {
		return nil, err
	}
	model, err := BuildAppModel(spec)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		filepath.ToSlash("objects/app/generated_oci_parity_flags.go"):                                          renderAppFlags(model),
		filepath.ToSlash("objects/app/generated_oci_parity_apply.go"):                                          renderAppApply(model),
		filepath.ToSlash("objects/app/generated_oci_parity_list.go"):                                           renderAppList(model),
		filepath.ToSlash("vendor/github.com/fnproject/fn_go/provider/oracle/shim/generated_app_oci_parity.go"): renderAppShim(model),
	}, nil
}

func BuildFnModel(spec *Spec) (*FnGenerationModel, error) {
	createSchema, err := spec.Schema("CreateFunctionDetails")
	if err != nil {
		return nil, err
	}
	updateSchema, err := spec.Schema("UpdateFunctionDetails")
	if err != nil {
		return nil, err
	}
	createSchema, err = ParseObjectOrRef(spec, createSchema)
	if err != nil {
		return nil, err
	}
	updateSchema, err = ParseObjectOrRef(spec, updateSchema)
	if err != nil {
		return nil, err
	}
	fields := map[string]GeneratedField{}
	collect := func(schema map[string]interface{}) {
		props, _ := schema["properties"].(map[string]interface{})
		for _, propName := range []string{"traceConfig", "timeoutInSeconds", "detachedModeTimeoutInSeconds", "successDestination", "failureDestination"} {
			if propNode, ok := props[propName].(map[string]interface{}); ok {
				kind := "string"
				switch propName {
				case "traceConfig", "successDestination", "failureDestination":
					kind = "json"
				case "timeoutInSeconds", "detachedModeTimeoutInSeconds":
					kind = "int"
				}
				flagName := toFlagName(propName)
				fields[propName] = GeneratedField{
					FlagName:    flagName,
					Property:    propName,
					Description: normalizeDescription(propNode["description"]),
					Kind:        kind,
				}
			}
		}
	}
	collect(createSchema)
	collect(updateSchema)
	ordered := make([]GeneratedField, 0, len(fields))
	for _, f := range fields {
		ordered = append(ordered, f)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].FlagName < ordered[j].FlagName })
	listFields := []GeneratedField{
		{FlagName: "display-name", Property: "displayName", Description: "Filter functions by exact display name", Kind: "string"},
		{FlagName: "id", Property: "id", Description: "Filter functions by OCID", Kind: "string"},
		{FlagName: "lifecycle-state", Property: "lifecycleState", Description: "Filter functions by lifecycle state", Kind: "string"},
		{FlagName: "sort-by", Property: "sortBy", Description: "Sort functions by a supported field", Kind: "string"},
		{FlagName: "sort-order", Property: "sortOrder", Description: "Sort order for list results", Kind: "string"},
	}
	return &FnGenerationModel{CreateUpdateFields: ordered, ListFields: listFields}, nil
}

func GenerateFnFiles(specPath string) (map[string]string, error) {
	spec, err := LoadSpec(specPath)
	if err != nil {
		return nil, err
	}
	model, err := BuildFnModel(spec)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		filepath.ToSlash("objects/fn/generated_oci_parity_flags.go"):                                          renderFnFlags(model),
		filepath.ToSlash("objects/fn/generated_oci_parity_apply.go"):                                          renderFnApply(model),
		filepath.ToSlash("objects/fn/generated_oci_parity_list.go"):                                           renderFnList(model),
		filepath.ToSlash("vendor/github.com/fnproject/fn_go/provider/oracle/shim/generated_fn_oci_parity.go"): renderFnShim(model),
	}, nil
}

func WriteGeneratedFiles(root, specPath string) error {
	files, err := GenerateAppFiles(specPath)
	if err != nil {
		return err
	}
	fnFiles, err := GenerateFnFiles(specPath)
	if err != nil {
		return err
	}
	for k, v := range fnFiles {
		files[k] = v
	}
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func normalizeDescription(value interface{}) string {
	if value == nil {
		return ""
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "<nil>" {
		return ""
	}
	return text
}
