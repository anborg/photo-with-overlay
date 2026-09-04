import {backend, byId, showStatus} from './dom.js';

const fields = {
  user: byId('user'),
  outputFolder: byId('output'),
  useManualLocation: byId('manual'),
  manualLatitude: byId('latitude'),
  manualLongitude: byId('longitude'),
  manualAddress: byId('address'),
  cameraId: byId('camera')
};

export function applySettings(settings) {
  for (const [name, element] of Object.entries(fields)) {
    element.type === 'checkbox'
      ? element.checked = Boolean(settings[name])
      : element.value = settings[name] ?? '';
  }
  updateLocationMode();
}

export function readSettings(watermarkState) {
  const settings = {};
  for (const [name, element] of Object.entries(fields)) {
    settings[name] = element.type === 'checkbox'
      ? element.checked
      : element.type === 'number' ? Number(element.value) : element.value;
  }
  return Object.assign(settings, {
    fontFamily: 'Arial',
    fontSize: 36,
    reverseGeocode: true,
    watermarkPosition: 'custom'
  }, watermarkState);
}

export async function saveSettings(watermarkState) {
  const settings = readSettings(watermarkState);
  await backend().SaveSettings(settings);
  return settings;
}

export function setupSettingsControls({watermarkState, refreshGallery, refreshAutomaticLocation}) {
  byId('manual').addEventListener('change', async () => {
    updateLocationMode();
    try {
      await saveSettings(watermarkState());
      if (!byId('manual').checked) refreshAutomaticLocation();
    } catch (error) {
      showStatus(error);
    }
  });
  byId('browse').addEventListener('click', async () => {
    try {
      const selected = await backend().SelectOutputFolder(byId('output').value);
      if (!selected) return;
      byId('output').value = selected;
      await saveSettings(watermarkState());
      await refreshGallery();
    } catch (error) {
      showStatus(error);
    }
  });
}

function updateLocationMode() {
  const manual = byId('manual').checked;
  byId('user').disabled = !manual;
  byId('latitude').disabled = !manual;
  byId('longitude').disabled = !manual;
  byId('address').disabled = !manual;
  byId('locationActivity').textContent = manual ? 'custom entry' : 'automatic GPS';
}
