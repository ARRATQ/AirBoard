/* eslint-env node */
module.exports = {
  root: true,
  env: { browser: true, es2022: true, node: true },
  parserOptions: { ecmaVersion: 'latest', sourceType: 'module' },
  extends: ['eslint:recommended', 'plugin:vue/vue3-essential'],
  ignorePatterns: ['dist/', 'node_modules/'],
  rules: {
    // Les vues portent des noms de page d'un mot (Home, Login, Notifications…).
    'vue/multi-word-component-names': 'off',
    // Code mort signalé sans bloquer : à nettoyer au fil des chantiers.
    'no-unused-vars': 'warn',
    'vue/no-unused-vars': 'warn',
  },
}
