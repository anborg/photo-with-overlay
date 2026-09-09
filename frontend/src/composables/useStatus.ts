import {ref} from 'vue';

const status = ref('Ready');

export function useStatus() {
  function showStatus(message: unknown) {
    status.value = (message as {message?: string})?.message || String(message);
  }

  return {status, showStatus};
}
