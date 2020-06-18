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
	"context"
	"fmt"

	"knative.dev/pkg/test/gcs"
)

// Example on how to override errors
const (
	bkt  = "NewBkt"
	proj = "NewProject"
)

func topFunction(c gcs.Client) error {
	ctx := context.Background()
	if err := c.NewStorageBucket(ctx, bkt, proj); err != nil {
		return err
	}

	// Should have returned error, but SetError override to nil
	if _, err := c.ReadObject(ctx, bkt, "non-existent-file"); err != nil {
		return err
	}

	if _, err := c.ListChildrenFiles(ctx, bkt, ""); err != nil {
		return err
	}

	if _, err := c.ListChildrenFiles(ctx, bkt, ""); err != nil {
		return err
	}

	// Should not have returned error, but SetError override to NewNoBucketError(bkt)
	if _, err := c.ListChildrenFiles(ctx, bkt, ""); err != nil {
		return err
	}

	return nil
}

func ExampleSetError() {
	mockClient := NewClientMocker()

	// Call to ReadObject, first call should return error, but returns nil
	// because it is overridden.
	mockClient.SetError(
		map[Method]*ReturnError{
			MethodReadObject: {
				NumCall: uint8(0),
				Err:     nil,
			},
			MethodListChildrenFiles: {
				NumCall: uint8(2),
				Err:     NewNoBucketError(bkt),
			},
		})

	fmt.Println(topFunction(mockClient))
	// Output:
	// no bucket "NewBkt"
}
