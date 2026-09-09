<script setup lang="ts">
import {useCamera} from '../composables/useCamera';
import {useLocation} from '../composables/useLocation';
import {useSettings} from '../composables/useSettings';

defineEmits<{
  'start-camera': [];
  'manual-changed': [];
  'browse-output-folder': [];
}>();

const {settings} = useSettings();
const {cameras, selectedCameraId, startLabel} = useCamera();
const {locationActivity} = useLocation();
</script>

<template>
  <aside id="settingsPanel">
    <div class="camera-row">
      <button id="startCamera" @click="$emit('start-camera')">{{ startLabel }}</button>
      <label>Camera
        <select id="camera" v-model="selectedCameraId">
          <option v-for="(camera, index) in cameras" :key="camera.deviceId" :value="camera.deviceId">
            {{ camera.label || `Camera ${index + 1}` }}
          </option>
        </select>
      </label>
    </div>
    <div class="manual-user-row">
      <label class="manual-field">Manual?
        <span class="check-box">
          <input type="checkbox" id="manual" v-model="settings.useManualLocation" @change="$emit('manual-changed')">
          <span>Custom</span>
        </span>
      </label>
      <label>User<input id="user" v-model="settings.user" :disabled="!settings.useManualLocation"></label>
    </div>
    <div class="coordinate-pair">
      <label>Latitude<input id="latitude" type="number" step="any" v-model.number="settings.manualLatitude" :disabled="!settings.useManualLocation"></label>
      <label>Longitude<input id="longitude" type="number" step="any" v-model.number="settings.manualLongitude" :disabled="!settings.useManualLocation"></label>
    </div>
    <label>Location (Reverse geocoded) <span id="locationActivity" role="status">{{ locationActivity }}</span>
      <input id="address" v-model="settings.manualAddress" :disabled="!settings.useManualLocation">
    </label>
    <label>Output folder
      <div class="folder">
        <input id="output" v-model="settings.outputFolder">
        <button id="browse" title="Browse" @click="$emit('browse-output-folder')">…</button>
      </div>
    </label>
  </aside>
</template>
