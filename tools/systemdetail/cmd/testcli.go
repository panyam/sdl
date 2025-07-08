package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/panyam/sdl/tools/shared/recipe"
	"github.com/panyam/sdl/tools/systemdetail"
)

func main() {
	fmt.Println("🧪 SystemDetailTool CLI Test")
	fmt.Println("============================")

	// Create tool
	tool := systemdetail.NewSystemDetailTool()

	// Set up callbacks
	tool.SetCallbacks(&systemdetail.Callbacks{
		OnError: func(msg string) {
			fmt.Printf("❌ ERROR: %s\n", msg)
		},
		OnInfo: func(msg string) {
			fmt.Printf("ℹ️  INFO: %s\n", msg)
		},
		OnSuccess: func(msg string) {
			fmt.Printf("✅ SUCCESS: %s\n", msg)
		},
	})

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "quit" || input == "exit" {
			break
		}

		switch {
		case input == "help":
			showHelp()
		case input == "load-bitly":
			loadBitlyExample(tool)
		case input == "load-simple":
			loadSimpleExample(tool)
		case input == "load-uber":
			loadUberExample(tool)
		case strings.HasPrefix(input, "use "):
			systemName := strings.TrimPrefix(input, "use ")
			useSystem(tool, systemName)
		case input == "info":
			showSystemInfo(tool)
		case input == "recipe":
			loadBitlyRecipe(tool)
		case input == "recipe-uber":
			loadUberRecipe(tool)
		case strings.HasPrefix(input, "recipe-test "):
			content := strings.TrimPrefix(input, "recipe-test ")
			testRecipeContent(content)
		case strings.HasPrefix(input, "sdl-test "):
			content := strings.TrimPrefix(input, "sdl-test ")
			testSDLContent(tool, content)
		case input == "status":
			showStatus(tool)
		case input == "examples":
			showExamples()
		case input == "clear":
			clearTool(tool)
		default:
			fmt.Printf("Unknown command: %s (type 'help' for commands)\n", input)
		}
	}
}

func showHelp() {
	fmt.Println(`
Commands:
  load-bitly      - Load Bitly SDL example with @stdlib imports
  load-simple     - Load simple SDL example without imports
  load-uber       - Load Uber SDL example with @stdlib imports
  use <system>    - Use/activate a compiled system
  info           - Show system information
  recipe         - Load Bitly recipe and parse it
  recipe-uber    - Load Uber recipe and parse it
  recipe-test <content> - Test recipe parsing with custom content
  sdl-test <content>    - Test SDL compilation with custom content
  status         - Show current tool status
  examples       - Show example SDL and recipe snippets
  clear          - Clear/reset the tool state
  help           - Show this help
  quit/exit      - Exit the CLI

Example commands:
  recipe-test echo "Hello World"
  recipe-test sdl load system.sdl
  sdl-test component Test { method Run() Bool { return true } }`)
}

