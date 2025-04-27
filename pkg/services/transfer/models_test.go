// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEndpoint(t *testing.T) {
	// Test creation of an Endpoint struct
	now := time.Now()
	endpoint := Endpoint{
		ID:               "endpoint-id",
		DisplayName:      "Test Endpoint",
		CanonicalName:    "test-endpoint",
		Description:      "A test endpoint",
		OwnerString:      "John Doe",
		OwnerID:          "user-id",
		Organization:     "Test Org",
		Department:       "Test Dept",
		Keywords:         []string{"test", "example"},
		ContactEmail:     "contact@example.com",
		ContactInfo:      "Contact info",
		Public:           true,
		Activated:        true,
		DeactivationTime: &now,
		LocalUserInfo: &LocalUserInfo{
			Username: "testuser",
			UID:      1000,
			GID:      1000,
			HomeDir:  "/home/testuser",
		},
		LocalUserInfoAvailable: true,
	}

	// Check that fields are set correctly
	if endpoint.ID != "endpoint-id" {
		t.Errorf("Endpoint.ID = %v, want %v", endpoint.ID, "endpoint-id")
	}
	if endpoint.DisplayName != "Test Endpoint" {
		t.Errorf("Endpoint.DisplayName = %v, want %v", endpoint.DisplayName, "Test Endpoint")
	}
	if endpoint.CanonicalName != "test-endpoint" {
		t.Errorf("Endpoint.CanonicalName = %v, want %v", endpoint.CanonicalName, "test-endpoint")
	}
	if endpoint.OwnerString != "John Doe" {
		t.Errorf("Endpoint.OwnerString = %v, want %v", endpoint.OwnerString, "John Doe")
	}
	if endpoint.OwnerID != "user-id" {
		t.Errorf("Endpoint.OwnerID = %v, want %v", endpoint.OwnerID, "user-id")
	}
	if endpoint.Organization != "Test Org" {
		t.Errorf("Endpoint.Organization = %v, want %v", endpoint.Organization, "Test Org")
	}
	if !endpoint.Public {
		t.Errorf("Endpoint.Public = %v, want %v", endpoint.Public, true)
	}
	if !endpoint.Activated {
		t.Errorf("Endpoint.Activated = %v, want %v", endpoint.Activated, true)
	}
	if endpoint.DeactivationTime.Unix() != now.Unix() {
		t.Errorf("Endpoint.DeactivationTime = %v, want %v", endpoint.DeactivationTime, now)
	}
	if endpoint.LocalUserInfo.Username != "testuser" {
		t.Errorf("Endpoint.LocalUserInfo.Username = %v, want %v", endpoint.LocalUserInfo.Username, "testuser")
	}
	if endpoint.LocalUserInfo.UID != 1000 {
		t.Errorf("Endpoint.LocalUserInfo.UID = %v, want %v", endpoint.LocalUserInfo.UID, 1000)
	}
	if endpoint.LocalUserInfo.GID != 1000 {
		t.Errorf("Endpoint.LocalUserInfo.GID = %v, want %v", endpoint.LocalUserInfo.GID, 1000)
	}
	if endpoint.LocalUserInfo.HomeDir != "/home/testuser" {
		t.Errorf("Endpoint.LocalUserInfo.HomeDir = %v, want %v", endpoint.LocalUserInfo.HomeDir, "/home/testuser")
	}
	if !endpoint.LocalUserInfoAvailable {
		t.Errorf("Endpoint.LocalUserInfoAvailable = %v, want %v", endpoint.LocalUserInfoAvailable, true)
	}
}

func TestCollection(t *testing.T) {
	// Test creation of a Collection struct
	collection := Collection{
		ID:               "collection-id",
		Name:             "Test Collection",
		Description:      "A test collection",
		BasePath:         "/path/to/collection",
		AdminReadACL:     true,
		AdminWriteACL:    true,
		IdentitiesRead:   true,
		IdentitiesWrite:  true,
		UserManageShares: true,
		UserMessageOnly:  false,
	}

	// Check that fields are set correctly
	if collection.ID != "collection-id" {
		t.Errorf("Collection.ID = %v, want %v", collection.ID, "collection-id")
	}
	if collection.Name != "Test Collection" {
		t.Errorf("Collection.Name = %v, want %v", collection.Name, "Test Collection")
	}
	if collection.Description != "A test collection" {
		t.Errorf("Collection.Description = %v, want %v", collection.Description, "A test collection")
	}
	if collection.BasePath != "/path/to/collection" {
		t.Errorf("Collection.BasePath = %v, want %v", collection.BasePath, "/path/to/collection")
	}
	if !collection.AdminReadACL {
		t.Errorf("Collection.AdminReadACL = %v, want %v", collection.AdminReadACL, true)
	}
	if !collection.AdminWriteACL {
		t.Errorf("Collection.AdminWriteACL = %v, want %v", collection.AdminWriteACL, true)
	}
	if !collection.IdentitiesRead {
		t.Errorf("Collection.IdentitiesRead = %v, want %v", collection.IdentitiesRead, true)
	}
	if !collection.IdentitiesWrite {
		t.Errorf("Collection.IdentitiesWrite = %v, want %v", collection.IdentitiesWrite, true)
	}
	if !collection.UserManageShares {
		t.Errorf("Collection.UserManageShares = %v, want %v", collection.UserManageShares, true)
	}
	if collection.UserMessageOnly {
		t.Errorf("Collection.UserMessageOnly = %v, want %v", collection.UserMessageOnly, false)
	}
}

