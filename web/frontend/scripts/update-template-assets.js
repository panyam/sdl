#!/usr/bin/env node

/**
 * Post-build script to update the Go template with the correct asset hashes
 * generated by Vite build and discover available WASM modules
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

// Paths
const distIndexPath = path.join(__dirname, '../../dist/frontend/index.html');
const templatePath = path.join(__dirname, '../../templates/base.html');
const wasmDistPath = path.join(__dirname, '../../dist/wasm');
const wasmConfigPath = path.join(__dirname, '../../dist/wasm-modules.json');

// Read the dist index.html
const distHtml = fs.readFileSync(distIndexPath, 'utf-8');

// Extract the asset references using regex
const jsMatch = distHtml.match(/src="\/assets\/index-([^"]+)\.js"/);
const cssMatch = distHtml.match(/href="\/assets\/index-([^"]+)\.css"/);

if (!jsMatch || !cssMatch) {
  console.error('❌ Could not find asset references in dist/index.html');
  process.exit(1);
}

const jsFile = jsMatch[0].match(/\/assets\/index-[^"]+\.js/)[0];
const cssFile = cssMatch[0].match(/\/assets\/index-[^"]+\.css/)[0];

console.log(`📦 Found assets:`);
console.log(`   JS:  ${jsFile}`);
console.log(`   CSS: ${cssFile}`);

// Read the template
let template = fs.readFileSync(templatePath, 'utf-8');

// Replace the asset references
template = template.replace(
  /<link rel="stylesheet" href="\/assets\/index-[^"]+\.css">/,
  `<link rel="stylesheet" href="${cssFile}">`
);

template = template.replace(
  /<script type="module" src="\/assets\/index-[^"]+\.js"><\/script>/,
  `<script type="module" src="${jsFile}"></script>`
);

// Write the updated template
fs.writeFileSync(templatePath, template);

console.log(`✅ Updated ${templatePath} with new asset references`);

// Discover and catalog WASM modules
function discoverWasmModules() {
  if (!fs.existsSync(wasmDistPath)) {
    console.log('📦 No WASM directory found, skipping WASM module discovery');
    return [];
  }

  const wasmFiles = fs.readdirSync(wasmDistPath)
    .filter(file => file.endsWith('.wasm'))
    .map(file => {
      const name = path.basename(file, '.wasm');
      return {
        name,
        path: `/dist/wasm/${file}`,
        size: fs.statSync(path.join(wasmDistPath, file)).size
      };
    });

  return wasmFiles;
}

// Generate WASM modules catalog
const wasmModules = discoverWasmModules();

if (wasmModules.length > 0) {
  console.log(`📦 Discovered ${wasmModules.length} WASM modules:`);
  wasmModules.forEach(module => {
    const sizeMB = (module.size / (1024 * 1024)).toFixed(2);
    console.log(`   ${module.name}: ${module.path} (${sizeMB}MB)`);
  });

  // Write WASM modules catalog
  fs.writeFileSync(wasmConfigPath, JSON.stringify(wasmModules, null, 2));
  console.log(`✅ Generated WASM modules catalog: ${wasmConfigPath}`);
} else {
  console.log('📦 No WASM modules found');
}