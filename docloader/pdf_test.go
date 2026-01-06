/*
 Copyright 2023 NanaFS Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package docloader

import (
	"testing"
)

func TestParsePDFDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantZero bool
	}{
		{"full date", "D:20240115123045", false},
		{"date without time", "D:20240115", false},
		{"invalid format", "2024-01-15", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePDFDate(tt.input)
			if tt.wantZero && result != 0 {
				t.Errorf("parsePDFDate(%q) = %d, want 0", tt.input, result)
			}
			if !tt.wantZero && result == 0 {
				t.Errorf("parsePDFDate(%q) = 0, want non-zero", tt.input)
			}
		})
	}
}

func TestExtractPDFMetadata_NilReader(t *testing.T) {
	// Should not panic
	result := extractPDFMetadata(nil)
	if result.Title != "" {
		t.Errorf("expected empty result for nil reader, got %+v", result)
	}
}
