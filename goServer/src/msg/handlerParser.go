package msg

import (
	"bytes"
	"fmt"
	"encoding/gob"
	"errors"
)

// parse send interfaces
func ParseSendInterfaces(msg *Message, data interface{}) error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	if err != nil {
		fmt.Println(err)
		return errors.New("unable to encode data")
	}
	
	msg.Data = buffer.Bytes()
	buffer.Reset()
		
	return nil
}


// parse receive interfaces
func ParseRcvInterfaces(msg *Message, realData interface{})(error) {
	buffer := bytes.NewBuffer(msg.Data)
	tmpdecoder := gob.NewDecoder(buffer)
	err := tmpdecoder.Decode(realData)	
	
	if err != nil {
		return errors.New("Unable to do conversion of data")
	}
	
	return nil
}