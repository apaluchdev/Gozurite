package blobclient

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

var blobClient *azblob.Client

func initializeBlobClient() (*azblob.Client, error) {
	// Define the connection string to the Azure Blob Storage account
	// Key is the default azurite key
	var connectionString = "AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;DefaultEndpointsProtocol=http;BlobEndpoint=http://azurite:10000/devstoreaccount1;QueueEndpoint=http://azurite:10001/devstoreaccount1;TableEndpoint=http://azurite:10002/devstoreaccount1;"

	// Create the Azure Blob client
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Blob client: %v", err)
	}
	return client, nil
}

func GetBlobClient() (*azblob.Client, error) {
	if blobClient != nil {
		return blobClient, nil
	}

	blobClient, err := initializeBlobClient()
	if err != nil {
		return nil, err
	}

	return blobClient, nil
}
