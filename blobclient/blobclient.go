package blobclient

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

const FILES_CONTAINER_NAME = "files"

var blobClient *azblob.Client

func initializeBlobClient() (*azblob.Client, error) {
	// Define the connection string to the Azure Blob Storage account
	// Key is the default azurite key
	var connectionString = "AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;DefaultEndpointsProtocol=http;BlobEndpoint=http://azurite:10000/devstoreaccount1;QueueEndpoint=http://azurite:10001/devstoreaccount1;TableEndpoint=http://azurite:10002/devstoreaccount1;" //os.Getenv("AZURITE_CONNECTION_STRING")

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

func GetFoldersInContainer(containerName string) ([]string, error) {
	blobClient, err := GetBlobClient()
	if err != nil {
		return nil, err
	}

	pager := blobClient.NewListBlobsFlatPager(containerName, nil)
	var folderNames []string

	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, blob := range resp.Segment.BlobItems {
			folderNames = append(folderNames, *blob.Name)
		}
	}

	return folderNames, nil
}

func DeleteFolderInContainer(containerName string, folderName string) error {
	blobClient, err := GetBlobClient()
	if err != nil {
		return err
	}

	folderPrefix := folderName + "/"

	containerClient := blobClient.ServiceClient().NewContainerClient(containerName)

	// List all blobs with the prefix (folder name)
	pager := containerClient.NewListBlobsFlatPager(&azblob.ListBlobsFlatOptions{
		Prefix: &folderPrefix, // Filter blobs with folder prefix
	})

	// Loop through blobs and delete each one
	for pager.More() {
		resp, err := pager.NextPage(context.TODO())
		if err != nil {
			log.Fatalf("Error listing blobs: %v", err)
		}
		for _, blob := range resp.Segment.BlobItems {
			// Get blob name and delete it
			blobName := *blob.Name
			blobClient := containerClient.NewBlobClient(blobName)
			_, err := blobClient.Delete(context.TODO(), nil)
			if err != nil {
				log.Printf("Failed to delete blob: %v", err)
				return err
			} else {
				log.Printf("Deleted blob: %s", blobName)
			}
		}
	}

	return nil
}

// Check if a container exists and create it if it doesn't
func CreateContainer(blobClient *azblob.Client, containerName string) {
	// Get a reference to the container
	container := blobClient.ServiceClient().NewContainerClient(containerName)

	// Create the container if it doesn't exist
	_, err := container.Create(context.Background(), nil)
	if err != nil {
		fmt.Printf("Failed to create container: %v\n", err)
	}

}

func CheckContainerExists(blobClient *azblob.Client, containerName string) bool {
	// Get a reference to the container
	serviceClient := blobClient.ServiceClient()

	// Create a client for the specific container
	containerClient := serviceClient.NewContainerClient(containerName)

	_, err := containerClient.Create(context.Background(), nil)
	if err != nil {
		log.Default().Println("Container already exists")
		return true
	}
	return false
}
