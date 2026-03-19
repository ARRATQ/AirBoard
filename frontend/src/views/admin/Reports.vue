<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-4 md:p-8">
    <div class="max-w-7xl mx-auto">
      <!-- Header -->
      <div class="flex flex-col md:flex-row md:items-center justify-between mb-8 gap-4">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
            <div class="p-2 bg-indigo-600 rounded-lg">
              <Icon icon="mdi:chart-timeline-variant" class="h-6 w-6 text-white" />
            </div>
            {{ $t('reports.title') }}
          </h1>
          <p class="mt-1 text-gray-500 dark:text-gray-400">{{ $t('reports.subtitle') }}</p>
        </div>

        <!-- Refresh button -->
        <button
          @click="loadCurrentTab"
          :disabled="isLoading"
          class="px-3 py-2 rounded-lg bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2 whitespace-nowrap"
        >
          <Icon :icon="isLoading ? 'mdi:loading' : 'mdi:refresh'" :class="['h-4 w-4', isLoading && 'animate-spin']" />
          {{ isLoading ? $t('common.loading') : $t('common.refresh') }}
        </button>

        <!-- Period selector -->
        <div class="flex items-center gap-2 flex-wrap">
          <button
            v-for="preset in periodPresets"
            :key="preset.value"
            @click="selectPreset(preset.value)"
            :class="[
              'px-3 py-1.5 text-sm rounded-lg font-medium transition-colors',
              activePeriod === preset.value
                ? 'bg-indigo-600 text-white'
                : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'
            ]"
          >
            {{ preset.label }}
          </button>
          <!-- Custom range -->
          <div class="flex items-center gap-1">
            <input
              type="date"
              v-model="customFrom"
              @change="applyCustomRange"
              class="text-sm border border-gray-200 dark:border-gray-700 rounded-lg px-2 py-1.5 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300"
            />
            <span class="text-gray-400">→</span>
            <input
              type="date"
              v-model="customTo"
              @change="applyCustomRange"
              class="text-sm border border-gray-200 dark:border-gray-700 rounded-lg px-2 py-1.5 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300"
            />
          </div>
        </div>
      </div>

      <!-- Tabs -->
      <div class="flex gap-1 mb-6 border-b border-gray-200 dark:border-gray-700">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          :class="[
            'px-5 py-2.5 text-sm font-medium transition-colors flex items-center gap-2',
            activeTab === tab.key
              ? 'text-indigo-600 border-b-2 border-indigo-600'
              : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
          ]"
        >
          <Icon :icon="tab.icon" class="h-4 w-4" />
          {{ tab.label }}
        </button>
      </div>

      <!-- Loading -->
      <div v-if="isLoading" class="flex flex-col items-center justify-center py-24">
        <Icon icon="mdi:loading" class="h-12 w-12 animate-spin text-indigo-500" />
        <p class="mt-4 text-gray-500 animate-pulse">{{ $t('common.loading') }}...</p>
      </div>

      <!-- Error state -->
      <div v-else-if="loadError" class="flex flex-col items-center justify-center py-24 text-center">
        <Icon icon="mdi:alert-circle-outline" class="h-12 w-12 text-red-400 mb-3" />
        <p class="text-gray-600 dark:text-gray-400 font-medium">Impossible de charger les données</p>
        <p class="text-sm text-gray-400 mt-1">{{ loadError }}</p>
        <button @click="loadCurrentTab" class="mt-4 px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm hover:bg-indigo-700">
          Réessayer
        </button>
      </div>

      <!-- ======================== -->
      <!-- TAB: PAR RÔLE           -->
      <!-- ======================== -->
      <div v-else-if="activeTab === 'roles' && roleData">

        <!-- Role KPI cards -->
        <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-5 mb-8">
          <div
            v-for="stat in roleData.role_stats"
            :key="stat.role"
            class="bg-white dark:bg-gray-800 rounded-xl p-5 shadow-sm border border-gray-100 dark:border-gray-700 hover:shadow-md transition-shadow"
          >
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-2">
                <div :class="['p-2 rounded-lg', roleColors[stat.role]?.bg]">
                  <Icon :icon="roleIcons[stat.role]" :class="['h-5 w-5', roleColors[stat.role]?.text]" />
                </div>
                <div>
                  <span class="font-semibold text-gray-700 dark:text-gray-300 capitalize">{{ $t('reports.roles.' + stat.role) }}</span>
                  <p v-if="stat.role === 'group_admin'" class="text-xs text-gray-400 mt-0.5">{{ $t('reports.groupAdminNote') }}</p>
                </div>
              </div>
              <span class="text-xs text-gray-400">{{ stat.member_count }} {{ $t('reports.members') }}</span>
            </div>

            <!-- Active rate -->
            <div class="mb-3">
              <div class="flex justify-between text-xs text-gray-500 dark:text-gray-400 mb-1">
                <span>{{ $t('reports.activeMembers') }}</span>
                <span class="font-medium">{{ stat.active_members }}/{{ stat.member_count }}</span>
              </div>
              <div class="w-full bg-gray-100 dark:bg-gray-700 h-1.5 rounded-full overflow-hidden">
                <div
                  :class="['h-full rounded-full transition-all', roleColors[stat.role]?.bar]"
                  :style="{ width: stat.member_count > 0 ? (stat.active_members / stat.member_count * 100) + '%' : '0%' }"
                ></div>
              </div>
            </div>

            <!-- Consumption metrics -->
            <div class="grid grid-cols-2 gap-2 text-xs">
              <div class="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-2">
                <p class="text-gray-400 mb-0.5">{{ $t('reports.appClicks') }}</p>
                <p class="font-bold text-gray-800 dark:text-white">{{ stat.app_clicks.toLocaleString() }}</p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-2">
                <p class="text-gray-400 mb-0.5">{{ $t('reports.newsRead') }}</p>
                <p class="font-bold text-gray-800 dark:text-white">{{ stat.news_read.toLocaleString() }}</p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-2">
                <p class="text-gray-400 mb-0.5">{{ $t('reports.reactions') }}</p>
                <p class="font-bold text-gray-800 dark:text-white">{{ stat.reactions_given.toLocaleString() }}</p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-2">
                <p class="text-gray-400 mb-0.5">{{ $t('reports.pollVotes') }}</p>
                <p class="font-bold text-gray-800 dark:text-white">{{ stat.poll_votes.toLocaleString() }}</p>
              </div>
            </div>

            <!-- Production metrics (editors/admins) -->
            <div v-if="['admin', 'editor', 'group_admin'].includes(stat.role)" class="mt-3 pt-3 border-t border-gray-100 dark:border-gray-700">
              <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2 uppercase tracking-wide">{{ $t('reports.production') }}</p>
              <div class="grid grid-cols-2 gap-1 text-xs mb-1">
                <div class="text-center">
                  <p class="font-bold text-gray-800 dark:text-white text-base">{{ stat.articles_published }}</p>
                  <p class="text-gray-400">{{ $t('reports.articles') }}</p>
                </div>
                <div class="text-center">
                  <p class="font-bold text-amber-600 dark:text-amber-400 text-base">
                    {{ stat.apps_created }}<span class="text-gray-400 font-normal text-xs"> / {{ stat.apps_created_total }}</span>
                  </p>
                  <p class="text-gray-400">{{ $t('reports.appsCreated') }}</p>
                </div>
              </div>
              <div class="grid grid-cols-2 gap-1 text-xs">
                <div class="text-center">
                  <p class="font-bold text-gray-800 dark:text-white text-base">{{ stat.total_views_generated.toLocaleString() }}</p>
                  <p class="text-gray-400">{{ $t('reports.views') }}</p>
                </div>
                <div class="text-center">
                  <p class="font-bold text-gray-800 dark:text-white text-base">{{ stat.total_reactions_earned }}</p>
                  <p class="text-gray-400">{{ $t('reports.earned') }}</p>
                </div>
              </div>
            </div>

            <!-- Top contributors -->
            <div v-if="stat.top_contributors && stat.top_contributors.length" class="mt-3 pt-3 border-t border-gray-100 dark:border-gray-700">
              <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2 uppercase tracking-wide">Top</p>
              <div class="space-y-1">
                <div v-for="(u, idx) in stat.top_contributors.slice(0, 3)" :key="u.user_id" class="flex items-center justify-between text-xs">
                  <div class="flex items-center gap-1.5">
                    <span class="text-gray-400">{{ idx + 1 }}.</span>
                    <span class="text-gray-700 dark:text-gray-300 truncate max-w-[100px]">{{ u.first_name || u.username }}</span>
                  </div>
                  <span class="font-semibold text-gray-600 dark:text-gray-400">{{ u.score }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Comparison table -->
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 overflow-hidden">
          <div class="px-6 py-4 border-b border-gray-100 dark:border-gray-700">
            <h2 class="font-semibold text-gray-900 dark:text-white">{{ $t('reports.comparisonTable') }}</h2>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead class="bg-gray-50 dark:bg-gray-700/50">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ $t('reports.role') }}</th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ $t('reports.members') }}</th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ $t('reports.activeRate') }}</th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ $t('reports.appClicks') }}</th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ $t('reports.newsRead') }}</th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ $t('reports.appsCreated') }}</th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ $t('reports.articles') }}</th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ $t('reports.views') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
                <tr v-for="stat in roleData.role_stats" :key="stat.role" class="hover:bg-gray-50 dark:hover:bg-gray-700/30">
                  <td class="px-6 py-3 font-medium text-gray-800 dark:text-gray-200 capitalize">
                    <div class="flex items-center gap-2">
                      <Icon :icon="roleIcons[stat.role]" :class="['h-4 w-4', roleColors[stat.role]?.text]" />
                      {{ $t('reports.roles.' + stat.role) }}
                    </div>
                  </td>
                  <td class="px-4 py-3 text-right text-gray-600 dark:text-gray-400">{{ stat.member_count }}</td>
                  <td class="px-4 py-3 text-right">
                    <span :class="['font-medium', stat.member_count > 0 && stat.active_members/stat.member_count >= 0.5 ? 'text-green-600' : 'text-orange-500']">
                      {{ stat.member_count > 0 ? Math.round(stat.active_members / stat.member_count * 100) : 0 }}%
                    </span>
                  </td>
                  <td class="px-4 py-3 text-right text-gray-600 dark:text-gray-400">{{ stat.app_clicks.toLocaleString() }}</td>
                  <td class="px-4 py-3 text-right text-gray-600 dark:text-gray-400">{{ stat.news_read.toLocaleString() }}</td>
                  <td class="px-4 py-3 text-right text-amber-600 dark:text-amber-400 font-medium">
                    {{ stat.apps_created }}<span class="text-gray-400 font-normal text-xs"> / {{ stat.apps_created_total }}</span>
                  </td>
                  <td class="px-4 py-3 text-right text-gray-600 dark:text-gray-400">{{ stat.articles_published }}</td>
                  <td class="px-4 py-3 text-right text-gray-600 dark:text-gray-400">{{ stat.total_views_generated.toLocaleString() }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- ======================== -->
      <!-- TAB: PAR GROUPE         -->
      <!-- ======================== -->
      <div v-else-if="activeTab === 'groups' && groupData">

        <!-- Recap table -->
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 overflow-hidden mb-6">
          <div class="px-6 py-4 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
            <h2 class="font-semibold text-gray-900 dark:text-white">{{ $t('reports.comparisonTable') }}</h2>
            <span class="text-xs text-gray-400">{{ sortedGroups.length }} {{ $t('reports.groups') }}</span>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead class="bg-gray-50 dark:bg-gray-700/50">
                <tr>
                  <th class="px-5 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ $t('reports.group') }}</th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 select-none" @click="sortGroupBy('member_count')">
                    {{ $t('reports.members') }} <Icon :icon="groupSortField === 'member_count' ? (groupSortDesc ? 'mdi:arrow-down' : 'mdi:arrow-up') : 'mdi:unfold-more-horizontal'" class="h-3 w-3 inline" />
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 select-none" @click="sortGroupBy('engagement_rate')">
                    {{ $t('reports.activeRate') }} <Icon :icon="groupSortField === 'engagement_rate' ? (groupSortDesc ? 'mdi:arrow-down' : 'mdi:arrow-up') : 'mdi:unfold-more-horizontal'" class="h-3 w-3 inline" />
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 select-none" @click="sortGroupBy('app_clicks')">
                    {{ $t('reports.appClicks') }} <Icon :icon="groupSortField === 'app_clicks' ? (groupSortDesc ? 'mdi:arrow-down' : 'mdi:arrow-up') : 'mdi:unfold-more-horizontal'" class="h-3 w-3 inline" />
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 select-none" @click="sortGroupBy('news_read')">
                    {{ $t('reports.newsRead') }} <Icon :icon="groupSortField === 'news_read' ? (groupSortDesc ? 'mdi:arrow-down' : 'mdi:arrow-up') : 'mdi:unfold-more-horizontal'" class="h-3 w-3 inline" />
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 select-none" @click="sortGroupBy('reactions_given')">
                    {{ $t('reports.reactions') }} <Icon :icon="groupSortField === 'reactions_given' ? (groupSortDesc ? 'mdi:arrow-down' : 'mdi:arrow-up') : 'mdi:unfold-more-horizontal'" class="h-3 w-3 inline" />
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 select-none" @click="sortGroupBy('poll_votes')">
                    {{ $t('reports.pollVotes') }} <Icon :icon="groupSortField === 'poll_votes' ? (groupSortDesc ? 'mdi:arrow-down' : 'mdi:arrow-up') : 'mdi:unfold-more-horizontal'" class="h-3 w-3 inline" />
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-amber-500 uppercase tracking-wider cursor-pointer hover:text-amber-600 select-none" @click="sortGroupBy('apps_created')">
                    {{ $t('reports.appsCreated') }} <Icon :icon="groupSortField === 'apps_created' ? (groupSortDesc ? 'mdi:arrow-down' : 'mdi:arrow-up') : 'mdi:unfold-more-horizontal'" class="h-3 w-3 inline" />
                  </th>
                  <th class="px-4 py-3 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ $t('reports.topApps') }}</th>
                  <th class="px-4 py-3"></th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
                <tr
                  v-for="group in sortedGroups"
                  :key="group.group_id"
                  class="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                >
                  <td class="px-5 py-3">
                    <div class="flex items-center gap-2.5">
                      <div class="w-3 h-3 rounded-full flex-shrink-0" :style="{ backgroundColor: group.group_color || '#6B7280' }"></div>
                      <span class="font-medium text-gray-800 dark:text-white">{{ group.group_name }}</span>
                    </div>
                  </td>
                  <td class="px-4 py-3 text-right text-gray-600 dark:text-gray-400">{{ group.member_count }}</td>
                  <td class="px-4 py-3 text-right">
                    <div class="flex items-center justify-end gap-2">
                      <div class="w-12 bg-gray-100 dark:bg-gray-700 h-1.5 rounded-full overflow-hidden">
                        <div class="bg-indigo-500 h-full rounded-full" :style="{ width: Math.min(group.engagement_rate, 100) + '%' }"></div>
                      </div>
                      <span :class="['font-medium text-xs w-8 text-right', group.engagement_rate >= 50 ? 'text-green-600' : 'text-orange-500']">
                        {{ group.engagement_rate.toFixed(0) }}%
                      </span>
                    </div>
                  </td>
                  <td class="px-4 py-3 text-right font-medium text-gray-700 dark:text-gray-300">{{ group.app_clicks.toLocaleString() }}</td>
                  <td class="px-4 py-3 text-right font-medium text-gray-700 dark:text-gray-300">{{ group.news_read.toLocaleString() }}</td>
                  <td class="px-4 py-3 text-right text-gray-600 dark:text-gray-400">{{ group.reactions_given.toLocaleString() }}</td>
                  <td class="px-4 py-3 text-right text-gray-600 dark:text-gray-400">{{ group.poll_votes.toLocaleString() }}</td>
                  <td class="px-4 py-3 text-right">
                    <span class="font-medium text-amber-600 dark:text-amber-400">{{ group.apps_created }}</span>
                    <span class="text-gray-400 text-xs"> / {{ group.apps_created_total }}</span>
                  </td>
                  <td class="px-4 py-3">
                    <div class="flex gap-1">
                      <div
                        v-for="app in (group.top_apps || []).slice(0, 3)"
                        :key="app.app_id"
                        class="w-5 h-5 rounded flex items-center justify-center text-white"
                        :style="{ backgroundColor: app.color || '#6B7280' }"
                        :title="app.app_name + ' (' + app.clicks + ')'"
                      >
                        <Icon v-if="app.icon" :icon="app.icon" class="h-3 w-3" />
                      </div>
                    </div>
                  </td>
                  <td class="px-4 py-3">
                    <button
                      @click="openGroupDetail(group)"
                      class="text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-200 transition-colors"
                    >
                      <Icon icon="mdi:eye" class="h-4 w-4" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- ======================== -->
      <!-- TAB: PAR UTILISATEUR    -->
      <!-- ======================== -->
      <div v-else-if="activeTab === 'users' && userData">
        <!-- Filters -->
        <div class="flex flex-wrap gap-3 mb-5">
          <input
            v-model="userSearch"
            type="text"
            :placeholder="$t('reports.searchUser')"
            class="flex-1 min-w-[200px] border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300"
          />
          <select
            v-model="userRoleFilter"
            class="border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300"
          >
            <option value="">{{ $t('reports.allRoles') }}</option>
            <option value="admin">Admin</option>
            <option value="group_admin">Group Admin</option>
            <option value="editor">Editor</option>
            <option value="user">User</option>
          </select>
          <div class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
            <Icon icon="mdi:account-group" class="h-4 w-4" />
            {{ filteredUsers.length }} {{ $t('reports.usersFound') }}
          </div>
        </div>

        <!-- Users table -->
        <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700 overflow-hidden">
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead class="bg-gray-50 dark:bg-gray-700/50">
                <tr>
                  <th class="px-5 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ $t('reports.user') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ $t('reports.role') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ $t('reports.groups') }}</th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer hover:text-gray-700" @click="sortBy('app_clicks')">
                    {{ $t('reports.appClicks') }} <Icon icon="mdi:unfold-more-horizontal" class="h-3 w-3 inline" />
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer hover:text-gray-700" @click="sortBy('news_read')">
                    {{ $t('reports.newsRead') }} <Icon icon="mdi:unfold-more-horizontal" class="h-3 w-3 inline" />
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer hover:text-gray-700" @click="sortBy('apps_created')">
                    {{ $t('reports.appsCreated') }} <Icon icon="mdi:unfold-more-horizontal" class="h-3 w-3 inline" />
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer hover:text-gray-700" @click="sortBy('articles_published')">
                    {{ $t('reports.articles') }} <Icon icon="mdi:unfold-more-horizontal" class="h-3 w-3 inline" />
                  </th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider cursor-pointer hover:text-gray-700" @click="sortBy('activity_score')">
                    Score <Icon icon="mdi:unfold-more-horizontal" class="h-3 w-3 inline" />
                  </th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ $t('reports.lastLogin') }}</th>
                  <th class="px-4 py-3"></th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
                <tr
                  v-for="user in paginatedUsers"
                  :key="user.user_id"
                  class="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
                >
                  <td class="px-5 py-3">
                    <div class="flex items-center gap-2.5">
                      <div class="w-8 h-8 rounded-full bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400 font-semibold text-xs uppercase">
                        {{ (user.first_name?.[0] || user.username?.[0] || '?') }}
                      </div>
                      <div>
                        <p class="font-medium text-gray-800 dark:text-white">{{ user.first_name }} {{ user.last_name }}</p>
                        <p class="text-xs text-gray-400">{{ user.username }}</p>
                      </div>
                    </div>
                  </td>
                  <td class="px-4 py-3">
                    <span :class="['text-xs px-2 py-0.5 rounded-full font-medium', roleColors[user.role]?.badge]">
                      {{ $t('reports.roles.' + user.role) }}
                    </span>
                  </td>
                  <td class="px-4 py-3">
                    <div class="flex flex-wrap gap-1">
                      <span v-for="g in (user.groups || []).slice(0, 2)" :key="g" class="text-xs bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 px-1.5 py-0.5 rounded">{{ g }}</span>
                      <span v-if="(user.groups || []).length > 2" class="text-xs text-gray-400">+{{ user.groups.length - 2 }}</span>
                    </div>
                  </td>
                  <td class="px-4 py-3 text-right font-medium text-gray-700 dark:text-gray-300">{{ user.app_clicks }}</td>
                  <td class="px-4 py-3 text-right font-medium text-gray-700 dark:text-gray-300">{{ user.news_read }}</td>
                  <td class="px-4 py-3 text-right">
                    <span class="font-medium text-amber-600 dark:text-amber-400">{{ user.apps_created }}</span>
                    <span class="text-gray-400 text-xs"> / {{ user.apps_created_total }}</span>
                  </td>
                  <td class="px-4 py-3 text-right">
                    <span v-if="user.articles_published > 0" class="font-medium text-indigo-600 dark:text-indigo-400">{{ user.articles_published }}</span>
                    <span v-else class="text-gray-300 dark:text-gray-600">—</span>
                  </td>
                  <td class="px-4 py-3 text-right">
                    <span :class="['font-bold text-sm', user.activity_score > 50 ? 'text-green-600' : user.activity_score > 10 ? 'text-blue-600' : 'text-gray-400']">
                      {{ user.activity_score }}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-xs text-gray-400">
                    {{ user.last_login ? formatDate(user.last_login) : $t('reports.never') }}
                  </td>
                  <td class="px-4 py-3">
                    <button
                      @click="openUserDetail(user.user_id)"
                      class="text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-200 transition-colors"
                    >
                      <Icon icon="mdi:eye" class="h-4 w-4" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Pagination -->
          <div class="px-5 py-3 border-t border-gray-100 dark:border-gray-700 flex items-center justify-between text-sm text-gray-500 dark:text-gray-400">
            <span>{{ (userPage - 1) * userPageSize + 1 }}-{{ Math.min(userPage * userPageSize, filteredUsers.length) }} / {{ filteredUsers.length }}</span>
            <div class="flex gap-1">
              <button @click="userPage--" :disabled="userPage === 1" class="px-2 py-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-30">
                <Icon icon="mdi:chevron-left" class="h-4 w-4" />
              </button>
              <button @click="userPage++" :disabled="userPage >= totalUserPages" class="px-2 py-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-30">
                <Icon icon="mdi:chevron-right" class="h-4 w-4" />
              </button>
            </div>
          </div>
        </div>
        <p class="mt-2 text-xs text-gray-400 dark:text-gray-500 italic px-1">
          * Score = clics apps ×1 &nbsp;+&nbsp; articles lus ×2 &nbsp;+&nbsp; réactions ×1 &nbsp;+&nbsp; votes sondage ×1 &nbsp;+&nbsp; applis créées ×5 &nbsp;+&nbsp; articles publiés ×10
        </p>
      </div>
    </div>

    <!-- ======================== -->
    <!-- GROUP DETAIL MODAL      -->
    <!-- ======================== -->
    <div v-if="groupDetail" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" @click.self="groupDetail = null">
      <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <!-- Header -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-100 dark:border-gray-700">
          <div class="flex items-center gap-3">
            <div class="w-4 h-4 rounded-full" :style="{ backgroundColor: groupDetail.group?.group_color || '#6B7280' }"></div>
            <div>
              <p class="font-semibold text-gray-800 dark:text-white">{{ groupDetail.group?.group_name }}</p>
              <p class="text-xs text-gray-400">{{ groupDetail.group?.member_count }} {{ $t('reports.members') }} · {{ groupDetail.group?.engagement_rate?.toFixed(0) }}% {{ $t('reports.activeRate').toLowerCase() }}</p>
            </div>
          </div>
          <button @click="groupDetail = null" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200">
            <Icon icon="mdi:close" class="h-5 w-5" />
          </button>
        </div>

        <!-- Loading -->
        <div v-if="groupDetailLoading" class="flex items-center justify-center py-16">
          <Icon icon="mdi:loading" class="h-8 w-8 animate-spin text-indigo-500" />
        </div>

        <div v-else class="p-6 space-y-6">
          <!-- KPI grid -->
          <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
            <div class="bg-indigo-50 dark:bg-indigo-900/20 rounded-lg p-3 text-center">
              <p class="text-2xl font-bold text-indigo-600 dark:text-indigo-400">{{ groupDetail.group?.active_members }}</p>
              <p class="text-xs text-indigo-500 mt-0.5">{{ $t('reports.activeMembers') }}</p>
            </div>
            <div class="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-3 text-center">
              <p class="text-2xl font-bold text-blue-600 dark:text-blue-400">{{ groupDetail.group?.app_clicks }}</p>
              <p class="text-xs text-blue-500 mt-0.5">{{ $t('reports.appClicks') }}</p>
            </div>
            <div class="bg-green-50 dark:bg-green-900/20 rounded-lg p-3 text-center">
              <p class="text-2xl font-bold text-green-600 dark:text-green-400">{{ groupDetail.group?.news_read }}</p>
              <p class="text-xs text-green-500 mt-0.5">{{ $t('reports.newsRead') }}</p>
            </div>
            <div class="bg-pink-50 dark:bg-pink-900/20 rounded-lg p-3 text-center">
              <p class="text-2xl font-bold text-pink-600 dark:text-pink-400">{{ groupDetail.group?.reactions_given }}</p>
              <p class="text-xs text-pink-500 mt-0.5">{{ $t('reports.reactions') }}</p>
            </div>
            <div class="bg-purple-50 dark:bg-purple-900/20 rounded-lg p-3 text-center">
              <p class="text-2xl font-bold text-purple-600 dark:text-purple-400">{{ groupDetail.group?.poll_votes }}</p>
              <p class="text-xs text-purple-500 mt-0.5">{{ $t('reports.pollVotes') }}</p>
            </div>
            <div class="bg-amber-50 dark:bg-amber-900/20 rounded-lg p-3 text-center">
              <p class="text-2xl font-bold text-amber-600 dark:text-amber-400">
                {{ groupDetail.group?.apps_created }}<span class="text-gray-400 font-normal text-base"> / {{ groupDetail.group?.apps_created_total }}</span>
              </p>
              <p class="text-xs text-amber-500 mt-0.5">{{ $t('reports.appsCreated') }}</p>
            </div>
          </div>

          <!-- Monthly activity chart -->
          <div v-if="groupDetail.monthly_activity">
            <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">{{ $t('reports.monthlyActivity') }}</h3>
            <div class="flex items-end gap-1 h-24">
              <div
                v-for="m in groupDetail.monthly_activity"
                :key="m.month"
                class="flex-1 flex flex-col items-center gap-0.5"
              >
                <div class="w-full flex flex-col-reverse gap-px">
                  <div
                    class="w-full bg-blue-400 rounded-sm"
                    :style="{ height: maxGroupMonthlyValue > 0 ? (m.app_clicks / maxGroupMonthlyValue * 64) + 'px' : '0' }"
                    :title="$t('reports.appClicks') + ': ' + m.app_clicks"
                  ></div>
                  <div
                    class="w-full bg-green-400 rounded-sm"
                    :style="{ height: maxGroupMonthlyValue > 0 ? (m.news_read / maxGroupMonthlyValue * 64) + 'px' : '0' }"
                    :title="$t('reports.newsRead') + ': ' + m.news_read"
                  ></div>
                  <div
                    class="w-full bg-pink-400 rounded-sm"
                    :style="{ height: maxGroupMonthlyValue > 0 ? (m.reactions_given / maxGroupMonthlyValue * 64) + 'px' : '0' }"
                    :title="$t('reports.reactions') + ': ' + m.reactions_given"
                  ></div>
                  <div
                    v-if="m.apps_created > 0"
                    class="w-full bg-amber-400 rounded-sm"
                    :style="{ height: maxGroupMonthlyValue > 0 ? (m.apps_created / maxGroupMonthlyValue * 64) + 'px' : '0' }"
                    :title="$t('reports.appsCreated') + ': ' + m.apps_created"
                  ></div>
                </div>
                <p class="text-gray-400 text-[9px] truncate w-full text-center">{{ m.month.slice(5) }}</p>
              </div>
            </div>
            <div class="flex flex-wrap gap-3 mt-2 text-xs text-gray-400">
              <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-blue-400 inline-block"></span>{{ $t('reports.appClicks') }}</span>
              <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-green-400 inline-block"></span>{{ $t('reports.newsRead') }}</span>
              <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-pink-400 inline-block"></span>{{ $t('reports.reactions') }}</span>
              <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-amber-400 inline-block"></span>{{ $t('reports.appsCreated') }}</span>
            </div>
          </div>

          <!-- Top apps & news -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div v-if="groupDetail.group?.top_apps && groupDetail.group.top_apps.length">
              <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">{{ $t('reports.topApps') }}</h3>
              <div class="space-y-2">
                <div v-for="(app, idx) in groupDetail.group.top_apps" :key="app.app_id" class="flex items-center gap-2 text-sm">
                  <span class="text-xs text-gray-400 w-4">{{ idx + 1 }}</span>
                  <div class="w-6 h-6 rounded flex items-center justify-center text-white" :style="{ backgroundColor: app.color || '#6B7280' }">
                    <Icon v-if="app.icon" :icon="app.icon" class="h-3.5 w-3.5" />
                  </div>
                  <span class="flex-1 text-gray-700 dark:text-gray-300 truncate">{{ app.app_name }}</span>
                  <span class="text-gray-500 font-medium">{{ app.clicks }}</span>
                </div>
              </div>
            </div>
            <div v-if="groupDetail.group?.top_news && groupDetail.group.top_news.length">
              <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">{{ $t('reports.topNews') }}</h3>
              <div class="space-y-2">
                <div v-for="(news, idx) in groupDetail.group.top_news" :key="news.news_id" class="flex items-start gap-2 text-sm">
                  <span class="text-xs text-gray-400 w-4 mt-0.5">{{ idx + 1 }}</span>
                  <span class="flex-1 text-gray-700 dark:text-gray-300 line-clamp-2">{{ news.title }}</span>
                  <span class="text-gray-500 font-medium whitespace-nowrap">{{ news.read_count }}x</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ======================== -->
    <!-- USER DETAIL MODAL       -->
    <!-- ======================== -->
    <div v-if="userDetail" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" @click.self="userDetail = null">
      <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-3xl max-h-[90vh] overflow-y-auto">
        <!-- Modal header -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-100 dark:border-gray-700">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-full bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600 font-bold text-sm uppercase">
              {{ (userDetail.user.first_name?.[0] || userDetail.user.username?.[0] || '?') }}
            </div>
            <div>
              <p class="font-semibold text-gray-800 dark:text-white">{{ userDetail.user.first_name }} {{ userDetail.user.last_name }}</p>
              <p class="text-xs text-gray-400">{{ userDetail.user.username }} · {{ userDetail.user.email }}</p>
            </div>
          </div>
          <button @click="userDetail = null" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200">
            <Icon icon="mdi:close" class="h-5 w-5" />
          </button>
        </div>

        <div class="p-6 space-y-6">
          <!-- KPI row -->
          <div class="grid grid-cols-2 sm:grid-cols-5 gap-3">
            <div class="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-3 text-center">
              <p class="text-xl font-bold text-blue-600">{{ userDetail.user.app_clicks }}</p>
              <p class="text-xs text-blue-400 mt-0.5">{{ $t('reports.appClicks') }}</p>
            </div>
            <div class="bg-green-50 dark:bg-green-900/20 rounded-lg p-3 text-center">
              <p class="text-xl font-bold text-green-600">{{ userDetail.user.news_read }}</p>
              <p class="text-xs text-green-400 mt-0.5">{{ $t('reports.newsRead') }}</p>
            </div>
            <div class="bg-pink-50 dark:bg-pink-900/20 rounded-lg p-3 text-center">
              <p class="text-xl font-bold text-pink-600">{{ userDetail.user.reactions_given }}</p>
              <p class="text-xs text-pink-400 mt-0.5">{{ $t('reports.reactions') }}</p>
            </div>
            <div class="bg-amber-50 dark:bg-amber-900/20 rounded-lg p-3 text-center">
              <p class="text-xl font-bold text-amber-600">
                {{ userDetail.user.apps_created }}<span class="text-gray-400 font-normal text-sm"> / {{ userDetail.user.apps_created_total }}</span>
              </p>
              <p class="text-xs text-amber-400 mt-0.5">{{ $t('reports.appsCreated') }}</p>
            </div>
            <div class="bg-indigo-50 dark:bg-indigo-900/20 rounded-lg p-3 text-center">
              <p class="text-xl font-bold text-indigo-600">{{ userDetail.user.activity_score }}</p>
              <p class="text-xs text-indigo-400 mt-0.5">Score</p>
            </div>
          </div>

          <!-- Monthly activity chart (bar) -->
          <div>
            <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">{{ $t('reports.monthlyActivity') }}</h3>
            <div class="flex items-end gap-1 h-24">
              <div
                v-for="m in userDetail.monthly_activity"
                :key="m.month"
                class="flex-1 flex flex-col items-center gap-0.5"
              >
                <div class="w-full flex flex-col-reverse gap-px">
                  <div
                    class="w-full bg-blue-400 rounded-sm"
                    :style="{ height: maxMonthlyValue > 0 ? (m.app_clicks / maxMonthlyValue * 64) + 'px' : '0' }"
                    :title="'Clics: ' + m.app_clicks"
                  ></div>
                  <div
                    class="w-full bg-green-400 rounded-sm"
                    :style="{ height: maxMonthlyValue > 0 ? (m.news_read / maxMonthlyValue * 64) + 'px' : '0' }"
                    :title="'Lus: ' + m.news_read"
                  ></div>
                  <div
                    v-if="m.apps_created > 0"
                    class="w-full bg-amber-400 rounded-sm"
                    :style="{ height: maxMonthlyValue > 0 ? (m.apps_created / maxMonthlyValue * 64) + 'px' : '0' }"
                    :title="$t('reports.appsCreated') + ': ' + m.apps_created"
                  ></div>
                </div>
                <p class="text-gray-400 text-[9px] truncate w-full text-center">{{ m.month.slice(5) }}</p>
              </div>
            </div>
            <div class="flex gap-3 mt-2 text-xs text-gray-400">
              <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-blue-400 inline-block"></span>{{ $t('reports.appClicks') }}</span>
              <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-green-400 inline-block"></span>{{ $t('reports.newsRead') }}</span>
              <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-amber-400 inline-block"></span>{{ $t('reports.appsCreated') }}</span>
            </div>
          </div>

          <!-- Top apps & news -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div v-if="userDetail.top_apps && userDetail.top_apps.length">
              <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">{{ $t('reports.topApps') }}</h3>
              <div class="space-y-2">
                <div v-for="(app, idx) in userDetail.top_apps" :key="app.app_id" class="flex items-center gap-2 text-sm">
                  <span class="text-xs text-gray-400 w-4">{{ idx + 1 }}</span>
                  <div class="w-6 h-6 rounded flex items-center justify-center text-white text-xs" :style="{ backgroundColor: app.color || '#6B7280' }">
                    <Icon v-if="app.icon" :icon="app.icon" class="h-3.5 w-3.5" />
                  </div>
                  <span class="flex-1 text-gray-700 dark:text-gray-300 truncate">{{ app.app_name }}</span>
                  <span class="text-gray-500 font-medium">{{ app.clicks }}</span>
                </div>
              </div>
            </div>
            <div v-if="userDetail.top_news && userDetail.top_news.length">
              <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">{{ $t('reports.topNews') }}</h3>
              <div class="space-y-2">
                <div v-for="(news, idx) in userDetail.top_news" :key="news.news_id" class="flex items-start gap-2 text-sm">
                  <span class="text-xs text-gray-400 w-4 mt-0.5">{{ idx + 1 }}</span>
                  <span class="flex-1 text-gray-700 dark:text-gray-300 line-clamp-2">{{ news.title }}</span>
                  <span class="text-gray-500 font-medium whitespace-nowrap">{{ news.read_count }}x</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Articles authored -->
          <div v-if="userDetail.articles_authored && userDetail.articles_authored.length">
            <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">{{ $t('reports.articlesAuthored') }}</h3>
            <div class="space-y-2">
              <div
                v-for="article in userDetail.articles_authored"
                :key="article.news_id"
                class="flex items-center justify-between text-sm bg-gray-50 dark:bg-gray-700/30 rounded-lg px-3 py-2"
              >
                <div class="flex items-center gap-2">
                  <Icon
                    :icon="article.is_published ? 'mdi:check-circle' : 'mdi:pencil'"
                    :class="['h-4 w-4', article.is_published ? 'text-green-500' : 'text-orange-400']"
                  />
                  <span class="text-gray-700 dark:text-gray-300 truncate max-w-[240px]">{{ article.title }}</span>
                </div>
                <div class="flex items-center gap-3 text-gray-400">
                  <span class="flex items-center gap-1"><Icon icon="mdi:eye" class="h-3.5 w-3.5" />{{ article.view_count }}</span>
                  <span class="flex items-center gap-1"><Icon icon="mdi:heart" class="h-3.5 w-3.5" />{{ article.reaction_count }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { reportsService } from '@/services/api'

const { t } = useI18n()

// ========================
// Period state
// ========================
const activePeriod = ref('30d')
const customFrom = ref('')
const customTo = ref('')

const periodPresets = computed(() => [
  { value: '7d', label: t('reports.last7days') },
  { value: '30d', label: t('reports.last30days') },
  { value: '90d', label: t('reports.last90days') },
  { value: 'year', label: t('reports.thisYear') },
])

function getPeriodDates(preset) {
  const now = new Date()
  const to = now.toISOString().split('T')[0]
  let from
  if (preset === '7d') {
    const d = new Date(); d.setDate(d.getDate() - 6); from = d.toISOString().split('T')[0]
  } else if (preset === '30d') {
    const d = new Date(); d.setDate(d.getDate() - 29); from = d.toISOString().split('T')[0]
  } else if (preset === '90d') {
    const d = new Date(); d.setDate(d.getDate() - 89); from = d.toISOString().split('T')[0]
  } else if (preset === 'year') {
    from = new Date(now.getFullYear(), 0, 1).toISOString().split('T')[0]
  }
  return { from, to }
}

function selectPreset(preset) {
  activePeriod.value = preset
  const { from, to } = getPeriodDates(preset)
  customFrom.value = from
  customTo.value = to
  loadCurrentTab()
}

function applyCustomRange() {
  if (customFrom.value && customTo.value) {
    activePeriod.value = 'custom'
    loadCurrentTab()
  }
}

// ========================
// Tabs
// ========================
const activeTab = ref('roles')

const tabs = computed(() => [
  { key: 'roles', label: t('reports.tabRoles'), icon: 'mdi:badge-account' },
  { key: 'groups', label: t('reports.tabGroups'), icon: 'mdi:account-group' },
  { key: 'users', label: t('reports.tabUsers'), icon: 'mdi:account-details' },
])

watch(activeTab, () => loadCurrentTab())

// ========================
// Data
// ========================
const isLoading = ref(false)
const loadError = ref(null)
const roleData = ref(null)
const groupData = ref(null)
const userData = ref(null)

async function loadCurrentTab() {
  isLoading.value = true
  loadError.value = null
  const params = { from: customFrom.value, to: customTo.value }
  try {
    if (activeTab.value === 'roles') {
      roleData.value = await reportsService.getRoleReport(params)
    } else if (activeTab.value === 'groups') {
      groupData.value = await reportsService.getGroupReport(params)
    } else if (activeTab.value === 'users') {
      userData.value = await reportsService.getUserReport({ ...params, role: userRoleFilter.value })
      userPage.value = 1
    }
  } catch (e) {
    console.error('Error loading report:', e)
    loadError.value = e?.response?.data?.message || e.message || 'Erreur inconnue'
  } finally {
    isLoading.value = false
  }
}

// ========================
// Role helpers
// ========================
const roleIcons = {
  admin: 'mdi:shield-crown',
  group_admin: 'mdi:shield-account',
  editor: 'mdi:pencil-circle',
  user: 'mdi:account',
}

const roleColors = {
  admin: { bg: 'bg-red-100 dark:bg-red-900/30', text: 'text-red-600 dark:text-red-400', bar: 'bg-red-500', badge: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' },
  group_admin: { bg: 'bg-orange-100 dark:bg-orange-900/30', text: 'text-orange-600 dark:text-orange-400', bar: 'bg-orange-500', badge: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400' },
  editor: { bg: 'bg-purple-100 dark:bg-purple-900/30', text: 'text-purple-600 dark:text-purple-400', bar: 'bg-purple-500', badge: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400' },
  user: { bg: 'bg-blue-100 dark:bg-blue-900/30', text: 'text-blue-600 dark:text-blue-400', bar: 'bg-blue-500', badge: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' },
}

// ========================
// Group detail modal + sort
// ========================
const groupDetail = ref(null)
const groupDetailLoading = ref(false)

async function openGroupDetail(group) {
  groupDetail.value = { group: { ...group }, monthly_activity: null }
  groupDetailLoading.value = true
  try {
    const params = { from: customFrom.value, to: customTo.value }
    groupDetail.value = await reportsService.getGroupDetailReport(group.group_id, params)
  } catch (e) {
    console.error('Error loading group detail:', e)
  } finally {
    groupDetailLoading.value = false
  }
}

const maxGroupMonthlyValue = computed(() => {
  if (!groupDetail.value?.monthly_activity) return 1
  return Math.max(1, ...groupDetail.value.monthly_activity.map(m =>
    Math.max(m.app_clicks, m.news_read, m.reactions_given, m.apps_created || 0)
  ))
})

const groupSortField = ref('app_clicks')
const groupSortDesc = ref(true)

function sortGroupBy(field) {
  if (groupSortField.value === field) groupSortDesc.value = !groupSortDesc.value
  else { groupSortField.value = field; groupSortDesc.value = true }
}

const sortedGroups = computed(() => {
  if (!groupData.value?.group_stats) return []
  const field = groupSortField.value
  return [...groupData.value.group_stats].sort((a, b) => {
    const va = a[field] ?? 0, vb = b[field] ?? 0
    return groupSortDesc.value ? vb - va : va - vb
  })
})

// ========================
// User table
// ========================
const userSearch = ref('')
const userRoleFilter = ref('')
const userSortField = ref('activity_score')
const userSortDesc = ref(true)
const userPage = ref(1)
const userPageSize = 20

watch(userRoleFilter, () => {
  loadCurrentTab()
})

function sortBy(field) {
  if (userSortField.value === field) userSortDesc.value = !userSortDesc.value
  else { userSortField.value = field; userSortDesc.value = true }
}

const filteredUsers = computed(() => {
  if (!userData.value?.users) return []
  let list = userData.value.users
  if (userSearch.value) {
    const q = userSearch.value.toLowerCase()
    list = list.filter(u =>
      u.username?.toLowerCase().includes(q) ||
      u.first_name?.toLowerCase().includes(q) ||
      u.last_name?.toLowerCase().includes(q) ||
      u.email?.toLowerCase().includes(q)
    )
  }
  const field = userSortField.value
  return [...list].sort((a, b) => {
    const va = a[field] ?? 0, vb = b[field] ?? 0
    return userSortDesc.value ? vb - va : va - vb
  })
})

const totalUserPages = computed(() => Math.ceil(filteredUsers.value.length / userPageSize))

const paginatedUsers = computed(() => {
  const start = (userPage.value - 1) * userPageSize
  return filteredUsers.value.slice(start, start + userPageSize)
})

// ========================
// User detail modal
// ========================
const userDetail = ref(null)
const detailLoading = ref(false)

async function openUserDetail(userId) {
  detailLoading.value = true
  userDetail.value = null
  try {
    const params = { from: customFrom.value, to: customTo.value }
    userDetail.value = await reportsService.getUserDetailReport(userId, params)
  } catch (e) {
    console.error('Error loading user detail:', e)
  } finally {
    detailLoading.value = false
  }
}

const maxMonthlyValue = computed(() => {
  if (!userDetail.value?.monthly_activity) return 1
  return Math.max(1, ...userDetail.value.monthly_activity.map(m => Math.max(m.app_clicks, m.news_read, m.apps_created || 0)))
})

// ========================
// Helpers
// ========================
function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString()
}

// ========================
// Init
// ========================
onMounted(() => {
  selectPreset('30d')
})
</script>
