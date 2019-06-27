/*
Package util provides utility functions for OSD servers and drivers.
Copyright 2017 Portworx

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
package util

import (
	"fmt"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume/drivers/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestVolumeFromNameFailedToLocateDueToTooManyVolumes(t *testing.T) {

	mc := gomock.NewController(t)
	defer mc.Finish()

	driver := mock.NewMockVolumeDriver(mc)

	// Setup calls
	name := "myvolume"
	gomock.InOrder(
		// Too many
		driver.
			EXPECT().
			Inspect([]string{name}).
			Return([]*api.Volume{
				&api.Volume{
					Id: name,
					Locator: &api.VolumeLocator{
						Name: name,
					},
				},
				&api.Volume{
					Id: "two",
					Locator: &api.VolumeLocator{
						Name: name,
					},
				},
			}, nil).
			Times(1),
		driver.
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return([]*api.Volume{}, nil).
			Times(1),
	)

	// Expect not found
	_, err := VolumeFromName(driver, name)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Cannot locate")

	// Setup calls
	gomock.InOrder(
		// Return that it was not found
		driver.
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		// Return too many volumes
		driver.
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: "one",
					Locator: &api.VolumeLocator{
						Name: name,
					},
				},
				&api.Volume{
					Id: "two",
					Locator: &api.VolumeLocator{
						Name: name,
					},
				},
			}, nil).
			Times(1),
	)

	// Expect not found
	_, err = VolumeFromName(driver, name)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Cannot locate")

	// Setup calls
	gomock.InOrder(
		driver.
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),

		driver.
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("Error")).
			Times(1),
	)

	// Expect not found
	_, err = VolumeFromName(driver, name)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Failed to locate")
}

func TestVolumeFromNameFailedToLocate(t *testing.T) {

	mc := gomock.NewController(t)
	defer mc.Finish()

	driver := mock.NewMockVolumeDriver(mc)

	// Setup calls
	name := "myvolume"
	gomock.InOrder(
		driver.
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),
		driver.
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return(nil, fmt.Errorf("not found")).
			Times(1),
	)

	// Expect not found
	_, err := VolumeFromName(driver, name)
	assert.NotNil(t, err)
}

func TestVolumeFromNameFoundFromInspect(t *testing.T) {

	mc := gomock.NewController(t)
	defer mc.Finish()

	driver := mock.NewMockVolumeDriver(mc)

	// Setup calls
	name := "myvolume"
	driver.
		EXPECT().
		Inspect([]string{name}).
		Return([]*api.Volume{
			&api.Volume{
				Id: name,
				Locator: &api.VolumeLocator{
					Name: "hello",
				},
			},
		}, nil).
		Times(1)

	// Expect not found
	v, err := VolumeFromName(driver, name)
	assert.Nil(t, err)
	assert.NotNil(t, v)
	assert.Equal(t, v.Id, name)
	assert.Equal(t, v.GetLocator().GetName(), "hello")
}

func TestVolumeFromNameFoundFromEnumerate(t *testing.T) {

	mc := gomock.NewController(t)
	defer mc.Finish()

	driver := mock.NewMockVolumeDriver(mc)

	// Setup calls
	name := "myvolume"
	gomock.InOrder(
		driver.
			EXPECT().
			Inspect([]string{name}).
			Return(nil, fmt.Errorf("not found")).
			Times(1),
		driver.
			EXPECT().
			Enumerate(&api.VolumeLocator{Name: name}, nil).
			Return([]*api.Volume{
				&api.Volume{
					Id: "myid",
					Locator: &api.VolumeLocator{
						Name: name,
					},
				},
			}, nil).
			Times(1),
	)

	// Expect not found
	v, err := VolumeFromName(driver, name)
	assert.Nil(t, err)
	assert.NotNil(t, v)
	assert.Equal(t, v.Id, "myid")
	assert.Equal(t, v.GetLocator().GetName(), name)
}
