<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6">
    <div class="w-full">
      <!-- Header -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">{{ $t('groupAdmin.title') }}</h1>
        <p class="mt-1 text-gray-600 dark:text-gray-400">
          {{ $t('groupAdmin.subtitle') }}
        </p>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center items-center h-64">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>

      <!-- Content -->
      <div v-else>
        <!-- Stats Cards -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <!-- Managed Groups -->
          <div class="card">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-gray-600 dark:text-gray-400">{{ $t('groupAdmin.managedGroups') }}</p>
                <p class="text-3xl font-bold text-gray-900 dark:text-white mt-2">
                  {{ managedGroups.length }}
                </p>
              </div>
              <div class="bg-blue-100 dark:bg-blue-900 p-3 rounded-lg">
                <Icon icon="mdi:account-group" class="h-8 w-8 text-blue-600 dark:text-blue-400" />
              </div>
            </div>
          </div>

          <!-- AppGroups -->
          <div class="card">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-gray-600 dark:text-gray-400">{{ $t('groupAdmin.appGroups') }}</p>
                <p class="text-3xl font-bold text-gray-900 dark:text-white mt-2">
                  {{ stats.appGroups }}
                </p>
              </div>
              <div class="bg-green-100 dark:bg-green-900 p-3 rounded-lg">
                <Icon icon="mdi:folder-multiple" class="h-8 w-8 text-green-600 dark:text-green-400" />
              </div>
            </div>
          </div>

          <!-- Applications -->
          <div class="card">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-gray-600 dark:text-gray-400">{{ $t('groupAdmin.applications') }}</p>
                <p class="text-3xl font-bold text-gray-900 dark:text-white mt-2">
                  {{ stats.applications }}
                </p>
              </div>
              <div class="bg-purple-100 dark:bg-purple-900 p-3 rounded-lg">
                <Icon icon="mdi:apps" class="h-8 w-8 text-purple-600 dark:text-purple-400" />
              </div>
            </div>
          </div>

          <!-- News -->
          <div class="card">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-gray-600 dark:text-gray-400">{{ $t('groupAdmin.news') }}</p>
                <p class="text-3xl font-bold text-gray-900 dark:text-white mt-2">
                  {{ stats.news }}
                </p>
              </div>
              <div class="bg-orange-100 dark:bg-orange-900 p-3 rounded-lg">
                <Icon icon="mdi:newspaper-variant" class="h-8 w-8 text-orange-600 dark:text-orange-400" />
              </div>
            </div>
          </div>
        </div>

        <!-- Managed Groups List -->
        <div class="card">
          <div class="mb-6">
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">{{ $t('groupAdmin.yourGroups') }}</h2>
            <p class="text-sm text-gray-600 dark:text-gray-400 mt-1">
              {{ $t('groupAdmin.yourGroupsSubtitle') }}
            </p>
          </div>

          <div v-if="managedGroups.length === 0" class="text-center py-12">
            <Icon icon="mdi:account-group-outline" class="h-16 w-16 text-gray-400 mx-auto mb-4" />
            <p class="text-gray-600 dark:text-gray-400">{{ $t('groupAdmin.noManagedGroups') }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-500 mt-2">
              {{ $t('groupAdmin.noManagedGroupsHelp') }}
            </p>
          </div>

          <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div
              v-for="group in managedGroups"
              :key="group.id"
              class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:shadow-md transition-shadow"
            >
              <div class="flex items-start justify-between mb-3">
                <div>
                  <h3 class="font-semibold text-gray-900 dark:text-white">{{ group.name }}</h3>
                  <p v-if="group.description" class="text-sm text-gray-600 dark:text-gray-400 mt-1">
                    {{ group.description }}
                  </p>
                </div>
                <span
                  v-if="group.is_active"
                  class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
                >
                  {{ $t('common.active') }}
                </span>
              </div>

              <div class="space-y-2">
                <div class="flex items-center text-sm text-gray-600 dark:text-gray-400">
                  <Icon icon="mdi:account-multiple" class="h-4 w-4 mr-2" />
                  {{ group.users?.length || 0 }} {{ $t('common.users') }}
                </div>
                <div class="flex items-center text-sm text-gray-600 dark:text-gray-400">
                  <Icon icon="mdi:folder-multiple" class="h-4 w-4 mr-2" />
                  {{ group.app_groups?.length || 0 }} {{ $t('common.appGroups') }}
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Quick Actions -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mt-8">
          <router-link
            to="/group-admin/app-groups"
            class="card hover:shadow-lg transition-shadow cursor-pointer"
          >
            <div class="flex items-center">
              <div class="bg-blue-100 dark:bg-blue-900 p-4 rounded-lg mr-4">
                <Icon icon="mdi:folder-multiple" class="h-8 w-8 text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <h3 class="font-semibold text-gray-900 dark:text-white">{{ $t('groupAdmin.manageAppGroups') }}</h3>
                <p class="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  {{ $t('groupAdmin.manageAppGroupsDesc') }}
                </p>
              </div>
            </div>
          </router-link>

          <router-link
            to="/group-admin/applications"
            class="card hover:shadow-lg transition-shadow cursor-pointer"
          >
            <div class="flex items-center">
              <div class="bg-purple-100 dark:bg-purple-900 p-4 rounded-lg mr-4">
                <Icon icon="mdi:apps" class="h-8 w-8 text-purple-600 dark:text-purple-400" />
              </div>
              <div>
                <h3 class="font-semibold text-gray-900 dark:text-white">{{ $t('groupAdmin.manageApplications') }}</h3>
                <p class="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  {{ $t('groupAdmin.manageApplicationsDesc') }}
                </p>
              </div>
            </div>
          </router-link>

          <router-link
            to="/group-admin/news"
            class="card hover:shadow-lg transition-shadow cursor-pointer"
          >
            <div class="flex items-center">
              <div class="bg-orange-100 dark:bg-orange-900 p-4 rounded-lg mr-4">
                <Icon icon="mdi:newspaper-variant" class="h-8 w-8 text-orange-600 dark:text-orange-400" />
              </div>
              <div>
                <h3 class="font-semibold text-gray-900 dark:text-white">{{ $t('groupAdmin.manageNews') }}</h3>
                <p class="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  {{ $t('groupAdmin.manageNewsDesc') }}
                </p>
              </div>
            </div>
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { groupAdminService } from '@/services/api'

const loading = ref(true)
const managedGroups = ref([])
const stats = ref({
  appGroups: 0,
  applications: 0,
  news: 0
})

const fetchData = async () => {
  try {
    loading.value = true

    // Fetch managed groups
    const groupsData = await groupAdminService.getManagedGroups()
    managedGroups.value = groupsData

    // Fetch stats
    const [appGroupsData, applicationsData] = await Promise.all([
      groupAdminService.getAppGroups(),
      groupAdminService.getApplications()
    ])

    stats.value.appGroups = appGroupsData.total || appGroupsData.length || 0
    stats.value.applications = applicationsData.total || applicationsData.length || 0

    // Note: News count would require a separate endpoint
    // For now, we'll leave it at 0 or fetch it separately if needed
  } catch (error) {
    console.error('Error fetching group admin data:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>
