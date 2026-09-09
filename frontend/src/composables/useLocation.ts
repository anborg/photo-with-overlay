import {ref} from 'vue';
import {GetCurrentLocation, ReverseGeocode} from '../../wailsjs/go/main/App';
import type {config} from '../../wailsjs/go/models';
import type {LocationResult} from '../types';
import {useStatus} from './useStatus';

const {showStatus} = useStatus();

const locationActivity = ref('automatic GPS');
let latestAutomaticLocation: LocationResult | null = null;
let pendingLookup: Promise<LocationResult | null> | null = null;

function isMacOS(): boolean {
  return navigator.userAgent.includes('Mac') || navigator.platform.includes('Mac');
}

async function tryNativeLocation() {
  if (!isMacOS() || typeof GetCurrentLocation !== 'function') return null;
  const position = await GetCurrentLocation();
  return {latitude: position.latitude, longitude: position.longitude, accuracy: position.accuracy ?? null, source: 'Automatic GPS'};
}

async function lookupBrowserLocation() {
  const position = await new Promise<GeolocationPosition>((resolve, reject) =>
    navigator.geolocation.getCurrentPosition(resolve, reject, {
      enableHighAccuracy: true,
      timeout: 20000,
      maximumAge: 10000
    })
  );
  return {
    latitude: position.coords.latitude,
    longitude: position.coords.longitude,
    accuracy: position.coords.accuracy,
    source: 'Automatic GPS'
  };
}

async function lookupAutomaticLocation(): Promise<LocationResult> {
  showStatus('Getting GPS location…');
  const nativeLocation = await tryNativeLocation();
  const position = nativeLocation || await lookupBrowserLocation();
  showStatus('Looking up address and nearby roads…');
  const details = await ReverseGeocode(position.latitude, position.longitude);
  return {
    latitude: position.latitude,
    longitude: position.longitude,
    accuracy: position.accuracy,
    address: details.address,
    roadClue: details.roadClue,
    source: position.source
  };
}

export function useLocation() {
  async function refreshAutomaticLocation(onUpdate?: (location: LocationResult) => Promise<unknown>): Promise<LocationResult | null> {
    if (pendingLookup) return pendingLookup;
    locationActivity.value = 'locating…';
    pendingLookup = lookupAutomaticLocation()
      .then(async location => {
        latestAutomaticLocation = location;
        locationActivity.value = 'GPS updated';
        if (onUpdate) await onUpdate(location);
        return location;
      })
      .catch(error => {
        locationActivity.value = 'GPS unavailable';
        showStatus(error);
        return null;
      })
      .finally(() => {
        pendingLookup = null;
      });
    return pendingLookup;
  }

  async function getLocation(settings: config.Settings): Promise<LocationResult> {
    if (settings.useManualLocation) {
      return {
        latitude: settings.manualLatitude,
        longitude: settings.manualLongitude,
        accuracy: null,
        address: settings.manualAddress,
        roadClue: '',
        source: 'Manual'
      };
    }
    if (latestAutomaticLocation) return latestAutomaticLocation;
    const location = await lookupAutomaticLocation();
    latestAutomaticLocation = location;
    return location;
  }

  return {locationActivity, refreshAutomaticLocation, getLocation};
}
