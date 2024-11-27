package main

import (
	"encoding/binary"
	"io"
	"os"
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

/*
// SendIQData IQ geschickt
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
*/
