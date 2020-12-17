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
	"bytes"
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/v3io/v3io-go/pkg/errors"
)

type Item map[string]interface{}

func (i Item) GetField(name string) interface{} {
	return i[name]
}

func (i Item) GetFieldInt(name string) (int, error) {
	fieldValue, fieldFound := i[name]
	if !fieldFound {
		return 0, v3ioerrors.ErrNotFound
	}

	switch typedField := fieldValue.(type) {
	case int:
		return typedField, nil
	case float64:
		return int(typedField), nil
	case string:
		return strconv.Atoi(typedField)
	default:
		return 0, v3ioerrors.ErrInvalidTypeConversion
	}
}

func (i Item) GetFieldString(name string) (string, error) {
	fieldValue, fieldFound := i[name]
	if !fieldFound {
		return "", v3ioerrors.ErrNotFound
	}

	switch typedField := fieldValue.(type) {
	case int:
		return strconv.Itoa(typedField), nil
	case float64:
		return strconv.FormatFloat(typedField, 'E', -1, 64), nil
	case string:
		return typedField, nil
	default:
		return "", v3ioerrors.ErrInvalidTypeConversion
	}
}

func (i Item) GetFieldUint64(name string) (uint64, error) {
	fieldValue, fieldFound := i[name]
	if !fieldFound {
		return 0, v3ioerrors.ErrNotFound
	}

	switch typedField := fieldValue.(type) {
	// TODO: properly handle uint64
	case int:
		return uint64(typedField), nil
	case uint64:
		return typedField, nil
	default:
		return 0, v3ioerrors.ErrInvalidTypeConversion
	}
}

// For internal use only - DO NOT USE!
func (i Item) GetShard() (int, []*ItemChunkData, *ItemChunkMetadata, *ItemCurrentChunkMetadata, error) {
	const streamDataPrefix = "__data_stream["
	const streamMetadataPrefix = "__data_stream_metadata["
	const offsetPrefix = "__data_stream[0000]["

	var chunkDataArray []*ItemChunkData
	chunkMetaData := ItemChunkMetadata{}
	currentChunkMetadata := ItemCurrentChunkMetadata{}
	var chunkID int

	for k, v := range i {
		if strings.HasPrefix(k, streamDataPrefix) {
			chunkID, _ = strconv.Atoi(k[len(streamDataPrefix):][:4])
			offset, _ := strconv.ParseUint(k[len(offsetPrefix):][:16], 10, 64)
			data, ok := v.([]byte)
			if !ok {
				return 0, nil, nil, nil, v3ioerrors.ErrInvalidTypeConversion
			}
			streamData := ItemChunkData{Offset: offset, Data: &data}
			chunkDataArray = append(chunkDataArray, &streamData)
		}

		if strings.HasPrefix(k, streamMetadataPrefix) {
			chunkID, _ = strconv.Atoi(k[len(streamMetadataPrefix):][:4])
			metadata, ok := v.([]byte)
			if !ok {
				return 0, nil, nil, nil, v3ioerrors.ErrInvalidTypeConversion
			}

			buf := bytes.NewBuffer(metadata[8:64])
			err := binary.Read(buf, binary.LittleEndian, &chunkMetaData)
			if err != nil {
				return 0, nil, nil, nil, err
			}

			buf = bytes.NewBuffer(metadata[0:1])
			var isCurrent bool
			err = binary.Read(buf, binary.LittleEndian, &isCurrent)
			if err != nil {
				return 0, nil, nil, nil, err
			}
			if isCurrent {
				buf = bytes.NewBuffer(metadata[64:110])
				err = binary.Read(buf, binary.LittleEndian, &currentChunkMetadata)
				if err != nil {
					return 0, nil, nil, nil, err
				}
			}
		}
	}
	return chunkID, chunkDataArray, &chunkMetaData, &currentChunkMetadata, nil
}
