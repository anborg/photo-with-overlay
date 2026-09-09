import {computed, ref} from 'vue';
import {DeletePhotos, ListPhotos, PhotoDataURL} from '../../wailsjs/go/main/App';
import type {photo} from '../../wailsjs/go/models';
import {useSettings} from './useSettings';
import {useStatus} from './useStatus';

const {settings} = useSettings();
const {showStatus} = useStatus();

const photos = ref<photo.Item[]>([]);
const selectedPaths = ref<Set<string>>(new Set());
const anchorPath = ref('');
const previewPath = ref('');
const previewName = ref('');
const previewImageSrc = ref('');
const previewDialogEl = ref<HTMLDialogElement | null>(null);

const selectionCount = computed(() => selectedPaths.value.size);
const selectionCountLabel = computed(() => selectionCount.value === 0 ? 'Select' : `${selectionCount.value} selected`);
const deleteButtonDisabled = computed(() => selectionCount.value === 0);
const deleteButtonTitle = computed(() => selectionCount.value === 0
  ? 'Select photos to delete'
  : `Delete ${selectionCount.value} selected photo${selectionCount.value === 1 ? '' : 's'}`);
const previewNavDisabled = computed(() => photos.value.length < 2);
const isEmpty = computed(() => photos.value.length === 0);

function isSelected(path: string): boolean {
  return selectedPaths.value.has(path);
}

function singleSelectedPath(): string {
  return selectedPaths.value.size === 1 ? [...selectedPaths.value][0] : '';
}

function photoRange(fromPath: string, toPath: string): photo.Item[] {
  const fromIndex = Math.max(0, photos.value.findIndex(item => item.path === fromPath));
  const toIndex = Math.max(0, photos.value.findIndex(item => item.path === toPath));
  const start = Math.min(fromIndex, toIndex);
  const end = Math.max(fromIndex, toIndex);
  return photos.value.slice(start, end + 1);
}

function clampIndex(index: number): number {
  return Math.max(0, Math.min(index, photos.value.length - 1));
}

function reconcileSelection() {
  const validPaths = new Set(photos.value.map(item => item.path));
  selectedPaths.value = new Set([...selectedPaths.value].filter(path => validPaths.has(path)));
  if (anchorPath.value && !validPaths.has(anchorPath.value)) anchorPath.value = '';
  if (previewPath.value && !validPaths.has(previewPath.value)) closePreview();
}

async function refreshGallery() {
  const currentFolder = settings.outputFolder;
  if (!currentFolder) return;
  try {
    photos.value = await ListPhotos(currentFolder);
    reconcileSelection();
  } catch (error) {
    showStatus(error);
  }
}

function toggleSelection(path: string) {
  const next = new Set(selectedPaths.value);
  if (next.has(path)) next.delete(path); else next.add(path);
  selectedPaths.value = next;
  anchorPath.value = path;
  if (next.size !== 1 || !next.has(previewPath.value)) {
    previewPath.value = next.size === 1 ? [...next][0] : '';
  }
}

function selectRangeTo(path: string) {
  const anchor = anchorPath.value || singleSelectedPath() || path;
  selectedPaths.value = new Set(photoRange(anchor, path).map(item => item.path));
  anchorPath.value = anchor;
  previewPath.value = selectedPaths.value.size === 1 ? path : '';
}

function closePreviewIfNeeded() {
  if (selectedPaths.value.size !== 1 || !selectedPaths.value.has(previewPath.value)) {
    closePreview();
  }
}

function handleSelectPress(path: string, event: MouseEvent) {
  if (event.shiftKey) selectRangeTo(path); else toggleSelection(path);
  closePreviewIfNeeded();
}

let longPressTimer: ReturnType<typeof setTimeout> | null = null;
let longPressTriggered = false;

function startLongPress(path: string, event: PointerEvent) {
  if (event.pointerType === 'mouse') return;
  cancelLongPress();
  longPressTimer = setTimeout(() => {
    longPressTriggered = true;
    toggleSelection(path);
    closePreviewIfNeeded();
  }, 450);
}

function cancelLongPress() {
  if (longPressTimer !== null) {
    clearTimeout(longPressTimer);
    longPressTimer = null;
  }
}