func TestTask(t *testing.T) {
	// Create a sample time for testing
	now := time.Now()
	completionTime := now.Add(time.Hour)
	deadline := now.Add(time.Hour * 24)
	cancelTime := now.Add(time.Minute * 30)

	// Test creation of a Task struct
	task := Task{
		DataType:               "task",
		TaskID:                 "task-id",
		Type:                   "TRANSFER",
		Status:                 "ACTIVE",
		Label:                  "Test Transfer",
		SourceEndpointID:       "source-endpoint-id",
		SourceEndpointDisplay:  "Source Endpoint",
		DestinationEndpointID:  "destination-endpoint-id",
		DestEndpointDisplay:    "Destination Endpoint",
		RequestTime:            now,
		CompletionTime:         &completionTime,
		Deadline:               &deadline,
		CancelTime:             &cancelTime,
		CreatorID:              "creator-id",
		OwnerID:                "owner-id",
		FilesTransferred:       10,
		FilesSkipped:           2,
		BytesTransferred:       1024,
		BytesSkipped:           256,
		Subtasks:               5,
		SubtasksSucceeded:      3,
		SubtasksFailed:         1,
		SubtasksPending:        1,
		SubtasksCanceled:       0,
		SubtasksExpired:        0,
		SyncLevel:              3,
		EncryptData:            true,
		VerifyChecksum:         true,
		DeleteDestinationExtra: false,
		Recursive:              true,
		UseSharing:             true,
		HistoryDeleted:         false,
		SkipSourceErrors:       false,
		FailOnQuotaErrors:      true,
		PreserveMtime:          true,
		UserMessage:            "User message",
	}

	// Check that fields are set correctly
	if task.TaskID != "task-id" {
		t.Errorf("Task.TaskID = %v, want %v", task.TaskID, "task-id")
	}
	if task.Type != "TRANSFER" {
		t.Errorf("Task.Type = %v, want %v", task.Type, "TRANSFER")
	}
	if task.Status != "ACTIVE" {
		t.Errorf("Task.Status = %v, want %v", task.Status, "ACTIVE")
	}
	if task.Label != "Test Transfer" {
		t.Errorf("Task.Label = %v, want %v", task.Label, "Test Transfer")
	}
	if task.SourceEndpointID != "source-endpoint-id" {
		t.Errorf("Task.SourceEndpointID = %v, want %v", task.SourceEndpointID, "source-endpoint-id")
	}
	if task.DestinationEndpointID != "destination-endpoint-id" {
		t.Errorf("Task.DestinationEndpointID = %v, want %v", task.DestinationEndpointID, "destination-endpoint-id")
	}
	if task.RequestTime.Unix() != now.Unix() {
		t.Errorf("Task.RequestTime = %v, want %v", task.RequestTime, now)
	}
	if task.CompletionTime.Unix() != completionTime.Unix() {
		t.Errorf("Task.CompletionTime = %v, want %v", task.CompletionTime, completionTime)
	}
	if task.Deadline.Unix() != deadline.Unix() {
		t.Errorf("Task.Deadline = %v, want %v", task.Deadline, deadline)
	}
	if task.CancelTime.Unix() != cancelTime.Unix() {
		t.Errorf("Task.CancelTime = %v, want %v", task.CancelTime, cancelTime)
	}
	if task.FilesTransferred != 10 {
		t.Errorf("Task.FilesTransferred = %v, want %v", task.FilesTransferred, 10)
	}
	if task.FilesSkipped != 2 {
		t.Errorf("Task.FilesSkipped = %v, want %v", task.FilesSkipped, 2)
	}
	if task.BytesTransferred != 1024 {
		t.Errorf("Task.BytesTransferred = %v, want %v", task.BytesTransferred, 1024)
	}
	if task.SyncLevel != 3 {
		t.Errorf("Task.SyncLevel = %v, want %v", task.SyncLevel, 3)
	}
	if !task.EncryptData {
		t.Errorf("Task.EncryptData = %v, want %v", task.EncryptData, true)
	}
	if !task.VerifyChecksum {
		t.Errorf("Task.VerifyChecksum = %v, want %v", task.VerifyChecksum, true)
	}
	if task.DeleteDestinationExtra {
		t.Errorf("Task.DeleteDestinationExtra = %v, want %v", task.DeleteDestinationExtra, false)
	}
	if !task.Recursive {
		t.Errorf("Task.Recursive = %v, want %v", task.Recursive, true)
	}
}

