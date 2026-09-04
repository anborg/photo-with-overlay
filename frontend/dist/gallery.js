import {backend, byId, escapeHTML, showStatus} from './dom.js';

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
  card.innerHTML = `<div class="thumb"><img alt=""><div class="name">${escapeHTML(photo.name)}</div></div>
    <div class="actions"><button class="show">◉ Show</button><button class="delete">✕ Delete</button></div>`;
  card.querySelector('img').src = await backend().Thumbnail(photo.path, folder);
  card.querySelector('.show').addEventListener('click', () => backend().ShowPhoto(photo.path, folder).catch(showStatus));
  card.querySelector('.delete').addEventListener('click', async () => {
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
