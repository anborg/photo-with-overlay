import {backend, byId, showStatus} from './dom.js';

let currentFolder = '';
let photos = [];
let selectedPaths = new Set();
let anchorPath = '';
let previewPath = '';
let longPressTimer = null;
let longPressTriggered = false;

export function setupGallery() {
  byId('deleteSelected').addEventListener('click', deleteSelectedPhotos);
  byId('previewClose').addEventListener('click', closePreview);
  byId('previewPrev').addEventListener('click', () => navigateSelection(-1));
  byId('previewNext').addEventListener('click', () => navigateSelection(1));
  byId('previewDialog').addEventListener('click', event => {
    if (event.target === byId('previewDialog')) closePreview();
  });
  document.addEventListener('keydown', handleDocumentKeydown);
  syncDeleteButton();
}

export async function refreshGallery() {
  currentFolder = byId('output').value;
  if (!currentFolder) return;
  try {
    photos = await backend().ListPhotos(currentFolder);
    reconcileSelection();

    const gallery = byId('gallery');
    gallery.innerHTML = '';
    byId('emptyGallery').hidden = photos.length > 0;
    for (const photo of photos) gallery.append(await createPhotoCard(photo, currentFolder));

    renderSelection();
    updatePreviewNav();
  } catch (error) {
    showStatus(error);
  }
}

async function createPhotoCard(photo, folder) {
  const card = document.createElement('div');
  card.className = 'photo';
  card.dataset.path = photo.path;
  card.innerHTML = `<button type="button" class="photo-open" aria-label="Open ${photo.name}" title="${photo.name}"><span class="thumb"><img alt="Captured field photo"></span></button><button type="button" class="photo-select-indicator" aria-label="Select photo for deletion" title="Select photo for deletion"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="m6 12 4 4 8-8"/></svg></button>`;

  card.querySelector('img').src = await backend().Thumbnail(photo.path, folder);

  card.querySelector('.photo-open').addEventListener('click', event => handlePhotoActivate(photo, event));
  const selectIndicator = card.querySelector('.photo-select-indicator');
  selectIndicator.addEventListener('click', event => {
    event.preventDefault();
    handleSelectPress(photo, event);
  });
  card.addEventListener('pointerdown', event => startLongPress(photo, event));
  card.addEventListener('pointerup', cancelLongPress);
  card.addEventListener('pointerleave', cancelLongPress);
  card.addEventListener('pointercancel', cancelLongPress);

  return card;
}

async function handlePhotoActivate(photo, event) {
  if (longPressTriggered) {
    longPressTriggered = false;
    return;
  }

  previewPath = photo.path;
  await openPreview(photo.path);
}

function handleSelectPress(photo, event) {
  if (event.shiftKey) {
    selectRangeTo(photo.path);
  } else if (event.metaKey || event.ctrlKey) {
    toggleSelection(photo.path);
  } else {
    toggleSelection(photo.path);
  }
  closePreviewIfNeeded();
  renderSelection();
}

function startLongPress(photo, event) {
  if (event.pointerType === 'mouse') return;
  cancelLongPress();
  longPressTimer = setTimeout(() => {
    longPressTriggered = true;
    toggleSelection(photo.path);
    closePreviewIfNeeded();
    renderSelection();
  }, 450);
}

function cancelLongPress() {
  if (longPressTimer !== null) {
    clearTimeout(longPressTimer);
    longPressTimer = null;
  }
}

function toggleSelection(path) {
  if (selectedPaths.has(path)) {
    selectedPaths.delete(path);
  } else {
    selectedPaths.add(path);
  }
  anchorPath = path;
  if (selectedPaths.size !== 1 || !selectedPaths.has(previewPath)) {
    previewPath = selectedPaths.size === 1 ? [...selectedPaths][0] : '';
  }
}

function selectRangeTo(path) {
  const anchor = anchorPath || singleSelectedPath() || path;
  const range = photoRange(anchor, path);
  selectedPaths = new Set(range.map(photo => photo.path));
  anchorPath = anchor;
  previewPath = selectedPaths.size === 1 ? path : '';
}

async function openPreview(path) {
  const photo = photos.find(item => item.path === path);
  if (!photo) return;
  try {
    byId('previewImage').src = await backend().PhotoDataURL(photo.path, currentFolder);
    byId('previewName').textContent = photo.name;
    previewPath = photo.path;
    if (!byId('previewDialog').open) byId('previewDialog').showModal();
    updatePreviewNav();
  } catch (error) {
    showStatus(error);
  }
}

function closePreview() {
  const dialog = byId('previewDialog');
  if (dialog.open) dialog.close();
  byId('previewImage').removeAttribute('src');
  byId('previewName').textContent = '';
  previewPath = '';
}

