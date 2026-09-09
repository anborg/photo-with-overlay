import {computed, ref} from 'vue';
import {BrowserOpenURL} from '../../wailsjs/runtime/runtime';

function savedState(key: string, fallback: boolean): boolean {
  const value = localStorage.getItem(key);
  return value === null ? fallback : value === 'true';
}

const compactScreen = matchMedia('(max-width: 700px)').matches;
const settingsClosed = ref(savedState('settingsDrawerClosed', compactScreen));
const galleryClosed = ref(savedState('galleryDrawerClosed', compactScreen));
const aboutDialogEl = ref<HTMLDialogElement | null>(null);
let userToggledDrawer = false;

function toggleLabel(closed: boolean, panelName: string) {
  const action = closed ? 'Show' : 'Hide';
  return `${action} ${panelName}`;
}

const settingsToggleLabel = computed(() => toggleLabel(settingsClosed.value, 'settings'));
const galleryToggleLabel = computed(() => toggleLabel(galleryClosed.value, 'gallery'));

setTimeout(() => {
  if (userToggledDrawer) return;
  settingsClosed.value = true;
  galleryClosed.value = true;
}, 3000);

export function useLayout() {
  function toggleSettings() {
    userToggledDrawer = true;
    settingsClosed.value = !settingsClosed.value;
    localStorage.setItem('settingsDrawerClosed', String(settingsClosed.value));
  }

  function toggleGallery() {
    userToggledDrawer = true;
    galleryClosed.value = !galleryClosed.value;
    localStorage.setItem('galleryDrawerClosed', String(galleryClosed.value));
  }

  function openAbout() {
    if (!aboutDialogEl.value?.open) aboutDialogEl.value?.showModal();
  }

  function closeAbout() {
    if (aboutDialogEl.value?.open) aboutDialogEl.value.close();
  }

  function openGoWebsite() {
    BrowserOpenURL('https://go.dev/');
  }

  return {
    settingsClosed, galleryClosed, aboutDialogEl,
    settingsToggleLabel, galleryToggleLabel,
    toggleSettings, toggleGallery, openAbout, closeAbout, openGoWebsite
  };
}
