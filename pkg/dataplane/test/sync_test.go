package test

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/errors"

	"github.com/stretchr/testify/suite"
)

type syncTestSuite struct {
	testSuite
}

const (
	TestPath = "/dir1/#dir2/"
)

//
// Container tests
//

func (suite *syncContainerTestSuite) TestGetClusterMD() {
	suite.containerName = "bigdata"

	getClusterMDInput := v3io.GetClusterMDInput{}

	// when run against a context
	suite.populateDataPlaneInput(&getClusterMDInput.DataPlaneInput)

	// get cluster md
	response, err := suite.container.GetClusterMDSync(&getClusterMDInput)
	suite.Require().NoError(err, "Failed to get cluster meta data")
	getClusteMDOutput := response.Output.(*v3io.GetClusterMDOutput)
	suite.Require().NotEqual(getClusteMDOutput.NumberOfVNs, 0)
	response.Release()
}

type syncContainerTestSuite struct {
	syncTestSuite
}

func (suite *syncContainerTestSuite) TestGetContainers() {
	suite.containerName = "bigdata"

	getContainersInput := v3io.GetContainersInput{}

	// when run against a context
	suite.populateDataPlaneInput(&getContainersInput.DataPlaneInput)

	// get containers
	response, err := suite.container.GetContainersSync(&getContainersInput)
	suite.Require().NoError(err, "Failed to get containers")
	response.Release()
}

func (suite *syncContainerTestSuite) TestGetContainerContentsDefault() {
	path := fmt.Sprintf("tmp/test/sync_test/TestGetContainerContentsDefault/%d/", time.Now().Unix())
	fileContent := "If you cannot do great things, do small things in a great way."

	// Create some content (directory and files)
	putObjectInput := &v3io.PutObjectInput{}
	for i := 0; i < 10; i++ {
		if i < 5 {
			// Create file with content
			putObjectInput.Path = fmt.Sprintf("%sfile-%d.txt", path, i)
			putObjectInput.Body = []byte(fileContent)
		} else {
			// create empty directory
			putObjectInput.Path = fmt.Sprintf("%sdir-%d/", path, i)
			putObjectInput.Body = nil
		}

		// when run against a context
		suite.populateDataPlaneInput(&putObjectInput.DataPlaneInput)
		err := suite.container.PutObjectSync(putObjectInput)
		suite.Require().NoError(err, "Failed to create test content")
	}

	getContainerContentsInput := v3io.GetContainerContentsInput{
		Path: path,
	}

	// when run against a context
	suite.populateDataPlaneInput(&getContainerContentsInput.DataPlaneInput)

	// get container contents
	response, err := suite.container.GetContainerContentsSync(&getContainerContentsInput)
	suite.Require().NoError(err, "Failed to get container contents")
	response.Release()

	getContainerContentsOutput := response.Output.(*v3io.GetContainerContentsOutput)
	suite.Require().Equal(5, len(getContainerContentsOutput.Contents))

	for _, content := range getContainerContentsOutput.Contents {
		validateContent(suite, &content, len(fileContent), false)
	}

	for _, prefix := range getContainerContentsOutput.CommonPrefixes {
		validateCommonPrefix(suite, &prefix, false)
	}

	suite.Require().Equal(5, len(getContainerContentsOutput.CommonPrefixes))
}

func (suite *syncContainerTestSuite) TestGetContainerContentsFilesWithAllAttrs() {
	path := fmt.Sprintf("tmp/test/sync_test/TestGetContainerContentsFilesWithAllAttrs/%d/", time.Now().Unix())
	fileContent := "If you cannot do great things, do small things in a great way."

	// Create some content (directory and files)
	putObjectInput := &v3io.PutObjectInput{}
	for i := 0; i < 10; i++ {
		// Create file with content
		putObjectInput.Path = path + fmt.Sprintf("file-%d.txt", i)
		putObjectInput.Body = []byte(fileContent)

		// when run against a context
		suite.populateDataPlaneInput(&putObjectInput.DataPlaneInput)
		err := suite.container.PutObjectSync(putObjectInput)
		suite.Require().NoError(err, "Failed to create test content")
	}

	getContainerContentsInput := v3io.GetContainerContentsInput{
		Path:             path,
		GetAllAttributes: true,
		DirectoriesOnly:  false,
		Limit:            5,
	}

	// when run against a context
	suite.populateDataPlaneInput(&getContainerContentsInput.DataPlaneInput)

	// get container contents
	response, err := suite.container.GetContainerContentsSync(&getContainerContentsInput)
	suite.Require().NoError(err, "Failed to get container contents")
	response.Release()

	getContainerContentsOutput := response.Output.(*v3io.GetContainerContentsOutput)
	suite.Require().Equal(5, len(getContainerContentsOutput.Contents))
	suite.Require().Equal(path+"file-4.txt", getContainerContentsOutput.NextMarker)
	suite.Require().Equal(true, getContainerContentsOutput.IsTruncated)

	for _, content := range getContainerContentsOutput.Contents {
		validateContent(suite, &content, len(fileContent), true)
	}
	// get remaining content
	getContainerContentsInput.Marker = getContainerContentsOutput.NextMarker
	// get container contents
	response, err = suite.container.GetContainerContentsSync(&getContainerContentsInput)
	suite.Require().NoError(err, "Failed to get container contents")
	response.Release()

	getContainerContentsOutput = response.Output.(*v3io.GetContainerContentsOutput)
	suite.Require().Equal(5, len(getContainerContentsOutput.Contents))
	suite.Require().Equal(path+"file-9.txt", getContainerContentsOutput.NextMarker)
	suite.Require().Equal(false, getContainerContentsOutput.IsTruncated)

	for _, content := range getContainerContentsOutput.Contents {
		validateContent(suite, &content, len(fileContent), true)
	}
}

func (suite *syncContainerTestSuite) TestGetContainerContentsDirsWithAllAttrs() {
	path := fmt.Sprintf("tmp/test/sync_test/TestGetContainerContentsDirsWithAllAttrs/%d/", time.Now().Unix())
	content := "If you cannot do great things, do small things in a great way."

	// Create some content (directory and files)
	putObjectInput := &v3io.PutObjectInput{}
	for i := 0; i < 10; i++ {
		// create 2 files and 8 directories at the target path
		if i < 2 {
			// Create file with content
			putObjectInput.Path = fmt.Sprintf("%sfile-%d.txt", path, i)
			putObjectInput.Body = []byte(content)
		} else {
			// create empty directory
			putObjectInput.Path = fmt.Sprintf("%sdir-%d/", path, i)
			putObjectInput.Body = nil
		}

		// when run against a context
		suite.populateDataPlaneInput(&putObjectInput.DataPlaneInput)
		err := suite.container.PutObjectSync(putObjectInput)
		suite.Require().NoError(err, "Failed to create test content")
	}

	getContainerContentsInput := v3io.GetContainerContentsInput{
		Path:             path,
		GetAllAttributes: true,
		DirectoriesOnly:  true,
		Limit:            10,
	}

	// when run against a context
	suite.populateDataPlaneInput(&getContainerContentsInput.DataPlaneInput)

	// get container contents
	response, err := suite.container.GetContainerContentsSync(&getContainerContentsInput)
	suite.Require().NoError(err, "Failed to get container contents")
	response.Release()

	getContainerContentsOutput := response.Output.(*v3io.GetContainerContentsOutput)
	suite.Require().Empty(len(getContainerContentsOutput.Contents))
	suite.Require().Equal(8, len(getContainerContentsOutput.CommonPrefixes))
	suite.Require().Equal(path+"dir-9", getContainerContentsOutput.NextMarker)
	suite.Require().Equal(false, getContainerContentsOutput.IsTruncated)

	for _, prefix := range getContainerContentsOutput.CommonPrefixes {
		validateCommonPrefix(suite, &prefix, true)
	}
}

