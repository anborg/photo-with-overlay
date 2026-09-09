import {computed, ref, watch} from 'vue';

const ZOOM_MIN = 0.7;
const ZOOM_MAX = 1.5;
const ZOOM_STEP = 0.1;
const THEMES = ['hc-light', 'hc-dark', 'pretty-light', 'pretty-dark'] as const;
const THEME_NAMES: Record<string, string> = {
  'hc-light': 'High contrast day',
  'hc-dark': 'High contrast night',
  'pretty-light': 'Indoor calm light',
  'pretty-dark': 'Indoor calm dark'
};

function migrateTheme(savedTheme: string | null): string {
  if (savedTheme === 'light') return 'hc-light';
  if (savedTheme === 'dark') return 'hc-dark';
  return savedTheme && (THEMES as readonly string[]).includes(savedTheme) ? savedTheme : 'hc-dark';
}

function clamp(value: number, minimum: number, maximum: number) {
  return Math.max(minimum, Math.min(maximum, value));
}

const zoom = ref(Number(localStorage.getItem('uiZoom')) || 1);
const theme = ref(migrateTheme(localStorage.getItem('uiTheme')));
// Bumped whenever zoom/theme settle, so layout consumers (drawer toggle
// labels) can react the same way they did to the old 'layoutchange' event.
const layoutChangeTick = ref(0);

const zoomCompact = computed(() => zoom.value > 1 && zoom.value < 1.3);
const zoomFocus = computed(() => zoom.value >= 1.3);
const themeName = computed(() => THEME_NAMES[theme.value]);
const nextThemeName = computed(() => THEME_NAMES[THEMES[(THEMES.indexOf(theme.value as typeof THEMES[number]) + 1) % THEMES.length]]);
const zoomInTitle = computed(() => zoomFocus.value ? 'Photo area maximized' : 'Enlarge photo area');

watch(zoom, value => {
  const rounded = clamp(Math.round(value * 10) / 10, ZOOM_MIN, ZOOM_MAX);
  if (rounded !== value) {
    zoom.value = rounded;
    return;
  }
  localStorage.setItem('uiZoom', String(rounded));
  layoutChangeTick.value++;
}, {immediate: true});

watch(theme, value => {
  document.documentElement.dataset.theme = value;
  localStorage.setItem('uiTheme', value);
  layoutChangeTick.value++;
}, {immediate: true});

export function useViewPreferences() {
  function zoomOut() {
    zoom.value = clamp(Math.round((zoom.value - ZOOM_STEP) * 10) / 10, ZOOM_MIN, ZOOM_MAX);
  }

  function zoomIn() {
    zoom.value = clamp(Math.round((zoom.value + ZOOM_STEP) * 10) / 10, ZOOM_MIN, ZOOM_MAX);
  }

  function cycleTheme() {
    theme.value = THEMES[(THEMES.indexOf(theme.value as typeof THEMES[number]) + 1) % THEMES.length];
  }

  return {
    zoom, theme, zoomCompact, zoomFocus, themeName, nextThemeName, zoomInTitle, layoutChangeTick,
    zoomOut, zoomIn, cycleTheme
  };
}
