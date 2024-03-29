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
// Code generated by capnpc-go. DO NOT EDIT.

package node_common_capnp

import (
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

type VnObjectItemsGetResponseHeader struct{ capnp.Struct }

// VnObjectItemsGetResponseHeader_TypeID is the unique identifier for the type VnObjectItemsGetResponseHeader.
const VnObjectItemsGetResponseHeader_TypeID = 0x85032dcfe77493da

func NewVnObjectItemsGetResponseHeader(s *capnp.Segment) (VnObjectItemsGetResponseHeader, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 32, PointerCount: 2})
	return VnObjectItemsGetResponseHeader{st}, err
}

func NewRootVnObjectItemsGetResponseHeader(s *capnp.Segment) (VnObjectItemsGetResponseHeader, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 32, PointerCount: 2})
	return VnObjectItemsGetResponseHeader{st}, err
}

func ReadRootVnObjectItemsGetResponseHeader(msg *capnp.Message) (VnObjectItemsGetResponseHeader, error) {
	root, err := msg.RootPtr()
	return VnObjectItemsGetResponseHeader{root.Struct()}, err
}

func (s VnObjectItemsGetResponseHeader) String() string {
	str, _ := text.Marshal(0x85032dcfe77493da, s.Struct)
	return str
}

func (s VnObjectItemsGetResponseHeader) Marker() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s VnObjectItemsGetResponseHeader) HasMarker() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s VnObjectItemsGetResponseHeader) MarkerBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s VnObjectItemsGetResponseHeader) SetMarker(v string) error {
	return s.Struct.SetText(0, v)
}

func (s VnObjectItemsGetResponseHeader) ScanState() (VnObjectItemsScanCookie, error) {
	p, err := s.Struct.Ptr(1)
	return VnObjectItemsScanCookie{Struct: p.Struct()}, err
}

func (s VnObjectItemsGetResponseHeader) HasScanState() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s VnObjectItemsGetResponseHeader) SetScanState(v VnObjectItemsScanCookie) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewScanState sets the scanState field to a newly
// allocated VnObjectItemsScanCookie struct, preferring placement in s's segment.
func (s VnObjectItemsGetResponseHeader) NewScanState() (VnObjectItemsScanCookie, error) {
	ss, err := NewVnObjectItemsScanCookie(s.Struct.Segment())
	if err != nil {
		return VnObjectItemsScanCookie{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

func (s VnObjectItemsGetResponseHeader) HasMore() bool {
	return s.Struct.Bit(0)
}

func (s VnObjectItemsGetResponseHeader) SetHasMore(v bool) {
	s.Struct.SetBit(0, v)
}

func (s VnObjectItemsGetResponseHeader) NumItems() uint64 {
	return s.Struct.Uint64(8)
}

func (s VnObjectItemsGetResponseHeader) SetNumItems(v uint64) {
	s.Struct.SetUint64(8, v)
}

func (s VnObjectItemsGetResponseHeader) NumKeys() uint64 {
	return s.Struct.Uint64(16)
}

func (s VnObjectItemsGetResponseHeader) SetNumKeys(v uint64) {
	s.Struct.SetUint64(16, v)
}

func (s VnObjectItemsGetResponseHeader) NumValues() uint64 {
	return s.Struct.Uint64(24)
}

func (s VnObjectItemsGetResponseHeader) SetNumValues(v uint64) {
	s.Struct.SetUint64(24, v)
}

// VnObjectItemsGetResponseHeader_List is a list of VnObjectItemsGetResponseHeader.
type VnObjectItemsGetResponseHeader_List struct{ capnp.List }

// NewVnObjectItemsGetResponseHeader creates a new list of VnObjectItemsGetResponseHeader.
func NewVnObjectItemsGetResponseHeader_List(s *capnp.Segment, sz int32) (VnObjectItemsGetResponseHeader_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 32, PointerCount: 2}, sz)
	return VnObjectItemsGetResponseHeader_List{l}, err
}

func (s VnObjectItemsGetResponseHeader_List) At(i int) VnObjectItemsGetResponseHeader {
	return VnObjectItemsGetResponseHeader{s.List.Struct(i)}
}

func (s VnObjectItemsGetResponseHeader_List) Set(i int, v VnObjectItemsGetResponseHeader) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s VnObjectItemsGetResponseHeader_List) String() string {
	str, _ := text.MarshalList(0x85032dcfe77493da, s.List)
	return str
}

