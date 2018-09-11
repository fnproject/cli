package common

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
)

const LatestYamlVersion = 20180708

const V20180708Schema = `{
    "title": "V20180708 func file schema",
    "type": "object",
    "properties": {
        "name": {
            "type":"string"
        },
        "schema_version": {
            "type":"integer"
        },
        "version": {
            "type":"string"
        },
        "runtime": {
            "type":"string"
        },
        "build_image": {
            "type":"string"
        },
        "run_image": {
            "type": "string"
        },
        "entrypoint": {
            "type":"string"
        },
        "content_type": {
            "type":"string"
        },
        "cmd": {
            "type":"string"
        },
        "memory": {
            "type":"integer"
        },
        "timeout": {
            "type":"integer"
        },
        "idle_timeout": {
            "type": "integer"
        },
        "config": {
            "type": "object"
        },
        "triggers": {
            "type": "array",
            "properties": {
                "name": {
                    "type":"string"
                },
                "type": {
                    "type":"string"
                },
                "source": {
                    "type":"string"
                }
            }
        }
    }
}`

func ValidateFileAgainstSchema(jsonFile, schema string) error {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewReferenceLoader(filepath.Join("file://", GetWd(), jsonFile))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		fmt.Println("The func.yaml is not valid. Please see errors: ")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
		return errors.New("Please update your func.yaml to satisfy the schema.")
	}

	return nil
}
