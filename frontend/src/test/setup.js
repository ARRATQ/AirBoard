// Mise en place commune des tests (jsdom).

// jsdom ne fournit pas ResizeObserver, utilisé par le Dialog de Headless UI.
if (typeof globalThis.ResizeObserver === 'undefined') {
  globalThis.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
}

// La directive globale v-focus-trap (enregistrée dans main.js) est aussi connue des montages de test.
import { config } from '@vue/test-utils'
import { focusTrap } from '@/directives/focusTrap'

config.global.directives = { ...config.global.directives, 'focus-trap': focusTrap }