// VnObjectItemsGetResponseHeader_Promise is a wrapper for a VnObjectItemsGetResponseHeader promised by a client call.
type VnObjectItemsGetResponseHeader_Promise struct{ *capnp.Pipeline }

func (p VnObjectItemsGetResponseHeader_Promise) Struct() (VnObjectItemsGetResponseHeader, error) {
	s, err := p.Pipeline.Struct()
	return VnObjectItemsGetResponseHeader{s}, err
}

func (p VnObjectItemsGetResponseHeader_Promise) ScanState() VnObjectItemsScanCookie_Promise {
	return VnObjectItemsScanCookie_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

type VnObjectItemsGetMappedKeyValuePair struct{ capnp.Struct }

// VnObjectItemsGetMappedKeyValuePair_TypeID is the unique identifier for the type VnObjectItemsGetMappedKeyValuePair.
const VnObjectItemsGetMappedKeyValuePair_TypeID = 0xd26f89592f5fc95f

func NewVnObjectItemsGetMappedKeyValuePair(s *capnp.Segment) (VnObjectItemsGetMappedKeyValuePair, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 0})
	return VnObjectItemsGetMappedKeyValuePair{st}, err
}

func NewRootVnObjectItemsGetMappedKeyValuePair(s *capnp.Segment) (VnObjectItemsGetMappedKeyValuePair, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 16, PointerCount: 0})
	return VnObjectItemsGetMappedKeyValuePair{st}, err
}

func ReadRootVnObjectItemsGetMappedKeyValuePair(msg *capnp.Message) (VnObjectItemsGetMappedKeyValuePair, error) {
	root, err := msg.RootPtr()
	return VnObjectItemsGetMappedKeyValuePair{root.Struct()}, err
}

func (s VnObjectItemsGetMappedKeyValuePair) String() string {
	str, _ := text.Marshal(0xd26f89592f5fc95f, s.Struct)
	return str
}

func (s VnObjectItemsGetMappedKeyValuePair) KeyMapIndex() uint64 {
	return s.Struct.Uint64(0)
}

func (s VnObjectItemsGetMappedKeyValuePair) SetKeyMapIndex(v uint64) {
	s.Struct.SetUint64(0, v)
}

func (s VnObjectItemsGetMappedKeyValuePair) ValueMapIndex() uint64 {
	return s.Struct.Uint64(8)
}

func (s VnObjectItemsGetMappedKeyValuePair) SetValueMapIndex(v uint64) {
	s.Struct.SetUint64(8, v)
}

// VnObjectItemsGetMappedKeyValuePair_List is a list of VnObjectItemsGetMappedKeyValuePair.
type VnObjectItemsGetMappedKeyValuePair_List struct{ capnp.List }

// NewVnObjectItemsGetMappedKeyValuePair creates a new list of VnObjectItemsGetMappedKeyValuePair.
func NewVnObjectItemsGetMappedKeyValuePair_List(s *capnp.Segment, sz int32) (VnObjectItemsGetMappedKeyValuePair_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 16, PointerCount: 0}, sz)
	return VnObjectItemsGetMappedKeyValuePair_List{l}, err
}

func (s VnObjectItemsGetMappedKeyValuePair_List) At(i int) VnObjectItemsGetMappedKeyValuePair {
	return VnObjectItemsGetMappedKeyValuePair{s.List.Struct(i)}
}

func (s VnObjectItemsGetMappedKeyValuePair_List) Set(i int, v VnObjectItemsGetMappedKeyValuePair) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s VnObjectItemsGetMappedKeyValuePair_List) String() string {
	str, _ := text.MarshalList(0xd26f89592f5fc95f, s.List)
	return str
}

