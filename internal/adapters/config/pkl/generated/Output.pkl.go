// Code generated from Pkl module `searchbench.config.Experiment`. DO NOT EDIT.
package generated

import "github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/reportformat"

type Output struct {
	ReportFormat reportformat.ReportFormat `pkl:"reportFormat"`

	BundleRoot string `pkl:"bundleRoot"`

	Traces Tracing `pkl:"traces"`
}
