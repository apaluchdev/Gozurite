package routes

import (
	"bytes"
	"context"
	"fmt"
	"gozurite/blobclient"
	"gozurite/expiryhelper"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/gin-gonic/gin"
)

func RegisterFileRoutes(router *gin.Engine) {
	fileRoutes := router.Group("/file")
	{
		fileRoutes.GET("/query/:pin", queryFiles)
		fileRoutes.GET("/query", queryFolders)
		fileRoutes.GET("/:pin/:id", getFile)
		fileRoutes.GET("/pin/:pin", getPinExpiry)
		fileRoutes.POST("", uploadFile)
		fileRoutes.DELETE("/:id", deleteFile)
		fileRoutes.DELETE("/pin/:pin", expirePin)
	}
}

func queryFolders(c *gin.Context) {
	log.Default().Println("Querying folders...")

	// Get the list of folders in the container
	folders, err := blobclient.GetFoldersInContainer(blobclient.FILES_CONTAINER_NAME)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the folder names as JSON
	c.JSON(http.StatusOK, gin.H{"folders": folders})
}

func queryFiles(c *gin.Context) {
	log.Default().Println("Querying files...")
	pin := c.Param("pin")

	blobClient, err := blobclient.GetBlobClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	containerName := "files" // Replace with your container name
	folderPrefix := pin + "/"

	//ensureContainerExists(blobClient, containerName)

	// Create a slice to hold blob names
	var blobNames []string

	// Use ListBlobsFlat to enumerate blobs in the container with a prefix
	pager := blobClient.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
		Prefix: &folderPrefix, // Filter blobs within the specified folder
	})

	// Iterate over the pager to retrieve all blobs
	for pager.More() {
		// Get the next page of blobs
		page, err := pager.NextPage(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Append the names of the blobs to the slice
		for _, blob := range page.Segment.BlobItems {
			blobNames = append(blobNames, *blob.Name)
		}
	}

	// Return the blob names as JSON
	c.JSON(http.StatusOK, gin.H{"blobs": blobNames})
}

func uploadFile(c *gin.Context) {
	log.Default().Println("Uploading file...")

	log.Default().Println("Checking for file...")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file"})
		return
	}
	defer file.Close()

	// Get the pin metadata from the form data
	log.Default().Println("Checking for pin...")
	pin := c.PostForm("pin")
	if pin == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PIN not specified"})
		return
	}

	// Check if pin is a numeric value
	log.Default().Println("Verifying pin...")
	if _, err := strconv.Atoi(pin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PIN must be a numeric value"})
		return
	}

	if len(pin) != 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PIN must be 5 digits"})
		return
	}

	// Get the pin metadata from the form data
	log.Default().Println("Checking for filename...")
	filename := c.PostForm("filename")
	if pin == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename not specified"})
		return
	}

	// Check if the filename contains any invalid characters
	chars := "<>:\"/\\|?*"
	if strings.ContainsAny(filename, chars) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid character in filename"})
		return
	}

	// Get the size of the file
	fileSize := header.Size
	fmt.Printf("File size: %d bytes\n", fileSize)

	// Create a byte slice of the appropriate size
	fileBytes := make([]byte, fileSize)

	// Read the file into the byte slice
	_, err = io.ReadFull(file, fileBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file into byte array"})
		return
	}

	// Do something with fileBytes
	fmt.Printf("File received successfully, size: %d bytes\n", len(fileBytes))

	blobClient, err := blobclient.GetBlobClient()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Default().Println("Uploading file to blob storage... " + pin + `/` + filename)

	_, err = blobClient.UploadBuffer(c, "files", pin+`/`+filename, fileBytes, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get the expiry time for the pin
	log.Default().Println("Checking for expiry time...")
	expiryInHours := c.PostForm("expiryInHours")
	if expiryInHours == "" {
		log.Default().Println("Expiry not specified, using default of 8 hours")
		expiryhelper.AddPinExpiry(pin, 8)
	} else {
		hours, err := strconv.Atoi(expiryInHours)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expiry time"})
			return
		}
		expiryhelper.AddPinExpiry(pin, hours)
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

func getFile(c *gin.Context) {
	filename := c.Param("id")
	pin := c.Param("pin")

	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename not specified"})
		return
	}

	// Download the file from Azurite and serve it
	err := serveFileFromBlob(c, pin, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func expirePin(c *gin.Context) {
	err := blobclient.DeleteFolderInContainer(blobclient.FILES_CONTAINER_NAME, c.Param("pin"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	expiryhelper.RemovePinExpiry(c.Param("pin"))

	c.JSON(200, gin.H{"message": "Pin expired, all files deleted"})
}

func getPinExpiry(c *gin.Context) {
	expiryTime, exists := expiryhelper.GetPinExpiry(c.Param("pin"))
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "PIN not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"expiryTime": expiryTime})
}

func deleteFile(c *gin.Context) {
	// Handle delete file logic
	c.JSON(200, gin.H{"message": "File deleted"})
}

func serveFileFromBlob(c *gin.Context, folder, filename string) error {
	client, err := blobclient.GetBlobClient()
	if err != nil {
		log.Default().Println("Error getting blob client: ", err)
		return err
	}

	// Construct the blob path using the folder path and filename
	blobPath := fmt.Sprintf("%s/%s", folder, filename)
	containerName := "files" // Replace with your container name

	// Download the blob
	get, err := client.DownloadStream(c, containerName, blobPath, nil)
	if err != nil {
		log.Default().Println("Error downloading file from blob storage: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file from blob storage"})
		return err
	}

	// Read the downloaded data
	downloadedData := bytes.Buffer{}
	retryReader := get.NewRetryReader(c, &azblob.RetryReaderOptions{MaxRetries: 3})
	_, err = downloadedData.ReadFrom(retryReader)
	if err != nil {
		return fmt.Errorf("failed to read file from response: %v", err)
	}

	err = retryReader.Close()
	if err != nil {
		return fmt.Errorf("failed to close retry reader: %v", err)
	}

	// Set the response headers for file download
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", downloadedData.Len()))

	// Write the buffer content to the response
	c.Data(http.StatusOK, "application/octet-stream", downloadedData.Bytes())

	return nil
}
