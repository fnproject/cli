/*
 * Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
)

const (
	V20180708         = 20180708
	LatestYamlVersion = V20180708
)

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
