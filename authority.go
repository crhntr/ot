package ot

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Authority struct {
	Document   string
	Operations [][]Applier
}

func (authority *Authority) Recieve(message Message) ([]Applier, error) {
	if message.Revision < 0 || message.Revision >= len(authority.Operations) {
		return nil, errors.New("revision not in history")
	}
	concurentOperations := authority.Operations[message.Revision:]
	operation := message.Operation
	var err error
	for _, authorityOp := range concurentOperations {
		operation, _, err = Transform(operation, authorityOp)
		if err != nil {
			return nil, err
		}
	}
	nextDoc, err := Apply(authority.Document, operation...)
	if err != nil {
		return nil, err
	}
	authority.Document = nextDoc
	authority.Operations = append(authority.Operations, operation)
	return operation, nil
}

type Message struct {
	Revision  int       `json:"revision"`
	Operation []Applier `json:"operation"`
}

func (message *Message) UnmarshalJSON(data []byte) error {
	var obj struct {
		Revision   int           `json:"revision"`
		Opperation []interface{} `json:"operation"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	message.Revision = obj.Revision
	for _, op := range obj.Opperation {
		switch o := op.(type) {
		case string:
			message.Operation = append(message.Operation, Insert(o))
		case float64:
			if o < 0 {
				message.Operation = append(message.Operation, Delete(int(o)))
			} else {
				message.Operation = append(message.Operation, Retain(int(o)))
			}
		default:
			return fmt.Errorf("unknown op type %v %t", o, o)
		}
	}
	return nil
}
