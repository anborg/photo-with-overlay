import {backend, byId, showStatus} from './dom.js';

export async function refreshGallery() {
  const folder = byId('output').value;
  if (!folder) return;
  try {
    const photos = await backend().ListPhotos(folder);
    const gallery = byId('gallery');
    gallery.innerHTML = '';
    byId('emptyGallery').hidden = photos.length > 0;
    for (const photo of photos) gallery.append(await createPhotoCard(photo, folder));
  } catch (error) {
    showStatus(error);
  }
}

async function createPhotoCard(photo, folder) {
  const card = document.createElement('div');
  card.className = 'photo';
  card.innerHTML = `<div class="thumb"><img alt="Captured field photo"></div>
    <div class="actions">
      <button class="show" type="button"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M2.5 12s3.5-6 9.5-6 9.5 6 9.5 6-3.5 6-9.5 6-9.5-6-9.5-6Z"/><circle cx="12" cy="12" r="3"/></svg></button>
      <button class="delete" type="button"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 7h16M9 7V4h6v3M7 7l1 13h8l1-13M10 11v5M14 11v5"/></svg></button>
    </div>`;
  card.querySelector('img').src = await backend().Thumbnail(photo.path, folder);
  const showButton = card.querySelector('.show');
  const deleteButton = card.querySelector('.delete');
  showButton.title = `View ${photo.name}`;
  showButton.setAttribute('aria-label', showButton.title);
  deleteButton.title = `Delete ${photo.name}`;
  deleteButton.setAttribute('aria-label', deleteButton.title);
  showButton.addEventListener('click', () => backend().ShowPhoto(photo.path, folder).catch(showStatus));
  deleteButton.addEventListener('click', async () => {
    if (!confirm(`Permanently delete this photo?\n\n${photo.name}`)) return;
    try {
      await backend().DeletePhoto(photo.path, folder);
      showStatus(`Deleted: ${photo.name}`);
      await refreshGallery();
    } catch (error) {
      showStatus(error);
    }
  });
  return card;
}
