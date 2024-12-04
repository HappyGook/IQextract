package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/gousb"
	"io"
	"log"
	"mime/multipart"
	"os"

	"time"
)

var (
	usbContext     *gousb.Context
	usbDevice      *gousb.Device
	usbEndpoint    *gousb.OutEndpoint
	transferCtx    context.Context
	cancelTransfer context.CancelFunc
	iqData         []byte
)

const (
	AirspyVID = gousb.ID(0x1D50)
	AirspyPID = gousb.ID(0x60A1)
)

func extractHandler(c *gin.Context) {

	result, err := ExtractIQData("file.wav")
	iqData = result
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

func usbSetup() error {
	var err error

	usbContext = gousb.NewContext()
	usbDevice, err = usbContext.OpenDeviceWithVIDPID(AirspyVID, AirspyPID)
	if err != nil {
		return fmt.Errorf("error opening device: %v", err)
	}

	if usbDevice == nil {
		return fmt.Errorf("USB device not found")
	}

	err = usbDevice.SetAutoDetach(true)
	if err != nil {
		return err
	}

	config, err := usbDevice.Config(1)
	if err != nil {
		return fmt.Errorf("error configuring device: %v", err)
	}

	intf, err := config.Interface(0, 0) //REPLACE
	if err != nil {
		return fmt.Errorf("error opening interface: %v", err)
	}
	defer intf.Close()

	usbEndpoint, err = intf.OutEndpoint(0x02) //REPLACE
	if err != nil {
		return fmt.Errorf("error opening endpoint: %v", err)
	}

	return nil
}

func startHandler(c *gin.Context) {
	// Check if file exists
	if _, err := os.Stat("file.wav"); os.IsNotExist(err) {
		c.JSON(400, gin.H{"error": "No data file found. Please upload and extract data first."})
		return
	}

	data := iqData

	// Start the transfer loop
	transferCtx, cancelTransfer = context.WithCancel(context.Background())
	go func(ctx context.Context, data []byte) {
		for {
			select {
			case <-ctx.Done():
				log.Println("Data transfer stopped.")
				return
			default:
				if err := SendIQData(usbEndpoint, data); err != nil {
					log.Println("Error sending data:", err)
				}
				time.Sleep(100 * time.Millisecond) // Adjust as necessary
			}
		}
	}(transferCtx, data)

	c.JSON(200, gin.H{"message": "Data transfer started"})
}

func stopHandler(c *gin.Context) {
	if cancelTransfer != nil {
		cancelTransfer()
		cancelTransfer = nil
		c.JSON(200, gin.H{"message": "Data transfer stopped"})
	} else {
		c.JSON(400, gin.H{"error": "No data transfer in progress"})
	}
}

func main() {
	err := usbSetup()
	if err != nil {
		log.Fatalf("Error by setup: %v", err)
	}

	defer func(usbContext *gousb.Context) {
		err := usbContext.Close()
		if err != nil {
			return
		}
	}(usbContext)

	defer func(usbDevice *gousb.Device) {
		err := usbDevice.Close()
		if err != nil {
			return
		}
	}(usbDevice)

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
	router.Use(static.Serve("/", static.LocalFile("./frontend/build", true)))

	router.POST("/api/upload", uploadHandler)
	router.GET("/api/extractHandler", extractHandler)
	router.POST("/api/start", startHandler)
	router.POST("/api/stop", stopHandler)

	err = router.Run("localhost:8080")
	if err != nil {
		return
	}
}
