package main

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
)

func extractHandler(c *gin.Context) {

	result, err := ExtractIQData("file.wav")
	if err != nil {
		c.JSON(200, gin.H{"Error extracting IQ data:": err})
	}

	c.JSON(200, gin.H{"Extracted Data": result})

}

func uploadHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Couldn't reach the file"})
		return
	}
	if filepath.Ext(file.Filename) != ".wav" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not a wav uploaded"})
		return
	}
	tempFile, err := os.CreateTemp("", "file.wav")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(tempFile.Name())

	if err := c.SaveUploadedFile(file, tempFile.Name()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file."})
		return
	}
}

func main() {
	//router for request-handling
	var router *gin.Engine = gin.Default()
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
