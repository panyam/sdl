package systemdetail

import (
	_ "embed"
	"slices"
	"testing"
)

// This is a copy of the examples/bitly/mvp.sdl
const bitlyMVPSdl = `
import Cache, HashIndex, NativeDisk, HttpStatusCode, delay, log from "@stdlib/common.sdl" ;

enum DBResult { FOUND, NOT_FOUND, INTERNAL_ERROR }
enum DuplicateCheck { FOUND, NOTFOUND, ERROR }

component AppServer {
  // Our main DB
  uses db Database
  param RetryCount = 3

  method Shorten() HttpStatusCode {
    for RetryCount {
      // 0.01% chance of a collision
      let foundDuplicate = sample dist {
          9999 => false
          1 => true
      }

      if not foundDuplicate {
        if db.Insert() {
          return HttpStatusCode.Ok
        } else {
          return HttpStatusCode.InternalError
        }
      }
    }

    // All retries elapsed - too many conflicts
    return HttpStatusCode.Conflict
  }

  method Redirect() HttpStatusCode {
    if self.db.Select() {
      return HttpStatusCode.Ok // or may be 302
    }

    // Simplified - can model more errors
    return HttpStatusCode.InternalError
  }
}

component Database {
    uses itemsById HashIndex()

    method Select() Bool {
      return itemsById.Find()
    }

    method Insert() Bool {
       return itemsById.Insert() 
    }
}

system Bitly {
    // Order of dependencies does not matter  They will be bound later
    // This allows cyclical links
    use app AppServer ( db = db )
    use db Database
}
`

func TestBityLoad(t *testing.T) {
	tool := NewSystemDetailTool()

	// Set up callback to capture output
	var lastError string
	var lastInfo string
	tool.SetCallbacks(&Callbacks{
		OnError: func(msg string) { lastError = msg },
		OnInfo:  func(msg string) { lastInfo = msg },
	})

	// Use variables to avoid unused warnings
	_ = lastError
	_ = lastInfo

	ensureSuccessfulLoad(t, tool, "Bitly", bitlyMVPSdl)

	// Now do some fun things.
}

func ensureSuccessfulLoad(t *testing.T, tool *SystemDetailTool, name string, sdl string) {
	err := tool.SetSDLContent(sdl)
	if err != nil {
		t.Fatalf("SetSDLContent failed: %v", err)
	}

	if tool.GetSDLContent() != sdl {
		t.Error("SDL content not stored correctly")
	}

	// Check that compilation was triggered
	result := tool.GetCompileResult()
	if result == nil {
		t.Fatal("Expected compile result after SetSDLContent")
	}

	if !result.Success {
		t.Errorf("Expected successful compilation, got errors: %v", result.Errors)
	}

	if len(result.Systems) == 0 {
		t.Error("Expected at least one system to be found")
	}

	// Should find the system
	found := slices.Contains(result.Systems, name)
	if !found {
		t.Errorf("Expected to find '%s' system, got: %v", name, result.Systems)
	}
}
