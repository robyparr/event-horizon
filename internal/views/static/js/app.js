document.addEventListener("DOMContentLoaded", function() {
  const timezoneEl = document.querySelector('[name="timezone"]');
  if (timezoneEl) {
    timezoneEl.value = Intl.DateTimeFormat().resolvedOptions().timeZone;
  }
});
