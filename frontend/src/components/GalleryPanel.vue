<script setup lang="ts">
import {useGallery} from '../composables/useGallery';
import {useSettings} from '../composables/useSettings';
import PhotoCard from './PhotoCard.vue';

const {
  photos, isEmpty, selectionCountLabel, deleteButtonDisabled, deleteButtonTitle,
  isSelected, handlePhotoActivate, handleSelectPress, startLongPress, cancelLongPress, deleteSelectedPhotos
} = useGallery();
const {settings} = useSettings();
</script>

<template>
  <div id="galleryPanel" class="gallery-wrap">
    <div id="gallerySelectionBar" class="gallery-selection-bar">
      <div class="selection-summary">
        <div id="selectionCount" aria-live="polite">{{ selectionCountLabel }}</div>
        <button
          id="deleteSelected"
          type="button"
          class="selection-action"
          :class="{'is-enabled': !deleteButtonDisabled}"
          :disabled="deleteButtonDisabled"
          :aria-disabled="deleteButtonDisabled"
          :aria-label="deleteButtonTitle"
          :title="deleteButtonTitle"
          @click="deleteSelectedPhotos"
        >
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 7h16M9 7V4h6v3M7 7l1 13h8l1-13M10 11v5M14 11v5"/></svg>
        </button>
      </div>
    </div>
    <div id="emptyGallery" :hidden="!isEmpty">Captured photos will appear here</div>
    <div id="gallery">
      <PhotoCard
        v-for="photo in photos"
        :key="photo.path"
        :item="photo"
        :folder="settings.outputFolder"
        :selected="isSelected(photo.path)"
        @activate="handlePhotoActivate"
        @select="handleSelectPress"
        @longpress="startLongPress"
        @cancel-longpress="cancelLongPress"
      />
    </div>
  </div>
</template>
