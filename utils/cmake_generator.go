package utils

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/saferwall/pe"
)

type CmakeProjectGenerator struct {
	ProjectName         string
	ExportFunction      []pe.ExportFunction
	FunctionForwardList []uint32
}

func NewCmakeProjectGenerator(projectName string, exports []pe.ExportFunction, functionForwardList []uint32) *CmakeProjectGenerator {
	return &CmakeProjectGenerator{
		ProjectName:         projectName,
		ExportFunction:      exports,
		FunctionForwardList: functionForwardList,
	}
}

func (c *CmakeProjectGenerator) fixFunctionName(name string) string {
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "@", "_")
	return name
}

func (c *CmakeProjectGenerator) Generate(outputDir string) error {
	// create output directory
	if err := os.MkdirAll(outputDir+"/src", 0755); err != nil {
		return err
	}
	err := c.generateCmakelists(outputDir)
	if err != nil {
		return err
	}
	err = c.generateOriginalDefFile(outputDir)
	if err != nil {
		return err
	}
	err = c.generateForwardDefFile(outputDir)
	if err != nil {
		return err
	}
	err = c.generateSourceFile(outputDir)
	if err != nil {
		return err
	}
	err = c.generateVscodeConfig(outputDir)
	if err != nil {
		return err
	}
	return nil
}

func (c *CmakeProjectGenerator) generateCmakelists(outputDir string) error {
	str := "cmake_minimum_required(VERSION 3.10)\n"
	str += "project(" + c.ProjectName + ")\n"
	str += "add_library(" + c.ProjectName + " SHARED src/" + c.ProjectName + ".cpp)\n"
	str += "set(CMAKE_EXPORT_COMPILE_COMMANDS ON)\n"
	str += "set(CMAKE_CXX_FLAGS \"${CMAKE_CXX_FLAGS} /utf-8\")\n"
	// compile original def to lib
	str += "execute_process(COMMAND\n"
	str += "    ${CMAKE_AR} /def:" + c.ProjectName + ".original.def /out:${CMAKE_BINARY_DIR}/" + c.ProjectName + ".original.lib /machine:x64 ${CMAKE_STATIC_LINKER_FLAGS}\n"
	str += "    WORKING_DIRECTORY ${PROJECT_SOURCE_DIR}/src\n"
	str += "    COMMAND_ERROR_IS_FATAL ANY\n"
	str += ")\n"
	str += "set(" + c.ProjectName + "_DEF ${CMAKE_CURRENT_SOURCE_DIR}/src/" + c.ProjectName + ".def)\n"
	str += "if(EXISTS ${" + c.ProjectName + "_DEF})\n"
	str += "    if(MSVC)\n"
	str += "        set_target_properties(" + c.ProjectName + " PROPERTIES LINK_FLAGS \"/DEF:${" + c.ProjectName + "_DEF}\")\n"
	str += "    else()\n"
	str += "        target_link_options(" + c.ProjectName + " PRIVATE \"-Wl,--input-def=${" + c.ProjectName + "_DEF}\")\n"
	str += "    endif()\n"
	str += "else()\n"
	str += "    message(FATAL_ERROR \"" + c.ProjectName + ".def not found\")\n"
	str += "endif()\n"
	// link original lib
	str += "target_link_libraries(" + c.ProjectName + " PRIVATE ${CMAKE_BINARY_DIR}/" + c.ProjectName + ".original.lib)\n"
	// set output directory and remove prefix/suffix
	str += "set_target_properties(" + c.ProjectName + " PROPERTIES RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR})\n"
	str += "set_target_properties(" + c.ProjectName + " PROPERTIES PREFIX \"\" SUFFIX \".dll\")\n"

	// write CMakeLists.txt
	if err := os.WriteFile(outputDir+"/CMakeLists.txt", []byte(str), 0644); err != nil {
		return err
	}
	return nil
}

