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

// A container interface allows perform actions against a container
type Container interface {

	// GetObject
	GetObject(*GetObjectInput, interface{}, chan *Response) (*Request, error)

	// GetObjectSync
	GetObjectSync(*GetObjectInput) (*Response, error)

	// PutObject
	PutObject(*PutObjectInput, interface{}, chan *Response) (*Request, error)

	// PutObjectSync
	PutObjectSync(*PutObjectInput) error

	// DeleteObject
	DeleteObject(*DeleteObjectInput, interface{}, chan *Response) (*Request, error)

	// DeleteObjectSync
	DeleteObjectSync(*DeleteObjectInput) error

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
	PutItemSync(*PutItemInput) error

	// PutItems
	PutItems(*PutItemsInput, interface{}, chan *Response) (*Request, error)

	// PutItemsSync
	PutItemsSync(*PutItemsInput) (*Response, error)

	// UpdateItem
	UpdateItem(*UpdateItemInput, interface{}, chan *Response) (*Request, error)

	// UpdateItemSync
	UpdateItemSync(*UpdateItemInput) error
}
