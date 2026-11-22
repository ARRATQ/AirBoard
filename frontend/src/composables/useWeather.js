import { ref } from 'vue'

// Météo de l'accueil (open-meteo + nom de ville via nominatim), mise en cache dans le navigateur :
// revenir sur l'accueil ne relance ni la géolocalisation ni les deux appels externes.
const WEATHER_KEY = 'airboard_weather'
const CITY_KEY = 'airboard_weather_city'
export const WEATHER_TTL_MS = 30 * 60 * 1000

const EMPTY_WEATHER = { temp: null, icon: 'lucide:cloud', city: '' }

/** Code météo WMO → icône mdi. */
export function wmoIcon(code) {
  if (code === 0) return 'lucide:sun'
  if (code <= 2) return 'lucide:cloud-sun'
  if (code <= 3) return 'lucide:cloud'
  if (code <= 49) return 'lucide:cloud-fog'
  if (code <= 59) return 'lucide:cloud-rain'
  if (code <= 69) return 'lucide:cloud-hail'
  if (code <= 79) return 'lucide:cloud-snow'
  if (code <= 82) return 'lucide:cloud-rain-wind'
  if (code <= 84) return 'lucide:cloud-hail'
  if (code <= 99) return 'lucide:cloud-lightning'
  return 'lucide:cloud'
}

function readJSON(key) {
  try {
    return JSON.parse(localStorage.getItem(key) || 'null')
  } catch {
    return null
  }
}

function writeJSON(key, value) {
  try {
    localStorage.setItem(key, JSON.stringify(value))
  } catch {
    /* stockage plein ou interdit : la météo marche sans cache */
  }
}

/** Météo en cache si elle a moins de WEATHER_TTL_MS. */
export function readCachedWeather(now = Date.now()) {
  const cached = readJSON(WEATHER_KEY)
  if (!cached || typeof cached.at !== 'number' || now - cached.at > WEATHER_TTL_MS) return null
  return { temp: cached.temp, icon: cached.icon, city: cached.city }
}

/** Ville connue pour cette zone (~1 km) : nominatim n'est rappelé que si l'on se déplace. */
async function resolveCity(lat, lon) {
  const zone = `${lat.toFixed(2)},${lon.toFixed(2)}`
  const cached = readJSON(CITY_KEY)
  if (cached?.zone === zone) return cached.city
  const geo = await fetch(
    `https://nominatim.openstreetmap.org/reverse?lat=${lat}&lon=${lon}&format=json&accept-language=fr`,
  ).then((r) => r.json())
  const city = geo.address?.city || geo.address?.town || geo.address?.village || geo.address?.county || ''
  writeJSON(CITY_KEY, { zone, city })
  return city
}

async function fetchWeather({ latitude: lat, longitude: lon }) {
  const [meteo, city] = await Promise.all([
    fetch(
      `https://api.open-meteo.com/v1/forecast?latitude=${lat}&longitude=${lon}&current_weather=true&temperature_unit=celsius`,
    ).then((r) => r.json()),
    resolveCity(lat, lon),
  ])
  return {
    temp: Math.round(meteo.current_weather.temperature),
    icon: wmoIcon(meteo.current_weather.weathercode),
    city,
  }
}

export function useWeather() {
  const weather = ref({ ...EMPTY_WEATHER })

  const loadWeather = () => {
    const cached = readCachedWeather()
    if (cached) {
      weather.value = cached
      return
    }
    if (!navigator.geolocation) return
    navigator.geolocation.getCurrentPosition(
      async ({ coords }) => {
        try {
          const fresh = await fetchWeather(coords)
          weather.value = fresh
          writeJSON(WEATHER_KEY, { ...fresh, at: Date.now() })
        } catch {
          /* service météo indisponible : l'accueil s'affiche sans météo */
        }
      },
      () => {
        /* géolocalisation refusée : pas de météo */
      },
    )
  }

  return { weather, loadWeather }
}
