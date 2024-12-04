import './App.css';
import React,{ useState } from "react"
import { useDropzone } from "react-dropzone";

function App() {
  const [selectedFile, setSelectedFile] = useState(null);
  const [uploadStatus, setUploadStatus] = useState("");
  const [result, setResult] = useState("");

  const onDrop = (acceptedFiles)=>{
      const file = acceptedFiles[0];

      if(file && (file.type==="audio/wav" || file.name.endsWith(".wav"))){
          setSelectedFile(file);
          setUploadStatus("");
      } else{
          setUploadStatus("Invalid File uploaded");
      }
  };



  const { getRootProps, getInputProps } = useDropzone({
        onDrop,
        accept: ".wav", // Only accept .wav files
        multiple: false, // Allow one file at a time
  });

  const callExtract = async () => {
        try {
            // Make the request to the backend API
            const response = await fetch("/api/extractHandler");

            if (!response.ok) {
                // If the response is not OK, display an error
                const errorData = await response.json();
                console.error("Error from backend:", errorData);
                setResult("Error: " + errorData.Error);  // Display the error message
                return;
            }

            // Parse the JSON response
            const data = await response.json();

            // Check if extracted data is available
            if (data.ExtractedData) {
                const decodedData = atob(data.ExtractedData);  // Decode the base64 string into byte data if needed
                setResult(decodedData);  // Display the decoded result (or further process it)
            } else {
                setResult("No extracted data available.");
            }
        } catch (error) {
            console.error("Error while fetching data:", error);
            setResult("Error fetching data.");
        }
  };

  const [isTransferRunning, setIsTransferRunning] = useState(false);

  const startTransfer = async () => {
        try {
            const response = await fetch("/api/start", { method: "POST" });

            if (response.ok) {
                setIsTransferRunning(true);  // Update the state to indicate transfer has started
            } else {
                const errorData = await response.json();
                console.error("Error starting transfer:", errorData);
            }
        } catch (error) {
            console.error("Error starting transfer:", error);
        }
  };

  const stopTransfer = async () => {
        try {
            const response = await fetch("/api/stop", { method: "POST" });

            if (response.ok) {
                setIsTransferRunning(false);  // Update the state to indicate transfer has stopped
            } else {
                const errorData = await response.json();
                console.error("Error stopping transfer:", errorData);
            }
        } catch (error) {
            console.error("Error while stopping transfer:", error);
        }
  };

  const upload = async ()=>{
      if(!selectedFile){
          setUploadStatus("No file selected");
          return;
      }
      const formData=new FormData();
      formData.append("file",selectedFile);

      try{
          const response = await fetch("/api/upload",{
              method: "POST",
              body: formData,
          })

          if (response.ok){
              setUploadStatus("Upload successful");
          } else{
              setUploadStatus(" Couldn't upload, try again");
          }
      } catch (error){
          console.error("Error uploading file:",error);
          setUploadStatus("An Error occurred, check the logs");
      }
  };



  return (
      <div style={{ textAlign: "center", margin: "50px" }}>
          <h1>Upload Your .wav File</h1>

          {/* Drag-and-drop area */}
          <div
              {...getRootProps()}
              style={{
                  border: "2px dashed #cccccc",
                  padding: "20px",
                  borderRadius: "10px",
                  cursor: "pointer",
                  margin: "20px auto",
                  maxWidth: "400px",
              }}
          >
              <input {...getInputProps()} />
              <p>Drag and drop your .wav file here, or click to browse.</p>
          </div>

          {/* Selected file info */}
          {selectedFile && <p>Selected file: {selectedFile.name}</p>}

          {/* Upload button */}
          <button onClick={upload} style={{ padding: "10px 20px", marginTop: "20px" }}>
              Upload File
          </button>

          {/* Status message */}
          {uploadStatus && <p>{uploadStatus}</p>}

          {/* Button to call extraction */}
          <button onClick={callExtract} style={{ padding: "10px 20px", marginTop: "20px" }}>
              Extract IQ Data
          </button>

          {/* Display extracted IQ data */}
          {result && (
              <div style={{
                  marginTop: "20px",
                  padding: "20px",
                  border: "1px solid #ccc",
                  borderRadius: "10px",
                  maxHeight: "300px",
                  overflowY: "auto",
                  whiteSpace: "pre-wrap",
                  background: "#f9f9f9"
              }}>
                  <h3>Extracted IQ Data:</h3>
                  <p>{result}</p>
              </div>
          )}
          {/* Button to start transfer */}
          <button
              onClick={startTransfer}
              style={{ padding: "10px 20px", marginTop: "20px" }}
              disabled={isTransferRunning}  // Disable button if transfer is running
          >
              Start Data Transfer
          </button>

          {/* Button to stop transfer */}
          <button
              onClick={stopTransfer}
              style={{ padding: "10px 20px", marginTop: "20px" }}
              disabled={!isTransferRunning}  // Disable button if transfer isn't running
          >
              Stop Data Transfer
          </button>
      </div>
  );
}

export default App;