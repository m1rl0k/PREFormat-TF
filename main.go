package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
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

	parser := hclparse.NewParser()
	f, diags := parser.ParseHCL(data, filename)
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
		tokens := tokensForExpr(attr.Expr, f)
		newBody.SetAttributeRaw(name, tokens)
	}

	for _, block := range content.Blocks {
		newBlock := newBody.AppendNewBlock(block.Type, block.Labels)
		blockContent, _ := block.Body.Content(&hcl.BodySchema{})
		for name, bAttr := range blockContent.Attributes {
			tokens := tokensForExpr(bAttr.Expr, f)
			newBlock.Body().SetAttributeRaw(name, tokens)
		}
		for _, subBlock := range blockContent.Blocks {
			writeBlock(subBlock, newBlock.Body())
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

func writeBlock(block *hcl.Block, body *hclwrite.Body) {
	newBlock := body.AppendNewBlock(block.Type, block.Labels)
	blockContent, _ := block.Body.Content(&hcl.BodySchema{})
	for name, bAttr := range blockContent.Attributes {
		tokens := tokensForExpr(bAttr.Expr, nil)
		newBlock.Body().SetAttributeRaw(name, tokens)
	}
	for _, subBlock := range blockContent.Blocks {
		writeBlock(subBlock, newBlock.Body())
	}
}

func tokensForAttr(attr *hclsyntax.Attribute, f *hcl.File) hclwrite.Tokens {
	var tokens hclwrite.Tokens

	tokens = append(tokens, hclwrite.Token{
		Type:  hclwrite.TokenIdent,
		Bytes: []byte(attr.Name),
	})

	return tokens
}

func tokensForExpr(expr hcl.Expression, f *hcl.File) hclwrite.Tokens {
	var tokens hclwrite.Tokens

	switch expr := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		tokens = append(tokens, hclwrite.Token{Type: hclwrite.TokenString, Bytes: []byte(expr.Val.AsBigFloat().String())})

	case *hclsyntax.TemplateExpr:
		parts := expr.Parts
		for _, part := range parts {
			switch part := part.(type) {
			case *hclsyntax.TemplateExpr:
				tokens = append(tokens, tokensForExpr(part, f)...)

			case *hclsyntax.TemplateWrapExpr:
				tokens = append(tokens, tokensForExpr(part.Wrapped, f)...)

			case *hclsyntax.LiteralValueExpr:
				tokens = append(tokens, hclwrite.Token{Type: hclwrite.TokenString, Bytes: []byte(part.Val.AsBigFloat().String())})
			}
		}

	case *hclsyntax.VariableExpr:
		tokens = append(tokens, hclwrite.Token{Type: hclwrite.TokenIdent, Bytes: []byte("$")})

		attrTokens := tokensForAttr(&hclsyntax.TraverseAttr{Name: expr.Name}, f)
		tokens = append(tokens, attrTokens...)
	// ... rest of the cases and default ...
	}

	return tokens
}


