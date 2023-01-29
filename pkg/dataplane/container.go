/*
Copyright 2019 Iguazio Systems Ltd.

Licensed under the Apache License, Version 2.0 (the "License") with
an addition restriction as set forth herein. You may not use this
file except in compliance with the License. You may obtain a copy of
the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing
permissions and limitations under the License.

In addition, you may not use the software for any purposes that are
illegal under applicable law, and the grant of the foregoing license
under the Apache 2.0 license is conditioned upon your compliance with
such restriction.
*/

package v3io

// A container interface allows perform actions against a container
type Container interface {
	//
	// Container
	//

	// GetContainers
	GetClusterMD(*GetClusterMDInput, interface{}, chan *Response) (*Request, error)

	// GetContainersSync
	GetClusterMDSync(*GetClusterMDInput) (*Response, error)

	// GetContainers
	GetContainers(*GetContainersInput, interface{}, chan *Response) (*Request, error)

	// GetContainersSync
	GetContainersSync(*GetContainersInput) (*Response, error)

	// GetContainers
	GetContainerContents(*GetContainerContentsInput, interface{}, chan *Response) (*Request, error)

	// GetContainerContentsSync
	GetContainerContentsSync(*GetContainerContentsInput) (*Response, error)

	//
	// Object
	//
	// CheckPathExists
	CheckPathExists(*CheckPathExistsInput, interface{}, chan *Response) (*Request, error)

	// CheckPathExistsSync
	CheckPathExistsSync(*CheckPathExistsInput) error

	// GetObject
	GetObject(*GetObjectInput, interface{}, chan *Response) (*Request, error)

	// GetObjectSync
	GetObjectSync(*GetObjectInput) (*Response, error)

	// PutObject
	PutObject(*PutObjectInput, interface{}, chan *Response) (*Request, error)

	// PutObjectSync
	PutObjectSync(*PutObjectInput) error

	// UpdateObjectSync
	UpdateObjectSync(*UpdateObjectInput) error

	// DeleteObject
	DeleteObject(*DeleteObjectInput, interface{}, chan *Response) (*Request, error)

	// DeleteObjectSync
	DeleteObjectSync(*DeleteObjectInput) error

	//
	// KV
	//

	// GetItem
	GetItem(*GetItemInput, interface{}, chan *Response) (*Request, error)

	// GetItemSync
	GetItemSync(*GetItemInput) (*Response, error)

	// GetItems
	GetItems(*GetItemsInput, interface{}, chan *Response) (*Request, error)

	// GetItemSync
	GetItemsSync(*GetItemsInput) (*Response, error)

	// PutItem
	PutItem(*PutItemInput, interface{}, chan *Response) (*Request, error)

	// PutItemSync
	PutItemSync(*PutItemInput) (*Response, error)

	// PutItems
	PutItems(*PutItemsInput, interface{}, chan *Response) (*Request, error)

	// PutItemsSync
	PutItemsSync(*PutItemsInput) (*Response, error)

	// UpdateItem
	UpdateItem(*UpdateItemInput, interface{}, chan *Response) (*Request, error)

	// UpdateItemSync
	UpdateItemSync(*UpdateItemInput) (*Response, error)

	//
	// Stream
	//

	// CreateStream
	CreateStream(*CreateStreamInput, interface{}, chan *Response) (*Request, error)

	// CreateStreamSync
	CreateStreamSync(*CreateStreamInput) error

	// DescribeStream
	DescribeStream(*DescribeStreamInput, interface{}, chan *Response) (*Request, error)

	// DescribeStreamSync
	DescribeStreamSync(*DescribeStreamInput) (*Response, error)

	// DeleteStream
	DeleteStream(*DeleteStreamInput, interface{}, chan *Response) (*Request, error)

	// DeleteStreamSync
	DeleteStreamSync(*DeleteStreamInput) error

	// SeekShard
	SeekShard(*SeekShardInput, interface{}, chan *Response) (*Request, error)

	// SeekShardSync
	SeekShardSync(*SeekShardInput) (*Response, error)

	// PutRecords
	PutRecords(*PutRecordsInput, interface{}, chan *Response) (*Request, error)

	// PutRecordsSync
	PutRecordsSync(*PutRecordsInput) (*Response, error)

	// PutChunk
	PutChunk(*PutChunkInput, interface{}, chan *Response) (*Request, error)

	// PutChunkSync
	PutChunkSync(input *PutChunkInput) error

	// GetRecords
	GetRecords(*GetRecordsInput, interface{}, chan *Response) (*Request, error)

	// GetRecordsSync
	GetRecordsSync(*GetRecordsInput) (*Response, error)

	//
	// OOS
	//

	// PutOOSObject
	PutOOSObject(*PutOOSObjectInput, interface{}, chan *Response) (*Request, error)

	// PutOOSObjectSync
	PutOOSObjectSync(*PutOOSObjectInput) error
}
