package main

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
	"os"
)

func extractHandler(c *gin.Context) {

	result, err := ExtractIQData("file.wav")
	if err != nil {

		c.JSON(400, gin.H{
			"Error": map[string]interface{}{
				"Op":   "open",
				"Path": "file.wav",
				"Err":  err.Error(),
			},
		})
		return
	}

	encodedResult := base64.StdEncoding.EncodeToString(result)

	c.JSON(200, gin.H{
		"ExtractedData": encodedResult,
	})
}

func uploadHandler(c *gin.Context) {
	// Retrieve the uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"Error": "Error retrieving file"})
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// Create a file named "file.wav" in the current directory
	out, err := os.Create("file.wav")
	if err != nil {
		c.JSON(500, gin.H{"Error": "Could not create file"})
		return
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {

		}
	}(out)

	// Write the uploaded file to disk
	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(500, gin.H{"Error": "Could not save file"})
		return
	}

	c.JSON(200, gin.H{"Message": "File uploaded successfully", "Filename": header.Filename})
}

func main() {

	defer func() {
		err := os.Remove("file.wav")
		if err == nil {
			fmt.Println("Temporary file 'file.wav' deleted.")
		} else if !os.IsNotExist(err) {
			fmt.Printf("Error deleting 'file.wav': %v\n", err)
		}
	}()

	//router for request-handling
	router := gin.Default()
	router.Use(static.Serve("/", static.LocalFile("./bg/build", true)))

	router.POST("/api/upload", uploadHandler)
	router.GET("/api/extractHandler", extractHandler)
	err := router.Run("localhost:8080")
	if err != nil {
		return
	}

	/*
		// Load IQ data from file
		iqData, err := ExtractIQData("datei.wav")
		if err != nil {
			log.Fatal("Error extracting IQ data:", err)
		}
		data := ConvertIQData(iqData)
		fmt.Println(data)

		// Initialize USB
		ctx := gousb.NewContext()
		defer ctx.Close()

		// Open USB device
		device, err := ctx.OpenDeviceWithVIDPID(0x1234, 0x5678)
		if err != nil {
			log.Fatal("Error opening USB device:", err)
		}
		defer device.Close()

		// Set up USB endpoint
		config, err := device.Config(1)
		if err != nil {
			log.Fatal("Error getting device configuration:", err)
		}
		defer config.Close()

		intf, err := config.Interface(0, 0)
		if err != nil {
			log.Fatal("Error setting up interface:", err)
		}
		defer intf.Close()

		outEndpoint, err := intf.OutEndpoint(1)
		if err != nil {
			log.Fatal("Error setting up OUT endpoint:", err)
		}

		// Start IQ data loop
		ctxLoop, cancel := context.WithCancel(context.Background())
		go LoopIQData(ctxLoop, outEndpoint, data)

		// Wait for user to stop
		fmt.Println("Press Ctrl+C to stop...")
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		cancel()

	*/

}
