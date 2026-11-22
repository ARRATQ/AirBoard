import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { getRoles, getScope } from '@/config/access'

/**
 * Périmètre d'une page d'administration : tout le site (admin, permission « tout le site »)
 * ou les groupes de la personne.
 *
 * @param {string} access niveau de la page (voir config/access.js)
 * @returns {{
 *   roles: import('vue').ComputedRef<{ isAdmin: boolean, isGroupAdmin: boolean, permissions: Record<string, string> }>,
 *   scope: import('vue').ComputedRef<'global'|'group'>,
 *   isGlobal: import('vue').ComputedRef<boolean>,
 *   isAdmin: import('vue').ComputedRef<boolean>
 * }}
 */
export function useManagementScope(access) {
  const authStore = useAuthStore()
  const roles = computed(() => getRoles(authStore.user))
  const scope = computed(() => getScope(access, roles.value))
  return {
    roles,
    scope,
    isGlobal: computed(() => scope.value === 'global'),
    isAdmin: computed(() => roles.value.isAdmin)
  }
}
