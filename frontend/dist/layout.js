import {byId} from './dom.js';

export function setupLayoutControls() {
  const main = document.querySelector('main');
  const brandToggle = byId('panelToggle');
  const settingsToggle = byId('settingsDrawerToggle');
  const galleryToggle = byId('galleryDrawerToggle');
  const about = byId('aboutDialog');
  const compactScreen = matchMedia('(max-width: 700px)').matches;

  main.classList.remove('panels-hidden');
  main.classList.toggle('settings-closed', savedState('settingsDrawerClosed', compactScreen));
  main.classList.toggle('gallery-closed', savedState('galleryDrawerClosed', compactScreen));

  const updateToggle = () => {
    const settingsClosed = main.classList.contains('settings-closed');
    const galleryClosed = main.classList.contains('gallery-closed');
    setToggleState(settingsToggle, settingsClosed, 'settings');
    setToggleState(brandToggle, settingsClosed, 'settings');
    setToggleState(galleryToggle, galleryClosed, 'gallery');
  };

  const toggleSettings = () => {
    main.classList.toggle('settings-closed');
    localStorage.setItem('settingsDrawerClosed', String(main.classList.contains('settings-closed')));
    updateToggle();
  };
  const toggleGallery = () => {
    main.classList.toggle('gallery-closed');
    localStorage.setItem('galleryDrawerClosed', String(main.classList.contains('gallery-closed')));
    updateToggle();
  };

  brandToggle.addEventListener('click', toggleSettings);
  settingsToggle.addEventListener('click', toggleSettings);
  galleryToggle.addEventListener('click', toggleGallery);
  document.addEventListener('layoutchange', updateToggle);
  updateToggle();

  byId('aboutOpen').addEventListener('click', () => about.showModal());
  const goLink = about.querySelector('.made-with strong');
  goLink.setAttribute('role', 'link');
  goLink.setAttribute('tabindex', '0');
  goLink.title = 'Visit the Go programming language website';
  const openGoWebsite = () => window.runtime.BrowserOpenURL('https://go.dev/');
  goLink.addEventListener('click', openGoWebsite);
  goLink.addEventListener('keydown', event => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      openGoWebsite();
      about.close();
    }
  });
  // A modal captures clicks made anywhere over the app, including its backdrop.
  // Closing on the dialog's click event therefore makes the whole screen dismiss it.
  about.addEventListener('click', () => about.close());
}

function savedState(key, fallback) {
  const value = localStorage.getItem(key);
  return value === null ? fallback : value === 'true';
}

function setToggleState(button, closed, panelName) {
  const action = closed ? 'Show' : 'Hide';
  button.setAttribute('aria-expanded', String(!closed));
  button.title = `${action} ${panelName}`;
  button.setAttribute('aria-label', button.title);
}
