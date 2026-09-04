import {byId, showStatus} from './dom.js';

const clamp = (value, minimum, maximum) => Math.max(minimum, Math.min(maximum, value));

export function createWatermarkController(saveSettings) {
  const marker = byId('watermarkMarker');
  const resizeHandle = byId('watermarkResize');
  const preview = document.querySelector('.preview');
  const state = {watermarkX: 0, watermarkY: 1, watermarkWidth: 0.42};

  function load(settings) {
    state.watermarkX = Number.isFinite(settings.watermarkX) ? settings.watermarkX : 0;
    state.watermarkY = Number.isFinite(settings.watermarkY) ? settings.watermarkY : 1;
    state.watermarkWidth = Number.isFinite(settings.watermarkWidth) && settings.watermarkWidth > 0
      ? settings.watermarkWidth : 0.42;
    positionMarker();
  }

  function refresh({roadClue = ''} = {}) {
    const latitude = Number(byId('latitude').value);
    const longitude = Number(byId('longitude').value);
    byId('watermarkTime').textContent = localTimestamp(new Date());
    byId('watermarkUser').textContent = `User: ${byId('user').value.trim() || '—'}`;
    byId('watermarkCoordinates').textContent = Number.isFinite(latitude) && Number.isFinite(longitude)
      ? `${latitude.toFixed(6)}, ${longitude.toFixed(6)}`
      : 'GPS coordinates unavailable';
    setOptionalText(byId('watermarkAddress'), byId('address').value.trim());
    setOptionalText(byId('watermarkRoadClue'), roadClue);
    positionMarker();
  }

  function positionMarker() {
    marker.style.width = `${Math.max(170, state.watermarkWidth * preview.clientWidth)}px`;
    const scale = clamp(marker.offsetWidth / 330, 0.72, 1.7);
    marker.style.setProperty('--marker-scale', scale);
    const maximumX = Math.max(0, preview.clientWidth - marker.offsetWidth);
    const maximumY = Math.max(0, preview.clientHeight - marker.offsetHeight);
    marker.style.left = `${state.watermarkX * maximumX}px`;
    marker.style.top = `${state.watermarkY * maximumY}px`;
    marker.setAttribute('aria-valuetext',
      `${Math.round(state.watermarkX * 100)}% from left, ${Math.round(state.watermarkY * 100)}% from top, ${Math.round(state.watermarkWidth * 100)}% width`);
  }

  function moveMarker(clientX, clientY, offsetX, offsetY) {
    const previewBounds = preview.getBoundingClientRect();
    const markerBounds = marker.getBoundingClientRect();
    const maximumX = Math.max(0, previewBounds.width - markerBounds.width);
    const maximumY = Math.max(0, previewBounds.height - markerBounds.height);
    state.watermarkX = maximumX ? clamp((clientX - previewBounds.left - offsetX) / maximumX, 0, 1) : 0;
    state.watermarkY = maximumY ? clamp((clientY - previewBounds.top - offsetY) / maximumY, 0, 1) : 0;
    positionMarker();
  }

  marker.addEventListener('pointerdown', event => {
    event.preventDefault();
    marker.setPointerCapture(event.pointerId);
    const bounds = marker.getBoundingClientRect();
    const offsetX = event.clientX - bounds.left;
    const offsetY = event.clientY - bounds.top;
    const drag = move => moveMarker(move.clientX, move.clientY, offsetX, offsetY);
    const done = async () => {
      marker.removeEventListener('pointermove', drag);
      marker.removeEventListener('pointerup', done);
      marker.removeEventListener('pointercancel', done);
      await persist('Watermark position saved');
    };
    marker.addEventListener('pointermove', drag);
    marker.addEventListener('pointerup', done);
    marker.addEventListener('pointercancel', done);
  });

  marker.addEventListener('keydown', async event => {
    const step = event.shiftKey ? 0.05 : 0.01;
    if (event.key === 'ArrowLeft') state.watermarkX -= step;
    else if (event.key === 'ArrowRight') state.watermarkX += step;
    else if (event.key === 'ArrowUp') state.watermarkY -= step;
    else if (event.key === 'ArrowDown') state.watermarkY += step;
    else return;
    event.preventDefault();
    state.watermarkX = clamp(state.watermarkX, 0, 1);
    state.watermarkY = clamp(state.watermarkY, 0, 1);
    positionMarker();
    await persist();
  });

  resizeHandle.addEventListener('pointerdown', event => {
    event.preventDefault();
    event.stopPropagation();
    resizeHandle.setPointerCapture(event.pointerId);
    const resize = move => {
      const previewBounds = preview.getBoundingClientRect();
      const markerBounds = marker.getBoundingClientRect();
      state.watermarkWidth = clamp((move.clientX - markerBounds.left) / previewBounds.width, 0.2, 0.75);
      positionMarker();
    };
    const done = async () => {
      resizeHandle.removeEventListener('pointermove', resize);
      resizeHandle.removeEventListener('pointerup', done);
      resizeHandle.removeEventListener('pointercancel', done);
      await persist('Watermark size saved');
    };
    resizeHandle.addEventListener('pointermove', resize);
    resizeHandle.addEventListener('pointerup', done);
    resizeHandle.addEventListener('pointercancel', done);
  });

  new ResizeObserver(positionMarker).observe(preview);

  async function persist(successMessage) {
    try {
      await saveSettings();
      if (successMessage) showStatus(successMessage);
    } catch (error) {
      showStatus(error);
    }
  }

  return {load, refresh, state: () => ({...state})};
}

function setOptionalText(element, value) {
  element.textContent = value;
  element.hidden = !value;
}

export function drawWatermark(context, canvas, settings, capturedAt, location) {
  const widthRatio = clamp(settings.watermarkWidth ?? 0.42, 0.2, 0.75);
  const mainSize = clamp(Math.round(canvas.width * widthRatio / 22), 12, 120);
  const smallSize = Math.max(10, Math.round(mainSize * 0.62));
  const lines = [
    {text: localTimestamp(capturedAt), size: mainSize},
    {text: `User: ${settings.user}`, size: mainSize},
    {text: `${location.latitude.toFixed(6)}, ${location.longitude.toFixed(6)}`, size: mainSize}
  ];
  if (location.address) lines.push({text: location.address, size: mainSize});
  if (location.roadClue) lines.push({text: location.roadClue, size: smallSize});

  const padding = Math.max(10, Math.round(mainSize * 0.55));
  const gap = Math.max(4, Math.round(mainSize * 0.17));
  for (const line of lines) {
    context.font = `bold ${line.size}px Arial`;
    line.width = context.measureText(line.text).width;
    line.height = line.size * 1.2;
  }
  const width = Math.max(...lines.map(line => line.width)) + padding * 2;
  const height = lines.reduce((total, line) => total + line.height, 0) + gap * (lines.length - 1) + padding * 2;
  const x = Math.round(clamp(settings.watermarkX ?? 0, 0, 1) * Math.max(0, canvas.width - width));
  const y = Math.round(clamp(settings.watermarkY ?? 1, 0, 1) * Math.max(0, canvas.height - height));

  context.fillStyle = 'rgba(0,0,0,.67)';
  context.fillRect(x, y, width, height);
  context.fillStyle = '#fff';
  context.textBaseline = 'top';
  let lineY = y + padding;
  for (const line of lines) {
    context.font = `bold ${line.size}px Arial`;
    context.fillText(line.text, x + padding, lineY);
    lineY += line.height + gap;
  }
}

function localTimestamp(date) {
  const pad = number => String(number).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}
