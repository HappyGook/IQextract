package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/google/gousb"
)

// WAV -> IQ
func ExtractIQData(filename string) ([]int16, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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

	return data, nil
}

// IQ -> Bytes
func ConvertIQData(iqData []int16) []byte {
	data := make([]byte, len(iqData)*2)
	for i, sample := range iqData {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(sample))
	}
	return data
}

// IQ geschickt
func SendIQData(device *gousb.OutEndpoint, data []byte) error {
	_, err := device.Write(data)
	return err
}

// LoopIQData - geschickt in einer Schleife
func LoopIQData(ctx context.Context, device *gousb.OutEndpoint, data []byte) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Loop stopped.")
			return
		default:
			if err := SendIQData(device, data); err != nil {
				log.Println("Error sending data:", err)
			}
			time.Sleep(100 * time.Millisecond) // Adjust for your data rate
		}
	}
}

func main() {
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

}
