/*
Copyright 2014 Google Inc. All rights reserved.

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

package build

import (
	"fmt"
	"time"
    
	"code.google.com/p/go-uuid/uuid"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/apiserver"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/build/buildapi"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/buildconfig"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/golang/glog"
)

// BuildRegistryStorage is an implementation of RESTStorage for the api server.
type BuildRegistryStorage struct {
	registry BuildRegistry
	buildConfigRegistry buildconfig.BuildConfigRegistry
	
}

func NewBuildRegistryStorage(registry BuildRegistry, buildConfigRegistry buildconfig.BuildConfigRegistry) apiserver.RESTStorage {
	return &BuildRegistryStorage{
		registry: registry,
		buildConfigRegistry: buildConfigRegistry,
	}
}

// List obtains a list of Builds that match selector.
func (storage *BuildRegistryStorage) List(selector labels.Selector) (interface{}, error) {
	result := buildapi.BuildList{}
	builds, err := storage.registry.ListBuilds()
	if err == nil {
		for _, build := range builds.Items {
			result.Items = append(result.Items, build)
		}
	}
	return result, err
}

// Get obtains the build specified by its id.
func (storage *BuildRegistryStorage) Get(id string) (interface{}, error) {
	build, err := storage.registry.GetBuild(id)
	if err != nil {
		return nil, err
	}
	return build, err
}

// Delete asynchronously deletes the Build specified by its id.
func (storage *BuildRegistryStorage) Delete(id string) (<-chan interface{}, error) {
	return apiserver.MakeAsync(func() (interface{}, error) {
		return api.Status{Status: api.StatusSuccess}, storage.registry.DeleteBuild(id)
	}), nil
}

// Extract deserializes user provided data into an buildapi.Build.
func (storage *BuildRegistryStorage) Extract(body []byte, queryParams map[string][]string) (interface{}, error) {
	result := buildapi.Build{}
	values := queryParams["plugin"]
	//if plugin specified then parse body using plugin
	if len(values) > 0 {
		
		
	} else {
		err := api.DecodeInto(body, &result)
		if err != nil {
			return nil, err
		}
	}
	
	values = queryParams["build_config_id"]
	if len(values) > 0 {
		buildConfigID := values[0]
		glog.Infof("config id %s", buildConfigID)
		buildConfig, err := storage.buildConfigRegistry.GetBuildConfig(buildConfigID)
		glog.Infof("config %#v", buildConfig)
		if err != nil {
			return nil, err
		}
		result.Config.Type = buildConfig.Type
		result.Config.SourceURI = buildConfig.SourceURI
		result.Config.ImageTag = buildConfig.ImageTag   
		result.Config.BuilderImage = buildConfig.BuilderImage
		result.Config.SourceRef = buildConfig.SourceRef
	}
	
	return result, nil
}

// Create registers a given new Build instance to storage.registry.
func (storage *BuildRegistryStorage) Create(obj interface{}) (<-chan interface{}, error) {
	build, ok := obj.(buildapi.Build)
	if !ok {
		return nil, fmt.Errorf("not a build: %#v", obj)
	}
	if len(build.ID) == 0 {
		build.ID = uuid.NewUUID().String()
	}
	if len(build.Status) == 0 {
		build.Status = buildapi.BuildNew
	}

	if build.CreationTimestamp == "" {
		build.CreationTimestamp = time.Now().Format(time.UnixDate)
	}

	return apiserver.MakeAsync(func() (interface{}, error) {
		err := storage.registry.CreateBuild(build)
		if err != nil {
			return nil, err
		}
		return build, nil
	}), nil
}

// Update replaces a given Build instance with an existing instance in storage.registry.
func (storage *BuildRegistryStorage) Update(obj interface{}) (<-chan interface{}, error) {
	build, ok := obj.(buildapi.Build)
	if !ok {
		return nil, fmt.Errorf("not a build: %#v", obj)
	}
	if len(build.ID) == 0 {
		return nil, fmt.Errorf("ID should not be empty: %#v", build)
	}
	return apiserver.MakeAsync(func() (interface{}, error) {
		err := storage.registry.UpdateBuild(build)
		if err != nil {
			return nil, err
		}
		return build, nil
	}), nil
}
