document.addEventListener("DOMContentLoaded", function() {
  const timezoneEl = document.querySelector('[name="timezone"]');
  if (timezoneEl) {
    timezoneEl.value = Intl.DateTimeFormat().resolvedOptions().timeZone;
  }
});

document.addEventListener('click', function(e) {
  const button = e.target
  const copyTargetEl = button.getAttribute('data-copy-target')
  if (!copyTargetEl) return

  const target = document.querySelector(copyTargetEl)
  if (!target) return

  navigator.clipboard.writeText(target.innerText).then(() => {
    button.classList.add('copied')
    setTimeout(() => button.classList.remove('copied'), 1500);
  });
});

document.addEventListener('click', function(e) {
  const confirmMsg = e.target.getAttribute('data-confirm')
  if (!confirmMsg) return

  const result = confirm(confirmMsg)
  if (!result) {
    e.preventDefault()
  }
})