// VnObjectItemsGetMappedKeyValuePair_Promise is a wrapper for a VnObjectItemsGetMappedKeyValuePair promised by a client call.
type VnObjectItemsGetMappedKeyValuePair_Promise struct{ *capnp.Pipeline }

func (p VnObjectItemsGetMappedKeyValuePair_Promise) Struct() (VnObjectItemsGetMappedKeyValuePair, error) {
	s, err := p.Pipeline.Struct()
	return VnObjectItemsGetMappedKeyValuePair{s}, err
}

type VnObjectItemsGetItem struct{ capnp.Struct }

// VnObjectItemsGetItem_TypeID is the unique identifier for the type VnObjectItemsGetItem.
const VnObjectItemsGetItem_TypeID = 0xeacf121c1705f7c7

func NewVnObjectItemsGetItem(s *capnp.Segment) (VnObjectItemsGetItem, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return VnObjectItemsGetItem{st}, err
}

func NewRootVnObjectItemsGetItem(s *capnp.Segment) (VnObjectItemsGetItem, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2})
	return VnObjectItemsGetItem{st}, err
}

func ReadRootVnObjectItemsGetItem(msg *capnp.Message) (VnObjectItemsGetItem, error) {
	root, err := msg.RootPtr()
	return VnObjectItemsGetItem{root.Struct()}, err
}

func (s VnObjectItemsGetItem) String() string {
	str, _ := text.Marshal(0xeacf121c1705f7c7, s.Struct)
	return str
}

func (s VnObjectItemsGetItem) Name() (string, error) {
	p, err := s.Struct.Ptr(0)
	return p.Text(), err
}

func (s VnObjectItemsGetItem) HasName() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s VnObjectItemsGetItem) NameBytes() ([]byte, error) {
	p, err := s.Struct.Ptr(0)
	return p.TextBytes(), err
}

func (s VnObjectItemsGetItem) SetName(v string) error {
	return s.Struct.SetText(0, v)
}

func (s VnObjectItemsGetItem) Attrs() (VnObjectItemsGetMappedKeyValuePair_List, error) {
	p, err := s.Struct.Ptr(1)
	return VnObjectItemsGetMappedKeyValuePair_List{List: p.List()}, err
}

func (s VnObjectItemsGetItem) HasAttrs() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s VnObjectItemsGetItem) SetAttrs(v VnObjectItemsGetMappedKeyValuePair_List) error {
	return s.Struct.SetPtr(1, v.List.ToPtr())
}

// NewAttrs sets the attrs field to a newly
// allocated VnObjectItemsGetMappedKeyValuePair_List, preferring placement in s's segment.
func (s VnObjectItemsGetItem) NewAttrs(n int32) (VnObjectItemsGetMappedKeyValuePair_List, error) {
	l, err := NewVnObjectItemsGetMappedKeyValuePair_List(s.Struct.Segment(), n)
	if err != nil {
		return VnObjectItemsGetMappedKeyValuePair_List{}, err
	}
	err = s.Struct.SetPtr(1, l.List.ToPtr())
	return l, err
}

// VnObjectItemsGetItem_List is a list of VnObjectItemsGetItem.
type VnObjectItemsGetItem_List struct{ capnp.List }

// NewVnObjectItemsGetItem creates a new list of VnObjectItemsGetItem.
func NewVnObjectItemsGetItem_List(s *capnp.Segment, sz int32) (VnObjectItemsGetItem_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 2}, sz)
	return VnObjectItemsGetItem_List{l}, err
}

func (s VnObjectItemsGetItem_List) At(i int) VnObjectItemsGetItem {
	return VnObjectItemsGetItem{s.List.Struct(i)}
}

func (s VnObjectItemsGetItem_List) Set(i int, v VnObjectItemsGetItem) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s VnObjectItemsGetItem_List) String() string {
	str, _ := text.MarshalList(0xeacf121c1705f7c7, s.List)
	return str
}