func TestTransferTaskRequest(t *testing.T) {
	// Create a sample deadline for testing
	deadline := time.Now().Add(time.Hour * 24)

	// Test creation of a TransferTaskRequest struct
	request := TransferTaskRequest{
		DataType:               "transfer",
		Label:                  "Test Transfer",
		SourceEndpointID:       "source-endpoint-id",
		DestinationEndpointID:  "destination-endpoint-id",
		VerifyChecksum:         true,
		Encrypt:                true,
		SyncLevel:              3,
		DeleteDestinationExtra: false,
		Deadline:               &deadline,
		NotifyOnSucceeded:      true,
		NotifyOnFailed:         true,
		NotifyOnInactive:       false,
		SkipSourceErrors:       false,
		FailOnQuotaErrors:      true,
		SubmissionID:           "submission-id",
		UseSharing:             true,
		PreserveMtime:          true,
		Items: []TransferItem{
			{
				SourcePath:      "/source/path",
				DestinationPath: "/destination/path",
				Recursive:       true,
				Checksum:        "md5:1234567890abcdef",
			},
		},
	}

	// Check that fields are set correctly
	if request.DataType != "transfer" {
		t.Errorf("TransferTaskRequest.DataType = %v, want %v", request.DataType, "transfer")
	}
	if request.Label != "Test Transfer" {
		t.Errorf("TransferTaskRequest.Label = %v, want %v", request.Label, "Test Transfer")
	}
	if request.SourceEndpointID != "source-endpoint-id" {
		t.Errorf("TransferTaskRequest.SourceEndpointID = %v, want %v", request.SourceEndpointID, "source-endpoint-id")
	}
	if request.DestinationEndpointID != "destination-endpoint-id" {
		t.Errorf("TransferTaskRequest.DestinationEndpointID = %v, want %v", request.DestinationEndpointID, "destination-endpoint-id")
	}
	if !request.VerifyChecksum {
		t.Errorf("TransferTaskRequest.VerifyChecksum = %v, want %v", request.VerifyChecksum, true)
	}
	if !request.Encrypt {
		t.Errorf("TransferTaskRequest.Encrypt = %v, want %v", request.Encrypt, true)
	}
	if request.SyncLevel != 3 {
		t.Errorf("TransferTaskRequest.SyncLevel = %v, want %v", request.SyncLevel, 3)
	}
	if request.Deadline.Unix() != deadline.Unix() {
		t.Errorf("TransferTaskRequest.Deadline = %v, want %v", request.Deadline, deadline)
	}
	if !request.NotifyOnSucceeded {
		t.Errorf("TransferTaskRequest.NotifyOnSucceeded = %v, want %v", request.NotifyOnSucceeded, true)
	}
	if !request.NotifyOnFailed {
		t.Errorf("TransferTaskRequest.NotifyOnFailed = %v, want %v", request.NotifyOnFailed, true)
	}
	if len(request.Items) != 1 {
		t.Fatalf("len(TransferTaskRequest.Items) = %v, want %v", len(request.Items), 1)
	}
	if request.Items[0].SourcePath != "/source/path" {
		t.Errorf("TransferTaskRequest.Items[0].SourcePath = %v, want %v", request.Items[0].SourcePath, "/source/path")
	}
	if request.Items[0].DestinationPath != "/destination/path" {
		t.Errorf("TransferTaskRequest.Items[0].DestinationPath = %v, want %v", request.Items[0].DestinationPath, "/destination/path")
	}
	if !request.Items[0].Recursive {
		t.Errorf("TransferTaskRequest.Items[0].Recursive = %v, want %v", request.Items[0].Recursive, true)
	}
	if request.Items[0].Checksum != "md5:1234567890abcdef" {
		t.Errorf("TransferTaskRequest.Items[0].Checksum = %v, want %v", request.Items[0].Checksum, "md5:1234567890abcdef")
	}
}

