function eh(action, data) {
  var data = data || {};

  var projectKey = document.querySelector('meta[name="eh_project_key"]')?.getAttribute('content');
  projectKey ||= window.EH_PROJECT_KEY;
  if (!projectKey) return;

  var xhr = new XMLHttpRequest();
  xhr.open('POST', 'https://localhost:4000/api/events', true);
  // xhr.open('POST', 'https://event-horizon.robyparr.com/api/events', true);
  xhr.setRequestHeader('Content-type', 'application/json; charset=UTF-8');
  xhr.setRequestHeader('Authorization', 'Bearer ' + projectKey);

  var jsonBody = JSON.stringify({ action: action, count: data.count || 1 })
  xhr.send(jsonBody);
}