async function handlePhotoActivate(path: string) {
  if (longPressTriggered) {
    longPressTriggered = false;
    return;
  }
  previewPath.value = path;
  await openPreview(path);
}

async function openPreview(path: string) {
  const item = photos.value.find(candidate => candidate.path === path);
  if (!item) return;
  try {
    previewImageSrc.value = await PhotoDataURL(item.path, settings.outputFolder);
    previewName.value = item.name;
    previewPath.value = item.path;
    if (!previewDialogEl.value?.open) previewDialogEl.value?.showModal();
  } catch (error) {
    showStatus(error);
  }
}

function closePreview() {
  if (previewDialogEl.value?.open) previewDialogEl.value.close();
  previewImageSrc.value = '';
  previewName.value = '';
  previewPath.value = '';
}

async function deleteSelectedPhotos() {
  const selection = photos.value.filter(item => selectedPaths.value.has(item.path));
  if (!selection.length) return;
  try {
    await DeletePhotos(selection.map(item => item.path), settings.outputFolder);
    showStatus(`Deleted ${selection.length} photo${selection.length === 1 ? '' : 's'}`);
    selectedPaths.value = new Set();
    anchorPath.value = '';
    closePreview();
    await refreshGallery();
  } catch (error) {
    showStatus(error);
  }
}

async function navigateSelection(direction: number) {
  if (!photos.value.length) return;
  if (!previewDialogEl.value?.open) return;
  const currentPath = previewPath.value || singleSelectedPath() || photos.value[0].path;
  const currentIndex = Math.max(0, photos.value.findIndex(item => item.path === currentPath));
  const nextIndex = (currentIndex + direction + photos.value.length) % photos.value.length;
  const nextPath = photos.value[nextIndex].path;
  selectedPaths.value = new Set([nextPath]);
  anchorPath.value = nextPath;
  await openPreview(nextPath);
}

function selectionEdgeIndex(direction: number): number {
  if (!selectedPaths.value.size) {
    return Math.max(0, photos.value.findIndex(item => item.path === (anchorPath.value || photos.value[0].path)));
  }
  const selectedIndexes = photos.value
    .map((item, index) => selectedPaths.value.has(item.path) ? index : -1)
    .filter(index => index >= 0);
  return direction > 0 ? selectedIndexes[selectedIndexes.length - 1] : selectedIndexes[0];
}

function extendSelectionByArrow(direction: number) {
  const basePath = anchorPath.value || singleSelectedPath() || photos.value[0].path;
  const baseIndex = Math.max(0, photos.value.findIndex(item => item.path === basePath));
  const edgeIndex = selectionEdgeIndex(direction);
  const nextIndex = clampIndex(edgeIndex + direction);
  selectedPaths.value = new Set(photoRange(basePath, photos.value[nextIndex].path).map(item => item.path));
  anchorPath.value = photos.value[baseIndex].path;
}

function handleDocumentKeydown(event: KeyboardEvent) {
  const target = event.target;
  if (target instanceof HTMLInputElement || target instanceof HTMLSelectElement || target instanceof HTMLTextAreaElement) return;
  if (!photos.value.length) return;
  if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return;

  const direction = event.key === 'ArrowRight' ? 1 : -1;
  if (event.shiftKey) {
    event.preventDefault();
    extendSelectionByArrow(direction);
    closePreview();
    return;
  }
  if (previewDialogEl.value?.open) {
    event.preventDefault();
    void navigateSelection(direction);
  }
}

// The vanilla app never tore this down either — it lives for the app's
// whole lifetime, so it's registered once at module load rather than
// through a component's mount/unmount hooks.
document.addEventListener('keydown', handleDocumentKeydown);

export function useGallery() {
  return {
    photos, isEmpty, previewPath, previewName, previewImageSrc, previewDialogEl,
    selectionCount, selectionCountLabel, deleteButtonDisabled, deleteButtonTitle, previewNavDisabled,
    isSelected, refreshGallery, handlePhotoActivate, handleSelectPress,
    startLongPress, cancelLongPress, deleteSelectedPhotos,
    navigateSelection, closePreview
  };
}
