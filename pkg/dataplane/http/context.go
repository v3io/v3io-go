package v3iohttp

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/errors"

	"github.com/nuclio/logger"
	"github.com/valyala/fasthttp"
)

// TODO: Request should have a global pool
var requestID uint64 = 0

type context struct {
	logger           logger.Logger
	requestChan      chan *v3io.Request
	httpClient       *fasthttp.HostClient
	clusterEndpoints []string
	numWorkers       int
}

func NewContext(parentLogger logger.Logger, clusterEndpoints []string, numWorkers int) (v3io.Context, error) {
	var hosts []string

	// iterate over endpoints which contain scheme
	for _, clusterEndpoint := range clusterEndpoints {
		parsedClusterEndpoint, err := url.Parse(clusterEndpoint)
		if err != nil {
			return nil, err
		}

		hosts = append(hosts, parsedClusterEndpoint.Host)
	}

	newContext := &context{
		logger: parentLogger.GetChild("context.http"),
		httpClient: &fasthttp.HostClient{
			Addr: strings.Join(hosts, ","),
		},
		clusterEndpoints: clusterEndpoints,
		requestChan:      make(chan *v3io.Request, 1024),
		numWorkers:       numWorkers,
	}

	for workerIndex := 0; workerIndex < numWorkers; workerIndex++ {
		go newContext.workerEntry(workerIndex)
	}

	return newContext, nil
}

// create a new session
func (c *context) NewSessionSync(newSessionInput *v3io.NewSessionInput) (v3io.Session, error) {
	return newSession(c.logger,
		c,
		newSessionInput.Username,
		newSessionInput.Password,
		newSessionInput.AccessKey)
}

// GetItem
func (c *context) GetItem(getItemInput *v3io.GetItemInput,
	context interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	return c.sendRequestToWorker(getItemInput, context, responseChan)
}

// GetItemSync
func (c *context) GetItemSync(getItemInput *v3io.GetItemInput) (*v3io.Response, error) {

	// no need to marshal, just sprintf
	body := fmt.Sprintf(`{"AttributesToGet": "%s"}`, strings.Join(getItemInput.AttributeNames, ","))

	response, err := c.sendRequest(&getItemInput.DataPlaneInput,
		http.MethodPut,
		getItemInput.Path,
		getItemHeaders,
		[]byte(body),
		false)

	if err != nil {
		return nil, err
	}

	// ad hoc structure that contains response
	item := struct {
		Item map[string]map[string]string
	}{}

	c.logger.DebugWithCtx(getItemInput.Ctx, "Body", "body", string(response.Body()))

	// unmarshal the body
	err = json.Unmarshal(response.Body(), &item)
	if err != nil {
		return nil, err
	}

	// decode the response
	attributes, err := c.decodeTypedAttributes(item.Item)
	if err != nil {
		return nil, err
	}

	// attach the output to the response
	response.Output = &v3io.GetItemOutput{Item: attributes}

	return response, nil
}

// GetItems
func (c *context) GetItems(getItemsInput *v3io.GetItemsInput,
	context interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	return c.sendRequestToWorker(getItemsInput, context, responseChan)
}

// GetItemSync
func (c *context) GetItemsSync(getItemsInput *v3io.GetItemsInput) (*v3io.Response, error) {

	// create GetItem Body
	body := map[string]interface{}{
		"AttributesToGet": strings.Join(getItemsInput.AttributeNames, ","),
	}

	if getItemsInput.Filter != "" {
		body["FilterExpression"] = getItemsInput.Filter
	}

	if getItemsInput.Marker != "" {
		body["Marker"] = getItemsInput.Marker
	}

	if getItemsInput.ShardingKey != "" {
		body["ShardingKey"] = getItemsInput.ShardingKey
	}

	if getItemsInput.Limit != 0 {
		body["Limit"] = getItemsInput.Limit
	}

	if getItemsInput.TotalSegments != 0 {
		body["TotalSegment"] = getItemsInput.TotalSegments
		body["Segment"] = getItemsInput.Segment
	}

	marshalledBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	response, err := c.sendRequest(&getItemsInput.DataPlaneInput,
		"PUT",
		getItemsInput.Path,
		getItemsHeaders,
		[]byte(marshalledBody),
		false)

	if err != nil {
		return nil, err
	}

	c.logger.DebugWithCtx(getItemsInput.Ctx, "Body", "body", string(response.Body()))

	getItemsResponse := struct {
		Items            []map[string]map[string]string
		NextMarker       string
		LastItemIncluded string
	}{}

	// unmarshal the body into an ad hoc structure
	err = json.Unmarshal(response.Body(), &getItemsResponse)
	if err != nil {
		return nil, err
	}

	getItemsOutput := v3io.GetItemsOutput{
		NextMarker: getItemsResponse.NextMarker,
		Last:       getItemsResponse.LastItemIncluded == "TRUE",
	}

	// iterate through the items and decode them
	for _, typedItem := range getItemsResponse.Items {

		item, err := c.decodeTypedAttributes(typedItem)
		if err != nil {
			return nil, err
		}

		getItemsOutput.Items = append(getItemsOutput.Items, item)
	}

	// attach the output to the response
	response.Output = &getItemsOutput

	return response, nil
}

