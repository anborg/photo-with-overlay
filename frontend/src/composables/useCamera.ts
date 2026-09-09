import {computed, ref} from 'vue';
import {useStatus} from './useStatus';

const {showStatus} = useStatus();

const cameras = ref<MediaDeviceInfo[]>([]);
const selectedCameraId = ref('');
const stream = ref<MediaStream | null>(null);
const cameraMessage = ref('Select Start camera');
const cameraMessageHidden = ref(false);
const startLabel = ref('Start camera');
const isRunning = computed(() => stream.value !== null);

export function useCamera() {
  async function enumerate(preferredId?: string) {
    const devices = await navigator.mediaDevices.enumerateDevices();
    cameras.value = devices.filter(device => device.kind === 'videoinput');
    if (preferredId) selectedCameraId.value = preferredId;
  }

  async function start(videoEl: HTMLVideoElement, onStarted: () => Promise<unknown>) {
    cameraMessage.value = 'Starting camera…';
    try {
      stream.value?.getTracks().forEach(track => track.stop());
      const video = selectedCameraId.value
        ? {deviceId: {exact: selectedCameraId.value}, width: {ideal: 1920}, height: {ideal: 1080}}
        : {width: {ideal: 1920}, height: {ideal: 1080}};
      stream.value = await navigator.mediaDevices.getUserMedia({video, audio: false});
      videoEl.srcObject = stream.value;
      cameraMessageHidden.value = true;
      startLabel.value = 'Restart camera';
      await enumerate(selectedCameraId.value);
      await onStarted();
      showStatus('Camera ready');
    } catch (error) {
      showStatus(`Camera error: ${(error as Error).message}`);
    }
  }

  return {
    cameras, selectedCameraId, cameraMessage, cameraMessageHidden, startLabel, isRunning,
    enumerate, start
  };
}
