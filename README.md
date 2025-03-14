## This code is serving a single (and a very specific) purpose of extracting the IQ information from a .wav file and sending it to a Raspberry Pi. 
### There are several preporational steps before actually extracting and sending: 
  First, the user uploads a wav file inside frontend, this file is temporarily saved for extraction and afterwards deleted. 
  The Raspberry Pi is also set up in a USBSetup function and made ready for transferring.

### This is how the extraction & transfer are made:
  The ExtractIQData function takes this wav file and parses its header. Then raw IQ data is extracted from it and converted into a byte array.
  The user can also see how the extracted IQ data array looks before actually sending it.
  
  After that, the array is gradually sent onto a Raspberry Pi in SendIQData function, handling possible errors along the way. 

  The program also features a simple React Frontend with one file field and a button.