// VnObjectItemsGetItem_Promise is a wrapper for a VnObjectItemsGetItem promised by a client call.
type VnObjectItemsGetItem_Promise struct{ *capnp.Pipeline }

func (p VnObjectItemsGetItem_Promise) Struct() (VnObjectItemsGetItem, error) {
	s, err := p.Pipeline.Struct()
	return VnObjectItemsGetItem{s}, err
}

type VnObjectItemsGetItemPtr struct{ capnp.Struct }

// VnObjectItemsGetItemPtr_TypeID is the unique identifier for the type VnObjectItemsGetItemPtr.
const VnObjectItemsGetItemPtr_TypeID = 0xf020cf2eadb0357c

func NewVnObjectItemsGetItemPtr(s *capnp.Segment) (VnObjectItemsGetItemPtr, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return VnObjectItemsGetItemPtr{st}, err
}

func NewRootVnObjectItemsGetItemPtr(s *capnp.Segment) (VnObjectItemsGetItemPtr, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return VnObjectItemsGetItemPtr{st}, err
}

func ReadRootVnObjectItemsGetItemPtr(msg *capnp.Message) (VnObjectItemsGetItemPtr, error) {
	root, err := msg.RootPtr()
	return VnObjectItemsGetItemPtr{root.Struct()}, err
}

func (s VnObjectItemsGetItemPtr) String() string {
	str, _ := text.Marshal(0xf020cf2eadb0357c, s.Struct)
	return str
}

func (s VnObjectItemsGetItemPtr) Item() (VnObjectItemsGetItem, error) {
	p, err := s.Struct.Ptr(0)
	return VnObjectItemsGetItem{Struct: p.Struct()}, err
}

