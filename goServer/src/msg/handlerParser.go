package msg

import (
	"bytes"
	"encoding/gob"
	"errors"
)

func ParseSendInterfaces(msg *Message, data interface{}) error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	if err != nil {
		return errors.New("unable to encode data")
	}
	msg.Data = buffer.Bytes()
	buffer.Reset()
		
	return nil
}

// ordinary parse for string
func ParseRcvString(msg *Message)(string, error) {
	var str string
	buffer := bytes.NewBuffer(msg.Data)
	tmpdecoder := gob.NewDecoder(buffer)
	err := tmpdecoder.Decode(&str)
	if err != nil {
		return str, errors.New("Unable to do conversion of data")
	}

	return str, nil
}


func ParseRcvMapStrings(msg *Message)(map[string]string, error) {
	var mapstrings map[string]string
	buffer := bytes.NewBuffer(msg.Data)
	tmpdecoder := gob.NewDecoder(buffer)
	err := tmpdecoder.Decode(&mapstrings)
	if err != nil {
		return mapstrings, errors.New("Unable to do conversion of data")
	}
	
	return mapstrings, nil
}