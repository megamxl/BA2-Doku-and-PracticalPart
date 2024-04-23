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

type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

func main() {

	endpoint := "192.168.76.198:9000"
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
	router.GET("/albums", getAlbums)

	router.GET("/getObject", func(c *gin.Context) {
		getFromBucketWithKey(c, minioClient)
	})

	router.POST("/putObject", func(c *gin.Context) {
		putObjectWithBucketWithKey(c, minioClient)
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
