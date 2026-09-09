import {reactive} from 'vue';
import {config} from '../../wailsjs/go/models';
import {SaveSettings} from '../../wailsjs/go/main/App';

export interface WatermarkState {
  watermarkX: number;
  watermarkY: number;
  watermarkWidth: number;
}

const settings = reactive({
  user: '',
  outputFolder: '',
  useManualLocation: false,
  manualLatitude: 0,
  manualLongitude: 0,
  manualAddress: '',
  cameraId: ''
});

export function useSettings() {
  function applySettings(loaded: config.Settings) {
    settings.user = loaded.user ?? '';
    settings.outputFolder = loaded.outputFolder ?? '';
    settings.useManualLocation = Boolean(loaded.useManualLocation);
    settings.manualLatitude = loaded.manualLatitude ?? 0;
    settings.manualLongitude = loaded.manualLongitude ?? 0;
    settings.manualAddress = loaded.manualAddress ?? '';
    settings.cameraId = loaded.cameraId ?? '';
  }

  function buildSettings(watermarkState: WatermarkState): config.Settings {
    return config.Settings.createFrom({
      ...settings,
      fontFamily: 'Arial',
      fontSize: 36,
      reverseGeocode: true,
      watermarkPosition: 'custom',
      ...watermarkState
    });
  }

  async function saveSettings(watermarkState: WatermarkState): Promise<config.Settings> {
    const built = buildSettings(watermarkState);
    await SaveSettings(built);
    return built;
  }

  return {settings, applySettings, buildSettings, saveSettings};
}
