package msg

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
)

// parse send interfaces
func ParseSendInterfaces(msg *Message, data interface{}) error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	
	//print information, for debug:
	fmt.Println("Send msg: ", msg.String())
	fmt.Println("Send: ", data)
	fmt.Println();
	
	if err != nil {
		fmt.Println(err)
		return errors.New("HandlerParser: unable to encode data")
	}

	msg.Data = buffer.Bytes()
	buffer.Reset()

	return nil
}

// parse receive interfaces
func ParseRcvInterfaces(msg *Message, realData interface{}) error {
	buffer := bytes.NewBuffer(msg.Data)
	tmpdecoder := gob.NewDecoder(buffer)
	err := tmpdecoder.Decode(realData)
	
	if err != nil {
		fmt.Println(err)
		return errors.New("HandlerParser: Unable to do conversion of data")
	}

	return nil
}