func loadBitlyExample(tool *systemdetail.SystemDetailTool) {
	fmt.Println("📄 Loading Bitly SDL example...")

	sdlContent, err := os.ReadFile("examples/bitly/mvp.sdl")
	if err != nil {
		fmt.Printf("❌ Failed to read Bitly SDL: %v\n", err)
		return
	}

	err = tool.Initialize("bitly", string(sdlContent), "")
	if err != nil {
		fmt.Printf("❌ Initialize failed: %v\n", err)
		return
	}

	err = tool.SetSDLContent(string(sdlContent))
	if err != nil {
		fmt.Printf("❌ SDL compilation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Loaded Bitly SDL (%d chars)\n", len(sdlContent))
}

func loadUberExample(tool *systemdetail.SystemDetailTool) {
	fmt.Println("📄 Loading Uber SDL example...")

	sdlContent, err := os.ReadFile("examples/uber/mvp.sdl")
	if err != nil {
		fmt.Printf("❌ Failed to read Uber SDL: %v\n", err)
		return
	}

	err = tool.Initialize("uber", string(sdlContent), "")
	if err != nil {
		fmt.Printf("❌ Initialize failed: %v\n", err)
		return
	}

	err = tool.SetSDLContent(string(sdlContent))
	if err != nil {
		fmt.Printf("❌ SDL compilation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Loaded Uber SDL (%d chars)\n", len(sdlContent))
}

func loadSimpleExample(tool *systemdetail.SystemDetailTool) {
	fmt.Println("📄 Loading simple SDL example...")

	sdlContent := `
component WebServer {
  method HandleRequest() Bool {
    return true
  }
}

system SimpleWeb {
  use server WebServer
}
`

	err := tool.Initialize("simple", sdlContent, "")
	if err != nil {
		fmt.Printf("❌ Initialize failed: %v\n", err)
		return
	}

	err = tool.SetSDLContent(sdlContent)
	if err != nil {
		fmt.Printf("❌ SDL compilation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Loaded simple SDL\n")
}

func useSystem(tool *systemdetail.SystemDetailTool, systemName string) {
	fmt.Printf("🎯 Using system: %s\n", systemName)

	err := tool.UseSystem(systemName)
	if err != nil {
		fmt.Printf("❌ Failed to use system: %v\n", err)
		return
	}

	fmt.Printf("✅ Successfully activated system: %s\n", systemName)
}

func showSystemInfo(tool *systemdetail.SystemDetailTool) {
	fmt.Println("📊 System Information:")

	info := tool.GetSystemInfo()
	for key, value := range info {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

func loadBitlyRecipe(tool *systemdetail.SystemDetailTool) {
	fmt.Println("📜 Loading Bitly recipe...")

	recipeContent, err := os.ReadFile("examples/bitly/mvp.recipe")
	if err != nil {
		fmt.Printf("❌ Failed to read Bitly recipe: %v\n", err)
		return
	}

	err = tool.SetRecipeContent(string(recipeContent))
	if err != nil {
		fmt.Printf("❌ Recipe parsing failed: %v\n", err)
		return
	}

	execState := tool.GetExecState()
	fmt.Printf("✅ Recipe parsed successfully: %d executable steps\n", execState.TotalSteps)

	// Show first few steps
	fmt.Println("\nFirst 10 steps:")
	for i, step := range execState.Steps {
		if i >= 10 {
			break
		}
		fmt.Printf("  %d. Line %d: %s %v\n", i+1, step.LineNumber, step.Command, step.Args)
	}
	if len(execState.Steps) > 10 {
		fmt.Printf("  ... and %d more steps\n", len(execState.Steps)-10)
	}
}

func loadUberRecipe(tool *systemdetail.SystemDetailTool) {
	fmt.Println("📜 Loading Uber recipe...")

	recipeContent, err := os.ReadFile("examples/uber/mvp.recipe")
	if err != nil {
		fmt.Printf("❌ Failed to read Uber recipe: %v\n", err)
		return
	}

	err = tool.SetRecipeContent(string(recipeContent))
	if err != nil {
		fmt.Printf("❌ Recipe parsing failed: %v\n", err)
		return
	}

	execState := tool.GetExecState()
	fmt.Printf("✅ Recipe parsed successfully: %d executable steps\n", execState.TotalSteps)

	// Show first few steps
	fmt.Println("\nFirst 10 steps:")
	for i, step := range execState.Steps {
		if i >= 10 {
			break
		}
		fmt.Printf("  %d. Line %d: %s %v\n", i+1, step.LineNumber, step.Command, step.Args)
	}
	if len(execState.Steps) > 10 {
		fmt.Printf("  ... and %d more steps\n", len(execState.Steps)-10)
	}
}

func testRecipeContent(content string) {
	fmt.Printf("🧪 Testing recipe content: %q\n", content)

	result := recipe.ParseRecipe(content)

	fmt.Printf("Commands found: %d\n", len(result.Commands))
	fmt.Printf("Validation errors: %d\n", len(result.Errors))

	if len(result.Errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range result.Errors {
			fmt.Printf("  Line %d: %s\n", err.LineNumber, err.Message)
		}
	}

	if len(result.Commands) > 0 {
		fmt.Println("Commands:")
		for i, cmd := range result.Commands {
			fmt.Printf("  %d. Line %d [%s]: %s\n", i+1, cmd.LineNumber, cmd.Type, cmd.RawLine)
		}
	}
}

func testSDLContent(tool *systemdetail.SystemDetailTool, content string) {
	fmt.Printf("🔧 Testing SDL content: %q\n", content)

	// Create a new tool instance for testing
	testTool := systemdetail.NewSystemDetailTool()

	// Set up callbacks
	testTool.SetCallbacks(&systemdetail.Callbacks{
		OnError: func(msg string) {
			fmt.Printf("  ❌ ERROR: %s\n", msg)
		},
		OnInfo: func(msg string) {
			fmt.Printf("  ℹ️  INFO: %s\n", msg)
		},
		OnSuccess: func(msg string) {
			fmt.Printf("  ✅ SUCCESS: %s\n", msg)
		},
	})

	err := testTool.SetSDLContent(content)
	if err != nil {
		fmt.Printf("❌ SDL compilation failed: %v\n", err)
		return
	}

	info := testTool.GetSystemInfo()
	fmt.Printf("✅ SDL compiled successfully\n")
	fmt.Printf("  Systems found: %v\n", info["systems"])
}

func showStatus(tool *systemdetail.SystemDetailTool) {
	fmt.Println("📋 Tool Status:")
	fmt.Printf("  System ID: %s\n", tool.GetSystemID())
	fmt.Printf("  SDL Content: %d chars\n", len(tool.GetSDLContent()))
	fmt.Printf("  Recipe Content: %d chars\n", len(tool.GetRecipeContent()))

	result := tool.GetCompileResult()
	if result != nil {
		fmt.Printf("  Compilation: %v\n", result.Success)
		if result.Success {
			fmt.Printf("  Systems: %v\n", result.Systems)
		} else {
			fmt.Printf("  Errors: %v\n", result.Errors)
		}
	} else {
		fmt.Println("  Compilation: Not attempted")
	}

	execState := tool.GetExecState()
	if execState != nil {
		fmt.Printf("  Recipe Steps: %d\n", execState.TotalSteps)
	}
}

func showExamples() {
	fmt.Println(`
SDL Examples:
  Simple component:
    component WebServer { method Process() Bool { return true } }
  
  System with @stdlib:
    import HttpStatusCode from "@stdlib/common.sdl";
    component API { method Call() HttpStatusCode { return HttpStatusCode.Ok } }
    system MyAPI { use api API }
  
  With parameters:
    component Cache { param HitRate Float method Read() Bool { return true } }

Recipe Examples:
  Valid commands:
    echo "Starting process"
    sdl load system.sdl
    sdl use MySystem
    read
    sdl set component.param 10.5
    sdl gen add traffic api.Process 100
  
  Invalid commands (will show errors):
    ls -la
    echo $HOME
    cat file.txt | grep pattern
    if [ condition ]; then echo "test"; fi

Try:
  sdl-test component Test { method Run() Bool { return true } }
  recipe-test echo "Hello World"
  recipe-test invalid | command`)
}

func clearTool(_ *systemdetail.SystemDetailTool) {
	fmt.Println("🧹 Clearing tool state...")

	/*
		// Create a new tool (simulating reset)
		newTool := systemdetail.NewSystemDetailTool()

		// Copy the callbacks
		tool.SetCallbacks(&systemdetail.Callbacks{
			OnError: func(msg string) {
				fmt.Printf("❌ ERROR: %s\n", msg)
			},
			OnInfo: func(msg string) {
				fmt.Printf("ℹ️  INFO: %s\n", msg)
			},
			OnSuccess: func(msg string) {
				fmt.Printf("✅ SUCCESS: %s\n", msg)
			},
		})
	*/

	fmt.Println("✅ Tool state cleared")
}
