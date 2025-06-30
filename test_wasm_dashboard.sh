#!/bin/bash

echo "Opening SDL WASM Dashboard..."
echo "Make sure the server is running on localhost:8080"
echo ""
echo "Testing URL: http://localhost:8080/canvases/default/?wasm=true"
echo ""

# Open in default browser
open "http://localhost:8080/canvases/default/?wasm=true"

echo "Dashboard opened in browser!"
echo ""
echo "Expected features:"
echo "- File Explorer panel on the left"
echo "- Monaco editor for SDL code"
echo "- Console panel at the bottom"
echo "- System Architecture panel on the right"
echo "- Traffic Generation controls"
echo "- Toolbar with Load/Save/Run buttons"