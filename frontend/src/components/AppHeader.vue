<script setup lang="ts">
import {useLayout} from '../composables/useLayout';
import {useViewPreferences} from '../composables/useViewPreferences';

const {settingsClosed, settingsToggleLabel, toggleSettings} = useLayout();
const {themeName, nextThemeName, zoomInTitle, zoomOut, zoomIn, cycleTheme} = useViewPreferences();
</script>

<template>
  <header class="app-header">
    <div class="brand">
      <button
        id="panelToggle"
        class="brand-toggle"
        type="button"
        aria-controls="settingsPanel"
        :aria-expanded="!settingsClosed"
        :title="settingsToggleLabel"
        :aria-label="settingsToggleLabel"
        @click="toggleSettings"
      >
        <img class="brand-icon" src="/appicon.png" alt="">
      </button>
      <h1>FIELD PHOTO</h1>
    </div>
    <div class="view-controls" aria-label="Display controls">
      <span id="themeName">{{ themeName }}</span>
      <button id="zoomOut" title="Show more controls and gallery" aria-label="Zoom out" @click="zoomOut">
        <svg class="zoom-icon" viewBox="0 0 24 24" aria-hidden="true"><circle cx="10.5" cy="10.5" r="6.5"/><path d="m15.5 15.5 5 5M7.5 10.5h6"/></svg>
      </button>
      <button id="zoomIn" :title="zoomInTitle" aria-label="Zoom in" @click="zoomIn">
        <svg class="zoom-icon" viewBox="0 0 24 24" aria-hidden="true"><circle cx="10.5" cy="10.5" r="6.5"/><path d="m15.5 15.5 5 5M7.5 10.5h6M10.5 7.5v6"/></svg>
      </button>
      <button id="themeToggle" :title="`${themeName}. Switch to ${nextThemeName}`" :aria-label="`${themeName}. Switch to ${nextThemeName}`" @click="cycleTheme">
        <svg class="theme-icon hc-light-icon" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.42 1.42M17.65 17.65l1.42 1.42M2 12h2M20 12h2M4.93 19.07l1.42-1.42M17.65 6.35l1.42-1.42"/></svg>
        <svg class="theme-icon hc-dark-icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M20.2 15.3A8.5 8.5 0 0 1 8.7 3.8 8.5 8.5 0 1 0 20.2 15.3Z"/></svg>
        <svg class="theme-icon pretty-light-icon" viewBox="0 0 24 24" aria-hidden="true"><path class="girl-hair" d="M5 21c0-4 2-6 4.4-7.2A6 6 0 0 1 6 8.4C6 4.8 8.6 2 12 2s6 2.8 6 6.4a6 6 0 0 1-3.4 5.4C17 15 19 17 19 21Z"/><circle class="girl-face" cx="12" cy="9" r="4"/><path class="girl-smile" d="M10 10.5c1.2 1 2.8 1 4 0"/></svg>
        <svg class="theme-icon pretty-dark-icon" viewBox="0 0 24 24" aria-hidden="true"><path class="girl-hair" d="M5 21c0-4 2-6 4.4-7.2A6 6 0 0 1 6 8.4C6 4.8 8.6 2 12 2s6 2.8 6 6.4a6 6 0 0 1-3.4 5.4C17 15 19 17 19 21Z"/><circle class="girl-face" cx="12" cy="9" r="4"/><path class="girl-smile" d="M10 10.5c1.2 1 2.8 1 4 0"/><path class="girl-star" d="m19 2 .6 1.4L21 4l-1.4.6L19 6l-.6-1.4L17 4l1.4-.6Z"/></svg>
      </button>
    </div>
  </header>
</template>