// PutItem
func (c *context) PutItem(putItemInput *v3io.PutItemInput,
	context interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	return c.sendRequestToWorker(putItemInput, context, responseChan)
}

// PutItemSync
func (c *context) PutItemSync(putItemInput *v3io.PutItemInput) error {

	// prepare the query path
	_, err := c.putItem(&putItemInput.DataPlaneInput,
		putItemInput.Path,
		putItemFunctionName,
		putItemInput.Attributes,
		putItemHeaders,
		nil)

	return err
}

// PutItems
func (c *context) PutItems(putItemsInput *v3io.PutItemsInput,
	context interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	return c.sendRequestToWorker(putItemsInput, context, responseChan)
}

// PutItemsSync
func (c *context) PutItemsSync(putItemsInput *v3io.PutItemsInput) (*v3io.Response, error) {

	response := c.allocateResponse()
	if response == nil {
		return nil, errors.New("Failed to allocate response")
	}

	putItemsOutput := v3io.PutItemsOutput{
		Success: true,
	}

	for itemKey, itemAttributes := range putItemsInput.Items {

		// try to post the item
		_, err := c.putItem(&putItemsInput.DataPlaneInput,
			putItemsInput.Path+"/"+itemKey,
			putItemFunctionName,
			itemAttributes,
			putItemHeaders,
			nil)

		// if there was an error, shove it to the list of errors
		if err != nil {

			// create the map to hold the errors since at least one exists
			if putItemsOutput.Errors == nil {
				putItemsOutput.Errors = map[string]error{}
			}

			putItemsOutput.Errors[itemKey] = err

			// clear success, since at least one error exists
			putItemsOutput.Success = false
		}
	}

	response.Output = &putItemsOutput

	return response, nil
}

// UpdateItem
func (c *context) UpdateItem(updateItemInput *v3io.UpdateItemInput,
	context interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	return c.sendRequestToWorker(updateItemInput, context, responseChan)
}

// UpdateItemSync
func (c *context) UpdateItemSync(updateItemInput *v3io.UpdateItemInput) error {
	var err error

	if updateItemInput.Attributes != nil {

		// specify update mode as part of body. "Items" will be injected
		body := map[string]interface{}{
			"UpdateMode": "CreateOrReplaceAttributes",
		}

		_, err = c.putItem(&updateItemInput.DataPlaneInput,
			updateItemInput.Path,
			putItemFunctionName,
			updateItemInput.Attributes,
			putItemHeaders,
			body)

	} else if updateItemInput.Expression != nil {

		_, err = c.updateItemWithExpression(&updateItemInput.DataPlaneInput,
			updateItemInput.Path,
			updateItemFunctionName,
			*updateItemInput.Expression,
			updateItemHeaders)
	}

	return err
}

// GetObject
func (c *context) GetObject(getObjectInput *v3io.GetObjectInput,
	context interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	return c.sendRequestToWorker(getObjectInput, context, responseChan)
}

// GetObjectSync
func (c *context) GetObjectSync(getObjectInput *v3io.GetObjectInput) (*v3io.Response, error) {
	return c.sendRequest(&getObjectInput.DataPlaneInput,
		http.MethodGet,
		getObjectInput.Path,
		nil,
		nil,
		false)
}

// PutObject
func (c *context) PutObject(putObjectInput *v3io.PutObjectInput,
	context interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	return c.sendRequestToWorker(putObjectInput, context, responseChan)
}

// PutObjectSync
func (c *context) PutObjectSync(putObjectInput *v3io.PutObjectInput) error {
	_, err := c.sendRequest(&putObjectInput.DataPlaneInput,
		http.MethodPut,
		putObjectInput.Path,
		nil,
		putObjectInput.Body,
		true)

	if err != nil {
		return err
	}

	return nil
}

