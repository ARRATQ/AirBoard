import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { WEATHER_TTL_MS, readCachedWeather, useWeather, wmoIcon } from '../useWeather'

const coords = { latitude: 34.0209, longitude: -6.8416 }
const flush = () => new Promise((resolve) => setTimeout(resolve, 0))

function mockServices() {
  const fetchMock = vi.fn((url) =>
    Promise.resolve({
      json: () =>
        Promise.resolve(
          url.includes('open-meteo')
            ? { current_weather: { temperature: 21.4, weathercode: 0 } }
            : { address: { city: 'Rabat' } },
        ),
    }),
  )
  vi.stubGlobal('fetch', fetchMock)
  const getCurrentPosition = vi.fn((success) => success({ coords }))
  Object.defineProperty(navigator, 'geolocation', { value: { getCurrentPosition }, configurable: true })
  return { fetchMock, getCurrentPosition }
}

describe('useWeather', () => {
  beforeEach(() => localStorage.clear())
  afterEach(() => vi.unstubAllGlobals())

  it('maps WMO codes to icons', () => {
    expect(wmoIcon(0)).toBe('mdi:weather-sunny')
    expect(wmoIcon(61)).toBe('mdi:weather-snowy-rainy')
    expect(wmoIcon(120)).toBe('mdi:weather-cloudy')
  })

  it('fetches once, then serves the cache without geolocation or network', async () => {
    const { fetchMock, getCurrentPosition } = mockServices()
    const first = useWeather()
    first.loadWeather()
    await flush()
    expect(first.weather.value).toEqual({ temp: 21, icon: 'mdi:weather-sunny', city: 'Rabat' })
    expect(fetchMock).toHaveBeenCalledTimes(2)

    const second = useWeather()
    second.loadWeather()
    expect(second.weather.value.city).toBe('Rabat')
    expect(getCurrentPosition).toHaveBeenCalledTimes(1)
    expect(fetchMock).toHaveBeenCalledTimes(2)
  })

  it('expires the weather but keeps the city of the same area', async () => {
    const { fetchMock } = mockServices()
    useWeather().loadWeather()
    await flush()
    expect(readCachedWeather(Date.now() + WEATHER_TTL_MS + 1)).toBeNull()

    localStorage.removeItem('airboard_weather')
    useWeather().loadWeather()
    await flush()
    const urls = fetchMock.mock.calls.map(([url]) => url)
    expect(urls.filter((url) => url.includes('open-meteo'))).toHaveLength(2)
    expect(urls.filter((url) => url.includes('nominatim'))).toHaveLength(1)
  })

  it('stays empty when the weather service fails', async () => {
    mockServices()
    vi.stubGlobal('fetch', vi.fn(() => Promise.reject(new Error('offline'))))
    const { weather, loadWeather } = useWeather()
    loadWeather()
    await flush()
    expect(weather.value.temp).toBeNull()
    expect(localStorage.getItem('airboard_weather')).toBeNull()
  })
})
