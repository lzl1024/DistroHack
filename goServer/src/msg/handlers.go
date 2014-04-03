package msg

import (
	"bytes"
	"encoding/gob"
	"errors"
)

var Handlers [NUMTYPES]Handler

type Sendhdlr func(*Message, interface{}) error
type Rcvhdlr func(*Message) (interface{}, error)

type Handler struct {
	Encode Sendhdlr
	Decode Rcvhdlr
}

func SendString(msg *Message, data interface{}) error {
	var buffer bytes.Buffer
	if msg.Kind == STRING {
		str, ok := data.(string)
		if !ok {
			return errors.New("data passed is not STRING")
		}
		encoder := gob.NewEncoder(&buffer)
		err := encoder.Encode(str)
		if err != nil {
			return errors.New("unable to encode STRING data")
		}
		msg.Data = buffer.Bytes()
		buffer.Reset()

		return nil
	}

	return errors.New("message Kind indicates not a STRING")
}

func RcvString(msg *Message) (interface{}, error) {
	var str string
	if msg.Kind == STRING {
		buffer := bytes.NewBuffer(msg.Data)
		tmpdecoder := gob.NewDecoder(buffer)
		err := tmpdecoder.Decode(&str)
		if err != nil {
			return nil, errors.New("Unable to do conversion of data")
		} else {
			return str, nil
		}
	}

	return nil, errors.New("message Kind indicates not a STRING")
}

func SendSnRank(msg *Message, data interface{}) error {
	var buffer bytes.Buffer

	if msg.Kind == SN_RANK {
		str, ok := data.([RankNum]Rank)
		if !ok {
			return errors.New("data passed is not SN_RANK")
		}
		encoder := gob.NewEncoder(&buffer)
		err := encoder.Encode(str)
		if err != nil {
			return errors.New("unable to encode SN_RANK data")
		}
		msg.Data = buffer.Bytes()
		buffer.Reset()

		return nil
	}

	return errors.New("message Kind indicates not a SN_RANK")
}

func RcvSnRank(msg *Message) (interface{}, error) {
	var rankList [RankNum]Rank
	if msg.Kind == SN_RANK {
		buffer := bytes.NewBuffer(msg.Data)
		tmpdecoder := gob.NewDecoder(buffer)
		err := tmpdecoder.Decode(&rankList)
		if err != nil {
			return nil, errors.New("Unable to do conversion of data")
		} else {
			return rankList, nil
		}
	}

	return nil, errors.New("message Kind indicates not a SN_RANK")
}

func SendSnOnSubmit(msg *Message, data interface{}) error {
	var buffer bytes.Buffer
	if msg.Kind == SN_RANK {
		str, ok := data.(string)
		if !ok {
			return errors.New("data passed is not SN_ON_SUBMIT")
		}
		encoder := gob.NewEncoder(&buffer)
		err := encoder.Encode(str)
		if err != nil {
			return errors.New("unable to encode SN_ON_SUBMIT data")
		}
		msg.Data = buffer.Bytes()
		buffer.Reset()

		return nil
	}

	return errors.New("message Kind indicates not a SN_ON_SUBMIT")
}

func RcvSnOnSubmit(msg *Message) (interface{}, error) {
	var str string
	if msg.Kind == SN_RANK {
		buffer := bytes.NewBuffer(msg.Data)
		tmpdecoder := gob.NewDecoder(buffer)
		err := tmpdecoder.Decode(&str)
		if err != nil {
			return nil, errors.New("Unable to do conversion of data")
		} else {
			return str, nil
		}
	}

	return nil, errors.New("message Kind indicates not a SN_ON_SUBMIT")
}

func SendSnOnSignIn(msg *Message, data interface{}) error {
	var buffer bytes.Buffer
	if msg.Kind == SN_ON_SIGNIN {
		str, ok := data.(string)
		if !ok {
			return errors.New("data passed is not SN_ON_SIGNIN")
		}
		encoder := gob.NewEncoder(&buffer)
		err := encoder.Encode(str)
		if err != nil {
			return errors.New("unable to encode SN_ON_SIGNIN data")
		}
		msg.Data = buffer.Bytes()
		buffer.Reset()

		return nil
	}

	return errors.New("message Kind indicates not a SN_ON_SIGNIN")
}

func RcvSnOnSignIn(msg *Message) (interface{}, error) {
	var str string
	if msg.Kind == SN_ON_SIGNIN {
		buffer := bytes.NewBuffer(msg.Data)
		tmpdecoder := gob.NewDecoder(buffer)
		err := tmpdecoder.Decode(&str)
		if err != nil {
			return nil, errors.New("Unable to do conversion of data")
		} else {
			return str, nil
		}
	}

	return nil, errors.New("message Kind indicates not a SN_ON_SIGNIN")
}