func TestFileListItem(t *testing.T) {
	// Test creation of a FileListItem struct
	fileItem := FileListItem{
		DataType:     "file",
		Name:         "test.txt",
		Type:         "file",
		Size:         1024,
		LastModified: "2023-01-01 12:00:00",
		Permissions:  "rw-r--r--",
		User:         "testuser",
		Group:        "testgroup",
	}

	dirItem := FileListItem{
		DataType:     "dir",
		Name:         "testdir",
		Type:         "dir",
		LastModified: "2023-01-01 12:00:00",
		Permissions:  "rwxr-xr-x",
		User:         "testuser",
		Group:        "testgroup",
	}

	linkItem := FileListItem{
		DataType:     "file",
		Name:         "testlink",
		Type:         "file",
		LastModified: "2023-01-01 12:00:00",
		Permissions:  "rwxrwxrwx",
		User:         "testuser",
		Group:        "testgroup",
		Link:         "test.txt",
	}

	// Check that fields are set correctly for file
	if fileItem.DataType != "file" {
		t.Errorf("FileListItem.DataType = %v, want %v", fileItem.DataType, "file")
	}
	if fileItem.Name != "test.txt" {
		t.Errorf("FileListItem.Name = %v, want %v", fileItem.Name, "test.txt")
	}
	if fileItem.Type != "file" {
		t.Errorf("FileListItem.Type = %v, want %v", fileItem.Type, "file")
	}
	if fileItem.Size != 1024 {
		t.Errorf("FileListItem.Size = %v, want %v", fileItem.Size, 1024)
	}
	if fileItem.LastModified != "2023-01-01 12:00:00" {
		t.Errorf("FileListItem.LastModified = %v, want %v", fileItem.LastModified, "2023-01-01 12:00:00")
	}
	if fileItem.Permissions != "rw-r--r--" {
		t.Errorf("FileListItem.Permissions = %v, want %v", fileItem.Permissions, "rw-r--r--")
	}
	if fileItem.User != "testuser" {
		t.Errorf("FileListItem.User = %v, want %v", fileItem.User, "testuser")
	}
	if fileItem.Group != "testgroup" {
		t.Errorf("FileListItem.Group = %v, want %v", fileItem.Group, "testgroup")
	}

	// Check that fields are set correctly for directory
	if dirItem.DataType != "dir" {
		t.Errorf("FileListItem(dir).DataType = %v, want %v", dirItem.DataType, "dir")
	}
	if dirItem.Name != "testdir" {
		t.Errorf("FileListItem(dir).Name = %v, want %v", dirItem.Name, "testdir")
	}
	if dirItem.Type != "dir" {
		t.Errorf("FileListItem(dir).Type = %v, want %v", dirItem.Type, "dir")
	}

	// Check that fields are set correctly for symlink
	if linkItem.Link != "test.txt" {
		t.Errorf("FileListItem(link).Link = %v, want %v", linkItem.Link, "test.txt")
	}
}

func TestJSONMarshaling(t *testing.T) {
	// Test JSON marshaling and unmarshaling of a TransferTaskRequest
	request := TransferTaskRequest{
		DataType:              "transfer",
		Label:                 "Test Transfer",
		SourceEndpointID:      "source-endpoint-id",
		DestinationEndpointID: "destination-endpoint-id",
		Items: []TransferItem{
			{
				SourcePath:      "/source/path",
				DestinationPath: "/destination/path",
				Recursive:       true,
			},
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal TransferTaskRequest: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaled TransferTaskRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal TransferTaskRequest: %v", err)
	}

	// Check that fields match
	if unmarshaled.DataType != request.DataType {
		t.Errorf("Unmarshaled.DataType = %v, want %v", unmarshaled.DataType, request.DataType)
	}
	if unmarshaled.Label != request.Label {
		t.Errorf("Unmarshaled.Label = %v, want %v", unmarshaled.Label, request.Label)
	}
	if unmarshaled.SourceEndpointID != request.SourceEndpointID {
		t.Errorf("Unmarshaled.SourceEndpointID = %v, want %v", unmarshaled.SourceEndpointID, request.SourceEndpointID)
	}
	if unmarshaled.DestinationEndpointID != request.DestinationEndpointID {
		t.Errorf("Unmarshaled.DestinationEndpointID = %v, want %v", unmarshaled.DestinationEndpointID, request.DestinationEndpointID)
	}
	if len(unmarshaled.Items) != len(request.Items) {
		t.Fatalf("len(Unmarshaled.Items) = %v, want %v", len(unmarshaled.Items), len(request.Items))
	}
	if unmarshaled.Items[0].SourcePath != request.Items[0].SourcePath {
		t.Errorf("Unmarshaled.Items[0].SourcePath = %v, want %v", unmarshaled.Items[0].SourcePath, request.Items[0].SourcePath)
	}
	if unmarshaled.Items[0].DestinationPath != request.Items[0].DestinationPath {
		t.Errorf("Unmarshaled.Items[0].DestinationPath = %v, want %v", unmarshaled.Items[0].DestinationPath, request.Items[0].DestinationPath)
	}
	if unmarshaled.Items[0].Recursive != request.Items[0].Recursive {
		t.Errorf("Unmarshaled.Items[0].Recursive = %v, want %v", unmarshaled.Items[0].Recursive, request.Items[0].Recursive)
	}
}
