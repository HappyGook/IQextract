package main

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/gousb"
	"io"
	"log"
	"mime/multipart"
	"os"
	"errors"
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

const packetSize = 512

const (
	AirspyVID = gousb.ID(0x1D50)
	AirspyPID = gousb.ID(0x60A1)
)

// ExtractIQData WAV -> IQ
func ExtractIQData(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// Parse WAV header (assumes a fixed 44-byte header for simplicity)
	header := make([]byte, 44)
	_, err = file.Read(header)
	if err != nil {
		return nil, err
	}

	// Extract raw IQ data
	var data []int16
	for {
		var sample int16
		err := binary.Read(file, binary.LittleEndian, &sample)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		data = append(data, sample)
	}

	dataByte := make([]byte, len(data)*2)
	for i, sample := range data {
		binary.LittleEndian.PutUint16(dataByte[i*2:], uint16(sample))
	}

	return dataByte, nil
}

// SendIQData - IQ geschickt
func SendIQData(device *gousb.OutEndpoint, data []byte) error {
	const chunkSize=512
	log.Printf("Endpoint address: %d, Max packet size: %d", device.Desc.Address, device.Desc.MaxPacketSize)
	log.Printf("Sending total data size: %d bytes",len(data))

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data) // Last chunk may be smaller
		}

		chunk := data[i:end]
		log.Printf("Sending chunk: %d bytes (from %d to %d)", len(chunk), i, end)

		n, err := device.Write(chunk)
		if err != nil {
			log.Printf("Error sending chunk at index %d: %v", i/packetSize, err)
			if errors.Unwrap(err) != nil {
				log.Printf("Root cause of transfer error: %v", errors.Unwrap(err)) // Print the underlying error
			}
		}
		log.Printf("Successfully sent %d bytes", n)
	}


	return nil
}

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

func usbSetup() (*gousb.Device, *gousb.OutEndpoint, error) {
	
	usbContext := gousb.NewContext()
	usbDevice, err := usbContext.OpenDeviceWithVIDPID(AirspyVID, AirspyPID)
	if err != nil {
		usbContext.Close()
		return nil, nil, fmt.Errorf("error opening device: %v", err)
	}

	if usbDevice == nil {
		usbContext.Close()
		return nil, nil, fmt.Errorf("USB device not found")
	}

	err = usbDevice.SetAutoDetach(true)
	if err != nil {
		usbContext.Close()
		usbDevice.Close()
		return nil, nil, fmt.Errorf("error setting auto-detach: %v", err)
	}

	config, err := usbDevice.Config(1)
	if err != nil {
		usbContext.Close()
		usbDevice.Close()
		return nil, nil, fmt.Errorf("error configuring device: %v", err)
	}

	intf, err := config.Interface(0, 0) 
	if err != nil {
		usbDevice.Close()
		usbContext.Close()
		return nil, nil, fmt.Errorf("error opening interface: %v", err)
	}
	

	log. Println("Interface opened successfully")

	usbEndpoint, err = intf.OutEndpoint(0x02)
	if err != nil {
		intf.Close()
		usbContext.Close()
		usbDevice.Close()
		return nil, nil, fmt.Errorf("error opening endpoint: %v", err)
	}

	log.Println("USB endpoint successfully opened")
	return usbDevice, usbEndpoint, nil
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
				err := SendIQData(usbEndpoint, data)
				if err != nil {
					log.Println("Error sending data: %v", err)
					time.Sleep(500*time.Millisecond)
				}
				time.Sleep(200 * time.Millisecond) // Adjust as necessary
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
	device, endpoint, err := usbSetup()
	if err != nil {
		log.Fatalf("Error by setup: %v", err)
	}
	defer device.Close()

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

	err = router.Run("0.0.0.0:8080")
	if err != nil {
		return
	}
}
