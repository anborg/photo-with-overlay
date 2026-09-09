<script setup lang="ts">
import {useLayout} from '../composables/useLayout';

const {aboutDialogEl, closeAbout, openGoWebsite} = useLayout();

function onGoLinkKeydown(event: KeyboardEvent) {
  if (event.key !== 'Enter' && event.key !== ' ') return;
  event.preventDefault();
  openGoWebsite();
  closeAbout();
}
</script>

<template>
  <dialog id="aboutDialog" ref="aboutDialogEl" aria-labelledby="aboutTitle" @click="closeAbout">
    <div class="about-heading">
      <svg class="gopher-icon" viewBox="0 0 120 100" role="img" aria-label="Go gopher">
        <ellipse cx="60" cy="58" rx="42" ry="36" fill="#6ad7e5"/>
        <circle cx="43" cy="31" r="17" fill="#fff" stroke="#333" stroke-width="3"/>
        <circle cx="77" cy="31" r="17" fill="#fff" stroke="#333" stroke-width="3"/>
        <circle cx="47" cy="34" r="5" fill="#111"/>
        <circle cx="73" cy="34" r="5" fill="#111"/>
        <ellipse cx="60" cy="49" rx="9" ry="7" fill="#f7a7a7" stroke="#333" stroke-width="2"/>
        <path d="M48 60q12 11 24 0" fill="none" stroke="#333" stroke-width="3"/>
        <path d="M53 65v11h7V66M60 66v10h7V65" fill="#fff" stroke="#333" stroke-width="2"/>
      </svg>
      <div>
        <h2 id="aboutTitle">Field Photo</h2>
        <p>Made for useful work outdoors—and hopefully a little fun too.</p>
      </div>
    </div>
    <dl>
      <dt>Author</dt>
      <dd>Prem</dd>
      <dt>License</dt>
      <dd>MIT License</dd>
    </dl>
    <p>This app was made for field use by arborists working for city and municipal governments, and by organizations that maintain resources in public spaces.</p>
    <p class="made-with">
      Made with
      <strong role="link" tabindex="0" title="Visit the Go programming language website" @click="openGoWebsite" @keydown="onGoLinkKeydown">Go</strong>
      <span aria-hidden="true">💙</span>
    </p>
    <p>Feel free to share the app with others. If it helps you, consider adding a star on GitHub when the project page becomes available.</p>
    <p class="future-link">GitHub link: coming soon</p>
    <p class="gopher-credit">Go gopher inspired by the mascot created by Renée French (CC BY 4.0).</p>
    <div class="dialog-actions">
      <button id="aboutClose" type="button">Close</button>
    </div>
  </dialog>
</template>
