package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func StripDefaults(fpath string, extraFieldsToRemove map[string]string) error {
	src, err := os.ReadFile(fpath)
	if err != nil {
		panic(err)
	}

	file, diags := hclwrite.ParseConfig(src, fpath, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		err := errors.New("an error occurred")
		if err != nil {
			return err
		}
	}

	hasChanges := false
	for _, block := range file.Body().Blocks() {
		if s := stripDefaultsFromBlock(block, extraFieldsToRemove); s {
			hasChanges = true
		}
	}
	if hasChanges {
		log.Printf("Updating file: %s\n", fpath)
		return os.WriteFile(fpath, file.Bytes(), 0644)
	}
	return nil
}

func stripDefaultsFromBlock(block *hclwrite.Block, extraFieldsToRemove map[string]string) bool {
	hasChanges := false
	for _, innblock := range block.Body().Blocks() {
		if s := stripDefaultsFromBlock(innblock, extraFieldsToRemove); s {
			hasChanges = true
		}
		if len(innblock.Body().Attributes()) == 0 && len(innblock.Body().Blocks()) == 0 {
			if rm := block.Body().RemoveBlock(innblock); rm {
				hasChanges = true
			}
		}
	}
	for name, attribute := range block.Body().Attributes() {
		if string(attribute.Expr().BuildTokens(nil).Bytes()) == " null" {
			if rm := block.Body().RemoveAttribute(name); rm != nil {
				hasChanges = true
			}
		}
		if string(attribute.Expr().BuildTokens(nil).Bytes()) == " {}" {
			if rm := block.Body().RemoveAttribute(name); rm != nil {
				hasChanges = true
			}
		}
		if string(attribute.Expr().BuildTokens(nil).Bytes()) == " []" {
			if rm := block.Body().RemoveAttribute(name); rm != nil {
				hasChanges = true
			}
		}
		for key, value := range extraFieldsToRemove {
			if name == key && string(attribute.Expr().BuildTokens(nil).Bytes()) == value {
				if rm := block.Body().RemoveAttribute(name); rm != nil {
					hasChanges = true
				}
			}
		}
	}
	return hasChanges
}

func AbstractDashboards(fpath string) error {
	path := filepath.Dir(fpath)
	outPath := filepath.Join(path, "files")

	src, err := os.ReadFile(fpath)
	if err != nil {
		panic(err)
	}

	file, diags := hclwrite.ParseConfig(src, fpath, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		err := errors.New("an error occurred")
		if err != nil {
			return err
		}
	}

	hasChanges := false
	dashboardJsons := map[string][]byte{}
	for _, block := range file.Body().Blocks() {
		labels := block.Labels()
		if len(labels) == 0 || labels[0] != "grafana_dashboard" {
			continue
		}

		dashboard, err := dashboardToJson(block)
		if err != nil {
			return err
		}

		if dashboard == nil {
			continue
		}

		writeTo := filepath.Join(outPath, fmt.Sprintf("%s.json", block.Labels()[1]))

		dashboardJsons[writeTo] = dashboard

		block.Body().SetAttributeRaw(
			"config_json",
			hclwrite.TokensForFunctionCall("file",
				hclwrite.TokensForValue(cty.StringVal(writeTo))))

		hasChanges = true
	}
	if hasChanges {
		log.Printf("Updating file: %s\n", fpath)
		os.Mkdir(outPath, 0755)
		for writeTo, dashboard := range dashboardJsons {
			err := os.WriteFile(writeTo, dashboard, 0644)
			if err != nil {
				panic(err)
			}
		}
		return os.WriteFile(fpath, file.Bytes(), 0644)
	}
	return nil
}

func dashboardToJson(block *hclwrite.Block) ([]byte, error) {
	s := string(block.Body().GetAttribute("config_json").Expr().BuildTokens(nil).Bytes())
	s = strings.TrimPrefix(s, " ")
	if !strings.HasPrefix(s, "\"") {
		// if expr is not a string, assume it's already converted, return (idempotency
		return nil, nil
	}
	s, err := strconv.Unquote(s)
	if err != nil {
		return nil, err
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal([]byte(s), &jsonMap)
	if err != nil {
		return nil, err
	}

	jsonMarshalled, err := json.MarshalIndent(jsonMap, "", "\t")
	if err != nil {
		return nil, err
	}

	return jsonMarshalled, nil

}
