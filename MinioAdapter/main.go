package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {

	endpoint := "localhost:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	router := gin.Default()

	router.GET("/getObject", func(c *gin.Context) {
		getFromBucketWithKey(c, minioClient)
	})

	router.POST("/putObject", func(c *gin.Context) {
		putObjectWithBucketWithKey(c, minioClient)
	})

	router.GET("/getFilesFromBucketWitPrefix", func(c *gin.Context) {
		getFilesInBucketByPrefix(c, minioClient)
	})

	router.Run("localhost:8085")
}

func getFromBucketWithKey(c *gin.Context, minioClient *minio.Client) {
	bucketName := c.Query("bucketName")
	key := c.Query("key")
	formatJson := c.Query("formatJson")

	if bucketName == "" || key == "" {
		// If any of the parameters are missing, return HTTP 500 with an error message
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Both 'bucketName' and 'key' must be provided.",
		})
		return
	}

	obj, err := minioClient.GetObject(context.Background(), bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Cant get what you asked for",
		})
		return
	}
	defer obj.Close()

	data, err := ioutil.ReadAll(obj)
	if err != nil {
		panic(err)
	}

	if formatJson == "true" {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse JSON data"})
			return
		}
		c.JSON(http.StatusOK, jsonData)
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}

func putObjectWithBucketWithKey(c *gin.Context, minioClient *minio.Client) {

	bucketName := c.Query("bucketName")
	key := c.Query("key")

	if bucketName == "" || key == "" {
		// If any of the parameters are missing, return HTTP 500 with an error message
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Both 'bucketName' and 'key' must be provided.",
		})
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read request body",
		})
		return
	}

	if len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No data provided in request body",
		})
		return
	}

	_, err = minioClient.PutObject(
		context.Background(),  // Context to control cancellations and timeouts
		bucketName,            // The name of the bucket
		key,                   // The object key (name of the file within the bucket)
		bytes.NewReader(body), // The data to upload
		int64(len(body)),      // The size of the data to upload
		minio.PutObjectOptions{
			ContentType: "application/json", // Explicitly setting the MIME type as JSON
		},
	)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Can't upload this",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Object successfully uploaded",
	})

}

func getFilesInBucketByPrefix(c *gin.Context, minioClient *minio.Client) {

	bucketName := c.Query("bucketName")
	key := c.Query("key")
	formatJson := c.Query("formatJson")

	if bucketName == "" {
		// If any of the parameters are missing, return HTTP 500 with an error message
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Both 'bucketName' and 'key' must be provided.",
		})
		return
	}

	objectCh := minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{
		Prefix:    key,
		Recursive: true,
	})

	var data []interface{}

	for object := range objectCh {
		if object.Err != nil {
			log.Println(object.Err)
			continue // Skip this object and move to the next
		}

		// Retrieve the object data.
		objectData, err := minioClient.GetObject(context.Background(), bucketName, object.Key, minio.GetObjectOptions{})
		if err != nil {
			log.Println(err)
			continue // Skip this object and move to the next
		}

		rawData, err := ioutil.ReadAll(objectData)
		if err != nil {
			log.Println(err)
			continue // Skip this object and move to the next
		}

		if formatJson == "true" {
			var parsedData map[string]interface{}
			if json.Unmarshal(rawData, &parsedData) == nil {
				data = append(data, parsedData)
			} else {
				log.Println("Failed to parse JSON:", err)
				continue // If JSON parsing fails, skip this object
			}
		} else {
			data = append(data, string(rawData))
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}