func (s VnObjectItemsGetItemPtr) HasItem() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s VnObjectItemsGetItemPtr) SetItem(v VnObjectItemsGetItem) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewItem sets the item field to a newly
// allocated VnObjectItemsGetItem struct, preferring placement in s's segment.
func (s VnObjectItemsGetItemPtr) NewItem() (VnObjectItemsGetItem, error) {
	ss, err := NewVnObjectItemsGetItem(s.Struct.Segment())
	if err != nil {
		return VnObjectItemsGetItem{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// VnObjectItemsGetItemPtr_List is a list of VnObjectItemsGetItemPtr.
type VnObjectItemsGetItemPtr_List struct{ capnp.List }

// NewVnObjectItemsGetItemPtr creates a new list of VnObjectItemsGetItemPtr.
func NewVnObjectItemsGetItemPtr_List(s *capnp.Segment, sz int32) (VnObjectItemsGetItemPtr_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return VnObjectItemsGetItemPtr_List{l}, err
}

func (s VnObjectItemsGetItemPtr_List) At(i int) VnObjectItemsGetItemPtr {
	return VnObjectItemsGetItemPtr{s.List.Struct(i)}
}

func (s VnObjectItemsGetItemPtr_List) Set(i int, v VnObjectItemsGetItemPtr) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s VnObjectItemsGetItemPtr_List) String() string {
	str, _ := text.MarshalList(0xf020cf2eadb0357c, s.List)
	return str
}

// VnObjectItemsGetItemPtr_Promise is a wrapper for a VnObjectItemsGetItemPtr promised by a client call.
type VnObjectItemsGetItemPtr_Promise struct{ *capnp.Pipeline }

func (p VnObjectItemsGetItemPtr_Promise) Struct() (VnObjectItemsGetItemPtr, error) {
	s, err := p.Pipeline.Struct()
	return VnObjectItemsGetItemPtr{s}, err
}

func (p VnObjectItemsGetItemPtr_Promise) Item() VnObjectItemsGetItem_Promise {
	return VnObjectItemsGetItem_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type VnObjectItemsGetResponseDataPayload struct{ capnp.Struct }

// VnObjectItemsGetResponseDataPayload_TypeID is the unique identifier for the type VnObjectItemsGetResponseDataPayload.
const VnObjectItemsGetResponseDataPayload_TypeID = 0xbb85e91da7f4c0a3

func NewVnObjectItemsGetResponseDataPayload(s *capnp.Segment) (VnObjectItemsGetResponseDataPayload, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return VnObjectItemsGetResponseDataPayload{st}, err
}

func NewRootVnObjectItemsGetResponseDataPayload(s *capnp.Segment) (VnObjectItemsGetResponseDataPayload, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1})
	return VnObjectItemsGetResponseDataPayload{st}, err
}

func ReadRootVnObjectItemsGetResponseDataPayload(msg *capnp.Message) (VnObjectItemsGetResponseDataPayload, error) {
	root, err := msg.RootPtr()
	return VnObjectItemsGetResponseDataPayload{root.Struct()}, err
}

func (s VnObjectItemsGetResponseDataPayload) String() string {
	str, _ := text.Marshal(0xbb85e91da7f4c0a3, s.Struct)
	return str
}

func (s VnObjectItemsGetResponseDataPayload) ValueMap() (VnObjectAttributeValueMap, error) {
	p, err := s.Struct.Ptr(0)
	return VnObjectAttributeValueMap{Struct: p.Struct()}, err
}

func (s VnObjectItemsGetResponseDataPayload) HasValueMap() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s VnObjectItemsGetResponseDataPayload) SetValueMap(v VnObjectAttributeValueMap) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewValueMap sets the valueMap field to a newly
// allocated VnObjectAttributeValueMap struct, preferring placement in s's segment.
func (s VnObjectItemsGetResponseDataPayload) NewValueMap() (VnObjectAttributeValueMap, error) {
	ss, err := NewVnObjectAttributeValueMap(s.Struct.Segment())
	if err != nil {
		return VnObjectAttributeValueMap{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

// VnObjectItemsGetResponseDataPayload_List is a list of VnObjectItemsGetResponseDataPayload.
type VnObjectItemsGetResponseDataPayload_List struct{ capnp.List }

// NewVnObjectItemsGetResponseDataPayload creates a new list of VnObjectItemsGetResponseDataPayload.
func NewVnObjectItemsGetResponseDataPayload_List(s *capnp.Segment, sz int32) (VnObjectItemsGetResponseDataPayload_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 1}, sz)
	return VnObjectItemsGetResponseDataPayload_List{l}, err
}

func (s VnObjectItemsGetResponseDataPayload_List) At(i int) VnObjectItemsGetResponseDataPayload {
	return VnObjectItemsGetResponseDataPayload{s.List.Struct(i)}
}

func (s VnObjectItemsGetResponseDataPayload_List) Set(i int, v VnObjectItemsGetResponseDataPayload) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s VnObjectItemsGetResponseDataPayload_List) String() string {
	str, _ := text.MarshalList(0xbb85e91da7f4c0a3, s.List)
	return str
}

// VnObjectItemsGetResponseDataPayload_Promise is a wrapper for a VnObjectItemsGetResponseDataPayload promised by a client call.
type VnObjectItemsGetResponseDataPayload_Promise struct{ *capnp.Pipeline }

func (p VnObjectItemsGetResponseDataPayload_Promise) Struct() (VnObjectItemsGetResponseDataPayload, error) {
	s, err := p.Pipeline.Struct()
	return VnObjectItemsGetResponseDataPayload{s}, err
}

func (p VnObjectItemsGetResponseDataPayload_Promise) ValueMap() VnObjectAttributeValueMap_Promise {
	return VnObjectAttributeValueMap_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

type VnObjectItemsGetResponseMetadataPayload struct{ capnp.Struct }

// VnObjectItemsGetResponseMetadataPayload_TypeID is the unique identifier for the type VnObjectItemsGetResponseMetadataPayload.
const VnObjectItemsGetResponseMetadataPayload_TypeID = 0xb4008849dd7a3304

func NewVnObjectItemsGetResponseMetadataPayload(s *capnp.Segment) (VnObjectItemsGetResponseMetadataPayload, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3})
	return VnObjectItemsGetResponseMetadataPayload{st}, err
}

