package model

import (
	"bytes"
	"encoding/gob"
)

type EUTxO struct {
	TxID      string `json:"TxID"`
	TxIDIndex int    `json:"TxIDIndex"`
}

func (a *EUTxO) Serialize() ([]byte, error) {
	var serialized bytes.Buffer
	encoder := gob.NewEncoder(&serialized)
	err := encoder.Encode(a)
	if err != nil {
		return nil, err
	}
	return serialized.Bytes(), nil
}

func DeserializeEUTxOByte(data []byte) (*EUTxO, error) {
	var a EUTxO
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