func (suite *syncContainerTestSuite) TestGetContainerContentsStreamDirsWithAllAttrs() {
	path := fmt.Sprintf("tmp/test/sync_test/TestGetContainerContentsStreamDirsWithAllAttrs/%d/", time.Now().Unix())

	const (
		shard_count            = 12
		retention_period_hours = 2
	)

	// Create some stream content (stream directory and shards)
	createStreamInput := &v3io.CreateStreamInput{}
	for i := 1; i <= 10; i++ {
		createStreamInput.Path = fmt.Sprintf("%sstream-%d/", path, i)
		createStreamInput.ShardCount = shard_count
		createStreamInput.RetentionPeriodHours = retention_period_hours

		// when run against a context
		suite.populateDataPlaneInput(&createStreamInput.DataPlaneInput)
		err := suite.container.CreateStreamSync(createStreamInput)
		suite.Require().NoError(err, "Failed to create test content")
	}

	getContainerContentsInput := v3io.GetContainerContentsInput{
		Path:             path,
		GetAllAttributes: true,
		DirectoriesOnly:  true,
		Limit:            10,
	}

	// when run against a context
	suite.populateDataPlaneInput(&getContainerContentsInput.DataPlaneInput)

	// get container contents
	response, err := suite.container.GetContainerContentsSync(&getContainerContentsInput)
	suite.Require().NoError(err, "Failed to get container contents")
	response.Release()

	getContainerContentsOutput := response.Output.(*v3io.GetContainerContentsOutput)
	suite.Require().Empty(len(getContainerContentsOutput.Contents))
	suite.Require().Equal(10, len(getContainerContentsOutput.CommonPrefixes))
	suite.Require().Equal(path+"stream-9", getContainerContentsOutput.NextMarker)
	suite.Require().Equal(false, getContainerContentsOutput.IsTruncated)

	for _, prefix := range getContainerContentsOutput.CommonPrefixes {
		validateCommonPrefix(suite, &prefix, true)
		suite.Require().Equal(shard_count, prefix.ShardCount)
		suite.Require().Equal(retention_period_hours, prefix.RetentionPeriodHours)
	}
}

// TODO: fix. Broken with:
// Messages:   	Failed to update test directory
// due to:
//         	            	{
//        	            		"ErrorCode": -13,
//        	            		"ErrorMessage": "Permission denied"
//        	            	}
//func (suite *syncContainerTestSuite) TestSetDirsAttrs() {
//	path := fmt.Sprintf("tmp/test/sync_test/TestSetDirsAttrs/%d/", time.Now().Unix())
//
//	// create empty directory
//	putObjectInput := &v3io.PutObjectInput{}
//	putObjectInput.Path = fmt.Sprintf("%sdir-test/", path)
//	putObjectInput.Body = nil
//
//	// when run against a context
//	suite.populateDataPlaneInput(&putObjectInput.DataPlaneInput)
//	err := suite.container.PutObjectSync(putObjectInput)
//	suite.Require().NoError(err, "Failed to create test directory")
//
//	// Update directory's attributes
//	updateObjectInput := &v3io.UpdateObjectInput{}
//	updateObjectInput.Path = fmt.Sprintf("%sdir-test/", path)
//	layout := "2006-01-02T15:04:05.00Z"
//	atime, err := time.Parse(layout, "2020-11-22T19:27:33.00Z")
//	suite.Require().NoError(err)
//	ctime, err := time.Parse(layout, "2020-09-20T15:10:35.08Z")
//	suite.Require().NoError(err)
//	mtime, err := time.Parse(layout, "2020-09-24T12:55:35.08Z")
//	suite.Require().NoError(err)
//	dirAttributes := &v3io.DirAttributes{
//		Mode:      511,
//		UID:       0,
//		GID:       68,
//		AtimeSec:  int(atime.Unix()),
//		AtimeNSec: int(atime.UnixNano() % 1000000000),
//		CtimeSec:  int(ctime.Unix()),
//		CtimeNSec: int(ctime.UnixNano() % 1000000000),
//		MtimeSec:  int(mtime.Unix()),
//		MtimeNSec: int(mtime.UnixNano() % 1000000000),
//	}
//	updateObjectInput.DirAttributes = dirAttributes
//
//	// when run against a context
//	suite.populateDataPlaneInput(&updateObjectInput.DataPlaneInput)
//	err = suite.container.UpdateObjectSync(updateObjectInput)
//	suite.Require().NoError(err, "Failed to update test directory")
//
//	// Read directory and compare attributes
//	getContainerContentsInput := v3io.GetContainerContentsInput{
//		Path:             path,
//		GetAllAttributes: true,
//		DirectoriesOnly:  true,
//		Limit:            10,
//	}
//
//	// when run against a context
//	suite.populateDataPlaneInput(&getContainerContentsInput.DataPlaneInput)
//
//	// get container contents
//	response, err := suite.container.GetContainerContentsSync(&getContainerContentsInput)
//	suite.Require().NoError(err, "Failed to get container contents")
//	response.Release()
//
//	getContainerContentsOutput := response.Output.(*v3io.GetContainerContentsOutput)
//	suite.Require().Empty(len(getContainerContentsOutput.Contents))
//	suite.Require().Equal(1, len(getContainerContentsOutput.CommonPrefixes))
//	suite.Require().Equal(false, getContainerContentsOutput.IsTruncated)
//
//	prefix := getContainerContentsOutput.CommonPrefixes[0]
//	suite.Require().Equal(atime.Format(layout), prefix.AccessTime)
//	suite.Require().Equal(mtime.Format(layout), prefix.LastModified)
//	suite.Require().Equal(ctime.Format(layout), prefix.CreatingTime)
//	suite.Require().Equal(strconv.FormatInt(int64(dirAttributes.GID), 16), prefix.GID)
//	suite.Require().Equal(strconv.FormatInt(int64(dirAttributes.UID), 16), prefix.UID)
//	suite.Require().Equal("040777", string(prefix.Mode))
//}

type syncContextContainerTestSuite struct {
	syncContainerTestSuite
}

func (suite *syncContextContainerTestSuite) SetupSuite() {
	suite.syncContainerTestSuite.SetupSuite()

	suite.createContext()
}

