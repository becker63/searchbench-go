// Code generated from Pkl module `searchbench.config.Experiment`. DO NOT EDIT.
package provider

import (
	"encoding"
	"fmt"
)

type Provider string

const (
	Openai     Provider = "openai"
	Openrouter Provider = "openrouter"
	Cerebras   Provider = "cerebras"
	Fake       Provider = "fake"
)

// String returns the string representation of Provider
func (rcv Provider) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(Provider)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for Provider.
func (rcv *Provider) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "openai":
		*rcv = Openai
	case "openrouter":
		*rcv = Openrouter
	case "cerebras":
		*rcv = Cerebras
	case "fake":
		*rcv = Fake
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid Provider`, str)
	}
	return nil
}