function closePreviewIfNeeded() {
  if (selectedPaths.size !== 1 || !selectedPaths.has(previewPath)) {
    closePreview();
  }
}

function reconcileSelection() {
  const validPaths = new Set(photos.map(photo => photo.path));
  selectedPaths = new Set([...selectedPaths].filter(path => validPaths.has(path)));
  if (anchorPath && !validPaths.has(anchorPath)) anchorPath = '';
  if (previewPath && !validPaths.has(previewPath)) closePreview();
}

function renderSelection() {
  for (const card of document.querySelectorAll('#gallery .photo')) {
    const path = card.dataset.path;
    const selected = selectedPaths.has(path);
    card.classList.toggle('is-selected', selected);
    card.setAttribute('aria-pressed', String(selected));
  }
  syncDeleteButton();
}

function syncDeleteButton() {
  const count = selectedPaths.size;
  const button = byId('deleteSelected');
  byId('selectionCount').textContent = count === 0 ? 'Select' : `${count} selected`;
  button.disabled = count === 0;
  button.classList.toggle('is-enabled', count > 0);
  button.setAttribute('aria-disabled', String(count === 0));
  button.title = count === 0 ? 'Select photos to delete' : `Delete ${count} selected photo${count === 1 ? '' : 's'}`;
  button.setAttribute('aria-label', button.title);
}

async function deleteSelectedPhotos() {
  const selection = photos.filter(photo => selectedPaths.has(photo.path));
  if (!selection.length) return;

  try {
    await backend().DeletePhotos(selection.map(photo => photo.path), currentFolder);
    showStatus(`Deleted ${selection.length} photo${selection.length === 1 ? '' : 's'}`);
    selectedPaths = new Set();
    anchorPath = '';
    closePreview();
    await refreshGallery();
  } catch (error) {
    showStatus(error);
  }
}

async function navigateSelection(direction) {
  if (!photos.length) return;
  const dialog = byId('previewDialog');
  if (!dialog.open) return;

  const currentPath = previewPath || singleSelectedPath() || photos[0].path;
  const currentIndex = Math.max(0, photos.findIndex(photo => photo.path === currentPath));
  const nextIndex = (currentIndex + direction + photos.length) % photos.length;
  const nextPath = photos[nextIndex].path;

  selectedPaths = new Set([nextPath]);
  anchorPath = nextPath;
  renderSelection();
  await openPreview(nextPath);
}

function handleDocumentKeydown(event) {
  const target = event.target;
  if (target instanceof HTMLInputElement || target instanceof HTMLSelectElement || target instanceof HTMLTextAreaElement) return;
  if (!photos.length) return;

  if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return;

  const direction = event.key === 'ArrowRight' ? 1 : -1;
  if (event.shiftKey) {
    event.preventDefault();
    extendSelectionByArrow(direction);
    closePreview();
    renderSelection();
    return;
  }

  if (byId('previewDialog').open) {
    event.preventDefault();
    void navigateSelection(direction);
  }
}

function extendSelectionByArrow(direction) {
  const basePath = anchorPath || singleSelectedPath() || photos[0].path;
  const baseIndex = Math.max(0, photos.findIndex(photo => photo.path === basePath));
  const edgeIndex = selectionEdgeIndex(direction);
  const nextIndex = clampIndex(edgeIndex + direction);
  selectedPaths = new Set(photoRange(basePath, photos[nextIndex].path).map(photo => photo.path));
  anchorPath = photos[baseIndex].path;
}

function selectionEdgeIndex(direction) {
  if (!selectedPaths.size) {
    return Math.max(0, photos.findIndex(photo => photo.path === (anchorPath || photos[0].path)));
  }
  const selectedIndexes = photos
    .map((photo, index) => selectedPaths.has(photo.path) ? index : -1)
    .filter(index => index >= 0);
  return direction > 0 ? selectedIndexes[selectedIndexes.length - 1] : selectedIndexes[0];
}

function clampIndex(index) {
  return Math.max(0, Math.min(index, photos.length - 1));
}

function photoRange(fromPath, toPath) {
  const fromIndex = Math.max(0, photos.findIndex(photo => photo.path === fromPath));
  const toIndex = Math.max(0, photos.findIndex(photo => photo.path === toPath));
  const start = Math.min(fromIndex, toIndex);
  const end = Math.max(fromIndex, toIndex);
  return photos.slice(start, end + 1);
}

function updatePreviewNav() {
  const disabled = photos.length < 2;
  byId('previewPrev').disabled = disabled;
  byId('previewNext').disabled = disabled;
}

function singleSelectedPath() {
  return selectedPaths.size === 1 ? [...selectedPaths][0] : '';
}
