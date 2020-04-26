// Copyright 2015 Light Code Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package casket

import (
	"os"
	"strings"
	"testing"
)

func TestAssetsPath(t *testing.T) {
	if actual := AssetsPath(); !strings.HasSuffix(actual, ".casket") {
		t.Errorf("Expected path to be a .casket folder, got: %v", actual)
	}

	err := os.Setenv("CASKETFILE", "testpath")
	if err != nil {
		t.Error("Could not set CASKETFILE")
	}
	if actual, expected := AssetsPath(), "testpath"; actual != expected {
		t.Errorf("Expected path to be %v, got: %v", expected, actual)
	}
	err = os.Setenv("CASKETFILE", "")
	if err != nil {
		t.Error("Could not set CASKETFILE")
	}
}
