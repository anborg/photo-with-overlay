<script setup lang="ts">
import {useGallery} from '../composables/useGallery';

const {previewDialogEl, previewImageSrc, previewName, previewNavDisabled, navigateSelection, closePreview} = useGallery();

function onDialogClick(event: MouseEvent) {
  if (event.target === previewDialogEl.value) closePreview();
}
</script>

<template>
  <dialog id="previewDialog" ref="previewDialogEl" class="preview-dialog" aria-labelledby="previewName" @click="onDialogClick">
    <div class="preview-shell">
      <button
        id="previewPrev"
        type="button"
        class="preview-nav"
        aria-label="Previous photo"
        title="Previous photo"
        :disabled="previewNavDisabled"
        @click="navigateSelection(-1)"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m15 6-6 6 6 6"/></svg>
      </button>
      <figure class="preview-figure">
        <img id="previewImage" :src="previewImageSrc || undefined" alt="Selected field photo">
        <figcaption id="previewName">{{ previewName }}</figcaption>
      </figure>
      <button
        id="previewNext"
        type="button"
        class="preview-nav"
        aria-label="Next photo"
        title="Next photo"
        :disabled="previewNavDisabled"
        @click="navigateSelection(1)"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m9 6 6 6-6 6"/></svg>
      </button>
    </div>
    <button id="previewClose" type="button" class="preview-close" aria-label="Close preview" title="Close preview" @click="closePreview">
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12M18 6 6 18"/></svg>
    </button>
  </dialog>
</template>
