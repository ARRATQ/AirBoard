import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { cpSync, existsSync, mkdirSync } from 'fs'

// Plugin to copy docs to public folder
function copyDocsPlugin() {
  return {
    name: 'copy-docs',
    buildStart() {
      const docsSource = resolve(__dirname, '../docs')
      const docsTarget = resolve(__dirname, 'public/docs')

      // Create target directory if it doesn't exist
      if (!existsSync(docsTarget)) {
        mkdirSync(docsTarget, { recursive: true })
      }

      // Copy docs folders (en, fr, es, ar)
      const locales = ['en', 'fr', 'es', 'ar']
      for (const locale of locales) {
        const sourceDir = resolve(docsSource, locale)
        const targetDir = resolve(docsTarget, locale)
        if (existsSync(sourceDir)) {
          try {
            cpSync(sourceDir, targetDir, { recursive: true })
          } catch (e) {
            console.warn(`Could not copy docs/${locale}:`, e.message)
          }
        }
      }
    }
  }
}

export default defineConfig({
  plugins: [vue(), copyDocsPlugin()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3000,
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
    },
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
    minify: 'esbuild',
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['vue', 'vue-router', 'pinia', 'vue-i18n'],
          ui: ['@headlessui/vue', '@heroicons/vue'],
          icons: ['@iconify/vue'],
        },
      },
    },
  },
  optimizeDeps: {
    include: ['vue', 'vue-router', 'pinia', '@iconify/vue', 'vue-i18n'],
  },
})