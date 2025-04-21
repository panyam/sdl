// tailwind.config.js
/** @type {import('tailwindcss').Config} */
const defaultTheme = require('tailwindcss/defaultTheme')
const colors = require('tailwindcss/colors')
module.exports = {
  content: [
    "./static/js/*.js",
    "./static/images/*.svg",
    "./static/icons/*.svg",
    "./templates/*.html",
    "./views/templates/*.html",
    "./views/templates/*.css",
    "./views/components/*.ts",
    "./views/components/*.tsx"
  ],
  darkMode: 'class',
  theme: {
  },
  plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')],
}