func (c *CmakeProjectGenerator) generateForwardDefFile(outputDir string) error {
	str := "LIBRARY \"" + c.ProjectName + ".dll\"\n"
	str += "EXPORTS\n"
	for _, exp := range c.ExportFunction {
		if exp.Name != "" {
			// if function is in the forward list, forward to original dll
			// else forward to our own function
			if slices.Contains(c.FunctionForwardList, exp.Ordinal) {
				name := c.fixFunctionName(exp.Name)
				str += fmt.Sprintf("    %s=forward_%s\n", exp.Name, name)
			} else {
				str += fmt.Sprintf("    %s=%s.original.dll.%s\n", exp.Name, c.ProjectName, exp.Name)
			}
		}
	}
	if err := os.WriteFile(outputDir+"/src/"+c.ProjectName+".def", []byte(str), 0644); err != nil {
		return err
	}
	return nil
}

func (c *CmakeProjectGenerator) generateOriginalDefFile(outputDir string) error {
	str := "LIBRARY \"" + c.ProjectName + ".dll\"\n"
	str += "EXPORTS\n"
	for _, exp := range c.ExportFunction {
		if exp.Name != "" {
			str += fmt.Sprintf("    %s\n", exp.Name)
		}
	}
	if err := os.WriteFile(outputDir+"/src/"+c.ProjectName+".original.def", []byte(str), 0644); err != nil {
		return err
	}
	return nil
}

func (c *CmakeProjectGenerator) generateSourceFile(outputDir string) error {
	str := "#include <windows.h>\n"
	str += "#include <string>\n\n"
	str += `static FARPROC GetOriginalFunction(const char* name, const char* func) {
    std::string original_name = std::string(name) + ".original.dll";
    HMODULE h = GetModuleHandleA(original_name.c_str());
    if (!h) {
        return nullptr;
    }
    return GetProcAddress(h, func);
}`
	str += "\n\n"
	if len(c.FunctionForwardList) > 0 {
		str += `extern "C" {`
		for _, exp := range c.ExportFunction {
			if exp.Name != "" && slices.Contains(c.FunctionForwardList, exp.Ordinal) {
				name := c.fixFunctionName(exp.Name)
				str += `
	__declspec(dllexport) void forward_` + name + `(void *m) {
        printf("[QQNT] Forwarding %s\n", __FUNCTION__);

        // Look up the symbol by name
        FARPROC proc = GetOriginalFunction("QQNT", "forward_` + name + `");
        if (!proc) {
            // symbol not found
            return;
        }

        // Example of forwarding a function using GetProcAddress
        typedef void (*ExampleFunction_t)(void *);

        // Call the original function
        ExampleFunction_t fn = (ExampleFunction_t)proc;
        fn(m);
    }`
			}
		}
		str += "}\n"
	}
	str += "BOOL APIENTRY DllMain(HMODULE hModule,\n"
	str += "                      DWORD  ul_reason_for_call,\n"
	str += "                      LPVOID lpReserved\n"
	str += ") {\n"
	str += "    switch (ul_reason_for_call) {\n"
	str += "    case DLL_PROCESS_ATTACH:\n"
	str += "        break;\n"
	str += "    case DLL_THREAD_ATTACH:\n"
	str += "        break;\n"
	str += "    case DLL_THREAD_DETACH:\n"
	str += "        break;\n"
	str += "    case DLL_PROCESS_DETACH:\n"
	str += "        break;\n"
	str += "    }\n"
	str += "    return TRUE;\n"
	str += "}\n"

	// write source file
	if err := os.WriteFile(outputDir+"/src/"+c.ProjectName+".cpp", []byte(str), 0644); err != nil {
		return err
	}
	return nil
}

func (c *CmakeProjectGenerator) generateVscodeConfig(outputDir string) error {
	str := "{\n"
	str += "    \"cmake.generator\": \"Ninja\"\n"
	str += "}\n"
	// create .vscode directory
	if err := os.MkdirAll(outputDir+"/.vscode", 0755); err != nil {
		return err
	}
	if err := os.WriteFile(outputDir+"/.vscode/settings.json", []byte(str), 0644); err != nil {
		return err
	}
	return nil
}
