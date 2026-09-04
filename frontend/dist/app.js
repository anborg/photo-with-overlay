import {backend, byId, showStatus} from './dom.js';
import {setupViewPreferences} from './view-preferences.js';
import {applySettings, readSettings, saveSettings, setupSettingsControls} from './settings.js';
import {setupCamera} from './camera.js';
import {getLocation, refreshAutomaticLocation} from './location.js';
import {createWatermarkController, drawWatermark} from './watermark.js';
import {refreshGallery} from './gallery.js';
import {setupLayoutControls} from './layout.js';

let currentSettings = null;
let watermark;
let currentRoadClue = '';

setupViewPreferences();
setupLayoutControls();

async function persistSettings() {
  currentSettings = await saveSettings(watermark.state());
  return currentSettings;
}

watermark = createWatermarkController(persistSettings);
for (const id of ['user', 'latitude', 'longitude', 'address']) {
  byId(id).addEventListener('input', () => watermark.refresh({roadClue: currentRoadClue}));
}
byId('manual').addEventListener('change', () => {
  currentRoadClue = '';
  watermark.refresh();
});
setInterval(() => watermark.refresh({roadClue: currentRoadClue}), 1000);
const camera = setupCamera({
  getSettings: () => currentSettings || readSettings(watermark.state()),
  saveSettings: persistSettings
});
setupSettingsControls({
  watermarkState: watermark.state,
  refreshGallery,
  refreshAutomaticLocation: () => refreshAutomaticLocation(populateAutomaticLocation)
});

window.addEventListener('DOMContentLoaded', async () => {
  try {
    currentSettings = await backend().LoadSettings();
    applySettings(currentSettings);
    watermark.load(currentSettings);
    watermark.refresh();
    await camera.enumerate(currentSettings);
    byId('cameraMessage').textContent = 'Starting camera…';
    await camera.start();
    await refreshGallery();
    if (!currentSettings.useManualLocation) refreshAutomaticLocation(populateAutomaticLocation);
  } catch (error) {
    showStatus(error);
  }
});

async function populateAutomaticLocation(location) {
  if (!location || byId('manual').checked) return;
  byId('latitude').value = location.latitude.toFixed(6);
  byId('longitude').value = location.longitude.toFixed(6);
  byId('address').value = location.address;
  currentRoadClue = location.roadClue || '';
  watermark.refresh({roadClue: currentRoadClue});
  await persistSettings();
  showStatus('GPS and location updated');
}

byId('capture').addEventListener('click', async () => {
  const button = byId('capture');
  button.disabled = true;
  try {
    const settings = await persistSettings();
    const location = await getLocation(settings);
    const capturedAt = new Date();
    const video = byId('video');
    const canvas = byId('canvas');
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    const context = canvas.getContext('2d');
    context.drawImage(video, 0, 0);
    drawWatermark(context, canvas, settings, capturedAt, location);

    showStatus('Writing photo and metadata…');
    const item = await backend().SavePhoto({
      jpegDataUrl: canvas.toDataURL('image/jpeg', 0.92),
      capturedAt: localISO(capturedAt),
      user: settings.user,
      latitude: location.latitude,
      longitude: location.longitude,
      accuracy: location.accuracy,
      location: location.address,
      locationSource: location.source,
      outputFolder: settings.outputFolder
    });
    showStatus(`Saved: ${item.path}`);
    await refreshGallery();
  } catch (error) {
    showStatus(`Capture failed: ${error.message || error}`);
  } finally {
    button.disabled = !camera.isRunning();
  }
});

function localISO(date) {
  const pad = (number, width = 2) => String(number).padStart(width, '0');
  const offset = -date.getTimezoneOffset();
  const sign = offset >= 0 ? '+' : '-';
  const offsetHours = pad(Math.floor(Math.abs(offset) / 60));
  const offsetMinutes = pad(Math.abs(offset) % 60);
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}` +
    `T${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}.` +
    `${pad(date.getMilliseconds(), 3)}${sign}${offsetHours}:${offsetMinutes}`;
}
