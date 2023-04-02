package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/pmezard/go-difflib/difflib"
)

func main() {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading directory:", err)
		os.Exit(1)
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".tf") {
			processTerraformFile(f.Name())
		}
	}
}

func processTerraformFile(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", filename, err)
		return
	}

	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL(data, filename)
	if diags.HasErrors() {
		fmt.Println("Error parsing file:", filename, diags)
		return
	}

	formattedData := hclwrite.Format(data)

	if string(data) != string(formattedData) {
		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(data)),
			B:        difflib.SplitLines(string(formattedData)),
			FromFile: "Original",
			ToFile:   "Formatted",
			Context:  3,
		}
		diffStr, _ := difflib.GetUnifiedDiffString(diff)
		fmt.Printf("\033[33m%s\033[0m", diffStr)

		fmt.Printf("\033[33m\nProposed changes for %s:\033[0m\n", filename)
		fmt.Println("-----------------------------------")
		fmt.Printf("\033[31mOriginal:\033[0m\n\n%s\n", string(data))
		fmt.Printf("\033[32mFormatted:\033[0m\n\n%s\n", string(formattedData))
		fmt.Println("-----------------------------------")
	} else {
		fmt.Printf("\033[32mNo changes needed for %s\n\033[0m", filename)
	}
}