// DeleteObject
func (c *context) DeleteObject(deleteObjectInput *v3io.DeleteObjectInput,
	context interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	return c.sendRequestToWorker(deleteObjectInput, context, responseChan)
}

// DeleteObjectSync
func (c *context) DeleteObjectSync(deleteObjectInput *v3io.DeleteObjectInput) error {
	_, err := c.sendRequest(&deleteObjectInput.DataPlaneInput,
		http.MethodDelete,
		deleteObjectInput.Path,
		nil,
		nil,
		true)

	if err != nil {
		return err
	}

	return nil
}

func (c *context) putItem(dataPlaneInput *v3io.DataPlaneInput,
	path string,
	functionName string,
	attributes map[string]interface{},
	headers map[string]string,
	body map[string]interface{}) (*v3io.Response, error) {

	// iterate over all attributes and encode them with their types
	typedAttributes, err := c.encodeTypedAttributes(attributes)
	if err != nil {
		return nil, err
	}

	// create an empty body if the user didn't pass anything
	if body == nil {
		body = map[string]interface{}{}
	}

	// set item in body (use what the user passed as a base)
	body["Item"] = typedAttributes

	jsonEncodedBodyContents, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return c.sendRequest(dataPlaneInput,
		http.MethodPut,
		path,
		headers,
		jsonEncodedBodyContents,
		false)
}

func (c *context) updateItemWithExpression(dataPlaneInput *v3io.DataPlaneInput,
	path string,
	functionName string,
	expression string,
	headers map[string]string) (*v3io.Response, error) {

	body := map[string]interface{}{
		"UpdateExpression": expression,
		"UpdateMode":       "CreateOrReplaceAttributes",
	}

	jsonEncodedBodyContents, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return c.sendRequest(dataPlaneInput,
		http.MethodPost,
		path,
		headers,
		jsonEncodedBodyContents,
		false)
}

func (c *context) sendRequest(dataPlaneInput *v3io.DataPlaneInput,
	method string,
	path string,
	headers map[string]string,
	body []byte,
	releaseResponse bool) (*v3io.Response, error) {

	var success bool
	var statusCode int
	var err error

	request := fasthttp.AcquireRequest()
	response := c.allocateResponse()

	// ggenerate URI
	uri := fmt.Sprintf("%s/%s/%s", c.clusterEndpoints[0], dataPlaneInput.ContainerName, path)

	// init request
	request.SetRequestURI(uri)
	request.Header.SetMethod(method)
	request.SetBody(body)

	// check if we need to an an authorization header
	if len(dataPlaneInput.AuthenticationToken) > 0 {

		// add authorization token
		request.Header.Set("Authorization", dataPlaneInput.AuthenticationToken)
	}

	if headers != nil {
		for headerName, headerValue := range headers {
			request.Header.Add(headerName, headerValue)
		}
	}

	c.logger.DebugWithCtx(dataPlaneInput.Ctx,
		"Tx",
		"uri", uri,
		"method", method,
		"body", string(request.Body()))

	if dataPlaneInput.Timeout <= 0 {
		err = c.httpClient.Do(request, response.HTTPResponse)
	} else {
		err = c.httpClient.DoTimeout(request, response.HTTPResponse, dataPlaneInput.Timeout)
	}

	if err != nil {
		goto cleanup
	}

	statusCode = response.HTTPResponse.StatusCode()

	c.logger.DebugWithCtx(dataPlaneInput.Ctx,
		"Rx",
		"statusCode", statusCode,
		"body", string(response.HTTPResponse.Body()))

	// did we get a 2xx response?
	success = statusCode >= 200 && statusCode < 300

	// make sure we got expected status
	if !success {
		err = v3ioerrors.NewErrorWithStatusCode(fmt.Errorf("Failed %s with status %d", method, statusCode), statusCode)
		goto cleanup
	}

cleanup:

	// we're done with the request - the response must be released by the user
	// unless there's an error
	fasthttp.ReleaseRequest(request)

	if err != nil {
		response.Release()
		return nil, err
	}

	// if the user doesn't need the response, release it
	if releaseResponse {
		response.Release()
		return nil, nil
	}

	return response, nil
}

func (c *context) allocateResponse() *v3io.Response {
	return &v3io.Response{
		HTTPResponse: fasthttp.AcquireResponse(),
	}
}

