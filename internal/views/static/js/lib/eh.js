function eh(action, data) {
  var data = data || {};

  var projectKey = document.querySelector('meta[name="eh_project_key"]')?.getAttribute("content");
  projectKey ||= window.EH_PROJECT_KEY;
  if (!projectKey) return;

  var baseURL = document.querySelector('meta[name="eh_url"]')?.getAttribute("content");
  baseURL ||= "https://eh.robyparr.com";

  var xhr = new XMLHttpRequest();
  xhr.open("POST", baseURL + '/api/events', true);
  xhr.setRequestHeader("Content-type", "application/json; charset=UTF-8");
  xhr.setRequestHeader("Authorization", "Bearer " + projectKey);

  var referrer = document.referrer === '' ? undefined : new URL(document.referrer).host;
  var jsonBody = JSON.stringify({
    action: action,
    count: data.count || 1,
    referrer: referrer === window.location.host ? undefined : referrer
  });
  xhr.send(jsonBody);
}