func NewRootVnObjectItemsGetResponseMetadataPayload(s *capnp.Segment) (VnObjectItemsGetResponseMetadataPayload, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3})
	return VnObjectItemsGetResponseMetadataPayload{st}, err
}

func ReadRootVnObjectItemsGetResponseMetadataPayload(msg *capnp.Message) (VnObjectItemsGetResponseMetadataPayload, error) {
	root, err := msg.RootPtr()
	return VnObjectItemsGetResponseMetadataPayload{root.Struct()}, err
}

func (s VnObjectItemsGetResponseMetadataPayload) String() string {
	str, _ := text.Marshal(0xb4008849dd7a3304, s.Struct)
	return str
}

func (s VnObjectItemsGetResponseMetadataPayload) ValueMap() (VnObjectAttributeValueMap, error) {
	p, err := s.Struct.Ptr(0)
	return VnObjectAttributeValueMap{Struct: p.Struct()}, err
}

func (s VnObjectItemsGetResponseMetadataPayload) HasValueMap() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s VnObjectItemsGetResponseMetadataPayload) SetValueMap(v VnObjectAttributeValueMap) error {
	return s.Struct.SetPtr(0, v.Struct.ToPtr())
}

// NewValueMap sets the valueMap field to a newly
// allocated VnObjectAttributeValueMap struct, preferring placement in s's segment.
func (s VnObjectItemsGetResponseMetadataPayload) NewValueMap() (VnObjectAttributeValueMap, error) {
	ss, err := NewVnObjectAttributeValueMap(s.Struct.Segment())
	if err != nil {
		return VnObjectAttributeValueMap{}, err
	}
	err = s.Struct.SetPtr(0, ss.Struct.ToPtr())
	return ss, err
}

func (s VnObjectItemsGetResponseMetadataPayload) KeyMap() (VnObjectAttributeKeyMap, error) {
	p, err := s.Struct.Ptr(1)
	return VnObjectAttributeKeyMap{Struct: p.Struct()}, err
}

func (s VnObjectItemsGetResponseMetadataPayload) HasKeyMap() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s VnObjectItemsGetResponseMetadataPayload) SetKeyMap(v VnObjectAttributeKeyMap) error {
	return s.Struct.SetPtr(1, v.Struct.ToPtr())
}

// NewKeyMap sets the keyMap field to a newly
// allocated VnObjectAttributeKeyMap struct, preferring placement in s's segment.
func (s VnObjectItemsGetResponseMetadataPayload) NewKeyMap() (VnObjectAttributeKeyMap, error) {
	ss, err := NewVnObjectAttributeKeyMap(s.Struct.Segment())
	if err != nil {
		return VnObjectAttributeKeyMap{}, err
	}
	err = s.Struct.SetPtr(1, ss.Struct.ToPtr())
	return ss, err
}

func (s VnObjectItemsGetResponseMetadataPayload) Items() (VnObjectItemsGetItemPtr_List, error) {
	p, err := s.Struct.Ptr(2)
	return VnObjectItemsGetItemPtr_List{List: p.List()}, err
}

func (s VnObjectItemsGetResponseMetadataPayload) HasItems() bool {
	p, err := s.Struct.Ptr(2)
	return p.IsValid() || err != nil
}

func (s VnObjectItemsGetResponseMetadataPayload) SetItems(v VnObjectItemsGetItemPtr_List) error {
	return s.Struct.SetPtr(2, v.List.ToPtr())
}

// NewItems sets the items field to a newly
// allocated VnObjectItemsGetItemPtr_List, preferring placement in s's segment.
func (s VnObjectItemsGetResponseMetadataPayload) NewItems(n int32) (VnObjectItemsGetItemPtr_List, error) {
	l, err := NewVnObjectItemsGetItemPtr_List(s.Struct.Segment(), n)
	if err != nil {
		return VnObjectItemsGetItemPtr_List{}, err
	}
	err = s.Struct.SetPtr(2, l.List.ToPtr())
	return l, err
}

