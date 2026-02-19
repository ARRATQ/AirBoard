<template>
  <div class="content-area">
    <!-- Header -->
    <div class="page-header">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="page-title">{{ $t('settings.title') }}</h1>
          <p class="page-subtitle">{{ $t('settings.subtitle') }}</p>
        </div>
        <div class="flex gap-3">
          <button @click="showPasswordModal = true" class="btn btn-secondary">
            <Icon icon="mdi:key" class="h-4 w-4 mr-2" />
            {{ $t('settings.changePassword') }}
          </button>
          <button @click="resetSettings" class="btn btn-secondary">
            <Icon icon="mdi:restore" class="h-4 w-4 mr-2" />
            {{ $t('settings.resetToDefault') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Settings Form -->
    <div class="max-w-4xl">
      <!-- Tabs Navigation -->
      <div class="flex border-b border-gray-700 mb-6 overflow-x-auto">
        <button 
          @click="activeTab = 'general'"
          class="px-4 py-2 font-medium text-sm border-b-2 transition-colors whitespace-nowrap"
          :class="activeTab === 'general' ? 'border-blue-500 text-blue-500' : 'border-transparent text-gray-400 hover:text-gray-300 hover:border-gray-600'"
        >
          <Icon icon="mdi:cog" class="inline-block mr-2 h-4 w-4" />
          {{ $t('settings.general') || 'General' }}
        </button>
        <button 
          @click="activeTab = 'hero'"
          class="px-4 py-2 font-medium text-sm border-b-2 transition-colors whitespace-nowrap"
          :class="activeTab === 'hero' ? 'border-green-500 text-green-500' : 'border-transparent text-gray-400 hover:text-gray-300 hover:border-gray-600'"
        >
          <Icon icon="mdi:format-quote-close" class="inline-block mr-2 h-4 w-4" />
          Dynamic Messages
        </button>
        <button 
          @click="activeTab = 'auth'"
          class="px-4 py-2 font-medium text-sm border-b-2 transition-colors whitespace-nowrap"
          :class="activeTab === 'auth' ? 'border-purple-500 text-purple-500' : 'border-transparent text-gray-400 hover:text-gray-300 hover:border-gray-600'"
        >
          <Icon icon="mdi:shield-account" class="inline-block mr-2 h-4 w-4" />
          Access & Auth
        </button>
        <button 
          @click="activeTab = 'danger'"
          class="px-4 py-2 font-medium text-sm border-b-2 transition-colors whitespace-nowrap"
          :class="activeTab === 'danger' ? 'border-red-500 text-red-500' : 'border-transparent text-gray-400 hover:text-gray-300 hover:border-gray-600'"
        >
          <Icon icon="mdi:alert" class="inline-block mr-2 h-4 w-4" />
          Danger Zone
        </button>
      </div>

      <div class="card">
        <form @submit.prevent="handleSubmit" class="space-y-8">
          
          <!-- GENERAL TAB -->
          <div v-show="activeTab === 'general'" class="space-y-8 animate-fade-in">
            <!-- Preview Section -->
            <div>
            <div class="section-header">
              <Icon icon="mdi:eye-outline" class="section-icon" />
              <h4 class="section-title">{{ $t('settings.livePreview') }}</h4>
            </div>
            
            <div class="bg-gray-700 rounded-lg p-6 border-2 border-dashed border-gray-600">
              <!-- Header Preview -->
              <div class="flex items-center gap-3 mb-4">
                <div class="h-10 w-10 bg-white rounded-lg flex items-center justify-center">
                  <Icon :icon="form.app_icon || 'mdi:view-dashboard'" class="h-6 w-6 text-black" />
                </div>
                <h1 class="text-xl font-bold text-white">{{ form.app_name || 'Airboard' }}</h1>
              </div>
              
              <!-- Dashboard Preview -->
              <div class="bg-gray-800 rounded-lg p-4">
                <h2 class="text-lg font-semibold text-white mb-2">{{ form.dashboard_title || $t('settings.previewDashboardTitle') }}</h2>
                <p class="text-gray-300 text-sm mb-2">{{ form.welcome_message || $t('settings.previewWelcome') }}</p>
                <p class="text-xs text-gray-400 italic">▲ Message Dashboard (page Applications)</p>
                <!-- Sample App Cards -->
                <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                  <div class="bg-gray-700 rounded-lg p-4 hover:bg-gray-600 transition-colors">
              <div class="flex items-center gap-3 mb-2">
                      <div class="h-8 w-8 bg-blue-500 rounded-lg flex items-center justify-center">
                        <Icon icon="mdi:git" class="h-4 w-4 text-white" />
                      </div>
                      <div>
                        <h3 class="font-medium text-white text-sm">GitLab</h3>
                      </div>
                    </div>
                    <p class="text-xs text-gray-400">Git repository management</p>
                  </div>
                  
                  <div class="bg-gray-700 rounded-lg p-4 hover:bg-gray-600 transition-colors">
                    <div class="flex items-center gap-3 mb-2">
                      <div class="h-8 w-8 bg-green-500 rounded-lg flex items-center justify-center">
                        <Icon icon="mdi:email" class="h-4 w-4 text-white" />
                      </div>
                      <div>
                        <h3 class="font-medium text-white text-sm">Mail</h3>
                      </div>
                    </div>
                    <p class="text-xs text-gray-400">Email management</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- App Information -->
          <div>
            <div class="section-header">
              <Icon icon="mdi:information-outline" class="section-icon" />
              <h4 class="section-title">{{ $t('settings.appInformation') }}</h4>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div class="form-group">
                <label for="app_name" class="form-label form-label-required">{{ $t('settings.appName') }}</label>
                <input
                  id="app_name"
                  v-model="form.app_name"
                  type="text"
                  required
                  class="form-input"
                  :placeholder="$t('settings.appNamePlaceholder')"
                />
                <p v-if="errors.app_name" class="form-error">{{ errors.app_name }}</p>
                <p class="form-help">This name appears in the header and browser title</p>
              </div>

              <IconInput
                v-model="form.app_icon"
                :label="$t('settings.appIcon')"
                placeholder="mdi:view-dashboard, lucide:layout-dashboard"
                category="general"
                :show-suggestions="true"
                required
              />
            </div>
          </div>

          <!-- Dashboard Customization -->
          <div>
            <div class="section-header">
              <Icon icon="mdi:view-dashboard-outline" class="section-icon" />
              <h4 class="section-title">{{ $t('settings.homePageCustomization') }}</h4>
            </div>

            <div class="space-y-4">
              <div class="form-group">
                <label for="dashboard_title" class="form-label form-label-required">{{ $t('settings.dashboardTitle') }}</label>
                <input
                  id="dashboard_title"
                  v-model="form.dashboard_title"
                  type="text"
                  required
                  class="form-input"
                  :placeholder="$t('settings.dashboardTitlePlaceholder')"
                />
                <p v-if="errors.dashboard_title" class="form-error">{{ errors.dashboard_title }}</p>
                <p class="form-help">Main title displayed on the dashboard page</p>
              </div>

              <div class="form-group">
                <label for="welcome_message" class="form-label form-label-required">{{ $t('settings.welcomeMessage') }}</label>
                <textarea
                  id="welcome_message"
                  v-model="form.welcome_message"
                  rows="3"
                  required
                  class="form-textarea"
                  :placeholder="$t('settings.welcomeMessagePlaceholder')"
                ></textarea>
                <p v-if="errors.welcome_message" class="form-error">{{ errors.welcome_message }}</p>
                <p class="form-help">Friendly message displayed on the applications dashboard page</p>
              </div>

            </div>
            </div>

          <!-- Hero Background Image -->
          <div>
            <div class="section-header">
              <Icon icon="mdi:image-outline" class="section-icon" />
              <h4 class="section-title">Image de fond du Hero</h4>
            </div>

            <div class="space-y-6">
              <p class="form-help">Personnalisez l'arrière-plan du bandeau d'accueil pour marquer des événements (Ramadan, fêtes, etc.). Laissez vide pour conserver le dégradé animé par défaut.</p>

              <!-- Light theme image -->
              <div class="space-y-3">
                <div class="flex items-center gap-2">
                  <Icon icon="mdi:white-balance-sunny" class="h-4 w-4 text-yellow-400" />
                  <span class="text-sm font-medium text-gray-300">Thème clair</span>
                </div>

                <!-- Preview light -->
                <div v-if="form.hero_image_url" class="relative rounded-lg overflow-hidden h-28 bg-gray-800 border border-gray-600">
                  <img
                    :src="form.hero_image_url"
                    alt="Aperçu image thème clair"
                    class="w-full h-full object-cover"
                    @error="form.hero_image_url = ''"
                  />
                  <div class="absolute inset-0 bg-gradient-to-r from-black/50 to-black/30 flex items-center justify-end p-3">
                    <button
                      type="button"
                      @click="form.hero_image_url = ''"
                      class="p-1.5 bg-red-600/80 hover:bg-red-600 rounded-lg text-white transition-colors"
                      title="Supprimer l'image"
                    >
                      <Icon icon="mdi:close" class="h-4 w-4" />
                    </button>
                  </div>
                </div>

                <!-- URL Input light -->
                <div class="form-group">
                  <div class="flex gap-2">
                    <input
                      v-model="form.hero_image_url"
                      type="text"
                      class="form-input flex-1"
                      placeholder="https://example.com/hero-light.jpg"
                    />
                    <label
                      class="btn btn-secondary cursor-pointer flex items-center gap-2 whitespace-nowrap"
                      :class="{ 'opacity-60 pointer-events-none': heroImageUploading }"
                    >
                      <Icon v-if="heroImageUploading" icon="mdi:loading" class="animate-spin h-4 w-4" />
                      <Icon v-else icon="mdi:upload" class="h-4 w-4" />
                      <span>{{ heroImageUploading ? 'Upload...' : 'Uploader' }}</span>
                      <input
                        type="file"
                        accept="image/*"
                        class="sr-only"
                        @change="handleHeroImageUpload"
                      />
                    </label>
                  </div>
                </div>
              </div>

              <!-- Dark theme image -->
              <div class="space-y-3">
                <div class="flex items-center gap-2">
                  <Icon icon="mdi:moon-waning-crescent" class="h-4 w-4 text-blue-400" />
                  <span class="text-sm font-medium text-gray-300">Thème sombre</span>
                  <span class="text-xs text-gray-500">(si vide, l'image du thème clair est utilisée)</span>
                </div>

                <!-- Preview dark -->
                <div v-if="form.hero_image_url_dark" class="relative rounded-lg overflow-hidden h-28 bg-gray-900 border border-gray-600">
                  <img
                    :src="form.hero_image_url_dark"
                    alt="Aperçu image thème sombre"
                    class="w-full h-full object-cover"
                    @error="form.hero_image_url_dark = ''"
                  />
                  <div class="absolute inset-0 bg-gradient-to-r from-black/50 to-black/30 flex items-center justify-end p-3">
                    <button
                      type="button"
                      @click="form.hero_image_url_dark = ''"
                      class="p-1.5 bg-red-600/80 hover:bg-red-600 rounded-lg text-white transition-colors"
                      title="Supprimer l'image sombre"
                    >
                      <Icon icon="mdi:close" class="h-4 w-4" />
                    </button>
                  </div>
                </div>

                <!-- URL Input dark -->
                <div class="form-group">
                  <div class="flex gap-2">
                    <input
                      v-model="form.hero_image_url_dark"
                      type="text"
                      class="form-input flex-1"
                      placeholder="https://example.com/hero-dark.jpg"
                    />
                    <label
                      class="btn btn-secondary cursor-pointer flex items-center gap-2 whitespace-nowrap"
                      :class="{ 'opacity-60 pointer-events-none': heroImageDarkUploading }"
                    >
                      <Icon v-if="heroImageDarkUploading" icon="mdi:loading" class="animate-spin h-4 w-4" />
                      <Icon v-else icon="mdi:upload" class="h-4 w-4" />
                      <span>{{ heroImageDarkUploading ? 'Upload...' : 'Uploader' }}</span>
                      <input
                        type="file"
                        accept="image/*"
                        class="sr-only"
                        @change="handleHeroImageDarkUpload"
                      />
                    </label>
                  </div>
                </div>
              </div>

              <p class="form-help">Format idéal : 1920×480 px. JPEG, PNG ou WebP recommandés.</p>

              <!-- Position selector (visible only when at least one image is set) -->
              <div v-if="form.hero_image_url || form.hero_image_url_dark" class="form-group">
                <label class="form-label">Point de cadrage</label>
                <div class="grid grid-cols-3 gap-1.5 w-36">
                  <button
                    v-for="pos in positions"
                    :key="pos.value"
                    type="button"
                    @click="form.hero_image_position = pos.value"
                    :title="pos.label"
                    class="h-9 rounded-md border transition-colors flex items-center justify-center"
                    :class="form.hero_image_position === pos.value
                      ? 'border-blue-500 bg-blue-500/20 text-blue-400'
                      : 'border-gray-600 bg-gray-800 text-gray-500 hover:border-gray-500 hover:text-gray-300'"
                  >
                    <Icon :icon="pos.icon" class="h-4 w-4" />
                  </button>
                </div>
                <p class="form-help mt-2">Choisissez quelle zone de l'image reste visible.</p>
              </div>
            </div>
          </div>
          </div>

          <!-- HERO MESSAGES TAB -->
          <div v-show="activeTab === 'hero'" class="animate-fade-in">
            <!-- Dynamic Hero Messages -->
            <div>
              <div class="section-header">
                <Icon icon="mdi:format-quote-close" class="section-icon" />
                <div class="flex-1 flex items-center justify-between">
                  <h4 class="section-title">Dynamic Hero Messages</h4>
                  <button type="button" @click="openHeroModal()" class="btn btn-secondary btn-sm">
                    <Icon icon="mdi:plus" class="h-4 w-4 mr-1" />
                    Add Message
                  </button>
                </div>
              </div>

              <div class="space-y-4">
                <p class="form-help mb-4">Manage the rotating quotes and messages displayed on the home page hero section.</p>
                
                <div v-if="heroMessages.length === 0" class="text-center py-8 bg-gray-800/50 rounded-lg border border-gray-700">
                  <Icon icon="mdi:format-quote-open" class="h-12 w-12 text-gray-600 mx-auto mb-2" />
                  <p class="text-gray-400">No dynamic messages configured. Default messages will be used.</p>
                </div>

                <div v-else class="grid grid-cols-1 gap-4">
                  <div v-for="hero in heroMessages" :key="hero.id" 
                    class="bg-gray-800 rounded-lg p-4 border border-gray-700 hover:border-blue-500/50 transition-colors"
                    :class="{ 'opacity-60': !hero.is_active }">
                    <div class="flex items-start justify-between gap-4">
                      <div class="flex-1">
                        <p class="text-white font-medium mb-1 line-clamp-2">"{{ hero.content }}"</p>
                        <p v-if="hero.author" class="text-xs text-gray-400">— {{ hero.author }}</p>
                      </div>
                      <div class="flex items-center gap-2">
                        <button type="button" @click="toggleHeroStatus(hero)" 
                          class="p-1.5 rounded-lg transition-colors"
                          :class="hero.is_active ? 'text-green-500 hover:bg-green-500/10' : 'text-gray-500 hover:bg-gray-500/10'"
                          :title="hero.is_active ? 'Deactivate' : 'Activate'">
                          <Icon :icon="hero.is_active ? 'mdi:eye' : 'mdi:eye-off'" class="h-5 w-5" />
                        </button>
                        <button type="button" @click="openHeroModal(hero)" class="p-1.5 text-blue-400 hover:bg-blue-500/10 rounded-lg transition-colors">
                          <Icon icon="mdi:pencil" class="h-5 w-5" />
                        </button>
                        <button type="button" @click="deleteHeroMessage(hero.id)" class="p-1.5 text-red-400 hover:bg-red-500/10 rounded-lg transition-colors">
                          <Icon icon="mdi:trash-can" class="h-5 w-5" />
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- AUTH TAB -->
          <div v-show="activeTab === 'auth'" class="space-y-8 animate-fade-in">
            <!-- Authentication Configuration -->
            <div>
            <div class="section-header">
              <Icon icon="mdi:shield-account" class="section-icon" />
              <h4 class="section-title">{{ $t('settings.authConfiguration') }}</h4>
            </div>

            <div class="space-y-4">
              <div class="form-group">
                <div class="flex items-center justify-between">
                  <div>
                    <label class="form-label">{{ $t('settings.signupEnabled') }}</label>
                    <p class="form-help">{{ $t('settings.signupEnabledHelp') }}</p>
                  </div>
                  <label class="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      v-model="form.signup_enabled"
                      class="sr-only peer"
                    />
                    <div class="w-11 h-6 bg-gray-700 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-green-800 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-green-600"></div>
                  </label>
                </div>
              </div>
            </div>
          </div>

          <!-- Default Group Configuration -->
          <div>
            <div class="section-header">
              <Icon icon="mdi:account-group" class="section-icon" />
              <h4 class="section-title">{{ $t('settings.defaultGroupConfiguration') }}</h4>
            </div>

            <div class="form-group">
              <label for="default_group" class="form-label">{{ $t('settings.defaultGroup') }}</label>
              <select
                id="default_group"
                v-model="form.default_group_id"
                class="form-input"
              >
                <option :value="null">{{ $t('settings.noDefaultGroup') }}</option>
                <option
                  v-for="group in groups"
                  :key="group.id"
                  :value="group.id"
                >
                  {{ group.name }}
                </option>
              </select>
              <p class="form-help">{{ $t('settings.defaultGroupHelp') }}</p>
            </div>
          </div>
          </div>


          <!-- DANGER ZONE TAB -->
          <div v-show="activeTab === 'danger'" class="space-y-8 animate-fade-in">
            <!-- Danger Zone -->
            <div>
            <div class="section-header">
              <Icon icon="mdi:alert-circle" class="section-icon text-red-500" />
              <h4 class="section-title text-red-600 dark:text-red-400">{{ $t('admin.dangerZone') }}</h4>
            </div>

            <!-- Danger Zone Content -->
            <div class="bg-red-50 dark:bg-red-900/20 border-2 border-red-200 dark:border-red-800 rounded-lg p-6">
              <div class="flex items-start gap-4">
                <div class="flex-shrink-0">
                  <div class="h-12 w-12 bg-red-100 dark:bg-red-900 rounded-lg flex items-center justify-center">
                    <Icon icon="mdi:database-refresh" class="h-6 w-6 text-red-600 dark:text-red-400" />
                  </div>
                </div>
                <div class="flex-1">
                  <h5 class="text-lg font-semibold text-red-900 dark:text-red-100 mb-2">
                    {{ $t('admin.resetDatabaseTitle') }}
                  </h5>
                  <p class="text-sm text-gray-700 dark:text-gray-300 mb-4">
                    {{ $t('admin.resetDatabaseTitle') }}
                    <strong class="text-red-600 dark:text-red-400">{{ $t('admin.resetDatabaseWarning') }}</strong>
                  </p>
                  <button
                    type="button"
                    @click="showResetDatabaseModal = true"
                    :disabled="isResetting"
                    class="px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-red-400 text-white rounded-lg font-medium transition-colors flex items-center gap-2"
                  >
                    <Icon icon="mdi:database-refresh" class="h-5 w-5" />
                    <span v-if="!isResetting">{{ $t('admin.resetDatabaseButton') }}</span>
                    <span v-else>{{ $t('admin.resetDatabaseInProgress') }}</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
          </div>

          <!-- Actions (Only show save/reset if not in Danger Zone or Hero tab which has auto-save/modals) -->
          <div v-if="activeTab === 'general' || activeTab === 'auth'" class="flex justify-end space-x-4 pt-6 border-t border-gray-700 mt-8">
            <button
              type="button"
              @click="loadSettings"
              class="btn btn-secondary"
            >
              <Icon icon="mdi:refresh" class="h-4 w-4 mr-2" />
              {{ $t('settings.reload') }}
            </button>
            <button
              type="submit"
              :disabled="loading"
              class="btn btn-primary"
            >
              <Icon v-if="loading" icon="mdi:loading" class="animate-spin h-4 w-4 mr-2" />
              {{ $t('settings.save') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Password Change Modal -->
    <div v-if="showPasswordModal" class="modal-overlay">
      <div class="modal-container">
        <div class="modal-panel sm:max-w-md sm:w-full">
          <div class="modal-header">
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <Icon icon="mdi:key" class="h-6 w-6 text-green-500" />
              </div>
              <div>
                <h3 class="modal-title">{{ $t('settings.changePassword') }}</h3>
                <p class="modal-subtitle">
                  {{ $t('settings.changePasswordSubtitle') }}
                </p>
              </div>
            </div>
          </div>
          
          <form @submit.prevent="handlePasswordChange" class="modal-content">
            <div class="space-y-4">
              <div class="form-group">
                <label for="old_password" class="form-label form-label-required">{{ $t('settings.currentPassword') }}</label>
                <input
                  id="old_password"
                  v-model="passwordForm.oldPassword"
                  type="password"
                  required
                  class="form-input"
                  :placeholder="$t('settings.currentPasswordPlaceholder')"
                />
                <p v-if="passwordErrors.oldPassword" class="form-error">{{ passwordErrors.oldPassword }}</p>
              </div>

              <div class="form-group">
                <label for="new_password" class="form-label form-label-required">{{ $t('settings.newPassword') }}</label>
                <input
                  id="new_password"
                  v-model="passwordForm.newPassword"
                  type="password"
                  required
                  class="form-input"
                  :placeholder="$t('settings.newPasswordPlaceholder')"
                  minlength="6"
                />
                <p v-if="passwordErrors.newPassword" class="form-error">{{ passwordErrors.newPassword }}</p>
                <p class="form-help">Minimum 6 characters</p>
              </div>

              <div class="form-group">
                <label for="confirm_password" class="form-label form-label-required">{{ $t('settings.confirmNewPassword') }}</label>
                <input
                  id="confirm_password"
                  v-model="passwordForm.confirmPassword"
                  type="password"
                  required
                  class="form-input"
                  :placeholder="$t('settings.confirmNewPasswordPlaceholder')"
                  minlength="6"
                />
                <p v-if="passwordErrors.confirmPassword" class="form-error">{{ passwordErrors.confirmPassword }}</p>
              </div>
            </div>

            <div class="modal-footer">
              <button
                type="button"
                @click="closePasswordModal"
                class="btn btn-secondary w-full sm:w-auto"
              >
                {{ $t('common.cancel') }}
              </button>
              <button
                type="submit"
                :disabled="passwordLoading"
                class="btn btn-primary w-full sm:w-auto"
              >
                <Icon v-if="passwordLoading" icon="mdi:loading" class="animate-spin h-4 w-4 mr-2" />
                {{ $t('settings.changePassword') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Reset Confirmation Modal -->
    <div v-if="showResetModal" class="modal-overlay">
      <div class="modal-container">
        <div class="modal-panel sm:max-w-lg sm:w-full">
          <div class="modal-header">
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <Icon icon="mdi:restore" class="h-6 w-6 text-yellow-500" />
              </div>
              <div>
                <h3 class="modal-title">{{ $t('settings.resetTitle') }}</h3>
                <p class="modal-subtitle">
                  {{ $t('settings.resetSubtitle') }}
                </p>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="closeResetModal" class="btn btn-secondary w-full sm:w-auto">
              {{ $t('common.cancel') }}
            </button>
            <button
              @click="confirmReset"
              :disabled="resetLoading"
              class="btn btn-danger w-full sm:w-auto"
            >
              <Icon v-if="resetLoading" icon="mdi:loading" class="animate-spin h-4 w-4 mr-2" />
              {{ $t('settings.resetButton') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Reset Database Confirmation Modal -->
    <div v-if="showResetDatabaseModal" class="modal-overlay">
      <div class="modal-container">
        <div class="modal-panel sm:max-w-md sm:w-full">
          <div class="modal-header">
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <div class="h-12 w-12 bg-red-100 dark:bg-red-900 rounded-full flex items-center justify-center">
                  <Icon icon="mdi:alert" class="h-6 w-6 text-red-600 dark:text-red-400" />
                </div>
              </div>
              <div>
                <h3 class="modal-title">{{ $t('admin.confirmResetTitle') }}</h3>
                <p class="modal-subtitle">
                  {{ $t('admin.confirmResetMessage') }}
                </p>
              </div>
            </div>
          </div>

          <div class="modal-content">
            <p class="text-gray-700 dark:text-gray-300">
              <strong class="text-red-600 dark:text-red-400 block mt-2">
                {{ $t('admin.confirmResetWarning') }}
              </strong>
            </p>
          </div>

          <div class="modal-footer">
            <button
              @click="showResetDatabaseModal = false"
              class="btn btn-secondary w-full sm:w-auto"
            >
              {{ $t('admin.confirmResetCancel') }}
            </button>
            <button
              @click="handleResetDatabase"
              :disabled="isResetting"
              class="btn btn-danger w-full sm:w-auto"
            >
              <Icon v-if="isResetting" icon="mdi:loading" class="animate-spin h-4 w-4 mr-2" />
              {{ isResetting ? $t('admin.resetInProgress') : $t('admin.confirmResetButton') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Hero Message Modal -->
    <div v-if="showHeroModal" class="modal-overlay">
      <div class="modal-container">
        <div class="modal-panel sm:max-w-lg sm:w-full">
          <div class="modal-header">
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <Icon icon="mdi:format-quote-close" class="h-6 w-6 text-blue-500" />
              </div>
              <div>
                <h3 class="modal-title">{{ editingHero ? 'Edit Message' : 'Add Hero Message' }}</h3>
                <p class="modal-subtitle">
                  Configure a quote or informational message for the homepage.
                </p>
              </div>
            </div>
          </div>
          
          <form @submit.prevent="handleHeroSubmit" class="modal-content">
            <div class="space-y-4">
              <div class="form-group">
                <label for="hero_content" class="form-label form-label-required">Message Content</label>
                <textarea
                  id="hero_content"
                  v-model="heroForm.content"
                  rows="3"
                  required
                  class="form-textarea"
                  placeholder="e.g. The only way to do great work is to love what you do."
                ></textarea>
              </div>

              <div class="form-group">
                <label for="hero_author" class="form-label">Author / Source (Optional)</label>
                <input
                  id="hero_author"
                  v-model="heroForm.author"
                  type="text"
                  class="form-input"
                  placeholder="e.g. Steve Jobs"
                />
              </div>

              <div class="form-group">
                <div class="flex items-center justify-between">
                  <div>
                    <label class="form-label">Active Status</label>
                    <p class="form-help">Visible messages will be included in the homepage rotation.</p>
                  </div>
                  <label class="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      v-model="heroForm.is_active"
                      class="sr-only peer"
                    />
                    <div class="w-11 h-6 bg-gray-700 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-800 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                  </label>
                </div>
              </div>
            </div>

            <div class="modal-footer">
              <button
                type="button"
                @click="showHeroModal = false"
                class="btn btn-secondary"
              >
                {{ $t('common.cancel') }}
              </button>
              <button
                type="submit"
                class="btn btn-primary"
              >
                {{ editingHero ? 'Update Message' : 'Add Message' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { adminService, uploadAdminMedia } from '@/services/api'
import { authService } from '@/services/api'
import { useAppStore } from '@/stores/app'
import IconInput from '@/components/ui/IconInput.vue'

const appStore = useAppStore()
const router = useRouter()

// State
const loading = ref(false)
const resetLoading = ref(false)
const showResetModal = ref(false)
const showPasswordModal = ref(false)
const passwordLoading = ref(false)
const passwordErrors = ref({})
const errors = ref({})
const showResetDatabaseModal = ref(false)
const isResetting = ref(false)
const activeTab = ref('general') // 'general', 'hero', 'auth', 'danger'
const heroMessages = ref([])
const heroLoading = ref(false)
const showHeroModal = ref(false)
const editingHero = ref(null)
const heroForm = reactive({
  content: '',
  author: '',
  is_active: true
})

const form = reactive({
  app_name: '',
  app_icon: '',
  dashboard_title: '',
  welcome_message: '',
  signup_enabled: true,
  default_group_id: null,
  hero_image_url: '',
  hero_image_url_dark: '',
  hero_image_position: 'center center'
})

const heroImageUploading = ref(false)
const heroImageDarkUploading = ref(false)

const positions = [
  { value: 'left top',     icon: 'mdi:arrow-top-left',     label: 'Haut gauche' },
  { value: 'center top',   icon: 'mdi:arrow-up',           label: 'Haut centre' },
  { value: 'right top',    icon: 'mdi:arrow-top-right',    label: 'Haut droite' },
  { value: 'left center',  icon: 'mdi:arrow-left',         label: 'Milieu gauche' },
  { value: 'center center',icon: 'mdi:circle-small',       label: 'Centre' },
  { value: 'right center', icon: 'mdi:arrow-right',        label: 'Milieu droite' },
  { value: 'left bottom',  icon: 'mdi:arrow-bottom-left',  label: 'Bas gauche' },
  { value: 'center bottom',icon: 'mdi:arrow-down',         label: 'Bas centre' },
  { value: 'right bottom', icon: 'mdi:arrow-bottom-right', label: 'Bas droite' },
]

const groups = ref([])

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// Methods
const loadGroups = async () => {
  try {
    const data = await adminService.getGroups()
    groups.value = data
  } catch (error) {
    // Ne pas afficher d'erreur si l'utilisateur n'est pas authentifié (lors de la déconnexion)
    if (authStore.isAuthenticated) {
      console.error('Error loading groups:', error)
      appStore.showError('Failed to load groups')
    }
  }
}

const handleHeroImageUpload = async (event) => {
  const file = event.target.files?.[0]
  if (!file) return

  heroImageUploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    const response = await uploadAdminMedia(formData)
    form.hero_image_url = response.data.media.url
    appStore.showSuccess('Image uploadée avec succès')
  } catch (error) {
    console.error('Error uploading hero image:', error)
    appStore.showError("Échec de l'upload de l'image")
  } finally {
    heroImageUploading.value = false
    event.target.value = ''
  }
}

const handleHeroImageDarkUpload = async (event) => {
  const file = event.target.files?.[0]
  if (!file) return

  heroImageDarkUploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    const response = await uploadAdminMedia(formData)
    form.hero_image_url_dark = response.data.media.url
    appStore.showSuccess('Image sombre uploadée avec succès')
  } catch (error) {
    console.error('Error uploading dark hero image:', error)
    appStore.showError("Échec de l'upload de l'image sombre")
  } finally {
    heroImageDarkUploading.value = false
    event.target.value = ''
  }
}

const loadSettings = async () => {
  try {
    appStore.setLoading(true)
    const data = await adminService.getAppSettings()

    Object.assign(form, {
      app_name: data.app_name || 'Airboard',
      app_icon: data.app_icon || 'mdi:view-dashboard',
      dashboard_title: data.dashboard_title || 'Dashboard',
      welcome_message: data.welcome_message || 'Welcome to your application portal',
      signup_enabled: data.signup_enabled !== undefined ? data.signup_enabled : true,
      default_group_id: data.default_group_id || null,
      hero_image_url: data.hero_image_url || '',
      hero_image_url_dark: data.hero_image_url_dark || '',
      hero_image_position: data.hero_image_position || 'center center'
    })
  } catch (error) {
    // Ne pas afficher d'erreur si l'utilisateur n'est pas authentifié (lors de la déconnexion)
    if (authStore.isAuthenticated) {
      console.error('Error loading settings:', error)
      appStore.showError('Failed to load settings')
    }
  } finally {
    appStore.setLoading(false)
  }
}

const validateForm = () => {
  errors.value = {}
  
  if (!form.app_name.trim()) {
    errors.value.app_name = 'Application name is required'
  }
  
  if (!form.app_icon.trim()) {
    errors.value.app_icon = 'Application icon is required'
  }
  
  if (!form.dashboard_title.trim()) {
    errors.value.dashboard_title = 'Dashboard title is required'
  }
  
  if (!form.welcome_message.trim()) {
    errors.value.welcome_message = 'Welcome message is required'
  }
  
  return Object.keys(errors.value).length === 0
}

const handleSubmit = async () => {
  if (!validateForm()) return
  
  loading.value = true
  try {
    await adminService.updateAppSettings(form)
    appStore.showSuccess('Settings updated successfully')
    // 🔧 FIX: Recharger les settings et mettre à jour le store global
    await loadSettings()
    appStore.invalidateSettings()
    // Optionnel: rafraîchir les settings dans le store global
    try {
      await appStore.refreshAppSettings()
    } catch (e) {
      console.warn('Could not refresh global settings:', e)
    }
  } catch (error) {
    console.error('Error updating settings:', error)
    appStore.showError('Failed to update settings')
  } finally {
    loading.value = false
  }
}

const resetSettings = () => {
  showResetModal.value = true
}

const closeResetModal = () => {
  showResetModal.value = false
}

const confirmReset = async () => {
  resetLoading.value = true
  try {
    await adminService.resetAppSettings()
    // 🔧 FIX: Recharger les settings et mettre à jour le store global
    await loadSettings()
    appStore.invalidateSettings()
    // Optionnel: rafraîchir les settings dans le store global
    try {
      await appStore.refreshAppSettings()
    } catch (e) {
      console.warn('Could not refresh global settings:', e)
    }
    appStore.showSuccess('Settings reset to defaults successfully')
    closeResetModal()
  } catch (error) {
    console.error('Error resetting settings:', error)
    appStore.showError('Failed to reset settings')
  } finally {
    resetLoading.value = false
  }
}

const validatePasswordForm = () => {
  passwordErrors.value = {}
  
  if (!passwordForm.oldPassword.trim()) {
    passwordErrors.value.oldPassword = 'Current password is required'
  }
  
  if (!passwordForm.newPassword.trim()) {
    passwordErrors.value.newPassword = 'New password is required'
  } else if (passwordForm.newPassword.length < 6) {
    passwordErrors.value.newPassword = 'Password must be at least 6 characters'
  }
  
  if (!passwordForm.confirmPassword.trim()) {
    passwordErrors.value.confirmPassword = 'Please confirm your password'
  } else if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    passwordErrors.value.confirmPassword = 'Passwords do not match'
  }
  
  return Object.keys(passwordErrors.value).length === 0
}

const handlePasswordChange = async () => {
  if (!validatePasswordForm()) return
  
  passwordLoading.value = true
  try {
    await authService.changePassword(passwordForm.oldPassword, passwordForm.newPassword)
    appStore.showSuccess('Password changed successfully')
    closePasswordModal()
  } catch (error) {
    console.error('Error changing password:', error)
    if (error.response?.status === 401) {
      passwordErrors.value.oldPassword = 'Incorrect current password'
    } else {
      appStore.showError('Failed to change password')
    }
  } finally {
    passwordLoading.value = false
  }
}

const closePasswordModal = () => {
  showPasswordModal.value = false
  Object.assign(passwordForm, {
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  })
  passwordErrors.value = {}
}

const handleResetDatabase = async () => {
  isResetting.value = true
  try {
    const response = await adminService.resetDatabase()

    // Afficher un message de succès
    alert('Base de données réinitialisée avec succès!\n\n' + response.message + '\n\nVous allez être déconnecté.')

    // Fermer le modal
    showResetDatabaseModal.value = false

    // Déconnecter l'utilisateur et rediriger vers la page de connexion
    localStorage.removeItem('airboard_token')
    localStorage.removeItem('airboard_refresh_token')
    localStorage.removeItem('airboard_user')

    // Rediriger vers la page de connexion après 1 seconde
    setTimeout(() => {
      router.push('/auth/login')
      // Recharger la page pour que le serveur recrée les données initiales
      window.location.reload()
    }, 1000)
  } catch (error) {
    console.error('Erreur lors de la réinitialisation:', error)
    alert('Erreur lors de la réinitialisation de la base de données: ' + (error.response?.data?.message || error.message))
  } finally {
    isResetting.value = false
  }
}

// Hero Messages Methods
const loadHeroMessages = async () => {
  heroLoading.value = true
  try {
    const data = await adminService.getHeroMessages()
    heroMessages.value = data
  } catch (error) {
    console.error('Error loading hero messages:', error)
    appStore.showError('Failed to load hero messages')
  } finally {
    heroLoading.value = false
  }
}

const openHeroModal = (hero = null) => {
  if (hero) {
    editingHero.value = hero
    Object.assign(heroForm, {
      content: hero.content,
      author: hero.author,
      is_active: hero.is_active
    })
  } else {
    editingHero.value = null
    Object.assign(heroForm, {
      content: '',
      author: '',
      is_active: true
    })
  }
  showHeroModal.value = true
}

const handleHeroSubmit = async () => {
  if (!heroForm.content.trim()) return

  try {
    if (editingHero.value) {
      await adminService.updateHeroMessage(editingHero.value.id, heroForm)
      appStore.showSuccess('Message updated successfully')
    } else {
      await adminService.createHeroMessage(heroForm)
      appStore.showSuccess('Message added successfully')
    }
    showHeroModal.value = false
    await loadHeroMessages()
  } catch (error) {
    console.error('Error saving hero message:', error)
    appStore.showError('Failed to save message')
  }
}

const deleteHeroMessage = async (id) => {
  if (!confirm('Are you sure you want to delete this message?')) return

  try {
    await adminService.deleteHeroMessage(id)
    appStore.showSuccess('Message deleted successfully')
    await loadHeroMessages()
  } catch (error) {
    console.error('Error deleting hero message:', error)
    appStore.showError('Failed to delete message')
  }
}

const toggleHeroStatus = async (hero) => {
  try {
    await adminService.updateHeroMessage(hero.id, {
      content: hero.content,
      author: hero.author,
      is_active: !hero.is_active
    })
    await loadHeroMessages()
  } catch (error) {
    console.error('Error toggling hero status:', error)
    appStore.showError('Failed to update message status')
  }
}

// Lifecycle
onMounted(async () => {
  await Promise.all([loadSettings(), loadGroups(), loadHeroMessages()])
})
</script>