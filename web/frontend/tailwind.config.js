/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
    "../console/templates/**/*.html",
    "../console/templates/*.html"
  ],
  darkMode: 'class',
  theme: {
    extend: {
      fontFamily: {
        'mono': ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'Liberation Mono', 'Courier New', 'monospace'],
      },
      colors: {
        gray: {
          850: '#1f2937',
          950: '#0f172a'
        }
      }
    },
  },
  plugins: [],
}