// VnObjectItemsGetResponseMetadataPayload_List is a list of VnObjectItemsGetResponseMetadataPayload.
type VnObjectItemsGetResponseMetadataPayload_List struct{ capnp.List }

// NewVnObjectItemsGetResponseMetadataPayload creates a new list of VnObjectItemsGetResponseMetadataPayload.
func NewVnObjectItemsGetResponseMetadataPayload_List(s *capnp.Segment, sz int32) (VnObjectItemsGetResponseMetadataPayload_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 3}, sz)
	return VnObjectItemsGetResponseMetadataPayload_List{l}, err
}

func (s VnObjectItemsGetResponseMetadataPayload_List) At(i int) VnObjectItemsGetResponseMetadataPayload {
	return VnObjectItemsGetResponseMetadataPayload{s.List.Struct(i)}
}

func (s VnObjectItemsGetResponseMetadataPayload_List) Set(i int, v VnObjectItemsGetResponseMetadataPayload) error {
	return s.List.SetStruct(i, v.Struct)
}

func (s VnObjectItemsGetResponseMetadataPayload_List) String() string {
	str, _ := text.MarshalList(0xb4008849dd7a3304, s.List)
	return str
}

// VnObjectItemsGetResponseMetadataPayload_Promise is a wrapper for a VnObjectItemsGetResponseMetadataPayload promised by a client call.
type VnObjectItemsGetResponseMetadataPayload_Promise struct{ *capnp.Pipeline }

func (p VnObjectItemsGetResponseMetadataPayload_Promise) Struct() (VnObjectItemsGetResponseMetadataPayload, error) {
	s, err := p.Pipeline.Struct()
	return VnObjectItemsGetResponseMetadataPayload{s}, err
}

func (p VnObjectItemsGetResponseMetadataPayload_Promise) ValueMap() VnObjectAttributeValueMap_Promise {
	return VnObjectAttributeValueMap_Promise{Pipeline: p.Pipeline.GetPipeline(0)}
}

func (p VnObjectItemsGetResponseMetadataPayload_Promise) KeyMap() VnObjectAttributeKeyMap_Promise {
	return VnObjectAttributeKeyMap_Promise{Pipeline: p.Pipeline.GetPipeline(1)}
}

