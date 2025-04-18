// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package build

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

// AzureBlobstoreConfig is an authentication and configuration struct containing
// the data needed by the Azure SDK to interact with a specific container in the
// blobstore.
// AzureBlobstoreConfig 是一个认证和配置结构体，包含 Azure SDK 与 blobstore 中特定容器交互所需的数据。
type AzureBlobstoreConfig struct {
	Account string // Account name to authorize API requests with
	// 账户名称，用于授权 API 请求
	Token string // Access token for the above account
	// 上述账户的访问令牌
	Container string // Blob container to upload files into
	// 用于上传文件的 Blob 容器
}

// AzureBlobstoreUpload uploads a local file to the Azure Blob Storage. Note, this
// method assumes a max file size of 64MB (Azure limitation). Larger files will
// need a multi API call approach implemented.
//
// See: https://msdn.microsoft.com/en-us/library/azure/dd179451.aspx#Anchor_3
// AzureBlobstoreUpload 将本地文件上传到 Azure Blob 存储。注意，此方法假定最大文件大小为 64MB（Azure 限制）。更大的文件需要实现多 API 调用方法。
func AzureBlobstoreUpload(path string, name string, config AzureBlobstoreConfig) error {
	if *DryRunFlag {
		fmt.Printf("would upload %q to %s/%s/%s\n", path, config.Account, config.Container, name)
		return nil
	}
	// Create an authenticated client against the Azure cloud
	// 创建一个针对 Azure 云的认证客户端
	credential, err := azblob.NewSharedKeyCredential(config.Account, config.Token)
	if err != nil {
		return err
	}
	a := fmt.Sprintf("https://%s.blob.core.windows.net/", config.Account)
	client, err := azblob.NewClientWithSharedKeyCredential(a, credential, nil)
	if err != nil {
		return err
	}
	// Stream the file to upload into the designated blobstore container
	// 将要上传的文件流式传输到指定的 blobstore 容器
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = client.UploadFile(context.Background(), config.Container, name, in, nil)
	return err
}

// AzureBlobstoreList lists all the files contained within an azure blobstore.
// AzureBlobstoreList 列出 Azure blobstore 中包含的所有文件。
func AzureBlobstoreList(config AzureBlobstoreConfig) ([]*container.BlobItem, error) {
	// Create an authenticated client against the Azure cloud
	// 创建一个针对 Azure 云的认证客户端
	credential, err := azblob.NewSharedKeyCredential(config.Account, config.Token)
	if err != nil {
		return nil, err
	}
	a := fmt.Sprintf("https://%s.blob.core.windows.net/", config.Account)
	client, err := azblob.NewClientWithSharedKeyCredential(a, credential, nil)
	if err != nil {
		return nil, err
	}
	pager := client.NewListBlobsFlatPager(config.Container, nil)

	var blobs []*container.BlobItem
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		blobs = append(blobs, page.Segment.BlobItems...)
	}
	return blobs, nil
}

// AzureBlobstoreDelete iterates over a list of files to delete and removes them
// from the blobstore.
// AzureBlobstoreDelete 遍历要删除的文件列表并从 blobstore 中移除它们。
func AzureBlobstoreDelete(config AzureBlobstoreConfig, blobs []*container.BlobItem) error {
	if *DryRunFlag {
		for _, blob := range blobs {
			fmt.Printf("would delete %s (%s) from %s/%s\n", *blob.Name, blob.Properties.LastModified, config.Account, config.Container)
		}
		return nil
	}
	// Create an authenticated client against the Azure cloud
	// 创建一个针对 Azure 云的认证客户端
	credential, err := azblob.NewSharedKeyCredential(config.Account, config.Token)
	if err != nil {
		return err
	}
	a := fmt.Sprintf("https://%s.blob.core.windows.net/", config.Account)
	client, err := azblob.NewClientWithSharedKeyCredential(a, credential, nil)
	if err != nil {
		return err
	}
	// Iterate over the blobs and delete them
	// 遍历 blobs 并删除它们
	for _, blob := range blobs {
		if _, err := client.DeleteBlob(context.Background(), config.Container, *blob.Name, nil); err != nil {
			return err
		}
		fmt.Printf("deleted  %s (%s)\n", *blob.Name, blob.Properties.LastModified)
	}
	return nil
}
