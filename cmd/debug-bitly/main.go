package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/panyam/sdl/tools/systemdetail"
)

func main() {
	fmt.Println("🔍 Debug SystemDetailTool with Real Bitly System")
	fmt.Println("================================================")

	// Read the real Bitly SDL file
	bitlyPath := "/Users/sri/personal/golang/sdl/examples/bitly/mvp.sdl"
	sdlContent, err := ioutil.ReadFile(bitlyPath)
	if err != nil {
		log.Fatalf("Failed to read Bitly SDL: %v", err)
	}

	// Read the real Bitly recipe file
	recipePath := "/Users/sri/personal/golang/sdl/examples/bitly/mvp.recipe"
	recipeContent, err := ioutil.ReadFile(recipePath)
	if err != nil {
		log.Fatalf("Failed to read Bitly recipe: %v", err)
	}

	fmt.Printf("📄 SDL Content (%d chars):\n", len(sdlContent))
	fmt.Printf("📜 Recipe Content (%d chars):\n", len(recipeContent))

	// Create SystemDetailTool
	tool := systemdetail.NewSystemDetailTool()

	// Set up callbacks to capture output
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

	// Initialize with Bitly system
	fmt.Println("\n🚀 Initializing SystemDetailTool...")
	err = tool.Initialize("bitly", string(sdlContent), string(recipeContent))
	if err != nil {
		log.Fatalf("Initialize failed: %v", err)
	}

	// Try to set SDL content (this will trigger compilation)
	fmt.Println("\n🔧 Setting SDL content and compiling...")
	err = tool.SetSDLContent(string(sdlContent))
	if err != nil {
		fmt.Printf("❌ SDL compilation failed: %v\n", err)
		
		// Show compilation result details
		result := tool.GetCompileResult()
		if result != nil {
			fmt.Printf("   Compilation errors: %v\n", result.Errors)
		}
		
		// This is expected since we don't have @stdlib support yet
		fmt.Println("\n💡 This is expected - @stdlib imports not supported yet")
		os.Exit(0)
	}

	// If we get here, compilation succeeded
	fmt.Println("\n✅ SDL compiled successfully!")
	
	result := tool.GetCompileResult()
	if result != nil {
		fmt.Printf("   Found systems: %v\n", result.Systems)
	}

	// Get system info
	info := tool.GetSystemInfo()
	fmt.Printf("\n📊 System Info: %+v\n", info)

	// Try recipe parsing
	fmt.Println("\n📜 Setting recipe content...")
	err = tool.SetRecipeContent(string(recipeContent))
	if err != nil {
		fmt.Printf("❌ Recipe parsing failed: %v\n", err)
	} else {
		execState := tool.GetExecState()
		fmt.Printf("✅ Recipe parsed successfully: %d steps\n", execState.TotalSteps)
	}

	fmt.Println("\n🎉 Debug session completed!")
}