const schema_dfe00955984fcb17 = "x\xda\xac\x94OH\\W\x14\xc6\xcfw\xef\x9by3" +
	"0\xad\xf3:\x03\xdaR\x99\xb6t\xa1\x15\xff\xdbEm" +
	"e\xdab\xa9S;8\xd7\xb6\x96\xae\xecu\xe6\x82V" +
	"\xe7\xcd\xf0\xe6\xd9v\xa4\xe0\xa6B\xbbh\xbb\xb0\x8b\xb6" +
	"T\x10B\x88!J\"($d\x11w\x12\xb2\x10\x12" +
	"\xb2\x88\x10\x92EB\xfe\xac\x12HH\\\xbdp\xc7q" +
	"\x9c\x0c\x09\x01\xcd\xea\xbdw\xee\xc7}\xbf\xef;\xe7\xde" +
	"\x8e4\xfb\x98u\xfa\x86\x82Db\xca\xe7\xf7\xb6\xe7\xdd" +
	"[[\xad|\x8eD\x0b\x0c\xaf\xfe\xc2\xd0?\xdf\x04\xaf" +
	"_#\x1f3\x89\xba\xff0\x86Yd\xd5\xd0\xaf+\xc6" +
	"\xb7 \x9c3\xbag\xae&~[\xb3ZP\xa5\xe5Z" +
	"\xe0\xf3\xff\xc7\"\xad~\x93(\xd2\xec?I\xf0\x8el" +
	"<8\xd6xg\xee,=-\x86\x16_\xf6\xcf\xb0\xc8" +
	"NI\xfc\xd0\x1f'x\xa3\xe7G\xdb\xbf\xfb=wQ" +
	"c\xb0}u\xe9\xd7\xaf\x9b\x0e\x8b\xf4\x99Z\xfc\x81\xf9" +
	"\x13\xc1\xdb|\xe4\xab\x7f\xf3\xb5\xad\xbb5;\x97\x90\x17" +
	"L\xc6\"gJ\xe2\xf5\x92\xf8\x97\xf7O\xad\xb4m\xbd" +
	"u\xefY\x18o\x07\xde`\x91O\x02Z\xdc\x17\x88\xd3" +
	"m\xcf\xceeT{:\x97\x0ddsv\xfb\x88=4" +
	"\xf6\x83J\xbb\x09We\x0b\x9f+wX\x15\xf29\xbb" +
	"\xa0\xda\xd22o\xe7{\x9f\xb7<\xa0dF\xc1I\x01" +
	"\xa2\x81\x1bD\x06\x88\xac\x7f{\x89\xc4\xdf\x1cb\x91\xc1" +
	"\x02\xa2\xd0\xc5\x85a\"\xf1?\x87Xb\x00\x8b\x82\x11" +
	"YG?%\x12\x8b\x1cb\x99\xc1\xe2\x88\x82\x13Y\xc7" +
	"\xbf \x12K\x1cb\x8d\xc12X\x14\x06\x91\xb5\xaa\x95" +
	"\xcb\x1c\xe24\x83\xe5\xe3Q\xf8\x88\xacu\xbd\xe5\x1a\x87" +
	"\xd8`\x88g\xa53\xa9\x1c\x84\x88!D\xf0\x0aii" +
	"\x7f\xe5J\x97\xa0\x10\xf6\xbe\x9f\xdf<\xd1\x18`\xbf\x12" +
	"\x01a\xc2\xec\xb8,$s\x8e\x02\x88\x01\x04\xcf\x9e\xce" +
	"\x96\x9c\x11\x11\x82\xc4\x10$\xcc\xda\xd3\xd9AU,\xec" +
	"}k\xcd\x88\x9c\x9aV\x84\xaaZ9\xc2\xe0!\"L" +
	"*Wf\xa4+S\xb28\x95\xe32\xa3\xb3\x0cU\xb2" +
	"\xfcL\xa7\xd1\xcf!RUY&u\xc0\x03\x1c\xe2k" +
	"\x06\x8b\x95\xc3\x14]D\xe2K\x0e1\xce\xe0\xfd\xa8I" +
	"\x932\xaf\x0d\x85\xbd\x9d\x96+7\xff\xdc\xbeq\xbfl" +
	"?>\xa9\x8aI\x99G\xd8kz\xfc\xe1;\xab\x97z" +
	"\xfe*/\xc4&4\x1b^%\xa48\x10\xde\x1f'\x82" +
	".\xbe\x94\x89\xe9\xd7V\xe3\xda\xeb\xaeU\xa3b\xf5\x15" +
	"m5\xc4!\x1a^d\xe0P I\x99\xcf\xab\xcc\xa0" +
	"*\xean\xc6TJN\x94\xc67P\xe1h\x1e#\x12" +
	"M\x1c\xa2\xa7*\xf2N\x87Htp\x88\x8f\x18\xbc\xdd" +
	"\xfc\x126\x99\x19\xf5se\x18*\xc8\xb1\x84]]\xdf" +
	"\x835\x0f\x00\xab\x9fT\x83\xf7\x1e\x91x\x97CtT" +
	"\xe1\xb5v\x95\x99\xfb\x19\xeal\x99U{\x07!&]" +
	"\xd7\xa9ji\xe5\xee\xa9i\xe9A\xe1R\xdcuj\xda" +
	"\xa8\xf9\x02\x1c\"\xcaP\xa7\x07\x0a\xe1\xfdKl\xb7\x7f" +
	"O\x02\x00\x00\xff\xff\xb4\xc6\x85\xbd"

func init() {
	schemas.Register(schema_dfe00955984fcb17,
		0x85032dcfe77493da,
		0xb4008849dd7a3304,
		0xbb85e91da7f4c0a3,
		0xd26f89592f5fc95f,
		0xeacf121c1705f7c7,
		0xf020cf2eadb0357c)
}
