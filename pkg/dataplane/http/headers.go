package v3iohttp

// function names
const (
	putItemFunctionName      = "PutItem"
	updateItemFunctionName   = "UpdateItem"
	getItemFunctionName      = "GetItem"
	getItemsFunctionName     = "GetItems"
	createStreamFunctionName = "CreateStream"
	putRecordsFunctionName   = "PutRecords"
	getRecordsFunctionName   = "GetRecords"
	seekShardsFunctionName   = "SeekShard"
	getGetClusterMD          = "GetClusterMD"
)

// headers for put item
var putItemHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": putItemFunctionName,
}

// headers for update item
var updateItemHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": updateItemFunctionName,
}

// headers for update item
var getItemHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": getItemFunctionName,
}

// headers for get item with captain-proto serialisation
var getItemsHeadersCapnp = map[string]string{
	"Content-Type":                 "application/json",
	"X-v3io-response-content-type": "capnp",
	"X-v3io-function":              getItemsFunctionName,
}

// headers for create stream
var createStreamHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": createStreamFunctionName,
}

// headers for put records
var putRecordsHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": putRecordsFunctionName,
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

// map between SeekShardInputType and its encoded counterpart
var seekShardsInputTypeToString = [...]string{
	"TIME",
	"SEQUENCE",
	"LATEST",
	"EARLIEST",
}

// headers get cluster metadata
var getGetClusterMetadataHeaders = map[string]string{
	"Content-Type":    "application/json",
	"X-v3io-function": getGetClusterMD,
}
