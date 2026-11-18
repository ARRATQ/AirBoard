import { ref, computed } from 'vue'
import { directoryService } from '@/services/api'

// Couleurs d'avatar déterministes (quand une personne n'a pas de photo)
const AVATAR_COLORS = ['#2563EB', '#7C3AED', '#DB2777', '#EA580C', '#059669', '#0891B2', '#4F46E5', '#B45309']

export const FIELD_TYPES = ['text', 'textarea', 'email', 'phone', 'url', 'date', 'select']

export const FIELD_TYPE_ICONS = {
  text: 'mdi:form-textbox',
  textarea: 'mdi:text-long',
  email: 'mdi:email-outline',
  phone: 'mdi:phone-outline',
  url: 'mdi:link-variant',
  date: 'mdi:calendar',
  select: 'mdi:format-list-bulleted',
}

export function fullName(person) {
  if (!person) return ''
  return `${person.first_name || ''} ${person.last_name || ''}`.trim()
}

export function initials(person) {
  if (!person) return '?'
  const f = (person.first_name || '').trim().charAt(0)
  const l = (person.last_name || '').trim().charAt(0)
  return (f + l).toUpperCase() || '?'
}

export function avatarColor(person) {
  const seed = fullName(person) || String(person?.id || '')
  let hash = 0
  for (let i = 0; i < seed.length; i++) hash = (hash * 31 + seed.charCodeAt(i)) >>> 0
  return AVATAR_COLORS[hash % AVATAR_COLORS.length]
}

export function apiError(error, fallback) {
  return error?.response?.data?.error || error?.response?.data?.message || fallback
}

export function notify(type, message) {
  if (window.$toast && typeof window.$toast[type] === 'function') {
    window.$toast[type](message)
  } else if (type === 'error') {
    window.alert(message)
  }
}

// Formatage d'une valeur de champ dynamique pour l'affichage
export function formatFieldValue(field, value, locale) {
  if (!value) return ''
  if (field.type === 'date') {
    const d = new Date(`${value}T00:00:00`)
    return isNaN(d) ? value : d.toLocaleDateString(locale)
  }
  return value
}

export function fieldHref(field, value) {
  if (!value) return null
  if (field.type === 'email') return `mailto:${value}`
  if (field.type === 'phone') return `tel:${value.replace(/\s+/g, '')}`
  if (field.type === 'url' && /^https?:\/\//i.test(value)) return value
  return null
}

/**
 * État partagé d'une page de trombinoscope : chargement des données
 * et index dérivés (arbre d'unités, postes par personne, etc.).
 */
export function useDirectory() {
  const loading = ref(true)
  const error = ref('')
  const units = ref([])
  const people = ref([])
  const fields = ref([])
  const canManage = ref(false)
  const canAdmin = ref(false)

  const load = async () => {
    loading.value = true
    error.value = ''
    try {
      const data = await directoryService.getDirectory()
      units.value = data.units || []
      people.value = data.people || []
      fields.value = data.fields || []
      canManage.value = !!data.can_manage
      canAdmin.value = !!data.can_admin
    } catch (e) {
      error.value = apiError(e, 'Erreur de chargement')
    } finally {
      loading.value = false
    }
  }

  const peopleById = computed(() => new Map(people.value.map((p) => [p.id, p])))
  const unitsById = computed(() => new Map(units.value.map((u) => [u.id, u])))

  // Arbre : unités racines + enfants triés (ordre puis nom)
  const childrenOf = computed(() => {
    const map = new Map()
    const ids = new Set(units.value.map((u) => u.id))
    for (const u of units.value) {
      const key = u.parent_id && ids.has(u.parent_id) ? u.parent_id : null
      if (!map.has(key)) map.set(key, [])
      map.get(key).push(u)
    }
    for (const list of map.values()) {
      list.sort((a, b) => a.order - b.order || a.name.localeCompare(b.name))
    }
    return map
  })

  const rootUnits = computed(() => childrenOf.value.get(null) || [])

  // Liste à plat, en profondeur, avec niveau (pour les sélecteurs)
  const flatUnits = computed(() => {
    const out = []
    const walk = (list, depth) => {
      for (const u of list) {
        out.push({ ...u, depth })
        walk(childrenOf.value.get(u.id) || [], depth + 1)
      }
    }
    walk(rootUnits.value, 0)
    return out
  })

  // Postes occupés par chaque personne : [{ position, unit, assignment }]
  const rolesByPerson = computed(() => {
    const map = new Map()
    for (const unit of units.value) {
      for (const position of unit.positions || []) {
        for (const assignment of position.assignments || []) {
          if (!map.has(assignment.person_id)) map.set(assignment.person_id, [])
          map.get(assignment.person_id).push({ position, unit, assignment })
        }
      }
    }
    for (const list of map.values()) {
      list.sort((a, b) => Number(b.position.is_head) - Number(a.position.is_head))
    }
    return map
  })

  // Chemin d'une unité depuis la racine (fil d'Ariane)
  const unitPath = (unitId) => {
    const path = []
    let current = unitsById.value.get(unitId)
    let guard = 0
    while (current && guard++ < 50) {
      path.unshift(current)
      current = current.parent_id ? unitsById.value.get(current.parent_id) : null
    }
    return path
  }

  // Ensemble des unités d'un sous-arbre (pour le filtre « inclure les sous-unités »)
  const descendantIds = (unitId) => {
    const ids = new Set([unitId])
    const stack = [unitId]
    while (stack.length) {
      const id = stack.pop()
      for (const child of childrenOf.value.get(id) || []) {
        if (!ids.has(child.id)) {
          ids.add(child.id)
          stack.push(child.id)
        }
      }
    }
    return ids
  }

  const countPeopleInSubtree = (unitId) => {
    const ids = descendantIds(unitId)
    const persons = new Set()
    for (const u of units.value) {
      if (!ids.has(u.id)) continue
      for (const p of u.positions || []) for (const a of p.assignments || []) persons.add(a.person_id)
    }
    return persons.size
  }

  return {
    loading,
    error,
    units,
    people,
    fields,
    canManage,
    canAdmin,
    load,
    peopleById,
    unitsById,
    childrenOf,
    rootUnits,
    flatUnits,
    rolesByPerson,
    unitPath,
    descendantIds,
    countPeopleInSubtree,
  }
}
