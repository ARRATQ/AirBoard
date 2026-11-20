// Ton de StatusBadge (ok | warn | bad | info | neutral) pour chaque statut du modèle.

const EVENT_STATUS_TONES = {
  confirmed: 'ok',
  tentative: 'warn',
  cancelled: 'bad',
}

const SUGGESTION_STATUS_TONES = {
  new: 'info',
  reviewing: 'warn',
  planned: 'info',
  in_progress: 'warn',
  implemented: 'ok',
  rejected: 'bad',
}

export const eventStatusTone = (status) => EVENT_STATUS_TONES[status] || 'neutral'
export const suggestionStatusTone = (status) => SUGGESTION_STATUS_TONES[status] || 'neutral'
