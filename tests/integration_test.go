package main

import (
	"bytes"
	"os/exec"
	"testing"
)

func TestCmdWithNoMappingReturnsEmptyOutput(t *testing.T) {
	catCmd := exec.Command("cat", "../examples/groceries.csv")

	var outb, errb bytes.Buffer

	catCmd.Stdout = &outb
	catCmd.Stderr = &errb
	err := catCmd.Run()

	if err != nil {
		t.Fatalf("%s", err)
	}

	csvCmd := exec.Command("csv-tr")
	csvCmd.Stdin = &outb
	var outbCsv, errbCsv bytes.Buffer

	csvCmd.Stdout = &outbCsv
	csvCmd.Stderr = &errbCsv
	err = csvCmd.Run()

	if err != nil {
		t.Fatalf("%s", err)
	}

	// A new line for each row
	if outbCsv.String() != "\n\n\n" {
		t.Fatalf("If no mapping is provided output should be empty rows. Output: %s", outbCsv.String())
	}
}

func TestCmdWithMappingReturnsMappedValues(t *testing.T) {
	catCmd := exec.Command("cat", "../examples/groceries.csv")

	var outb, errb bytes.Buffer

	catCmd.Stdout = &outb
	catCmd.Stderr = &errb
	err := catCmd.Run()

	if err != nil {
		t.Fatalf("%s", err)
	}

	csvCmd := exec.Command("csv-tr", "1=0", "0=1")
	csvCmd.Stdin = &outb
	var outbCsv, errbCsv bytes.Buffer

	csvCmd.Stdout = &outbCsv
	csvCmd.Stderr = &errbCsv
	err = csvCmd.Run()

	if err != nil {
		t.Fatalf("%s", err)
	}

	expected := "oranges,7\nloaf of bread,1\npotatoes,4\n"
	if outbCsv.String() != expected {
		t.Fatalf("Columns did not map as expected output: %s.\nActual output: %s", expected, outbCsv.String())
	}
}

func TestCmdWithFilterOut(t *testing.T) {
	catCmd := exec.Command("cat", "../examples/groceries.csv")

	var outb, errb bytes.Buffer

	catCmd.Stdout = &outb
	catCmd.Stderr = &errb
	err := catCmd.Run()

	if err != nil {
		t.Fatalf("%s", err)
	}

	csvCmd := exec.Command("csv-tr", "1=0", "0=1", "-f", "1=potatoes")
	csvCmd.Stdin = &outb
	var outbCsv, errbCsv bytes.Buffer

	csvCmd.Stdout = &outbCsv
	csvCmd.Stderr = &errbCsv
	err = csvCmd.Run()

	if err != nil {
		t.Fatalf("%s", err)
	}

	expected := "oranges,7\nloaf of bread,1\n"
	if outbCsv.String() != expected {
		t.Fatalf("Columns did not map as expected output: %s.\nActual output: %s", expected, outbCsv.String())
	}
}
