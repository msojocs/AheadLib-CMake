package utils_test

import (
	"aheadlib/utils"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	p := utils.NewParser("D:/Program Files/Tencent/QQNT/versions/9.9.21-39038/QQNT.dll")
	if err := p.Parse(); err != nil {
		t.Errorf("Failed to parse PE file: %v", err)
	}
	defer p.Close()
	exports := p.GetExportInfo()
	g := utils.NewCmakeProjectGenerator(strings.Replace(exports.Name, ".dll", "", -1), exports.Functions, nil)
	err := g.Generate("D:/temp/qqnt")
	if err != nil {
		t.Errorf("Failed to generate CMake project: %v", err)
	}
	t.Logf("Successfully parsed PE file and generated CMake project")
}
