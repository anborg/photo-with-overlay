import {backend, byId, showStatus} from './dom.js';

let latestAutomaticLocation = null;
let pendingLookup = null;

export async function refreshAutomaticLocation(onUpdate) {
  if (pendingLookup) return pendingLookup;
  byId('locationActivity').textContent = 'locating…';
  pendingLookup = lookupAutomaticLocation()
    .then(async location => {
      latestAutomaticLocation = location;
      byId('locationActivity').textContent = 'GPS updated';
      if (onUpdate) await onUpdate(location);
      return location;
    })
    .catch(error => {
      byId('locationActivity').textContent = 'GPS unavailable';
      showStatus(error);
      return null;
    })
    .finally(() => { pendingLookup = null; });
  return pendingLookup;
}

export async function getLocation(settings) {
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

async function lookupAutomaticLocation() {
  showStatus('Getting GPS location…');
  const position = await new Promise((resolve, reject) =>
    navigator.geolocation.getCurrentPosition(resolve, reject, {
      enableHighAccuracy: true,
      timeout: 20000,
      maximumAge: 10000
    })
  );
  showStatus('Looking up address and nearby roads…');
  const details = await backend().ReverseGeocode(position.coords.latitude, position.coords.longitude);
  return {
    latitude: position.coords.latitude,
    longitude: position.coords.longitude,
    accuracy: position.coords.accuracy,
    address: details.address,
    roadClue: details.roadClue,
    source: 'Windows Location'
  };
}
