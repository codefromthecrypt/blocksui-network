package abi

import (
	ethgoAbi "github.com/umbracle/ethgo/abi"
)

type AbiIO struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type AbiMember struct {
	Name            string  `json:"name"`
	Inputs          []AbiIO `json:"inputs"`
	Outputs         []AbiIO `json:"outputs"`
	Constant        bool    `json:"constant"`
	StateMutability string  `json:"stateMutability"`
}

func ToAbiIOGroup(t *ethgoAbi.Type) []AbiIO {
	io := make([]AbiIO, 0)

	if t.Kind() == ethgoAbi.KindTuple {
		tes := t.TupleElems()
		for _, te := range tes {
			io = append(io, AbiIO{
				Name: te.Name,
				Type: te.Elem.String(),
			})
		}
	}

	// TODO: support more types

	return io
}

func MethodToMember(m *ethgoAbi.Method) AbiMember {
	return AbiMember{
		Inputs:          ToAbiIOGroup(m.Inputs),
		Name:            m.Name,
		Outputs:         ToAbiIOGroup(m.Outputs),
		StateMutability: "view",
	}
}
