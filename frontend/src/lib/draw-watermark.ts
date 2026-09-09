import type {config} from '../../wailsjs/go/models';
import type {LocationResult} from '../types';

function clamp(value: number, minimum: number, maximum: number) {
  return Math.max(minimum, Math.min(maximum, value));
}

function localTimestamp(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

interface WatermarkLine {
  text: string;
  size: number;
  width?: number;
  height?: number;
}

export function drawWatermark(
  context: CanvasRenderingContext2D,
  canvas: HTMLCanvasElement,
  settings: config.Settings,
  capturedAt: Date,
  location: LocationResult
) {
  const widthRatio = clamp(settings.watermarkWidth ?? 0.42, 0.2, 0.75);
  const mainSize = clamp(Math.round(canvas.width * widthRatio / 22), 12, 120);
  const smallSize = Math.max(10, Math.round(mainSize * 0.62));
  const lines: WatermarkLine[] = [
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
  const width = Math.max(...lines.map(line => line.width ?? 0)) + padding * 2;
  const height = lines.reduce((total, line) => total + (line.height ?? 0), 0) + gap * (lines.length - 1) + padding * 2;
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
    lineY += (line.height ?? 0) + gap;
  }
}
