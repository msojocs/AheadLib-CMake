package main

import (
	"aheadlib/utils"
	"context"
	"fmt"
	"strings"

	"github.com/saferwall/pe"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) SelectInputFile() string {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择目标文件",
	})
	if err != nil {
		return ""
	}
	return file
}

func (a *App) SelectOutputDirectory() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择输出目录",
	})
	if err != nil {
		return ""
	}
	return dir
}

func (a *App) GetExportInfo(filePath string) []pe.ExportFunction {
	p := utils.NewParser(filePath)
	if err := p.Parse(); err != nil {
		fmt.Println("Failed to parse PE file:", err)
		return nil
	}
	defer p.Close()
	exports := p.GetExportInfo()
	return exports.Functions
}

func (a *App) GenerateCmakeProject(filePath string, outputDir string, functionForwardList []uint32) error {
	p := utils.NewParser(filePath)
	if err := p.Parse(); err != nil {
		fmt.Println("Failed to parse PE file:", err)
		return err
	}
	defer p.Close()
	exports := p.GetExportInfo()
	g := utils.NewCmakeProjectGenerator(strings.Replace(exports.Name, ".dll", "", -1), exports.Functions, functionForwardList)
	return g.Generate(outputDir)
}
