// Ce que l'accueil met en avant : les derniers articles publiés et l'agenda qui vient
// (événements du jour puis à venir). Fonctions pures, sans modifier les listes reçues.

const publishedTime = (news) => new Date(news.published_at || news.created_at || 0).getTime()

/**
 * Les `count` articles publiés le plus récemment (épinglés ou non).
 * @param {object[] | undefined} news
 * @param {number} count
 * @returns {object[]}
 */
export function latestNews (news, count) {
  return [...(news || [])].sort((a, b) => publishedTime(b) - publishedTime(a)).slice(0, count)
}

/**
 * Événements du jour puis à venir, triés par début, sans doublon ; `isToday` marque ceux
 * du jour.
 * @param {object[] | undefined | null} today
 * @param {object[] | undefined | null} upcoming
 * @param {number} count
 * @returns {object[]}
 */
export function agendaAhead (today, upcoming, count) {
  const todayIds = new Set((today || []).map((e) => e.id))
  const seen = new Set()
  return [...(today || []), ...(upcoming || [])]
    .filter((e) => (seen.has(e.id) ? false : seen.add(e.id)))
    .map((e) => ({ ...e, isToday: todayIds.has(e.id) }))
    .sort((a, b) => new Date(a.start_date) - new Date(b.start_date))
    .slice(0, count)
}

/**
 * Type d'un article parmi les types connus de l'accueil : par `news_type_id`, sinon par le
 * slug historique `type`. `null` s'il est inconnu.
 * @param {{ news_type_id?: number, type?: string }} article
 * @param {Array<{ id: number, slug: string } | null | undefined>} types
 * @returns {object | null}
 */
export function newsTypeOf (article, types) {
  const known = (types || []).filter(Boolean)
  return (
    known.find((t) => article.news_type_id != null && t.id === article.news_type_id) ||
    known.find((t) => article.type && t.slug === article.type) ||
    null
  )
}
