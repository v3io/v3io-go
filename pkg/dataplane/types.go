/*
Copyright 2018 The v3io Authors.

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

package v3io

import (
	"context"
	"time"
)

type NewSessionInput struct {
	Username  string
	Password  string
	AccessKey string
}

type NewContainerInput struct {
	ContainerName string
}

type DataPlaneInput struct {
	Ctx                 context.Context
	ContainerName       string
	AuthenticationToken string
	AccessKey           string
	Timeout             time.Duration
}

type DataPlaneOutput struct {
	ctx context.Context
}

type PutItemInput struct {
	DataPlaneInput
	Path       string
	Attributes map[string]interface{}
}

type PutItemsInput struct {
	DataPlaneInput
	Path  string
	Items map[string]map[string]interface{}
}

type PutItemsOutput struct {
	DataPlaneOutput
	Success bool
	Errors  map[string]error
}

type UpdateItemInput struct {
	DataPlaneInput
	Path       string
	Attributes map[string]interface{}
	Expression *string
}

type GetItemInput struct {
	DataPlaneInput
	Path           string
	AttributeNames []string
}

type GetItemOutput struct {
	DataPlaneOutput
	Item Item
}

type GetItemsInput struct {
	DataPlaneInput
	Path           string
	AttributeNames []string
	Filter         string
	Marker         string
	ShardingKey    string
	Limit          int
	Segment        int
	TotalSegments  int
}

type GetItemsOutput struct {
	DataPlaneOutput
	Last       bool
	NextMarker string
	Items      []Item
}

type GetObjectInput struct {
	DataPlaneInput
	Path     string
	Offset   int
	NumBytes int
}

type PutObjectInput struct {
	DataPlaneInput
	Path   string
	Offset int
	Body   []byte
}

type DeleteObjectInput struct {
	DataPlaneInput
	Path string
}
