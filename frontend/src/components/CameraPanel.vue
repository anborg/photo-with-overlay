<script setup lang="ts">
import {onMounted, ref} from 'vue';
import {useCamera} from '../composables/useCamera';
import {useLayout} from '../composables/useLayout';
import {useWatermark} from '../composables/useWatermark';
import GalleryPanel from './GalleryPanel.vue';

const props = defineProps<{persist: (successMessage?: string) => Promise<void>}>();

const videoEl = ref<HTMLVideoElement | null>(null);
const previewEl = ref<HTMLElement | null>(null);
const markerEl = ref<HTMLElement | null>(null);
const resizeHandleEl = ref<HTMLElement | null>(null);

const {cameraMessage, cameraMessageHidden} = useCamera();
const {galleryClosed, galleryToggleLabel, toggleGallery} = useLayout();
const {
  previewTime, previewUser, previewCoordinates, previewAddress, roadClue,
  markerStyle, markerValueText, load, setRoadClue, state, observePreview,
  onMarkerPointerDown, onMarkerKeydown, onResizePointerDown
} = useWatermark(previewEl, markerEl, resizeHandleEl, (msg?: string) => props.persist(msg));

onMounted(() => observePreview());

defineExpose({videoEl, loadWatermark: load, watermarkState: state, setRoadClue});
</script>

<template>
  <section class="camera-panel">
    <div ref="previewEl" class="preview">
      <video ref="videoEl" id="video" autoplay muted playsinline></video>
      <div id="cameraMessage" :hidden="cameraMessageHidden">{{ cameraMessage }}</div>
      <div
        id="watermarkMarker"
        ref="markerEl"
        tabindex="0"
        role="slider"
        aria-label="Watermark position and size. Drag to move; use the corner grip to resize."
        title="Drag to position the photo watermark"
        :aria-valuetext="markerValueText"
        :style="markerStyle"
        @pointerdown="onMarkerPointerDown"
        @keydown="onMarkerKeydown"
      >
        <strong id="watermarkTime">{{ previewTime }}</strong>
        <span id="watermarkUser">{{ previewUser }}</span>
        <span id="watermarkCoordinates">{{ previewCoordinates }}</span>
        <small id="watermarkAddress" :hidden="!previewAddress">{{ previewAddress }}</small>
        <small id="watermarkRoadClue" :hidden="!roadClue">{{ roadClue }}</small>
        <button
          id="watermarkResize"
          ref="resizeHandleEl"
          type="button"
          aria-label="Resize watermark"
          title="Drag to resize watermark"
          @pointerdown="onResizePointerDown"
        ></button>
      </div>
    </div>
    <GalleryPanel />
    <button
      id="galleryDrawerToggle"
      class="drawer-handle gallery-drawer-handle"
      type="button"
      aria-controls="galleryPanel"
      :aria-expanded="!galleryClosed"
      :title="galleryToggleLabel"
      :aria-label="galleryToggleLabel"
      @click="toggleGallery"
    >
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m6 15 6-6 6 6"/></svg><span>Gallery</span>
    </button>
  </section>
</template>
