package expire

import (
	"fmt"
)

type DryRunReporter interface {
	ReportAction(message string)
}

type ConsoleDryRunReporter struct{}

func (r ConsoleDryRunReporter) ReportAction(formatString string, arg ...interface{}) {
	fmt.Printf(formatString+"\n", arg)
}

var dryRunReporter = ConsoleDryRunReporter{}