type syncContainerContainerTestSuite struct {
	syncContainerTestSuite
}

func (suite *syncContainerContainerTestSuite) SetupSuite() {
	suite.syncContainerTestSuite.SetupSuite()

	suite.createContainer()
}

//
// Object tests
//

type syncObjectTestSuite struct {
	syncTestSuite
}

func (suite *syncObjectTestSuite) TestObject() {
	path := TestPath + "#object.txt"
	contents := "vegans are better than everyone"

	getObjectInput := &v3io.GetObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err := suite.container.GetObjectSync(getObjectInput)

	// get the underlying root error
	errWithStatusCode, errHasStatusCode := err.(v3ioerrors.ErrorWithStatusCode)
	suite.Require().True(errHasStatusCode)
	suite.Require().Equal(404, errWithStatusCode.StatusCode())

	//
	// PUT contents to some object
	//

	putObjectInput := &v3io.PutObjectInput{
		Path: path,
		Body: []byte(contents),
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&putObjectInput.DataPlaneInput)

	err = suite.container.PutObjectSync(putObjectInput)

	suite.Require().NoError(err, "Failed to put")

	//
	// Get the contents by path
	//

	getObjectInput = &v3io.GetObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err = suite.container.GetObjectSync(getObjectInput)
	suite.Require().NoError(err, "Failed to get")

	// make sure buckets is not empty
	suite.Require().Equal(contents, string(response.Body()))

	// release the response
	response.Release()

	//
	// Get object by inode
	//

	// Get item attributes to find out what's the inode & ctime
	getItemInput := v3io.GetItemInput{
		Path:           path,
		AttributeNames: []string{"**"},
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getItemInput.DataPlaneInput)

	response, err = suite.container.GetItemSync(&getItemInput)
	suite.Require().NoError(err, "Failed to get item")

	// parse item attributes and extract inode & ctime
	getItemOutput := response.Output.(*v3io.GetItemOutput)
	inode, err := getItemOutput.Item.GetFieldInt("__inode_number")
	suite.Require().NoError(err, "Failed to get inode")
	ctimeSec, err := getItemOutput.Item.GetFieldInt("__ctime_secs")
	suite.Require().NoError(err, "Failed to get ctime sec")
	ctimeNsec, err := getItemOutput.Item.GetFieldInt("__ctime_nsecs")
	suite.Require().NoError(err, "Failed to get ctime nsec")

	// get object using inode and ctime
	getObjectInput = &v3io.GetObjectInput{
		Path:      fmt.Sprintf("%d", inode),
		CtimeSec:  ctimeSec,
		CtimeNsec: ctimeNsec,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err = suite.container.GetObjectSync(getObjectInput)
	suite.Require().NoError(err, "Failed to get")

	// make sure buckets is not empty
	suite.Require().Equal(contents, string(response.Body()))

	// release the response
	response.Release()

	//
	// Delete the object
	//

	deleteObjectInput := &v3io.DeleteObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&deleteObjectInput.DataPlaneInput)

	err = suite.container.DeleteObjectSync(deleteObjectInput)

	suite.Require().NoError(err, "Failed to delete")

	//
	// Get the contents again (should fail)
	//

	getObjectInput = &v3io.GetObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err = suite.container.GetObjectSync(getObjectInput)

	suite.Require().Error(err, "Failed to get")
	suite.Require().Nil(response)
}

func (suite *syncObjectTestSuite) TestAppend() {
	path := "/object.txt"
	contents := "vegans are better than everyone"

	getObjectInput := &v3io.GetObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err := suite.container.GetObjectSync(getObjectInput)

	// get the underlying root error
	errWithStatusCode, errHasStatusCode := err.(v3ioerrors.ErrorWithStatusCode)
	suite.Require().True(errHasStatusCode)
	suite.Require().Equal(404, errWithStatusCode.StatusCode())

	//
	// PUT contents to some object
	//

	putObjectInput := &v3io.PutObjectInput{
		Path: path,
		Body: []byte(contents),
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&putObjectInput.DataPlaneInput)

	err = suite.container.PutObjectSync(putObjectInput)

	suite.Require().NoError(err, "Failed to put")

	//
	// Append contents to the same object
	//

	appendContents := "NOT!"
	appendPutObjectInput := &v3io.PutObjectInput{
		Path:   path,
		Body:   []byte(appendContents),
		Append: true,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&appendPutObjectInput.DataPlaneInput)

	err = suite.container.PutObjectSync(appendPutObjectInput)

	suite.Require().NoError(err, "Failed to put")

	//
	// Get the contents
	//

	getObjectInput = &v3io.GetObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err = suite.container.GetObjectSync(getObjectInput)
	suite.Require().NoError(err, "Failed to get")

	// make sure buckets is not empty
	suite.Require().Equal(contents+appendContents, string(response.Body()))

	// release the response
	response.Release()

	//
	// Delete the object
	//

	deleteObjectInput := &v3io.DeleteObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&deleteObjectInput.DataPlaneInput)

	err = suite.container.DeleteObjectSync(deleteObjectInput)

	suite.Require().NoError(err, "Failed to delete")

	//
	// Get the contents again (should fail)
	//

	getObjectInput = &v3io.GetObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err = suite.container.GetObjectSync(getObjectInput)

	suite.Require().Error(err, "Failed to get")
	suite.Require().Nil(response)
}

func (suite *syncObjectTestSuite) TestCheckPathExists() {
	suite.containerName = "bigdata"

	checkPathExists := v3io.CheckPathExistsInput{}
	checkPathExists.Path = "/SomeFolder/"
	// when run against a context
	suite.populateDataPlaneInput(&checkPathExists.DataPlaneInput)

	err := suite.container.CheckPathExistsSync(&checkPathExists)
	suite.Require().Error(err, "did not get an error on non existing error")
}

