<script setup lang="ts">
import {onMounted, ref} from 'vue';
import {Thumbnail} from '../../wailsjs/go/main/App';
import type {photo} from '../../wailsjs/go/models';

const props = defineProps<{item: photo.Item; folder: string; selected: boolean}>();
const emit = defineEmits<{
  activate: [path: string];
  select: [path: string, event: MouseEvent];
  longpress: [path: string, event: PointerEvent];
  cancelLongpress: [];
}>();

const thumbSrc = ref('');

onMounted(async () => {
  thumbSrc.value = await Thumbnail(props.item.path, props.folder);
});
</script>

<template>
  <div
    class="photo"
    :class="{'is-selected': selected}"
    :data-path="item.path"
    :aria-pressed="selected"
    @pointerdown="emit('longpress', item.path, $event)"
    @pointerup="emit('cancelLongpress')"
    @pointerleave="emit('cancelLongpress')"
    @pointercancel="emit('cancelLongpress')"
  >
    <button
      type="button"
      class="photo-open"
      :aria-label="`Open ${item.name}`"
      :title="item.name"
      @click="emit('activate', item.path)"
    >
      <span class="thumb"><img :src="thumbSrc || undefined" alt="Captured field photo"></span>
    </button>
    <button
      type="button"
      class="photo-select-indicator"
      aria-label="Select photo for deletion"
      title="Select photo for deletion"
      @click.prevent="emit('select', item.path, $event)"
    >
      <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m6 12 4 4 8-8"/></svg>
    </button>
  </div>
</template>
