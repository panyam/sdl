# Tailwind CSS Setup

## Overview

This document describes the unified Tailwind CSS setup used across both server-rendered pages and client-side components in the SDL Canvas project.

## Configuration

### Tailwind Config (`web/tailwind.config.js`)

```javascript
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
```

### Key Points

1. **Unified Build**: Single CSS build serves both server and client pages
2. **Content Paths**: Includes both web components and console templates
3. **Dark Mode**: Class-based dark mode with `.dark` class on root element
4. **Custom Colors**: Extended gray palette for better dark mode support
5. **Monospace Font**: Consistent font stack for technical content

## Build Process

### Development
```bash
cd web
npm run dev  # Vite watches and rebuilds CSS automatically
```

### Production
```bash
cd web
npm run build  # Creates optimized CSS in dist/assets/
```

## Usage

### Server-Rendered Pages (Templar Templates)

```html
<!-- In base.html -->
<link rel="stylesheet" href="/assets/index-[hash].css">
```

### Client-Side Components

CSS is imported via `style.css` which includes Tailwind directives:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

## Theme Switching

### Implementation

1. **HTML Structure**: Theme switcher in `base.html` template
2. **JavaScript Handler**: `system-listing-handlers.ts` manages theme
3. **Storage**: Theme preference saved to localStorage
4. **System Detection**: Respects OS dark mode preference

### Theme Classes

- Light Mode: No class (default)
- Dark Mode: `dark` class on `<html>` element
- System Mode: Follows OS preference

## Component Styles

### Reusable Classes

Defined in `style.css` using `@layer components`:

```css
.panel {
  @apply bg-gray-800 border border-gray-700 rounded-lg p-4;
}

.btn-primary {
  @apply bg-blue-600 hover:bg-blue-700 text-white;
}
```

### Dark Mode Support

All components use dark mode variants:

```html
<div class="bg-white dark:bg-gray-800">
```

## Best Practices

1. **Consistency**: Use the same color palette across all pages
2. **Responsive**: Mobile-first approach with responsive modifiers
3. **Accessibility**: Sufficient color contrast in both themes
4. **Performance**: Purge unused styles in production build
5. **Maintainability**: Use semantic class names and components

## Migration from CDN

Previously used: `<script src="https://cdn.tailwindcss.com"></script>`

Benefits of build process:
- Smaller file size (only used styles)
- Better performance (no runtime compilation)
- Version control and consistency
- Custom configuration support
- Production optimizations

## Troubleshooting

### CSS Not Updating

1. Check Vite is running (`npm run dev`)
2. Verify content paths in config
3. Clear browser cache
4. Check for typos in class names

### Dark Mode Not Working

1. Ensure `darkMode: 'class'` in config
2. Check `dark` class is toggled on `<html>`
3. Verify dark variants in templates
4. Check JavaScript theme handler is loaded