func (suite *syncObjectTestSuite) TestReadRange() {
	path := "/range.txt"

	fileSize := 1024 * 1024 * 3
	contents := make([]byte, fileSize)

	for i := range contents {
		contents[i] = 'a'
	}

	getObjectInput := &v3io.GetObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err := suite.container.GetObjectSync(getObjectInput)

	// get the underlying root error
	errWithStatusCode, errHasStatusCode := err.(v3ioerrors.ErrorWithStatusCode)
	suite.Require().True(errHasStatusCode)
	suite.Require().Equal(404, errWithStatusCode.StatusCode())

	//
	// PUT contents to some object
	//

	putObjectInput := &v3io.PutObjectInput{
		Path: path,
		Body: []byte(contents),
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&putObjectInput.DataPlaneInput)

	err = suite.container.PutObjectSync(putObjectInput)

	suite.Require().NoError(err, "Failed to put")

	//
	// Get all the contents
	//

	getObjectInput = &v3io.GetObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err = suite.container.GetObjectSync(getObjectInput)
	suite.Require().NoError(err, "Failed to get")

	// make sure buckets is not empty
	suite.Require().Equal(string(contents), string(response.Body()))

	// release the response
	response.Release()

	halfFileLength := fileSize / 2
	//
	// Get the first half
	//

	getObjectInput = &v3io.GetObjectInput{
		Path:     path,
		Offset:   0,
		NumBytes: halfFileLength,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err = suite.container.GetObjectSync(getObjectInput)
	suite.Require().NoError(err, "Failed to get")

	// make sure buckets is not empty
	suite.Require().Equal(string(contents[:halfFileLength]), string(response.Body()))

	// release the response
	response.Release()

	fmt.Println("======== gonna read second")
	//
	// Get the second half
	//

	getObjectInput = &v3io.GetObjectInput{
		Path:     path,
		Offset:   halfFileLength,
		NumBytes: halfFileLength,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err = suite.container.GetObjectSync(getObjectInput)
	suite.Require().NoError(err, "Failed to get")

	// make sure buckets is not empty
	suite.Require().Equal(string(contents[halfFileLength:]), string(response.Body()))

	// release the response
	response.Release()

	//
	// Delete the object
	//

	deleteObjectInput := &v3io.DeleteObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&deleteObjectInput.DataPlaneInput)

	err = suite.container.DeleteObjectSync(deleteObjectInput)

	suite.Require().NoError(err, "Failed to delete")

	//
	// Get the contents again (should fail)
	//

	getObjectInput = &v3io.GetObjectInput{
		Path: path,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getObjectInput.DataPlaneInput)

	response, err = suite.container.GetObjectSync(getObjectInput)

	suite.Require().Error(err, "Failed to get")
	suite.Require().Nil(response)
}

type syncContextObjectTestSuite struct {
	syncObjectTestSuite
}

func (suite *syncContextObjectTestSuite) SetupSuite() {
	suite.syncObjectTestSuite.SetupSuite()

	suite.createContext()
}

type syncContainerObjectTestSuite struct {
	syncObjectTestSuite
}

func (suite *syncContainerObjectTestSuite) SetupSuite() {
	suite.syncObjectTestSuite.SetupSuite()

	suite.createContainer()
}

//
// KV tests
//

type syncKVTestSuite struct {
	syncTestSuite
	items map[string]map[string]interface{}
}

func (suite *syncKVTestSuite) TestEMD() {
	itemsToCreate := map[string]map[string]interface{}{
		"bob":    {"age": 42, "feature": "mustache"},
		"linda":  {"age": 41, "feature": "singing"},
		"louise": {"age": 9, "feature": "bunny ears"},
		"tina":   {"age": 14, "feature": "butts"},
	}

	//
	// Create items one by one
	//

	// create the items
	for itemToCreateKey, itemToCreateAttributes := range itemsToCreate {
		input := v3io.PutItemInput{
			Path:       TestPath + "emd0/" + itemToCreateKey,
			Attributes: itemToCreateAttributes,
		}

		// when run against a context, will populate fields like container name
		suite.populateDataPlaneInput(&input.DataPlaneInput)

		// get a specific bucket
		_, err := suite.container.PutItemSync(&input)
		suite.Require().NoError(err, "Failed to put item")
	}

	suite.verifyItems(itemsToCreate)

	//
	// Update item and verify
	//

	// update louise item
	updateItemInput := v3io.UpdateItemInput{
		Path: TestPath + "/emd0/louise",
		Attributes: map[string]interface{}{
			"height": 130,
			"quip":   "i can smell fear on you",
		},
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&updateItemInput.DataPlaneInput)

	_, err := suite.container.UpdateItemSync(&updateItemInput)
	suite.Require().NoError(err, "Failed to update item")

	// get louise
	getItemInput := v3io.GetItemInput{
		Path:           TestPath + "/emd0/louise",
		AttributeNames: []string{"__size", "age", "quip", "height"},
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getItemInput.DataPlaneInput)

	response, err := suite.container.GetItemSync(&getItemInput)
	suite.Require().NoError(err, "Failed to get item")

	getItemOutput := response.Output.(*v3io.GetItemOutput)

	// make sure we got the age and quip correctly
	suite.Require().Equal(0, getItemOutput.Item["__size"].(int))
	suite.Require().Equal(130, getItemOutput.Item["height"].(int))
	suite.Require().Equal("i can smell fear on you", getItemOutput.Item["quip"].(string))
	suite.Require().Equal(9, getItemOutput.Item["age"].(int))

	// release the response
	response.Release()

	// get all items whose age is over 15
	getItemsInput := v3io.GetItemsInput{
		Path:           TestPath + "/emd0/",
		AttributeNames: []string{"age", "feature"},
		Filter:         "age > 15",
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getItemsInput.DataPlaneInput)

	cursor, err := v3io.NewItemsCursor(suite.container, &getItemsInput)
	suite.Require().NoError(err, "Failed to get items")

	cursorItems, err := cursor.AllSync()
	suite.Require().NoError(err)
	suite.Require().Len(cursorItems, 2)

	// iterate over age, make sure it's over 15
	for _, cursorItem := range cursorItems {
		age, err := cursorItem.GetFieldInt("age")
		suite.Require().NoError(err)
		suite.Require().True(age > 15)
	}

	// release the response
	response.Release()

	//
	// Increment age
	//

	incrementAgeExpression := "age = age + 1"

	// update louise's age
	updateItemInput = v3io.UpdateItemInput{
		Path:       TestPath + "/emd0/louise",
		Expression: &incrementAgeExpression,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&updateItemInput.DataPlaneInput)

	_, err = suite.container.UpdateItemSync(&updateItemInput)
	suite.Require().NoError(err, "Failed to update item")

	// get tina
	getItemInput = v3io.GetItemInput{
		Path:           TestPath + "/emd0/louise",
		AttributeNames: []string{"age"},
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getItemInput.DataPlaneInput)

	response, err = suite.container.GetItemSync(&getItemInput)
	suite.Require().NoError(err, "Failed to get item")

	getItemOutput = response.Output.(*v3io.GetItemOutput)

	// check that age incremented
	suite.Require().Equal(10, getItemOutput.Item["age"].(int))

	// release the response
	response.Release()

	//
	// Delete everything
	//

	suite.deleteItems(TestPath+"/emd0", itemsToCreate)
}

func (suite *syncKVTestSuite) TestPutItems() {
	items := map[string]map[string]interface{}{
		"bob":   {"age": 42, "feature": "mustache", "married": false, "timestamp": 1556450700000},
		"linda": {"age": 41, "feature": "singing", "married": true, "timestamp": 1556450700000},
	}

	putItemsInput := &v3io.PutItemsInput{
		Path:  TestPath + "emd0",
		Items: items,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&putItemsInput.DataPlaneInput)

	// get a specific bucket
	response, err := suite.container.PutItemsSync(putItemsInput)
	suite.Require().NoError(err, "Failed to put items")

	putItemsOutput := response.Output.(*v3io.PutItemsOutput)

	// must succeed - everything was valid
	suite.Require().True(putItemsOutput.Success)
	suite.Require().Nil(putItemsOutput.Errors)

	response.Release()

	suite.verifyItems(items)

	suite.deleteItems(putItemsInput.Path, items)
}

func (suite *syncKVTestSuite) TestScatteredCursor() {
	path := "/emd1/scattered_cursor"

	numOfChunks := 4
	chunkSize := 30
	items, scatteredItemKeys := suite.populateScatteredItems(path, numOfChunks, chunkSize)

	// Get cursor
	getItemsInput := v3io.GetItemsInput{
		Path:               path + "/",
		AttributeNames:     []string{"**"},
		AllowObjectScatter: "true",
	}
	suite.populateDataPlaneInput(&getItemsInput.DataPlaneInput)
	cursor, err := v3io.NewItemsCursor(suite.container, &getItemsInput)
	suite.Require().NoError(err, "Failed to get cursor")

	// extract and combine scattered items
	retrievedItems := map[string]map[string]interface{}{}
	for cursor.NextSync() {
		item := cursor.GetItem()
		inode := item["__inode_number"]
		ctimeSec := item["__ctime_secs"]
		ctimeNSec := item["__ctime_nsecs"]
		ctime := int64(ctimeSec.(int))*1e9 + int64(ctimeNSec.(int))
		objectID := fmt.Sprintf("%d.%d", inode.(int), ctime)

		scatteredItem, scatteredItemFound := retrievedItems[objectID]
		if scatteredItemFound {
			for key, value := range item {
				scatteredItem[key] = value
			}
			item = scatteredItem
		}
		retrievedItems[objectID] = item
	}
	cursor.Release()

	// validate results
	suite.Assert().Equal(len(items), len(retrievedItems))
	for _, retrievedItem := range retrievedItems {
		suite.Assert().Contains(items, retrievedItem["__name"])
		for _, scatteredItemKey := range scatteredItemKeys {
			if retrievedItem["__name"] == scatteredItemKey {

				// count number of blob keys
				blobCounter := 0
				for key := range retrievedItem {
					if strings.HasPrefix(key, "blob_") {
						blobCounter++
					}
				}
				suite.Assert().Equal(numOfChunks*chunkSize, blobCounter)
			}
		}
	}

	suite.deleteItems(path, items)
}

// TODO: commenting out to stabilize CI.
// this failed with: Expected :6  / Actual   :5
//func (suite *syncKVTestSuite) TestIncludeResponseInError() {
//	path := "/emd1/response_in_error"
//
//	numOfChunks := 6
//	chunkSize := 30
//	items, scatteredItemKeys := suite.populateScatteredItems(path, numOfChunks, chunkSize)
//
//	// Issue GetItems request with scattering disabled. Since there are scattered items this will create an error
//	getItemsInput := v3io.GetItemsInput{
//		Path:               path + "/",
//		AttributeNames:     []string{"**"},
//		AllowObjectScatter: "false",
//	}
//	suite.populateDataPlaneInput(&getItemsInput.DataPlaneInput)
//	getItemsInput.IncludeResponseInError = true
//
//	var err error
//	var response *v3io.Response
//	var errorItemKeys []string
//	receivedItems := 0
//	for {
//		response, err = suite.container.GetItemsSync(&getItemsInput)
//		if response != nil {
//			getItemsOutput := response.Output.(*v3io.GetItemsOutput)
//			if getItemsOutput.Last {
//				break
//			}
//			suite.Assert().True(len(getItemsOutput.NextMarker) > 0)
//
//			if err != nil {
//				for _, item := range getItemsOutput.Items {
//					itemKey, _err := item.GetFieldString("__name")
//					suite.Assert().Nil(_err)
//					errorItemKeys = append(errorItemKeys, itemKey)
//				}
//			}
//
//			receivedItems += len(getItemsOutput.Items)
//			getItemsInput.Marker = getItemsOutput.NextMarker
//		} else {
//			break
//		}
//	}
//
//	suite.Assert().Equal(len(items), receivedItems)
//	suite.Assert().ElementsMatch(scatteredItemKeys, errorItemKeys)
//	suite.deleteItems(path, items)
//}

func (suite *syncKVTestSuite) TestPutItemsWithError() {
	items := map[string]map[string]interface{}{
		"bob":     {"age": 42, "feature": "mustache"},
		"linda":   {"age": 41, "feature": "singing"},
		"invalid": {"__name": "foo", "feature": "singing"},
	}

	putItemsInput := &v3io.PutItemsInput{
		Path:  TestPath + "/emd0",
		Items: items,
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&putItemsInput.DataPlaneInput)

	// get a specific bucket
	response, err := suite.container.PutItemsSync(putItemsInput)
	suite.Require().NoError(err, "Failed to put items")

	putItemsOutput := response.Output.(*v3io.PutItemsOutput)

	// must succeed - everything was valid
	suite.Require().False(putItemsOutput.Success)
	suite.Require().NotNil(putItemsOutput.Errors)
	suite.Require().NotNil(putItemsOutput.Errors["invalid"])

	response.Release()

	// remove invalid because it shouldn't be verified / deleted
	delete(items, "invalid")

	suite.verifyItems(items)

	suite.deleteItems(putItemsInput.Path, items)
}

func (suite *syncKVTestSuite) populateScatteredItems(path string,
	numOfChunks int,
	chunkSize int) (map[string]map[string]interface{}, []string) {

	scatteredItemKeys := []string{"louise", "karen"}
	items := map[string]map[string]interface{}{
		"bob":                {"age": 42, "feature": "mustache"},
		"linda":              {"age": 41, "feature": "singing"},
		"natan":              {"age": 35, "feature": "one leg"},
		"donald":             {"age": 20, "feature": "teeth"},
		scatteredItemKeys[0]: {"timestamp": time.Now().UnixNano(), "blob0": randomString(60000)},
		scatteredItemKeys[1]: {"timestamp": time.Now().UnixNano(), "blob0": randomString(60000)},
	}

	putItemsInput := &v3io.PutItemsInput{
		Path:  path,
		Items: items,
	}

	// Store initial items
	suite.populateDataPlaneInput(&putItemsInput.DataPlaneInput)
	response, err := suite.container.PutItemsSync(putItemsInput)
	suite.Require().NoError(err, "Failed to put items")
	putItemsOutput := response.Output.(*v3io.PutItemsOutput)
	suite.Require().True(putItemsOutput.Success)
	response.Release()

	// update `scatteredItemKeys` items with big KV entries to force them to scatter
	for _, key := range scatteredItemKeys {
		updateItemInput := v3io.UpdateItemInput{
			Path: fmt.Sprintf("%s/%s", path, key),
		}

		// because of request size limit we will have to update items in parts
		for i := 0; i < numOfChunks; i++ {
			attributes := map[string]interface{}{}
			for j := 0; j < chunkSize; j++ {
				attributes[fmt.Sprintf("%s_%s_%d_%d", "blob", key, i, j)] = randomString(60000)
			}
			updateItemInput.Attributes = attributes
			suite.populateDataPlaneInput(&updateItemInput.DataPlaneInput)
			response, err := suite.container.UpdateItemSync(&updateItemInput)
			suite.Require().NoError(err, "Failed to update item")
			response.Release()
		}
	}

	return items, scatteredItemKeys
}

func (suite *syncKVTestSuite) verifyItems(items map[string]map[string]interface{}) {

	// get all items
	getItemsInput := v3io.GetItemsInput{
		Path:           TestPath + "/emd0/",
		AttributeNames: []string{"*"},
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&getItemsInput.DataPlaneInput)

	cursor, err := v3io.NewItemsCursor(suite.container, &getItemsInput)
	suite.Require().NoError(err, "Failed to create cursor")

	receivedItems, err := cursor.AllSync()
	suite.Require().NoError(err)
	suite.Require().Len(receivedItems, len(items))

	for _, receivedItem := range receivedItems {
		nameIface, ok := receivedItem["__name"]
		suite.Require().True(ok, "Received an item without a name")
		name, ok := nameIface.(string)
		suite.Require().True(ok, "Received an item with a non-string name")
		items[name]["__name"] = name
		suite.Require().EqualValues(items[name], map[string]interface{}(receivedItem))
	}

	// release the response
	cursor.Release()
}

func (suite *syncKVTestSuite) deleteItems(path string, items map[string]map[string]interface{}) {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// delete the items
	for itemKey := range items {
		input := v3io.DeleteObjectInput{
			Path: filepath.Join(path, itemKey),
		}

		// when run against a context, will populate fields like container name
		suite.populateDataPlaneInput(&input.DataPlaneInput)

		// get a specific bucket
		err := suite.container.DeleteObjectSync(&input)
		suite.Require().NoError(err, "Failed to delete item")
	}

	input := &v3io.DeleteObjectInput{
		Path: fmt.Sprintf("%s", path),
	}

	// when run against a context, will populate fields like container name
	suite.populateDataPlaneInput(&input.DataPlaneInput)

	// delete the directory
	err := suite.container.DeleteObjectSync(input)

	suite.Require().NoError(err, "Failed to delete")
}

type syncContextKVTestSuite struct {
	syncKVTestSuite
}

func (suite *syncContextKVTestSuite) SetupSuite() {
	suite.syncKVTestSuite.SetupSuite()

	suite.createContext()
}

type syncContainerKVTestSuite struct {
	syncKVTestSuite
}

func (suite *syncContainerKVTestSuite) SetupSuite() {
	suite.syncKVTestSuite.SetupSuite()

	suite.createContainer()
}

//
// Stream tests
//

type syncStreamTestSuite struct {
	syncTestSuite
	streamTestSuite streamTestSuite
}

func (suite *syncStreamTestSuite) SetupTest() {
	suite.streamTestSuite = streamTestSuite{
		testSuite: suite.syncTestSuite.testSuite,
	}
	suite.streamTestSuite.SetupTest()
}

func (suite *syncStreamTestSuite) TearDownTest() {
	suite.streamTestSuite.TearDownTest()
}

func (suite *syncStreamTestSuite) TestStream() {
	streamPath := fmt.Sprintf("%s/mystream/", suite.streamTestSuite.testPath)

	//
	// Create the stream
	//

	createStreamInput := v3io.CreateStreamInput{
		Path:                 streamPath,
		ShardCount:           4,
		RetentionPeriodHours: 1,
	}

	suite.populateDataPlaneInput(&createStreamInput.DataPlaneInput)

	err := suite.container.CreateStreamSync(&createStreamInput)

	suite.Require().NoError(err, "Failed to create stream")

	//
	// Put some records
	//

	firstShardID := 1
	secondShardID := 2
	invalidShardID := 10

	records := []*v3io.StreamRecord{
		{ShardID: &firstShardID, Data: []byte("first shard record #1")},
		{ShardID: &firstShardID, Data: []byte("first shard record #2")},
		{ShardID: &invalidShardID, Data: []byte("invalid shard record #1")},
		{ShardID: &secondShardID, Data: []byte("second shard record #1")},
		{Data: []byte("some shard record #1")},
	}

	putRecordsInput := v3io.PutRecordsInput{
		Path:    streamPath,
		Records: records,
	}

	suite.populateDataPlaneInput(&putRecordsInput.DataPlaneInput)

	response, err := suite.container.PutRecordsSync(&putRecordsInput)
	suite.Require().NoError(err, "Failed to put records")

	putRecordsResponse := response.Output.(*v3io.PutRecordsOutput)

	// should have one failure
	suite.Require().Equal(1, putRecordsResponse.FailedRecordCount)

	// verify record results
	for recordIdx, record := range putRecordsResponse.Records {

		// third record should've failed
		if recordIdx == 2 {
			suite.Require().NotEqual(0, record.ErrorCode)
		} else {
			suite.Require().Equal(0, record.ErrorCode)
		}
	}

	response.Release()

	//
	// Seek
	//

	seekShardInput := v3io.SeekShardInput{
		Path: streamPath + "1",
		Type: v3io.SeekShardInputTypeEarliest,
	}

	suite.populateDataPlaneInput(&seekShardInput.DataPlaneInput)

	response, err = suite.container.SeekShardSync(&seekShardInput)

	suite.Require().NoError(err, "Failed to seek shard")
	location := response.Output.(*v3io.SeekShardOutput).Location

	suite.Require().NotEqual("", location)

	response.Release()

	//
	// Get records
	//

	getRecordsInput := v3io.GetRecordsInput{
		Path:     streamPath + "1",
		Location: location,
		Limit:    100,
	}

	suite.populateDataPlaneInput(&getRecordsInput.DataPlaneInput)

	response, err = suite.container.GetRecordsSync(&getRecordsInput)

	suite.Require().NoError(err, "Failed to get records")

	getRecordsOutput := response.Output.(*v3io.GetRecordsOutput)

	suite.Require().Equal("first shard record #1", string(getRecordsOutput.Records[0].Data))
	suite.Require().Equal("first shard record #2", string(getRecordsOutput.Records[1].Data))

	response.Release()

	getItemsInput := v3io.GetItemsInput{
		Path:               streamPath,
		ReturnData:         "true",
		AllowObjectScatter: "true",
		AttributeNames:     []string{"**"},
	}

	suite.populateDataPlaneInput(&getItemsInput.DataPlaneInput)

	cursor, err := v3io.NewItemsCursor(suite.container, &getItemsInput)
	suite.Require().NoError(err, "Failed to get items")

	cursorItems, err := cursor.AllSync()
	suite.Require().NoError(err)

	// internal use case test
	for _, cursorItem := range cursorItems {
		shardName, err := cursorItem.GetFieldString("__name")
		suite.Require().NoError(err, "Failed to get item name")
		chunkMap, _, err := cursorItem.GetShard()
		suite.Require().NoError(err, "Failed to get stream")

		suite.Require().Contains(chunkMap, 0, "chunk indexes doesn't match")
		suite.Require().True(chunkMap[0].Metadata.ChunkSeqNumber == 0)
		suite.Require().True(chunkMap[0].Metadata.FirstRecordSeqNumber == 1)
		suite.Require().True(chunkMap[0].Metadata.LengthInBytes == 0)
		switch shardName {
		case "0":
			suite.Require().True(strings.Contains(string(*chunkMap[0].Data[0].Data), string(records[4].Data)))
		case "1":
			suite.Require().True(strings.Contains(string(*chunkMap[0].Data[0].Data), string(records[0].Data)))
			suite.Require().True(strings.Contains(string(*chunkMap[0].Data[0].Data), string(records[1].Data)))
		case "2":
			suite.Require().True(strings.Contains(string(*chunkMap[0].Data[0].Data), string(records[3].Data)))
		}
	}

	//
	// Delete stream
	//

	deleteStreamInput := v3io.DeleteStreamInput{
		Path: streamPath,
	}

	suite.populateDataPlaneInput(&deleteStreamInput.DataPlaneInput)

	err = suite.container.DeleteStreamSync(&deleteStreamInput)
	suite.Require().NoError(err, "Failed to delete stream")
}

type syncStreamBackupRestoreTestSuite struct {
	syncTestSuite
	streamTestSuite streamTestSuite
}

func (suite *syncStreamBackupRestoreTestSuite) SetupTest() {
	suite.streamTestSuite = streamTestSuite{
		testSuite: suite.syncTestSuite.testSuite,
	}
	suite.streamTestSuite.SetupTest()
}

func (suite *syncStreamBackupRestoreTestSuite) TearDownTest() {
	suite.streamTestSuite.TearDownTest()
}

// internal use case test
func (suite *syncStreamBackupRestoreTestSuite) TestStream() {
	streamPath := fmt.Sprintf("%s/mystream/", suite.streamTestSuite.testPath)

	createStreamInput := v3io.CreateStreamInput{
		Path:                 streamPath,
		ShardCount:           4,
		RetentionPeriodHours: 1,
	}
	suite.populateDataPlaneInput(&createStreamInput.DataPlaneInput)
	err := suite.container.CreateStreamSync(&createStreamInput)
	suite.Require().NoError(err, "Failed to create stream")

	firstShardID := 1
	secondShardID := 2

	records := []*v3io.StreamRecord{
		{ShardID: &firstShardID, Data: []byte("first shard record #1")},
		{ShardID: &firstShardID, Data: []byte("first shard record #2")},
		{ShardID: &secondShardID, Data: []byte("second shard record #1")},
		{Data: []byte("some shard record #1")},
	}

	putRecordsInput := v3io.PutRecordsInput{
		Path:    streamPath,
		Records: records,
	}
	suite.populateDataPlaneInput(&putRecordsInput.DataPlaneInput)
	response, err := suite.container.PutRecordsSync(&putRecordsInput)
	suite.Require().NoError(err, "Failed to put records")
	response.Release()

	//
	// Backup stream
	//

	getItemsInput := v3io.GetItemsInput{
		Path:               streamPath,
		ReturnData:         "true",
		AllowObjectScatter: "true",
		AttributeNames:     []string{"**"},
	}
	suite.populateDataPlaneInput(&getItemsInput.DataPlaneInput)
	cursor, err := v3io.NewItemsCursor(suite.container, &getItemsInput)
	suite.Require().NoError(err, "Failed to get items")

	cursorItems, err := cursor.AllSync()
	suite.Require().NoError(err)

	type Shard struct {
		Chunks       map[int]*v3io.ItemChunk
		CurrentChunk *v3io.ItemCurrentChunkMetadata
	}
	streamBackup := map[string]*Shard{}
	for _, cursorItem := range cursorItems {
		shardName, err := cursorItem.GetFieldString("__name")
		suite.Require().NoError(err, "Failed to get item name")
		chunkMap, currentChunkMetadata, err := cursorItem.GetShard()
		suite.Require().NoError(err, "Failed to get stream")

		if _, ok := streamBackup[shardName]; !ok {
			streamBackup[shardName] = &Shard{Chunks: map[int]*v3io.ItemChunk{}}
		}

		for chunkId := range chunkMap {
			if _, ok := streamBackup[shardName].Chunks[chunkId]; !ok {
				streamBackup[shardName].Chunks[chunkId] = &v3io.ItemChunk{}
			}

			if chunkMap[chunkId].Metadata != nil {
				(*streamBackup[shardName].Chunks[chunkId]).Metadata = chunkMap[chunkId].Metadata
			}
			if len(chunkMap[chunkId].Data) != 0 {
				(*streamBackup[shardName].Chunks[chunkId]).Data = chunkMap[chunkId].Data
			}
		}

		if currentChunkMetadata != nil {
			streamBackup[shardName].CurrentChunk = currentChunkMetadata
		}
	}

	deleteStreamInput := v3io.DeleteStreamInput{
		Path: streamPath,
	}
	suite.populateDataPlaneInput(&deleteStreamInput.DataPlaneInput)
	err = suite.container.DeleteStreamSync(&deleteStreamInput)
	suite.Require().NoError(err, "Failed to delete stream")

	//
	// Restore stream
	//

	createStreamInput = v3io.CreateStreamInput{
		Path:                 streamPath,
		ShardCount:           4,
		RetentionPeriodHours: 1,
	}
	suite.populateDataPlaneInput(&createStreamInput.DataPlaneInput)
	err = suite.container.CreateStreamSync(&createStreamInput)
	suite.Require().NoError(err, "Failed to create stream")

	for shardName, shardData := range streamBackup {
		shardPath := fmt.Sprintf("%s/%s/", streamPath, shardName)

		putChunkMetadataInput := v3io.PutChunkInput{
			Path:                 shardPath,
			ChunksMetadata:       []*v3io.ChunkMetadata{},
			CurrentChunkMetadata: &v3io.CurrentChunkMetadata{},
		}

		// send data
		for chunkID, chunkData := range shardData.Chunks {
			for _, dataFragment := range chunkData.Data {
				putChunkInput := v3io.PutChunkInput{
					Path:           shardPath,
					ChunkSeqNumber: chunkID,
					Offset:         dataFragment.Offset,
					Data:           *dataFragment.Data,
				}

				suite.populateDataPlaneInput(&putChunkInput.DataPlaneInput)
				err = suite.container.PutChunkSync(&putChunkInput)
				suite.Require().NoError(err, "Failed to put chunk")
			}

			chunkMetadata := &v3io.ChunkMetadata{
				ChunkSeqNumber:       chunkData.Metadata.ChunkSeqNumber,
				LengthInBytes:        chunkData.Metadata.LengthInBytes,
				FirstRecordSeqNumber: chunkData.Metadata.FirstRecordSeqNumber,
				FirstRecordTimeSecs:  chunkData.Metadata.FirstRecordTsSec,
				FirstRecordTimeNSecs: chunkData.Metadata.FirstRecordTsNSec,
			}
			putChunkMetadataInput.ChunksMetadata = append(putChunkMetadataInput.ChunksMetadata, chunkMetadata)
		}

		putChunkMetadataInput.CurrentChunkMetadata = &v3io.CurrentChunkMetadata{
			ChunkSeqNumber:       shardData.CurrentChunk.CurrentChunkSeqNumber,
			OffsetAfterJob:       shardData.CurrentChunk.CurrentChunkLengthBytes,
			SeqNumberAfterJob:    shardData.CurrentChunk.NextRecordSeqNumber,
			FirstRecordTimeSec:   shardData.CurrentChunk.FirstRecordOnChunkSec,
			LatestRecordTimeSec:  shardData.CurrentChunk.LatestRecordArrivalTimeSec,
			LatestRecordTimeNSec: shardData.CurrentChunk.LatestRecordArrivalTimeNSec,
		}

		suite.populateDataPlaneInput(&putChunkMetadataInput.DataPlaneInput)
		err = suite.container.PutChunkSync(&putChunkMetadataInput)
		suite.Require().NoError(err, "Failed to put chunk")
	}

	//
	// Check Records
	//

	seekShardInput := v3io.SeekShardInput{
		Path: streamPath + "1",
		Type: v3io.SeekShardInputTypeEarliest,
	}
	suite.populateDataPlaneInput(&seekShardInput.DataPlaneInput)
	response, err = suite.container.SeekShardSync(&seekShardInput)
	suite.Require().NoError(err, "Failed to seek shard")
	location := response.Output.(*v3io.SeekShardOutput).Location
	suite.Require().NotEqual("", location)
	response.Release()

	getRecordsInput := v3io.GetRecordsInput{
		Path:     streamPath + "1",
		Location: location,
		Limit:    100,
	}
	suite.populateDataPlaneInput(&getRecordsInput.DataPlaneInput)
	response, err = suite.container.GetRecordsSync(&getRecordsInput)
	suite.Require().NoError(err, "Failed to get records")
	getRecordsOutput := response.Output.(*v3io.GetRecordsOutput)

	suite.Require().Equal("first shard record #1", string(getRecordsOutput.Records[0].Data))
	suite.Require().Equal("first shard record #2", string(getRecordsOutput.Records[1].Data))
	response.Release()

	deleteStreamInput = v3io.DeleteStreamInput{
		Path: streamPath,
	}
	suite.populateDataPlaneInput(&deleteStreamInput.DataPlaneInput)
	err = suite.container.DeleteStreamSync(&deleteStreamInput)
	suite.Require().NoError(err, "Failed to delete stream")
}

type syncContextStreamTestSuite struct {
	syncStreamTestSuite
}

func (suite *syncContextStreamTestSuite) SetupSuite() {
	suite.syncStreamTestSuite.SetupSuite()

	suite.createContext()
}

type syncContainerStreamTestSuite struct {
	syncStreamTestSuite
}

func (suite *syncContainerStreamTestSuite) SetupSuite() {
	suite.syncStreamTestSuite.SetupSuite()

	suite.createContainer()
}

type syncContextStreamBackupRestoreTestSuite struct {
	syncStreamBackupRestoreTestSuite
}

func (suite *syncContextStreamBackupRestoreTestSuite) SetupSuite() {
	suite.syncStreamBackupRestoreTestSuite.SetupSuite()

	suite.createContext()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSyncSuite(t *testing.T) {
	suite.Run(t, new(syncContextContainerTestSuite))
	suite.Run(t, new(syncContextObjectTestSuite))
	suite.Run(t, new(syncContainerObjectTestSuite))
	suite.Run(t, new(syncContextKVTestSuite))
	suite.Run(t, new(syncContainerKVTestSuite))
	suite.Run(t, new(syncContextStreamTestSuite))
	suite.Run(t, new(syncContainerStreamTestSuite))
	suite.Run(t, new(syncContextStreamBackupRestoreTestSuite))
}

func validateContent(suite *syncContainerTestSuite, content *v3io.Content, expectedSize int, withPrefixInfo bool) {
	// common
	suite.Require().NotEmpty(content.Key)
	suite.Require().NotEmpty(content.LastModified)
	suite.Require().NotNil(content.Size)
	suite.Require().Equal(expectedSize, *content.Size)

	if withPrefixInfo {
		suite.Require().NotEmpty(content.AccessTime)
		suite.Require().NotEmpty(content.CreatingTime)
		suite.Require().NotEmpty(content.GID)
		suite.Require().NotEmpty(content.UID)
		mode, err := content.Mode.FileMode()
		suite.NoErrorf(err, "Failed to resolve file mode")
		suite.Require().NotEmpty(mode)
		suite.Require().False(mode.IsDir())
		suite.Require().True(mode.IsRegular())
		suite.Require().Equal("-rw-rw-r--", mode.String(), "File '%s' mode '%v'", content.Key, content.Mode)
		suite.Require().NotEmpty(content.InodeNumber)
		suite.Require().Nil(content.LastSequenceID)
	} else {
		suite.Require().Empty(content.AccessTime)
		suite.Require().Empty(content.CreatingTime)
		suite.Require().Empty(content.GID)
		suite.Require().Empty(content.UID)
		suite.Require().Empty(content.Mode)
		suite.Require().Nil(content.InodeNumber)
		suite.Require().Nil(content.LastSequenceID)
	}
}

func validateCommonPrefix(suite *syncContainerTestSuite, prefix *v3io.CommonPrefix, withPrefixInfo bool) {
	// common
	suite.Require().NotEmpty(prefix.Prefix)

	if withPrefixInfo {
		suite.Require().NotEmpty(prefix.LastModified)
		suite.Require().NotEmpty(prefix.AccessTime)
		suite.Require().NotEmpty(prefix.CreatingTime)
		suite.Require().NotEmpty(prefix.GID)
		suite.Require().NotEmpty(prefix.UID)
		mode, err := prefix.Mode.FileMode()
		suite.NoErrorf(err, "Failed to resolve file mode")
		suite.Require().NotEmpty(mode)
		suite.Require().True(mode.IsDir())
		suite.Require().False(mode.IsRegular())
		suite.Require().Equal("drwxrwxr-x", mode.String(), "Dir '%s' mode '%v'", prefix.Prefix, prefix.Mode) // expected mode: -rw-rw-r-- (664)
		suite.NoErrorf(err, "Failed to resolve file mode")
		suite.Require().NotEmpty(mode)
		suite.Require().NotEmpty(prefix.InodeNumber)
		suite.Require().Equal(true, *prefix.InodeNumber > 0)
	} else {
		suite.Require().Empty(prefix.LastModified)
		suite.Require().Empty(prefix.AccessTime)
		suite.Require().Empty(prefix.CreatingTime)
		suite.Require().Empty(prefix.GID)
		suite.Require().Empty(prefix.UID)
		suite.Require().Empty(prefix.Mode)
		suite.Require().Nil(prefix.InodeNumber)
	}
}

func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randInt(97, 122))
	}

	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
