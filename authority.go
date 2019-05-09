// +build !js !wasm

package ot

import "errors"

type Authority struct {
	Document   string
	Operations [][]Applier
}

func (authority *Authority) Recieve(update Update) ([]Applier, error) {
	if update.Revision < 0 || update.Revision >= len(authority.Operations) {
		return nil, errors.New("revision not in history")
	}
	concurentOperations := authority.Operations[update.Revision:]
	operation := update.Operation
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
