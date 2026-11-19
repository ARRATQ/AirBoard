// Son court joué à l'arrivée d'un message de chat.
// Le choix de l'utilisateur est lu à chaque appel (bascule dans le menu du chat).
export const CHAT_NOTIFICATIONS_KEY = 'chat_notifications'

export function chatNotificationsEnabled() {
  try {
    return localStorage.getItem(CHAT_NOTIFICATIONS_KEY) !== 'false'
  } catch {
    return true
  }
}

export async function playChatNotificationSound() {
  if (!chatNotificationsEnabled()) return
  try {
    const audioCtx = new (window.AudioContext || window.webkitAudioContext)()
    if (audioCtx.state === 'suspended') await audioCtx.resume()
    const oscillator = audioCtx.createOscillator()
    const gainNode = audioCtx.createGain()
    oscillator.connect(gainNode)
    gainNode.connect(audioCtx.destination)
    oscillator.type = 'sine'
    oscillator.frequency.setValueAtTime(880, audioCtx.currentTime)
    gainNode.gain.setValueAtTime(0, audioCtx.currentTime)
    gainNode.gain.linearRampToValueAtTime(0.15, audioCtx.currentTime + 0.01)
    gainNode.gain.exponentialRampToValueAtTime(0.01, audioCtx.currentTime + 0.4)
    oscillator.start()
    oscillator.stop(audioCtx.currentTime + 0.4)
    setTimeout(() => audioCtx.close(), 500)
  } catch (e) {
    console.warn('Audio notification error:', e)
  }
}
