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

package v3iohttp

// function names
const (
	putItemFunctionName        = "PutItem"
	updateItemFunctionName     = "UpdateItem"
	getItemFunctionName        = "GetItem"
	getItemsFunctionName       = "GetItems"
	createStreamFunctionName   = "CreateStream"
	describeStreamFunctionName = "DescribeStream"
	putRecordsFunctionName     = "PutRecords"
	getRecordsFunctionName     = "GetRecords"
	seekShardsFunctionName     = "SeekShard"
	getClusterMDFunctionName   = "GetClusterMD"
	putOOSObjectFunctionName   = "OosRun"
	PutChunkFunctionName       = "PutChunk"
)

// headers for put item
var putItemHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": putItemFunctionName,
}

// headers for GetClusterMD
var getClusterMDHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": getClusterMDFunctionName,
}

// headers for update item
var updateItemHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": updateItemFunctionName,
}

// headers for get item
var getItemHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": getItemFunctionName,
}

// headers for get items
var getItemsHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": getItemsFunctionName,
}

// headers for get items requesting captain-proto response
var getItemsHeadersCapnp = map[string]string{
	"Content-Type":                 "application/json",
	"X-v3io-function":              getItemsFunctionName,
	"X-v3io-response-content-type": "capnp",
}

// headers for create stream
var createStreamHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": createStreamFunctionName,
}

// headers for get records
var describeStreamHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": describeStreamFunctionName,
}

// headers for put records
var putRecordsHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": putRecordsFunctionName,
}

// headers for put chunks
var putChunkHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": PutChunkFunctionName,
}

// headers for put records
var getRecordsHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": getRecordsFunctionName,
}

// headers for seek records
var seekShardsHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": seekShardsFunctionName,
}

// headers for OOS put object
var putOOSObjectHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": putOOSObjectFunctionName,
}

// map between SeekShardInputType and its encoded counterpart
var seekShardsInputTypeToString = [...]string{
	"TIME",
	"SEQUENCE",
	"LATEST",
	"EARLIEST",
}
