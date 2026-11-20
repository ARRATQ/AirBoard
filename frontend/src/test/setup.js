// Mise en place commune des tests (jsdom).

// jsdom ne fournit pas ResizeObserver, utilisé par le Dialog de Headless UI.
if (typeof globalThis.ResizeObserver === 'undefined') {
  globalThis.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
}
