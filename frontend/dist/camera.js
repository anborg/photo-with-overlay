import {backend, byId, escapeHTML, showStatus} from './dom.js';

export function setupCamera({getSettings, saveSettings}) {
  let stream = null;

  async function enumerate(settings) {
    const devices = await navigator.mediaDevices.enumerateDevices();
    const cameras = devices.filter(device => device.kind === 'videoinput');
    byId('camera').innerHTML = cameras.map((camera, index) =>
      `<option value="${escapeHTML(camera.deviceId)}">${escapeHTML(camera.label || `Camera ${index + 1}`)}</option>`
    ).join('');
    if (settings.cameraId) byId('camera').value = settings.cameraId;
  }

  async function start() {
    try {
      stream?.getTracks().forEach(track => track.stop());
      const selectedCamera = byId('camera').value;
      const video = selectedCamera
        ? {deviceId: {exact: selectedCamera}, width: {ideal: 1920}, height: {ideal: 1080}}
        : {width: {ideal: 1920}, height: {ideal: 1080}};
      stream = await navigator.mediaDevices.getUserMedia({video, audio: false});
      byId('video').srcObject = stream;
      byId('cameraMessage').hidden = true;
      byId('capture').disabled = false;
      byId('startCamera').textContent = 'Restart camera';
      await enumerate({...getSettings(), cameraId: selectedCamera});
      await saveSettings();
      showStatus('Camera ready');
    } catch (error) {
      showStatus(`Camera error: ${error.message}`);
    }
  }

  byId('startCamera').addEventListener('click', start);

  return {
    enumerate,
    start,
    isRunning: () => Boolean(stream)
  };
}
