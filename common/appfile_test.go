/*
 * Copyright (c) 2026 Oracle and/or its affiliates. All rights reserved.
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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppfileYAMLNestedAnnotations(t *testing.T) {
	dir, err := ioutil.TempDir("", "appfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	content := `name: myapp
annotations:
  oracle.com/oci/freeformTags:
    team: blue
  oracle.com/oci/definedTags:
    Operations:
      CostCenter: "42"
`
	if err := ioutil.WriteFile(filepath.Join(dir, "app.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	af, err := LoadAppfile(dir)
	if err != nil {
		t.Fatal(err)
	}

	freeform, err := freeformTagsFromAnnotationValue(af.Annotations[AnnotationOCIResourceFreeformTags])
	if err != nil {
		t.Fatalf("freeform tags: %v", err)
	}
	if freeform["team"] != "blue" {
		t.Fatalf("unexpected freeform tags: %v", freeform)
	}

	defined, err := definedTagsFromAnnotationValue(af.Annotations[AnnotationOCIResourceDefinedTags])
	if err != nil {
		t.Fatalf("defined tags: %v", err)
	}
	if defined["Operations"]["CostCenter"] != "42" {
		t.Fatalf("unexpected defined tags: %v", defined)
	}
}
