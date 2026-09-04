export const byId = id => document.getElementById(id);
export const backend = () => window.go.main.App;

export function showStatus(message) {
  byId('status').textContent = message?.message || String(message);
}

export function escapeHTML(value) {
  const replacements = {'&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;'};
  return String(value).replace(/[&<>'"]/g, character => replacements[character]);
}
