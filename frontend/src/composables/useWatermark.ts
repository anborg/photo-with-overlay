import {computed, type Ref, ref} from 'vue';
import type {config} from '../../wailsjs/go/models';
import {useSettings} from './useSettings';

const {settings} = useSettings();

function clamp(value: number, minimum: number, maximum: number) {
  return Math.max(minimum, Math.min(maximum, value));
}

function localTimestamp(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

const watermarkX = ref(0);
const watermarkY = ref(1);
const watermarkWidth = ref(0.42);
const roadClue = ref('');
const now = ref(new Date());
setInterval(() => { now.value = new Date(); }, 1000);

// preview/marker element sizes aren't reactive refs, so markerStyle can't see
// a plain resize — bumping this tick is what makes it recompute on resize.
const resizeTick = ref(0);
let resizeObserver: ResizeObserver | null = null;

const previewTime = computed(() => localTimestamp(now.value));
const previewUser = computed(() => `User: ${settings.user.trim() || '—'}`);
const previewCoordinates = computed(() => {
  const latitude = Number(settings.manualLatitude);
  const longitude = Number(settings.manualLongitude);
  return Number.isFinite(latitude) && Number.isFinite(longitude)
    ? `${latitude.toFixed(6)}, ${longitude.toFixed(6)}`
    : 'GPS coordinates unavailable';
});
const previewAddress = computed(() => settings.manualAddress.trim());

export function useWatermark(
  previewEl: Ref<HTMLElement | null>,
  markerEl: Ref<HTMLElement | null>,
  resizeHandleEl: Ref<HTMLElement | null>,
  persist: (successMessage?: string) => Promise<void>
) {
  function load(loaded: config.Settings) {
    watermarkX.value = Number.isFinite(loaded.watermarkX) ? loaded.watermarkX : 0;
    watermarkY.value = Number.isFinite(loaded.watermarkY) ? loaded.watermarkY : 1;
    watermarkWidth.value = Number.isFinite(loaded.watermarkWidth) && loaded.watermarkWidth > 0
      ? loaded.watermarkWidth : 0.42;
  }

  function setRoadClue(value: string) {
    roadClue.value = value;
  }

  const markerStyle = computed(() => {
    resizeTick.value;
    const preview = previewEl.value;
    const marker = markerEl.value;
    if (!preview || !marker) return {};
    const width = Math.max(170, watermarkWidth.value * preview.clientWidth);
    const scale = clamp(width / 330, 0.72, 1.7);
    const maximumX = Math.max(0, preview.clientWidth - marker.offsetWidth);
    const maximumY = Math.max(0, preview.clientHeight - marker.offsetHeight);
    return {
      width: `${width}px`,
      '--marker-scale': String(scale),
      left: `${watermarkX.value * maximumX}px`,
      top: `${watermarkY.value * maximumY}px`
    };
  });

  const markerValueText = computed(() =>
    `${Math.round(watermarkX.value * 100)}% from left, ${Math.round(watermarkY.value * 100)}% from top, ${Math.round(watermarkWidth.value * 100)}% width`
  );

  function observePreview() {
    const preview = previewEl.value;
    if (!preview) return;
    resizeObserver?.disconnect();
    resizeObserver = new ResizeObserver(() => { resizeTick.value++; });
    resizeObserver.observe(preview);
  }

  function moveMarker(clientX: number, clientY: number, offsetX: number, offsetY: number) {
    const preview = previewEl.value;
    const marker = markerEl.value;
    if (!preview || !marker) return;
    const previewBounds = preview.getBoundingClientRect();
    const markerBounds = marker.getBoundingClientRect();
    const maximumX = Math.max(0, previewBounds.width - markerBounds.width);
    const maximumY = Math.max(0, previewBounds.height - markerBounds.height);
    watermarkX.value = maximumX ? clamp((clientX - previewBounds.left - offsetX) / maximumX, 0, 1) : 0;
    watermarkY.value = maximumY ? clamp((clientY - previewBounds.top - offsetY) / maximumY, 0, 1) : 0;
  }

  function onMarkerPointerDown(event: PointerEvent) {
    const marker = markerEl.value;
    if (!marker) return;
    event.preventDefault();
    marker.setPointerCapture(event.pointerId);
    const bounds = marker.getBoundingClientRect();
    const offsetX = event.clientX - bounds.left;
    const offsetY = event.clientY - bounds.top;
    const drag = (move: PointerEvent) => moveMarker(move.clientX, move.clientY, offsetX, offsetY);
    const done = async () => {
      marker.removeEventListener('pointermove', drag);
      marker.removeEventListener('pointerup', done);
      marker.removeEventListener('pointercancel', done);
      await persist('Watermark position saved');
    };
    marker.addEventListener('pointermove', drag);
    marker.addEventListener('pointerup', done);
    marker.addEventListener('pointercancel', done);
  }

  async function onMarkerKeydown(event: KeyboardEvent) {
    const step = event.shiftKey ? 0.05 : 0.01;
    if (event.key === 'ArrowLeft') watermarkX.value -= step;
    else if (event.key === 'ArrowRight') watermarkX.value += step;
    else if (event.key === 'ArrowUp') watermarkY.value -= step;
    else if (event.key === 'ArrowDown') watermarkY.value += step;
    else return;
    event.preventDefault();
    watermarkX.value = clamp(watermarkX.value, 0, 1);
    watermarkY.value = clamp(watermarkY.value, 0, 1);
    await persist();
  }

  function onResizePointerDown(event: PointerEvent) {
    const resizeHandle = resizeHandleEl.value;
    const preview = previewEl.value;
    const marker = markerEl.value;
    if (!resizeHandle || !preview || !marker) return;
    event.preventDefault();
    event.stopPropagation();
    resizeHandle.setPointerCapture(event.pointerId);
    const resize = (move: PointerEvent) => {
      const previewBounds = preview.getBoundingClientRect();
      const markerBounds = marker.getBoundingClientRect();
      watermarkWidth.value = clamp((move.clientX - markerBounds.left) / previewBounds.width, 0.2, 0.75);
    };
    const done = async () => {
      resizeHandle.removeEventListener('pointermove', resize);
      resizeHandle.removeEventListener('pointerup', done);
      resizeHandle.removeEventListener('pointercancel', done);
      await persist('Watermark size saved');
    };
    resizeHandle.addEventListener('pointermove', resize);
    resizeHandle.addEventListener('pointerup', done);
    resizeHandle.addEventListener('pointercancel', done);
  }

  function state(): {watermarkX: number; watermarkY: number; watermarkWidth: number} {
    return {watermarkX: watermarkX.value, watermarkY: watermarkY.value, watermarkWidth: watermarkWidth.value};
  }

  return {
    watermarkX, watermarkY, watermarkWidth, roadClue,
    previewTime, previewUser, previewCoordinates, previewAddress,
    markerStyle, markerValueText,
    load, setRoadClue, state, observePreview,
    onMarkerPointerDown, onMarkerKeydown, onResizePointerDown
  };
}
