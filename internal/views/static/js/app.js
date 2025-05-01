document.addEventListener("DOMContentLoaded", function () {
  setTimezoneFieldDefault();
  initializeCharts();
});

document.addEventListener("click", function (e) {
  const button = e.target;
  const copyTargetEl = button.getAttribute("data-copy-target");
  if (!copyTargetEl) return;

  const target = document.querySelector(copyTargetEl);
  if (!target) return;

  navigator.clipboard.writeText(target.innerText).then(() => {
    button.classList.add("copied");
    setTimeout(() => button.classList.remove("copied"), 1500);
  });
});

document.addEventListener("click", function (e) {
  const confirmMsg = e.target.getAttribute("data-confirm");
  if (!confirmMsg) return;

  const result = confirm(confirmMsg);
  if (!result) {
    e.preventDefault();
  }
});

function setTimezoneFieldDefault() {
  const timezoneEl = document.querySelector('[name="timezone"]');
  if (timezoneEl) {
    timezoneEl.value = Intl.DateTimeFormat().resolvedOptions().timeZone;
  }
}

function initializeCharts() {
  const datasetOptions = {
    line: {
      fill: true,
    },
    doughnut: {
      hoverOffset: 10,
    },
  };

  document.querySelectorAll("canvas[data-chart-data]").forEach((el) => {
    const data = JSON.parse(el.dataset.chartData);
    let chartType = el.dataset.chartType;
    const chartOptions = {
      plugins: {
        legend: { display: false },
        tooltip: { displayColors: false }
      }
    }
    if (chartType === 'horizontal-bar') {
      chartType = 'bar'
      chartOptions.indexAxis = 'y'
    }

    new Chart(el, {
      type: chartType,
      options: chartOptions,
      data: {
        labels: Object.keys(data),
        datasets: [
          {
            label: el.dataset.chartLabel,
            data: Object.values(data),
            ...datasetOptions[chartType],
          },
        ],
      },
    });
  });
}
