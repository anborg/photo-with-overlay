import {byId} from './dom.js';

const ZOOM_MIN = 0.7;
const ZOOM_MAX = 1.5;
const ZOOM_STEP = 0.1;
const THEMES = ['hc-light', 'hc-dark', 'pretty-light', 'pretty-dark'];
const THEME_NAMES = {
  'hc-light': 'High contrast day',
  'hc-dark': 'High contrast night',
  'pretty-light': 'Indoor calm light',
  'pretty-dark': 'Indoor calm dark'
};

export function setupViewPreferences() {
  let zoom = Number(localStorage.getItem('uiZoom')) || 1;
  let theme = migrateTheme(localStorage.getItem('uiTheme'));

  const apply = () => {
    zoom = clamp(Math.round(zoom * 10) / 10, ZOOM_MIN, ZOOM_MAX);
    const main = document.querySelector('main');
    main.classList.toggle('zoom-compact', zoom > 1 && zoom < 1.3);
    main.classList.toggle('zoom-focus', zoom >= 1.3);
    document.documentElement.dataset.theme = theme;
    byId('themeName').textContent = THEME_NAMES[theme];

    const nextTheme = THEMES[(THEMES.indexOf(theme) + 1) % THEMES.length];
    const description = `${THEME_NAMES[theme]}. Switch to ${THEME_NAMES[nextTheme]}`;
    byId('themeToggle').title = description;
    byId('themeToggle').setAttribute('aria-label', description);
    byId('zoomOut').title = 'Show more controls and gallery';
    byId('zoomIn').title = zoom >= 1.3 ? 'Photo area maximized' : 'Enlarge photo area';
    document.dispatchEvent(new CustomEvent('layoutchange'));
    localStorage.setItem('uiZoom', String(zoom));
    localStorage.setItem('uiTheme', theme);
  };

  byId('zoomOut').addEventListener('click', () => { zoom -= ZOOM_STEP; apply(); });
  byId('zoomIn').addEventListener('click', () => { zoom += ZOOM_STEP; apply(); });
  byId('themeToggle').addEventListener('click', () => {
    theme = THEMES[(THEMES.indexOf(theme) + 1) % THEMES.length];
    apply();
  });
  apply();
}

function migrateTheme(savedTheme) {
  if (savedTheme === 'light') return 'hc-light';
  if (savedTheme === 'dark') return 'hc-dark';
  return THEMES.includes(savedTheme) ? savedTheme : 'hc-dark';
}

function clamp(value, minimum, maximum) {
  return Math.max(minimum, Math.min(maximum, value));
}
