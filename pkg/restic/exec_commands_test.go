/*
Copyright 2019 the Velero contributors.

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

package restic

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu/velero/pkg/test"
	"github.com/vmware-tanzu/velero/pkg/util/filesystem"
)

func Test_getSummaryLine(t *testing.T) {
	summaryLine := `{"message_type":"summary","files_new":0,"files_changed":0,"files_unmodified":3,"dirs_new":0,"dirs_changed":0,"dirs_unmodified":0,"data_blobs":0,"tree_blobs":0,"data_added":0,"total_files_processed":3,"total_bytes_processed":13238272000,"total_duration":0.319265105,"snapshot_id":"38515bb5"}`
	tests := []struct {
		name    string
		output  string
		wantErr bool
	}{
		{"no summary", `{"message_type":"status","percent_done":0,"total_files":1,"total_bytes":10485760000}
{"message_type":"status","percent_done":0,"total_files":3,"files_done":1,"total_bytes":13238272000}
`, true},
		{"no newline after summary", `{"message_type":"status","percent_done":0,"total_files":1,"total_bytes":10485760000}
{"message_type":"status","percent_done":0,"total_files":3,"files_done":1,"total_bytes":13238272000}
{"message_type":"summary","files_new":0,"files_changed":0,"files_unmodified":3,"dirs_new":0`, true},
		{"summary at end", `{"message_type":"status","percent_done":0,"total_files":1,"total_bytes":10485760000}
{"message_type":"status","percent_done":0,"total_files":3,"files_done":1,"total_bytes":13238272000}
{"message_type":"status","percent_done":1,"total_files":3,"files_done":3,"total_bytes":13238272000,"bytes_done":13238272000}
{"message_type":"summary","files_new":0,"files_changed":0,"files_unmodified":3,"dirs_new":0,"dirs_changed":0,"dirs_unmodified":0,"data_blobs":0,"tree_blobs":0,"data_added":0,"total_files_processed":3,"total_bytes_processed":13238272000,"total_duration":0.319265105,"snapshot_id":"38515bb5"}
`, false},
		{"summary before status", `{"message_type":"status","percent_done":0,"total_files":1,"total_bytes":10485760000}
{"message_type":"status","percent_done":0,"total_files":3,"files_done":1,"total_bytes":13238272000}
{"message_type":"summary","files_new":0,"files_changed":0,"files_unmodified":3,"dirs_new":0,"dirs_changed":0,"dirs_unmodified":0,"data_blobs":0,"tree_blobs":0,"data_added":0,"total_files_processed":3,"total_bytes_processed":13238272000,"total_duration":0.319265105,"snapshot_id":"38515bb5"}
{"message_type":"status","percent_done":1,"total_files":3,"files_done":3,"total_bytes":13238272000,"bytes_done":13238272000}
`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, err := getSummaryLine([]byte(tt.output))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, summaryLine, string(summary))
			}
		})
	}
}

func Test_getLastLine(t *testing.T) {
	tests := []struct {
		output []byte
		want   string
	}{
		{[]byte(`last line
`), "last line"},
		{[]byte(`first line
second line
third line
`), "third line"},
		{[]byte(""), ""},
		{nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, []byte(tt.want), getLastLine([]byte(tt.output)))
		})
	}
}

func Test_getVolumeSize(t *testing.T) {
	files := map[string][]byte{
		"/file1.txt":              []byte("file1"),
		"/file2.txt":              []byte("file2"),
		"/file3.txt":              []byte("file3"),
		"/files/file4.txt":        []byte("file4"),
		"/files/nested/file5.txt": []byte("file5"),
	}
	fakefs := test.NewFakeFileSystem()

	var expectedSize int64
	for path, content := range files {
		fakefs.WithFile(path, content)
		expectedSize += int64(len(content))
	}

	fileSystem = fakefs
	defer func() { fileSystem = filesystem.NewFileSystem() }()

	actualSize, err := getVolumeSize("/")

	assert.NoError(t, err)
	assert.Equal(t, expectedSize, actualSize)
}

func Test_lsFile(t *testing.T) {
	tests := []struct {
		output  string
		want    string
		wantErr bool
	}{
		{`found 8 old cache directories in /Users/youtube/Library/Caches/restic, run 'restic cache --cleanup' to remove them
{"time":"2022-10-27T16:52:24.843831+08:00","tree":"a46d6d682f5b5795ff4e2cbb15a1f35bce82de2f147a2fe1f05f64231c65fa75","paths":["/tmp/123123"],"hostname":"MacBook-Pro.local","username":"youtube","uid":501,"gid":20,"tags":["di=ddd"],"id":"7be14766b57eae6c75d3a747a35622756b2a3437f24fd0d355283c2e621a311d","short_id":"7be14766","struct_type":"snapshot"}
{"name":"Package.swift","type":"file","path":"/123123.swiftpm/Package.swift","uid":501,"gid":20,"size":1074,"mode":420,"permissions":"-rw-r--r--","mtime":"2022-05-16T11:36:02.068797131+08:00","atime":"2022-05-16T11:36:02.068797131+08:00","ctime":"2022-05-16T11:36:02.069074195+08:00","struct_type":"node"}
`, "Package.swift", false},
		{`found 8 old cache directories in /Users/youtube/Library/Caches/restic, run 'restic cache --cleanup' to remove them
{"time":"2022-10-27T16:52:24.843831+08:00","tree":"a46d6d682f5b5795ff4e2cbb15a1f35bce82de2f147a2fe1f05f64231c65fa75","paths":["/tmp/123123"],"hostname":"MacBook-Pro.local","username":"youtube","uid":501,"gid":20,"tags":["di=ddd"],"id":"7be14766b57eae6c75d3a747a35622756b2a3437f24fd0d355283c2e621a311d","short_id":"7be14766","struct_type":"snapshot"}
`, "", true},
		{`repository 498109fb opened successfully, password is correct
found 8 old cache directories in /Users/youtube/Library/Caches/restic, run 'restic cache --cleanup' to remove them
Ignoring "17be14766": no matching ID found for prefix "17be14766"
`, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			res, err := parseLsContent(tt.output)
			t.Logf("want:%s wantErr:%v content:%s", tt.want, tt.wantErr, tt.output)
			t.Logf("res: %s, err:%v", res, err)
			assert.Equal(t, tt.want, res)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
