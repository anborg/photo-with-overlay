<script setup lang="ts">
import {useLayout} from '../composables/useLayout';
import {useStatus} from '../composables/useStatus';

defineProps<{captureDisabled: boolean}>();
defineEmits<{capture: []}>();

const {status} = useStatus();
const {openAbout} = useLayout();
</script>

<template>
  <footer>
    <div id="status">{{ status }}</div>
    <button id="aboutOpen" class="about-link" type="button" @click="openAbout">About</button>
    <button id="capture" :disabled="captureDisabled" @click="$emit('capture')">CAPTURE PHOTO</button>
  </footer>
</template>
