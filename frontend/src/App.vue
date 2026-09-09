<script setup lang="ts">
import {onMounted, ref} from 'vue';
import {LoadSettings, SavePhoto, SelectOutputFolder} from '../wailsjs/go/main/App';
import {main, type config} from '../wailsjs/go/models';
import AboutDialog from './components/AboutDialog.vue';
import AppHeader from './components/AppHeader.vue';
import CameraPanel from './components/CameraPanel.vue';
import Footer from './components/Footer.vue';
import PreviewDialog from './components/PreviewDialog.vue';
import SettingsPanel from './components/SettingsPanel.vue';
import {useCamera} from './composables/useCamera';
import {useGallery} from './composables/useGallery';
import {useLayout} from './composables/useLayout';
import {useLocation} from './composables/useLocation';
import {useSettings} from './composables/useSettings';
import {useStatus} from './composables/useStatus';
import {useViewPreferences} from './composables/useViewPreferences';
import {drawWatermark} from './lib/draw-watermark';
import {localISO} from './lib/local-iso';
import type {LocationResult} from './types';

const {showStatus} = useStatus();
const {zoomCompact, zoomFocus} = useViewPreferences();
const {settingsClosed, galleryClosed, settingsToggleLabel, toggleSettings} = useLayout();
const {settings, applySettings, saveSettings} = useSettings();
const camera = useCamera();
const locationApi = useLocation();
const gallery = useGallery();

const canvasEl = ref<HTMLCanvasElement | null>(null);
const cameraPanelRef = ref<InstanceType<typeof CameraPanel> | null>(null);
const capturing = ref(false);

async function persistSettings(): Promise<config.Settings> {
  const watermarkState = cameraPanelRef.value!.watermarkState();
  return saveSettings(watermarkState);
}

async function onWatermarkPersist(successMessage?: string) {
  try {
    await persistSettings();
    if (successMessage) showStatus(successMessage);
  } catch (error) {
    showStatus(error);
  }
}

async function populateAutomaticLocation(loc: LocationResult | null) {
  if (!loc || settings.useManualLocation) return;
  settings.manualLatitude = Number(loc.latitude.toFixed(6));
  settings.manualLongitude = Number(loc.longitude.toFixed(6));
  settings.manualAddress = loc.address;
  cameraPanelRef.value?.setRoadClue(loc.roadClue || '');
  await persistSettings();
  showStatus('GPS and location updated');
}

onMounted(async () => {
  try {
    const loaded = await LoadSettings();
    applySettings(loaded);
    cameraPanelRef.value?.loadWatermark(loaded);
    locationApi.locationActivity.value = loaded.useManualLocation ? 'custom entry' : 'automatic GPS';
    await camera.enumerate(loaded.cameraId);
    if (cameraPanelRef.value?.videoEl) {
      await camera.start(cameraPanelRef.value.videoEl, persistSettings);
    }
    await gallery.refreshGallery();
    if (!loaded.useManualLocation) await locationApi.refreshAutomaticLocation(populateAutomaticLocation);
  } catch (error) {
    showStatus(error);
  }
});

async function onStartCamera() {
  if (!cameraPanelRef.value?.videoEl) return;
  await camera.start(cameraPanelRef.value.videoEl, persistSettings);
}

async function onManualChanged() {
  locationApi.locationActivity.value = settings.useManualLocation ? 'custom entry' : 'automatic GPS';
  cameraPanelRef.value?.setRoadClue('');
  try {
    await persistSettings();
    if (!settings.useManualLocation) await locationApi.refreshAutomaticLocation(populateAutomaticLocation);
  } catch (error) {
    showStatus(error);
  }
}

async function onBrowseOutputFolder() {
  try {
    const selected = await SelectOutputFolder(settings.outputFolder);
    if (!selected) return;
    settings.outputFolder = selected;
    await persistSettings();
    await gallery.refreshGallery();
  } catch (error) {
    showStatus(error);
  }
}

async function capture() {
  capturing.value = true;
  try {
    const savedSettings = await persistSettings();
    const loc = await locationApi.getLocation(savedSettings);
    const capturedAt = new Date();
    const video = cameraPanelRef.value?.videoEl;
    if (!video || !canvasEl.value) return;
    canvasEl.value.width = video.videoWidth;
    canvasEl.value.height = video.videoHeight;
    const context = canvasEl.value.getContext('2d');
    if (!context) return;
    context.drawImage(video, 0, 0);
    drawWatermark(context, canvasEl.value, savedSettings, capturedAt, loc);

    showStatus('Writing photo and metadata…');
    const item = await SavePhoto(main.SaveRequest.createFrom({
      jpegDataUrl: canvasEl.value.toDataURL('image/jpeg', 0.92),
      capturedAt: localISO(capturedAt),
      user: savedSettings.user,
      latitude: loc.latitude,
      longitude: loc.longitude,
      accuracy: loc.accuracy ?? undefined,
      location: loc.address,
      locationSource: loc.source,
      outputFolder: savedSettings.outputFolder
    }));
    showStatus(`Saved: ${item.path}`);
    await gallery.refreshGallery();
  } catch (error: any) {
    showStatus(`Capture failed: ${error.message || error}`);
  } finally {
    capturing.value = false;
  }
}
</script>

<template>
  <main :class="{'settings-closed': settingsClosed, 'gallery-closed': galleryClosed, 'zoom-compact': zoomCompact, 'zoom-focus': zoomFocus}">
    <AppHeader />
    <CameraPanel ref="cameraPanelRef" :persist="onWatermarkPersist" />
    <SettingsPanel @start-camera="onStartCamera" @manual-changed="onManualChanged" @browse-output-folder="onBrowseOutputFolder" />
    <button
      id="settingsDrawerToggle"
      class="drawer-handle settings-drawer-handle"
      type="button"
      aria-controls="settingsPanel"
      :aria-expanded="!settingsClosed"
      :title="settingsToggleLabel"
      :aria-label="settingsToggleLabel"
      @click="toggleSettings"
    >
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m15 6-6 6 6 6"/></svg><span>Settings</span>
    </button>
    <Footer :capture-disabled="capturing || !camera.isRunning.value" @capture="capture" />
  </main>
  <canvas ref="canvasEl" id="canvas" hidden></canvas>
  <PreviewDialog />
  <AboutDialog />
</template>