// {"age": 30, "name": "foo"} -> {"age": {"N": 30}, "name": {"S": "foo"}}
func (c *context) encodeTypedAttributes(attributes map[string]interface{}) (map[string]map[string]string, error) {
	typedAttributes := make(map[string]map[string]string)

	for attributeName, attributeValue := range attributes {
		typedAttributes[attributeName] = make(map[string]string)
		switch value := attributeValue.(type) {
		default:
			return nil, fmt.Errorf("Unexpected attribute type for %s: %T", attributeName, reflect.TypeOf(attributeValue))
		case int:
			typedAttributes[attributeName]["N"] = strconv.Itoa(value)
			// this is a tmp bypass to the fact Go maps Json numbers to float64
		case float64:
			typedAttributes[attributeName]["N"] = strconv.FormatFloat(value, 'E', -1, 64)
		case string:
			typedAttributes[attributeName]["S"] = value
		case []byte:
			typedAttributes[attributeName]["B"] = base64.StdEncoding.EncodeToString(value)
		}
	}

	return typedAttributes, nil
}

// {"age": {"N": 30}, "name": {"S": "foo"}} -> {"age": 30, "name": "foo"}
func (c *context) decodeTypedAttributes(typedAttributes map[string]map[string]string) (map[string]interface{}, error) {
	var err error
	attributes := map[string]interface{}{}

	for attributeName, typedAttributeValue := range typedAttributes {

		// try to parse as number
		if numberValue, ok := typedAttributeValue["N"]; ok {

			// try int
			if intValue, err := strconv.Atoi(numberValue); err != nil {

				// try float
				floatValue, err := strconv.ParseFloat(numberValue, 64)
				if err != nil {
					return nil, fmt.Errorf("Value for %s is not int or float: %s", attributeName, numberValue)
				}

				// save as float
				attributes[attributeName] = floatValue
			} else {
				attributes[attributeName] = intValue
			}
		} else if stringValue, ok := typedAttributeValue["S"]; ok {
			attributes[attributeName] = stringValue
		} else if byteSliceValue, ok := typedAttributeValue["B"]; ok {
			attributes[attributeName], err = base64.StdEncoding.DecodeString(byteSliceValue)
			if err != nil {
				return nil, err
			}
		}
	}

	return attributes, nil
}

func (c *context) sendRequestToWorker(input interface{},
	context interface{},
	responseChan chan *v3io.Response) (*v3io.Request, error) {
	id := atomic.AddUint64(&requestID, 1)

	// create a request/response (TODO: from pool)
	requestResponse := &v3io.RequestResponse{
		Request: v3io.Request{
			ID:                  id,
			Input:               input,
			Context:             context,
			ResponseChan:        responseChan,
			SendTimeNanoseconds: time.Now().UnixNano(),
		},
	}

	// point to container
	requestResponse.Request.RequestResponse = requestResponse

	// send the request to the request channel
	c.requestChan <- &requestResponse.Request

	return &requestResponse.Request, nil
}

func (c *context) workerEntry(workerIndex int) {
	for {
		var response *v3io.Response
		var err error

		// read a request
		request := <-c.requestChan

		// according to the input type
		switch typedInput := request.Input.(type) {
		case *v3io.PutObjectInput:
			err = c.PutObjectSync(typedInput)
		case *v3io.GetObjectInput:
			response, err = c.GetObjectSync(typedInput)
		case *v3io.DeleteObjectInput:
			err = c.DeleteObjectSync(typedInput)
		case *v3io.GetItemInput:
			response, err = c.GetItemSync(typedInput)
		case *v3io.GetItemsInput:
			response, err = c.GetItemsSync(typedInput)
		case *v3io.PutItemInput:
			err = c.PutItemSync(typedInput)
		case *v3io.PutItemsInput:
			response, err = c.PutItemsSync(typedInput)
		case *v3io.UpdateItemInput:
			err = c.UpdateItemSync(typedInput)
		default:
			c.logger.ErrorWith("Got unexpected request type", "request", request)
		}

		// TODO: have the sync interfaces somehow use the pre-allocated response
		if response != nil {
			request.RequestResponse.Response = *response
		}

		response = &request.RequestResponse.Response

		response.ID = request.ID
		response.Error = err
		response.RequestResponse = request.RequestResponse
		response.Context = request.Context

		// write to response channel
		request.ResponseChan <- &request.RequestResponse.Response
	}
}
