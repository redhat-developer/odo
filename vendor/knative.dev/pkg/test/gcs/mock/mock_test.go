/*
Copyright 2020 The Knative Authors

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

package mock

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"testing"
)

func TestSetError(t *testing.T) {
	ctx := context.Background()
	bkt := "dummy"
	project := "dummy"
	dirPath := "/"

	testCases := []struct {
		testname string
		m        map[Method]*ReturnError //error map to load into mockClient
	}{
		{
			testname: "set errors for methods",
			m: map[Method]*ReturnError{
				MethodNewStorageBucket: {
					NumCall: 2,
					Err:     fmt.Errorf("MethodNewStorageBucket Error"),
				},
				MethodDeleteStorageBucket: {
					NumCall: 1,
					Err:     fmt.Errorf("MethodDeleteStorageBucketError"),
				},
				MethodListChildrenFiles: {
					NumCall: 0,
					Err:     fmt.Errorf("MethodListChildrenFilesError"),
				},
				MethodListDirectChildren: {
					NumCall: 1,
					Err:     fmt.Errorf("MethodListDirectChildrenError"),
				},
				MethodAttrObject: {
					NumCall: 2,
					Err:     fmt.Errorf("MethodAttrObjectError"),
				},
				MethodCopyObject: {
					NumCall: 3,
					Err:     fmt.Errorf("MethodCopyObjectError"),
				},
				MethodReadObject: {
					NumCall: 2,
					Err:     fmt.Errorf("MethodReadObjectError"),
				},
				MethodWriteObject: {
					NumCall: 1,
					Err:     fmt.Errorf("MethodWriteObjectError"),
				},
				MethodDeleteObject: {
					NumCall: 0,
					Err:     fmt.Errorf("MethodDeleteObjectError"),
				},
				MethodDownload: {
					NumCall: 1,
					Err:     fmt.Errorf("MethodDownload"),
				},
				MethodUpload: {
					NumCall: 2,
					Err:     fmt.Errorf("MethodUpload"),
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			mockClient := NewClientMocker()
			mockClient.SetError(tt.m)

			for k, v := range tt.m {
				switch numCall := v.NumCall; k {
				case MethodNewStorageBucket:
					for i := uint8(0); i < numCall; i++ {
						mockClient.NewStorageBucket(ctx, bkt, project)
					}

					if err := mockClient.NewStorageBucket(ctx, bkt, project); err == nil {
						t.Errorf("expected error %v", v.Err)
					} else if err.Error() != v.Err.Error() {
						t.Errorf("expected error %v, got error %v", v.Err, err)
					}
				case MethodDeleteStorageBucket:
					for i := uint8(0); i < numCall; i++ {
						mockClient.DeleteStorageBucket(ctx, bkt, true)
					}

					if err := mockClient.DeleteStorageBucket(ctx, bkt, true); err == nil {
						t.Errorf("expected error %v", v.Err)
					} else if err.Error() != v.Err.Error() {
						t.Errorf("expected error %v, got error %v", v.Err, err)
					}
				case MethodListChildrenFiles:
					for i := uint8(0); i < numCall; i++ {
						mockClient.ListChildrenFiles(ctx, bkt, dirPath)
					}

					if _, err := mockClient.ListChildrenFiles(ctx, bkt, dirPath); err == nil {
						t.Errorf("expected error %v", v.Err)
					} else if err.Error() != v.Err.Error() {
						t.Errorf("expected error %v, got error %v", v.Err, err)
					}
				case MethodListDirectChildren:
					for i := uint8(0); i < numCall; i++ {
						mockClient.ListDirectChildren(ctx, bkt, dirPath)
					}

					if _, err := mockClient.ListDirectChildren(ctx, bkt, dirPath); err == nil {
						t.Errorf("expected error %v", v.Err)
					} else if err.Error() != v.Err.Error() {
						t.Errorf("expected error %v, got error %v", v.Err, err)
					}
				case MethodAttrObject:
					for i := uint8(0); i < numCall; i++ {
						mockClient.AttrObject(ctx, bkt, dirPath)
					}

					if _, err := mockClient.AttrObject(ctx, bkt, dirPath); err == nil {
						t.Errorf("expected error %v", v.Err)
					} else if err.Error() != v.Err.Error() {
						t.Errorf("expected error %v, got error %v", v.Err, err)
					}
				case MethodCopyObject:
					for i := uint8(0); i < numCall; i++ {
						mockClient.CopyObject(ctx, bkt, dirPath, bkt, dirPath)
					}

					if err := mockClient.CopyObject(ctx, bkt, dirPath, bkt, dirPath); err == nil {
						t.Errorf("expected error %v", v.Err)
					} else if err.Error() != v.Err.Error() {
						t.Errorf("expected error %v, got error %v", v.Err, err)
					}
				case MethodReadObject:
					for i := uint8(0); i < numCall; i++ {
						mockClient.ReadObject(ctx, bkt, dirPath)
					}

					if _, err := mockClient.ReadObject(ctx, bkt, dirPath); err == nil {
						t.Errorf("expected error %v", v.Err)
					} else if err.Error() != v.Err.Error() {
						t.Errorf("expected error %v, got error %v", v.Err, err)
					}
				case MethodWriteObject:
					for i := uint8(0); i < numCall; i++ {
						mockClient.WriteObject(ctx, bkt, dirPath, []byte{})
					}

					if _, err := mockClient.WriteObject(ctx, bkt, dirPath, []byte{}); err == nil {
						t.Errorf("expected error %v", v.Err)
					} else if err.Error() != v.Err.Error() {
						t.Errorf("expected error %v, got error %v", v.Err, err)
					}
				case MethodDeleteObject:
					for i := uint8(0); i < numCall; i++ {
						mockClient.DeleteObject(ctx, bkt, dirPath)
					}

					if err := mockClient.DeleteObject(ctx, bkt, dirPath); err == nil {
						t.Errorf("expected error %v", v.Err)
					} else if err.Error() != v.Err.Error() {
						t.Errorf("expected error %v, got error %v", v.Err, err)
					}
				case MethodDownload:
					for i := uint8(0); i < numCall; i++ {
						mockClient.Download(ctx, bkt, dirPath, dirPath)
					}

					if err := mockClient.Download(ctx, bkt, dirPath, dirPath); err == nil {
						t.Errorf("expected error %v", v.Err)
					} else if err.Error() != v.Err.Error() {
						t.Errorf("expected error %v, got error %v", v.Err, err)
					}
				case MethodUpload:
					for i := uint8(0); i < numCall; i++ {
						mockClient.Upload(ctx, bkt, dirPath, dirPath)
					}

					if err := mockClient.Upload(ctx, bkt, dirPath, dirPath); err == nil {
						t.Errorf("expected error %v", v.Err)
					} else if err.Error() != v.Err.Error() {
						t.Errorf("expected error %v, got error %v", v.Err, err)
					}
				default:
					t.Errorf("unknown method")
				}
			}
		})
	}
}

func TestClearError(t *testing.T) {
	ctx := context.Background()
	bkt := "dummy"
	project := "dummy"
	dirPath := "/"

	testCases := []struct {
		testname string
		m        map[Method]*ReturnError //error map to load into mockClient
	}{
		{
			testname: "set errors for methods",
			m: map[Method]*ReturnError{
				MethodNewStorageBucket: {
					NumCall: 0,
					Err:     fmt.Errorf("MethodNewStorageBucket Error"),
				},
				MethodDeleteStorageBucket: {
					NumCall: 0,
					Err:     fmt.Errorf("MethodDeleteStorageBucketError"),
				},
				MethodListChildrenFiles: {
					NumCall: 0,
					Err:     fmt.Errorf("MethodListChildrenFilesError"),
				},
				MethodListDirectChildren: {
					NumCall: 0,
					Err:     fmt.Errorf("MethodListDirectChildrenError"),
				},
				MethodAttrObject: {
					NumCall: 0,
					Err:     fmt.Errorf("MethodAttrObjectError"),
				},
				MethodCopyObject: {
					NumCall: 0,
					Err:     fmt.Errorf("MethodCopyObjectError"),
				},
				MethodReadObject: {
					NumCall: 0,
					Err:     fmt.Errorf("MethodReadObjectError"),
				},
				MethodWriteObject: {
					NumCall: 0,
					Err:     fmt.Errorf("MethodWriteObjectError"),
				},
				MethodDeleteObject: {
					NumCall: 0,
					Err:     fmt.Errorf("MethodDeleteObjectError"),
				},
				MethodDownload: {
					NumCall: 0,
					Err:     fmt.Errorf("MethodDownload"),
				},
				MethodUpload: {
					NumCall: 0,
					Err:     fmt.Errorf("MethodUpload"),
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			mockClient := NewClientMocker()
			mockClient.SetError(tt.m)
			mockClient.ClearError()

			for k, v := range tt.m {
				switch k {
				case MethodNewStorageBucket:
					if err := mockClient.NewStorageBucket(ctx, bkt, project); err != nil && err.Error() == v.Err.Error() {
						t.Errorf("error %v should have been cleared", v.Err)
					}
				case MethodDeleteStorageBucket:
					if err := mockClient.DeleteStorageBucket(ctx, bkt, true); err != nil && err.Error() == v.Err.Error() {
						t.Errorf("error %v should have been cleared", v.Err)
					}
				case MethodListChildrenFiles:
					if _, err := mockClient.ListChildrenFiles(ctx, bkt, dirPath); err != nil && err.Error() == v.Err.Error() {
						t.Errorf("error %v should have been cleared", v.Err)
					}
				case MethodListDirectChildren:
					if _, err := mockClient.ListDirectChildren(ctx, bkt, dirPath); err != nil && err.Error() == v.Err.Error() {
						t.Errorf("error %v should have been cleared", v.Err)
					}
				case MethodAttrObject:
					if _, err := mockClient.AttrObject(ctx, bkt, dirPath); err != nil && err.Error() == v.Err.Error() {
						t.Errorf("error %v should have been cleared", v.Err)
					}
				case MethodCopyObject:
					if err := mockClient.CopyObject(ctx, bkt, dirPath, bkt, dirPath); err != nil && err.Error() == v.Err.Error() {
						t.Errorf("error %v should have been cleared", v.Err)
					}
				case MethodReadObject:
					if _, err := mockClient.ReadObject(ctx, bkt, dirPath); err != nil && err.Error() == v.Err.Error() {
						t.Errorf("error %v should have been cleared", v.Err)
					}
				case MethodWriteObject:
					if _, err := mockClient.WriteObject(ctx, bkt, dirPath, []byte{}); err != nil && err.Error() == v.Err.Error() {
						t.Errorf("error %v should have been cleared", v.Err)
					}
				case MethodDeleteObject:
					if err := mockClient.DeleteObject(ctx, bkt, dirPath); err != nil && err.Error() == v.Err.Error() {
						t.Errorf("error %v should have been cleared", v.Err)
					}
				case MethodDownload:
					if err := mockClient.Download(ctx, bkt, dirPath, dirPath); err != nil && err.Error() == v.Err.Error() {
						t.Errorf("error %v should have been cleared", v.Err)
					}
				case MethodUpload:
					if err := mockClient.Upload(ctx, bkt, dirPath, dirPath); err != nil && err.Error() == v.Err.Error() {
						t.Errorf("error %v should have been cleared", v.Err)
					}
				default:
					t.Errorf("unknown method")
				}
			}
		})
	}

}

func TestNewStorageBucket(t *testing.T) {
	ctx := context.Background()
	mockClient := NewClientMocker()
	bktName1 := "test-bucket1"
	project1 := "test-project1"
	bktName2 := "test-bucket2"
	project2 := "test-project2"

	testCases := []struct {
		testname    string
		bkt         string
		projectName string
		err         error
	}{
		{
			testname:    "createNewBucket",
			bkt:         bktName1,
			projectName: project1,
			err:         nil,
		},
		{
			testname:    "existingNewBucket",
			bkt:         bktName1,
			projectName: project1,
			err:         NewBucketExistError(bktName1),
		},
		{
			testname:    "existingNewBucketDifferentProject",
			bkt:         bktName1,
			projectName: project2,
			err:         NewBucketExistError(bktName1),
		},
		{
			testname:    "secondNewBucket",
			bkt:         bktName2,
			projectName: project1,
			err:         nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			err := mockClient.NewStorageBucket(ctx, tt.bkt, tt.projectName)
			if (tt.err == nil || err == nil) && err != tt.err {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			} else if (tt.err != nil && err != nil) && tt.err.Error() != err.Error() {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			}

			if tt.err != nil {
				return
			}

			if p, ok := mockClient.revIndex[bucket(tt.bkt)]; !ok {
				t.Fatalf("expected revIndex to contain key %v", bucket(tt.bkt))
			} else if p != project(tt.projectName) {
				t.Fatalf("expected revIndex value %v, got %v", project(tt.projectName), p)
			}

			if p, ok := mockClient.gcp[project(tt.projectName)]; !ok {
				t.Fatalf("expected gcp to contain key %v", project(tt.projectName))
			} else if _, ok := p.bkt[bucket(tt.bkt)]; !ok {
				t.Fatalf("expected gcp.bucket to contain key %v", bucket(tt.bkt))
			}
		})
	}
}

func TestDeleteStorageBucket(t *testing.T) {
	ctx := context.Background()
	mockClient := NewClientMocker()
	bktName1 := "test-bucket1"
	project1 := "test-project1"

	mockClient.NewStorageBucket(ctx, bktName1, project1)
	mockClient.WriteObject(ctx, bktName1, "object1", []byte("Hello"))

	testCases := []struct {
		testname string
		bkt      string
		force    bool
		err      error
	}{
		{
			testname: "deleteBucket",
			bkt:      bktName1,
			force:    false,
			err:      NewNotEmptyBucketError(bktName1),
		},
		{
			testname: "deleteBucket",
			bkt:      bktName1,
			force:    true,
			err:      nil,
		},
		{
			testname: "deleteNonExistentBucket",
			bkt:      bktName1,
			err:      NewNoBucketError(bktName1),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			err := mockClient.DeleteStorageBucket(ctx, tt.bkt, tt.force)
			if (tt.err == nil || err == nil) && err != tt.err {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			} else if (tt.err != nil && err != nil) && tt.err.Error() != err.Error() {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			}

			if tt.err != nil {
				return
			}

			if _, ok := mockClient.revIndex[bucket(tt.bkt)]; ok {
				t.Fatalf("bucket %v should have been deleted", bucket(tt.bkt))
			}

			if _, ok := mockClient.gcp[project(project1)].bkt[bucket(tt.bkt)]; ok {
				t.Fatalf("bucket %v should have been deleted", bucket(tt.bkt))
			}
		})
	}
}

func TestExists(t *testing.T) {
	ctx := context.Background()
	mockClient := NewClientMocker()
	bktName1 := "test-bucket1"
	project1 := "test-project1"
	object1 := "object1"
	dir1 := "dir"
	content := []byte("Hello World")

	mockClient.NewStorageBucket(ctx, bktName1, project1)
	mockClient.WriteObject(ctx, bktName1, path.Join(dir1, object1), content)

	testCases := []struct {
		testname string
		bkt      string
		objpath  string
		exist    bool
	}{
		{
			testname: "existObject",
			bkt:      bktName1,
			objpath:  path.Join(dir1, object1),
			exist:    true,
		},
		{
			testname: "existBucket",
			bkt:      bktName1,
			objpath:  "",
			exist:    true,
		},
		{
			testname: "existDir",
			bkt:      bktName1,
			objpath:  "dir",
			exist:    true,
		},
		{
			testname: "nonexistentObject",
			bkt:      bktName1,
			objpath:  "badobjectpath",
			exist:    false,
		},
		{
			testname: "nonexistentBkt",
			bkt:      "non-existent-bucket",
			exist:    false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			if exist := mockClient.Exists(ctx, tt.bkt, tt.objpath); exist != tt.exist {
				t.Fatalf("expected exist %v to return %v, got %v", tt.objpath, tt.exist, exist)
			}
		})
	}
}

func TestListChildrenFiles(t *testing.T) {
	ctx := context.Background()
	mockClient := NewClientMocker()
	bktName1 := "test-bucket1"
	project1 := "test-project1"
	dir1 := "dir"
	dir2 := "dir/subdir"
	object1 := path.Join(dir1, "object1")
	object2 := path.Join(dir1, "object2")
	object3 := path.Join(dir2, "object3")
	content := []byte("Hello World")

	mockClient.NewStorageBucket(ctx, bktName1, project1)
	mockClient.WriteObject(ctx, bktName1, object1, content)
	mockClient.WriteObject(ctx, bktName1, object2, content)
	mockClient.WriteObject(ctx, bktName1, object3, content)

	testCases := []struct {
		testname string
		bkt      string
		dir      string
		expected []string
		err      error
	}{
		{
			testname: "listAllChildrenObjects",
			bkt:      bktName1,
			dir:      "dir",
			expected: []string{object1, object2, object3},
			err:      nil,
		},
		{
			testname: "listAllChildrenObjects",
			bkt:      bktName1,
			dir:      "",
			expected: []string{object1, object2, object3},
			err:      nil,
		},
		{
			testname: "badBucket",
			bkt:      "non-existent-bucket",
			err:      NewNoBucketError("non-existent-bucket"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			children, err := mockClient.ListChildrenFiles(ctx, tt.bkt, tt.dir)
			if (tt.err == nil || err == nil) && err != tt.err {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			} else if (tt.err != nil && err != nil) && tt.err.Error() != err.Error() {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			}

			sort.Strings(children)
			sort.Strings(tt.expected)
			if !reflect.DeepEqual(children, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, children)
			}

		})
	}
}

func TestListDirectChildren(t *testing.T) {
	ctx := context.Background()
	mockClient := NewClientMocker()
	bktName1 := "test-bucket1"
	project1 := "test-project1"
	dir1 := "dir"
	dir2 := "dir/subdir"
	object1 := path.Join(dir1, "object1")
	object2 := path.Join(dir1, "object2")
	object3 := path.Join(dir2, "object3")
	object4 := "object4"
	content := []byte("Hello World")

	mockClient.NewStorageBucket(ctx, bktName1, project1)
	mockClient.WriteObject(ctx, bktName1, object1, content)
	mockClient.WriteObject(ctx, bktName1, object2, content)
	mockClient.WriteObject(ctx, bktName1, object3, content)
	mockClient.WriteObject(ctx, bktName1, object4, content)

	testCases := []struct {
		testname string
		bkt      string
		dir      string
		expected []string
		err      error
	}{
		{
			testname: "listAllChildrenObjects",
			bkt:      bktName1,
			dir:      "dir",
			expected: []string{object1, object2},
			err:      nil,
		},
		{
			testname: "listAllChildrenObjects",
			bkt:      bktName1,
			dir:      "",
			expected: []string{object4},
			err:      nil,
		},
		{
			testname: "badBucket",
			bkt:      "non-existent-bucket",
			err:      NewNoBucketError("non-existent-bucket"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			children, err := mockClient.ListDirectChildren(ctx, tt.bkt, tt.dir)
			if (tt.err == nil || err == nil) && err != tt.err {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			} else if (tt.err != nil && err != nil) && tt.err.Error() != err.Error() {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			}

			sort.Strings(children)
			sort.Strings(tt.expected)
			if !reflect.DeepEqual(children, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, children)
			}

		})
	}
}

func TestAttrObject(t *testing.T) {
	ctx := context.Background()
	mockClient := NewClientMocker()
	bktName1 := "test-bucket1"
	project1 := "test-project1"
	object1 := "dir/object1"
	content := []byte("Hello World")

	mockClient.NewStorageBucket(ctx, bktName1, project1)
	mockClient.WriteObject(ctx, bktName1, object1, content)

	testCases := []struct {
		testname string
		bkt      string
		objpath  string
		size     int64
		err      error
	}{
		{
			testname: "existObjectAttr",
			bkt:      bktName1,
			objpath:  object1,
			size:     int64(len(content)),
		},
		{
			testname: "badObject",
			bkt:      bktName1,
			objpath:  "badobjectpath",
			err:      NewNoObjectError(bktName1, "badobjectpath", ""),
		},
		{
			testname: "badBucket",
			bkt:      "non-existent-bucket",
			err:      NewNoBucketError("non-existent-bucket"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			objAttr, err := mockClient.AttrObject(ctx, tt.bkt, tt.objpath)
			if (tt.err == nil || err == nil) && err != tt.err {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			} else if (tt.err != nil && err != nil) && tt.err.Error() != err.Error() {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			}

			if tt.err != nil {
				return
			}

			if objAttr.Bucket != tt.bkt {
				t.Fatalf("expected content %v, got content %v", tt.bkt, objAttr.Bucket)
			} else if objAttr.Name != tt.objpath {
				t.Fatalf("expected content %v, got content %v", tt.objpath, objAttr.Name)
			} else if objAttr.Size != tt.size {
				t.Fatalf("expected content %v, got content %v", tt.size, objAttr.Size)
			}
		})
	}
}

func TestCopyObject(t *testing.T) {
	ctx := context.Background()
	mockClient := NewClientMocker()
	bktName1 := "test-bucket1"
	bktName2 := "test-bucket2"
	project1 := "test-project1"
	object1 := "dir/object1"
	content := []byte("Hello World")

	mockClient.NewStorageBucket(ctx, bktName1, project1)
	mockClient.NewStorageBucket(ctx, bktName2, project1)
	mockClient.WriteObject(ctx, bktName1, object1, content)

	testCases := []struct {
		testname   string
		srcBkt     string
		srcObjPath string
		dstBkt     string
		dstObjPath string
		err        error
	}{
		{
			testname:   "copySameBucket",
			srcBkt:     bktName1,
			srcObjPath: object1,
			dstBkt:     bktName1,
			dstObjPath: "dir/object2",
			err:        nil,
		},
		{
			testname:   "copyAnotherBucket",
			srcBkt:     bktName1,
			srcObjPath: object1,
			dstBkt:     bktName2,
			dstObjPath: "dir/object2",
			err:        nil,
		},
		{
			testname:   "badSrcObject",
			srcBkt:     bktName1,
			srcObjPath: "badobjectpath",
			dstBkt:     bktName2,
			dstObjPath: "dir/object2",
			err:        NewNoObjectError(bktName1, "badobjectpath", ""),
		},
		{
			testname:   "badDstObject",
			srcBkt:     bktName1,
			srcObjPath: object1,
			dstBkt:     bktName1,
			dstObjPath: "badobjectpath/",
			err:        NewNoObjectError(bktName1, "", "badobjectpath/"),
		},
		{
			testname: "badSrcBucket",
			srcBkt:   "non-existent-bucket",
			dstBkt:   bktName1,
			err:      NewNoBucketError("non-existent-bucket"),
		},
		{
			testname: "badDstBucket",
			srcBkt:   bktName1,
			dstBkt:   "non-existent-bucket",
			err:      NewNoBucketError("non-existent-bucket"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			err := mockClient.CopyObject(ctx, tt.srcBkt, tt.srcObjPath, tt.dstBkt, tt.dstObjPath)
			if (tt.err == nil || err == nil) && err != tt.err {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			} else if (tt.err != nil && err != nil) && tt.err.Error() != err.Error() {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			}

			if tt.err != nil {
				return
			}

			objContent, err := mockClient.ReadObject(ctx, tt.dstBkt, tt.dstObjPath)
			if err != nil {
				t.Fatalf("cannot read %v from bucket %v, got error %v", tt.dstObjPath, tt.dstBkt, err)
			}

			if !bytes.Equal(objContent, content) {
				t.Fatalf("expected copied content %v, got content %v", content, objContent)
			}
		})
	}
}

func TestReadObject(t *testing.T) {
	ctx := context.Background()
	mockClient := NewClientMocker()
	bktName1 := "test-bucket1"
	project1 := "test-project1"
	object1 := "object1"
	content := []byte("Hello World")
	badBkt := "non-existent-bucket"

	mockClient.NewStorageBucket(ctx, bktName1, project1)
	mockClient.WriteObject(ctx, bktName1, object1, content)

	testCases := []struct {
		testname string
		bkt      string
		objpath  string
		err      error
	}{
		{
			testname: "readObject",
			bkt:      bktName1,
			objpath:  path.Join(object1),
			err:      nil,
		},
		{
			testname: "ReadObjectBadPath",
			bkt:      bktName1,
			objpath:  object1 + "/",
			err:      NewNoObjectError(bktName1, "", object1+"/"),
		},
		{
			testname: "ReadObjectBadBucket",
			bkt:      badBkt,
			err:      NewNoBucketError(badBkt),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			objContent, err := mockClient.ReadObject(ctx, tt.bkt, tt.objpath)
			if (tt.err == nil || err == nil) && err != tt.err {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			} else if (tt.err != nil && err != nil) && tt.err.Error() != err.Error() {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			}

			if tt.err != nil {
				return
			}

			if !bytes.Equal(content, objContent) {
				t.Fatalf("expected content %v, got content %v", content, objContent)
			}
		})
	}
}

func TestWriteObject(t *testing.T) {
	ctx := context.Background()
	mockClient := NewClientMocker()
	bktName1 := "test-bucket1"
	project1 := "test-project1"
	badBkt := "non-existent-bucket"

	mockClient.NewStorageBucket(ctx, bktName1, project1)

	testCases := []struct {
		testname string
		bkt      string
		objpath  string
		content  []byte
		err      error
	}{
		{
			testname: "writeObject",
			bkt:      bktName1,
			objpath:  "testing/object",
			content:  []byte("Hello World"),
			err:      nil,
		},
		{
			testname: "writeObjectBadPath",
			bkt:      bktName1,
			objpath:  "testing/",
			err:      NewNoObjectError(bktName1, "", "testing/"),
		},
		{
			testname: "writeObjectBadBucket",
			bkt:      badBkt,
			err:      NewNoBucketError(badBkt),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			n, err := mockClient.WriteObject(ctx, tt.bkt, tt.objpath, tt.content)
			if (tt.err == nil || err == nil) && err != tt.err {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			} else if (tt.err != nil && err != nil) && tt.err.Error() != err.Error() {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			}

			if tt.err != nil {
				return
			}

			if n != len(tt.content) {
				t.Fatalf("content has length %v, wrote only %v bytes", len(tt.content), n)
			}

			if content, err := mockClient.ReadObject(ctx, tt.bkt, tt.objpath); err != nil {
				t.Fatalf("read object returned error %v", err)
			} else if !bytes.Equal(content, tt.content) {
				t.Fatalf("expected content %v, got content %v", tt.content, content)
			}
		})
	}
}

func TestDeleteObject(t *testing.T) {
	ctx := context.Background()
	mockClient := NewClientMocker()
	bktName1 := "test-bucket1"
	project1 := "test-project1"
	object1 := "dir/object1"

	mockClient.NewStorageBucket(ctx, bktName1, project1)
	mockClient.WriteObject(ctx, bktName1, object1, []byte("Hello World"))

	testCases := []struct {
		testname string
		bkt      string
		objpath  string
		err      error
	}{
		{
			testname: "DeleteObject",
			bkt:      bktName1,
			objpath:  object1,
			err:      nil,
		},
		{
			testname: "DeleteNonExistentObject",
			bkt:      bktName1,
			objpath:  "non-existent-object",
			err:      nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			err := mockClient.DeleteObject(ctx, tt.bkt, tt.objpath)
			if (tt.err == nil || err == nil) && err != tt.err {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			} else if (tt.err != nil && err != nil) && tt.err.Error() != err.Error() {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			}

			if tt.err != nil {
				return
			}

			if mockClient.Exists(ctx, tt.bkt, tt.objpath) {
				t.Fatalf("%v in %v should not exist (deleted)", tt.objpath, tt.bkt)
			}
		})
	}
}

func TestDownload(t *testing.T) {
	ctx := context.Background()
	mockClient := NewClientMocker()
	bktName1 := "test-bucket1"
	project1 := "test-project1"
	object1 := "dir/object1"
	file := "test/download"
	defer os.Remove(file)
	content := []byte("Hello World")

	mockClient.NewStorageBucket(ctx, bktName1, project1)
	mockClient.WriteObject(ctx, bktName1, object1, content)

	testCases := []struct {
		testname string
		bkt      string
		objPath  string
		err      error
	}{
		{
			testname: "downloadObject",
			bkt:      bktName1,
			objPath:  object1,
			err:      nil,
		},
		{
			testname: "badObject",
			bkt:      bktName1,
			objPath:  "badobjectpath",
			err:      NewNoObjectError(bktName1, "badobjectpath", ""),
		},
		{
			testname: "badBucket",
			bkt:      "non-existent-bucket",
			err:      NewNoBucketError("non-existent-bucket"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			err := mockClient.Download(ctx, tt.bkt, tt.objPath, file)
			if (tt.err == nil || err == nil) && err != tt.err {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			} else if (tt.err != nil && err != nil) && tt.err.Error() != err.Error() {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			}

			if tt.err != nil {
				return
			}

			fileContent, err := ioutil.ReadFile(file)
			if err != nil {
				t.Fatalf("cannot read content %v, error %v", file, err)
			}
			if !bytes.Equal(fileContent, content) {
				t.Fatalf("expected copied content %v, got content %v", content, fileContent)
			}
		})
	}
}

func TestUpload(t *testing.T) {
	ctx := context.Background()
	mockClient := NewClientMocker()
	bktName1 := "test-bucket1"
	project1 := "test-project1"
	object1 := "dir/object1"
	file := "test/upload"
	content, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("cannot read content %v, error %v", file, err)
	}

	mockClient.NewStorageBucket(ctx, bktName1, project1)

	testCases := []struct {
		testname string
		bkt      string
		objPath  string
		err      error
	}{
		{
			testname: "uploadObject",
			bkt:      bktName1,
			objPath:  object1,
			err:      nil,
		},
		{
			testname: "badObject",
			bkt:      bktName1,
			objPath:  "badobjectpath/",
			err:      NewNoObjectError(bktName1, "", "badobjectpath/"),
		},
		{
			testname: "badBucket",
			bkt:      "non-existent-bucket",
			err:      NewNoBucketError("non-existent-bucket"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testname, func(t *testing.T) {
			err := mockClient.Upload(ctx, tt.bkt, tt.objPath, file)
			if (tt.err == nil || err == nil) && err != tt.err {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			} else if (tt.err != nil && err != nil) && tt.err.Error() != err.Error() {
				t.Fatalf("expected error %v, got error %v", tt.err, err)
			}

			if tt.err != nil {
				return
			}

			objContent, err := mockClient.ReadObject(ctx, tt.bkt, tt.objPath)
			if err != nil {
				t.Fatalf("cannot read content %v in bucket %v, error %v", tt.objPath, tt.bkt, err)
			}
			if !bytes.Equal(objContent, content) {
				t.Fatalf("expected copied content %v, got content %v", content, objContent)
			}
		})
	}
}
