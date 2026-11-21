/**
 * Règle de pluriel arabe (CLDR) pour vue-i18n. Un message arabe à six formes s'écrit
 * « zéro | un | deux | 3 à 10 | 11 à 99 | 100 et au-delà », par exemple
 * « لا رسائل | رسالة واحدة | رسالتان | {count} رسائل | {count} رسالة | {count} رسالة ».
 * Les messages à deux ou trois formes gardent la règle par défaut de vue-i18n.
 *
 * @param {number} choice nombre à accorder
 * @param {number} choicesLength nombre de formes du message
 * @param {(choice: number, choicesLength: number) => number} defaultRule règle par défaut
 * @returns {number} index de la forme à afficher
 */
export function arabicPluralRule (choice, choicesLength, defaultRule) {
  if (choicesLength !== 6) return defaultRule(choice, choicesLength)
  const n = Math.abs(choice)
  if (n === 0) return 0
  if (n === 1) return 1
  if (n === 2) return 2
  const rest = n % 100
  if (rest >= 3 && rest <= 10) return 3
  if (rest >= 11 && rest <= 99) return 4
  return 5
}
