import { defineConfig } from 'vitest/config';
import { resolve } from 'path';

export default defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/__tests__/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'src/__tests__/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/mockServiceWorker.js',
        'dist/'
      ]
    },
    include: ['src/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}'],
    deps: {
      inline: ['monaco-editor']
    }
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
      '@proto': resolve(__dirname, './src/proto'),
      'monaco-editor': resolve(__dirname, './src/__tests__/mocks/monaco-editor.ts')
    }
  }
});