package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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

	f, diags := hclsyntax.ParseConfig(data, filename, hcl.InitialPos)
	if diags.HasErrors() {
		fmt.Println("Error parsing file:", filename, diags)
		return
	}

	file := hclwrite.NewEmptyFile()
	newBody := file.Body()

	content, diags := f.Body.Content(&hcl.BodySchema{})
	if diags.HasErrors() {
		fmt.Println("Error decoding body:", filename, diags)
		return
	}

	for name, attr := range content.Attributes {
		tokens := tokensForExpr(attr.Expr)
		newBody.SetAttributeRaw(name, tokens)
	}

	for _, block := range content.Blocks {
		newBlock := newBody.AppendNewBlock(block.Type, block.Labels)
		blockContent, _ := block.Body.Content(&hcl.BodySchema{})
		for name, bAttr := range blockContent.Attributes {
			tokens := tokensForExpr(bAttr.Expr)
			newBlock.Body().SetAttributeRaw(name, tokens)
		}
	}

	formattedData := file.Bytes()

	if string(data) != string(formattedData) {
		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(data)),
			B:        difflib.SplitLines(string(formattedData)),
			FromFile: "Original",
			ToFile:   "Formatted",
			Context:  3,
		}
		diffStr, _ := difflib.GetUnifiedDiffString(diff)
		color.HiYellow(diffStr)

		fmt.Printf("\nProposed changes for %s:\n", filename)
		fmt.Println("-----------------------------------")
		fmt.Println(string(formattedData))
		fmt.Println("-----------------------------------")
	} else {
		fmt.Printf("No changes needed for %s\n", filename)
	}
}

func tokensForExpr(expr hcl.Expression) hclwrite.Tokens {
	return hclwrite.Tokens{
		{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte(expr.Range().Start.ByteSlice(data)),
			Pos: hclwrite.Pos{
				Filename: expr.SyntaxNode().Range().Filename,
				Offset:   int(expr.Range().Start.Byte),
			},
		},
	}
}
