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

import (
	"context"
	"reflect"
	"runtime"
	"time"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type ItemsCursor struct {
	currentItem     Item
	currentError    error
	currentResponse *Response
	nextMarker      string
	moreItemsExist  bool
	itemIndex       int
	items           []Item
	getItemsInput   *GetItemsInput
	container       Container
	scattered       bool

	logger        logger.Logger
	retryAttempts int
	retryInterval time.Duration
}

func getFunctionName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

func retryFuncWithResult(ctx context.Context,
	loggerInstance logger.Logger,
	attempts int,
	retryInterval time.Duration,
	fn func() (interface{}, bool, error)) (interface{}, error) {

	// If retries are not defined, execute the function once
	if attempts <= 0 {
		attempts = 1
	}

	var err error
	var retry bool
	var result interface{}

	for attempt := 1; attempt <= attempts; attempt++ {
		result, retry, err = fn()

		// if there's no need to retry - we're done
		if !retry {
			return result, err
		}

		if ctx.Err() != nil {
			loggerInstance.WarnWithCtx(ctx,
				"Context error detected during retries",
				"ctxErr", ctx.Err(),
				"previousErr", err,
				"function", getFunctionName(fn),
				"attempt", attempt)

			// return the error if one was provided
			if err != nil {
				return result, err
			}

			return result, ctx.Err()
		}

		// not final attempt
		if attempt < attempts {

			// don't over log, no output
			loggerInstance.DebugWithCtx(ctx,
				"Failed an attempt to invoke function",
				"function", getFunctionName(fn),
				"err", errors.GetErrorStackString(err, 10),
				"attempt", attempt)
		}

		time.Sleep(retryInterval)
	}

	// attempts exhausted and we're unsuccessful
	// Return the original error for later checking
	loggerInstance.WarnWithCtx(ctx,
		"Failed final attempt to invoke function",
		"function", getFunctionName(fn),
		"err", errors.GetErrorStackString(err, 10),
		"attempts", attempts)

	// this shouldn't happen
	if err == nil {
		loggerInstance.ErrorWithCtx(ctx,
			"Failed final attempt to invoke function, but error is nil. This shouldn't happen",
			"function", getFunctionName(fn),
			"err", errors.GetErrorStackString(err, 10),
			"attempts", attempts)
		return result, errors.New("Failed final attempt to invoke function without proper error supplied")
	}
	return result, err
}

func executeV3ioRequestWithRetriesAndResult(ctx context.Context,
	logger logger.Logger,
	retries int,
	interval time.Duration,
	fn func() (interface{}, error)) (interface{}, error) {

	result, err := retryFuncWithResult(ctx,
		logger,
		retries,
		interval,
		func() (interface{}, bool, error) {
			result, err := fn()
			return result, err != nil, err
		})
	return result, err
}
func (ic *ItemsCursor) getItemsSync(container Container, getItemsInput *GetItemsInput) (*Response, error) {
	if getItemsInput.RetryAttempts < 2 {
		return container.GetItemsSync(getItemsInput)
	}
	responseInterface, err := executeV3ioRequestWithRetriesAndResult(context.TODO(),
		getItemsInput.Logger,
		getItemsInput.RetryAttempts,
		getItemsInput.RetryInterval,
		func() (interface{}, error) {
			response, err := container.GetItemsSync(getItemsInput)
			return response, err
		})
	return responseInterface.(*Response), err
}

func NewItemsCursor(container Container, getItemsInput *GetItemsInput) (*ItemsCursor, error) {
	newItemsCursor := &ItemsCursor{
		container:     container,
		getItemsInput: getItemsInput,
	}
	response, err := newItemsCursor.getItemsSync(container, getItemsInput)
	if err != nil {
		return nil, err
	}

	newItemsCursor.setResponse(response)

	return newItemsCursor, nil
}

// Err returns the last error
func (ic *ItemsCursor) Err() error {
	return ic.currentError
}

// Release releases a cursor and its underlying resources
func (ic *ItemsCursor) Release() {
	if ic.currentResponse != nil {
		ic.currentResponse.Release()
	}
}

// Next gets the next matching item. this may potentially block as this lazy loads items from the collection
func (ic *ItemsCursor) NextSync() bool {
	item, err := ic.NextItemSync()

	if item == nil || err != nil {
		return false
	}

	return true
}

// NextItem gets the next matching item. this may potentially block as this lazy loads items from the collection
func (ic *ItemsCursor) NextItemSync() (Item, error) {

	// are there any more items left in the previous response we received?
	if ic.itemIndex < len(ic.items) {
		ic.currentItem = ic.items[ic.itemIndex]
		ic.currentError = nil

		// next time we'll give next item
		ic.itemIndex++

		return ic.currentItem, nil
	}

	// are there any more items up stream?
	if !ic.moreItemsExist {
		ic.currentError = nil
		return nil, nil
	}

	// get the previous request input and modify it with the marker
	ic.getItemsInput.Marker = ic.nextMarker

	if ic.getItemsInput.ChokeGetItemsMS != 0 {
		time.Sleep(time.Duration(ic.getItemsInput.ChokeGetItemsMS) * time.Millisecond)
	}

	// invoke get items
	newResponse, err := ic.getItemsSync(ic.container, ic.getItemsInput)
	if err != nil {
		ic.currentError = err
		return nil, err
	}

	// release the previous response
	ic.currentResponse.Release()

	// set the new response - read all the sub information from it
	ic.setResponse(newResponse)

	// and recurse into next now that we repopulated response
	return ic.NextItemSync()
}

// gets all items
func (ic *ItemsCursor) AllSync() ([]Item, error) {
	var items []Item

	for ic.NextSync() {
		items = append(items, ic.GetItem())
	}

	if ic.Err() != nil {
		return nil, ic.Err()
	}

	return items, nil
}

func (ic *ItemsCursor) GetField(name string) interface{} {
	return ic.currentItem[name]
}

func (ic *ItemsCursor) GetFieldInt(name string) (int, error) {
	return ic.currentItem.GetFieldInt(name)
}

func (ic *ItemsCursor) GetFieldString(name string) (string, error) {
	return ic.currentItem.GetFieldString(name)
}

func (ic *ItemsCursor) GetFields() map[string]interface{} {
	return ic.currentItem
}

func (ic *ItemsCursor) GetItem() Item {
	return ic.currentItem
}

func (ic *ItemsCursor) Scattered() bool {

	// scattered flag applies only to the last item
	return ic.scattered && ic.itemIndex == len(ic.items)
}

func (ic *ItemsCursor) setResponse(response *Response) {
	ic.currentResponse = response

	getItemsOutput := response.Output.(*GetItemsOutput)

	ic.moreItemsExist = !getItemsOutput.Last
	ic.nextMarker = getItemsOutput.NextMarker
	ic.items = getItemsOutput.Items
	ic.itemIndex = 0
	ic.scattered = getItemsOutput.Scattered
}
