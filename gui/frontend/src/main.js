// State
let selectedPath = "";
let isFolder = false;

// DOM elements
const ffmpegWarning = document.getElementById("ffmpeg-warning");
const dropZone = document.getElementById("drop-zone");
const btnFile = document.getElementById("btn-file");
const btnFolder = document.getElementById("btn-folder");
const selectedInput = document.getElementById("selected-input");
const inputIcon = document.getElementById("input-icon");
const inputPath = document.getElementById("input-path");
const btnClear = document.getElementById("btn-clear");
const settings = document.getElementById("settings");
const crfSlider = document.getElementById("crf");
const crfValue = document.getElementById("crf-value");
const presetSelect = document.getElementById("preset");
const workersSetting = document.getElementById("workers-setting");
const workersInput = document.getElementById("workers");
const actions = document.getElementById("actions");
const btnCompress = document.getElementById("btn-compress");
const btnCancel = document.getElementById("btn-cancel");
const progress = document.getElementById("progress");
const progressText = document.getElementById("progress-text");
const progressCount = document.getElementById("progress-count");
const progressFill = document.getElementById("progress-fill");
const progressFile = document.getElementById("progress-file");
const results = document.getElementById("results");
const resultsList = document.getElementById("results-list");
const resultsErrors = document.getElementById("results-errors");

// Check ffmpeg on startup
async function init() {
  try {
    const ok = await window.go.main.App.CheckFFmpeg();
    if (!ok) {
      ffmpegWarning.classList.remove("hidden");
    }
  } catch (e) {
    ffmpegWarning.classList.remove("hidden");
  }
}

// CRF slider
crfSlider.addEventListener("input", () => {
  crfValue.textContent = crfSlider.value;
});

// File picker
btnFile.addEventListener("click", async (e) => {
  e.stopPropagation();
  try {
    const path = await window.go.main.App.SelectFile();
    if (path) selectInput(path, false);
  } catch (e) {
    console.error(e);
  }
});

// Folder picker
btnFolder.addEventListener("click", async (e) => {
  e.stopPropagation();
  try {
    const path = await window.go.main.App.SelectFolder();
    if (path) selectInput(path, true);
  } catch (e) {
    console.error(e);
  }
});

// Clear selection
btnClear.addEventListener("click", () => {
  clearSelection();
});

function selectInput(path, folder) {
  selectedPath = path;
  isFolder = folder;
  inputIcon.textContent = folder ? "📁" : "📄";
  inputPath.textContent = path;
  selectedInput.classList.remove("hidden");
  settings.classList.remove("hidden");
  actions.classList.remove("hidden");
  if (folder) {
    workersSetting.classList.remove("hidden");
  } else {
    workersSetting.classList.add("hidden");
  }
  results.classList.add("hidden");
  resultsErrors.classList.add("hidden");
}

function clearSelection() {
  selectedPath = "";
  isFolder = false;
  selectedInput.classList.add("hidden");
  settings.classList.add("hidden");
  actions.classList.add("hidden");
  results.classList.add("hidden");
  progress.classList.add("hidden");
}

// Compress
btnCompress.addEventListener("click", async () => {
  if (!selectedPath) return;

  const crf = parseInt(crfSlider.value);
  const preset = presetSelect.value;

  // Show progress, hide results
  btnCompress.classList.add("hidden");
  btnCancel.classList.remove("hidden");
  progress.classList.remove("hidden");
  results.classList.add("hidden");
  progressFill.style.width = "0%";
  progressFile.textContent = "";
  progressCount.textContent = "";

  try {
    let response;
    if (isFolder) {
      const workers = parseInt(workersInput.value) || 2;
      progressText.textContent = "Compressing folder...";
      response = await window.go.main.App.CompressFolder(selectedPath, crf, preset, workers);
    } else {
      progressText.textContent = "Compressing...";
      progressCount.textContent = "0%";
      response = await window.go.main.App.CompressFile(selectedPath, crf, preset);
    }
    showResults(response);
  } catch (e) {
    showError(e.toString());
  } finally {
    btnCancel.classList.add("hidden");
    btnCompress.classList.remove("hidden");
    progress.classList.add("hidden");
  }
});

// Cancel
btnCancel.addEventListener("click", () => {
  try {
    window.go.main.App.Cancel();
  } catch (e) {
    console.error(e);
  }
});

// Progress events from backend
window.runtime.EventsOn("compress:progress", (data) => {
  const { file, done, total, status } = data;
  progressFile.textContent = file;

  if (total > 0) {
    const pct = Math.round((done / total) * 100);
    progressFill.style.width = pct + "%";
    progressCount.textContent = done + " / " + total;
  }

  if (status === "compressing") {
    progressText.textContent = total > 1 ? "Compressing folder..." : "Compressing...";
  }
});

// Per-file percent progress from ffmpeg
window.runtime.EventsOn("compress:percent", (data) => {
  const { file, percent } = data;
  const pct = Math.round(percent);
  progressFile.textContent = file;
  progressFill.style.width = pct + "%";
  progressCount.textContent = pct + "%";
});

function showResults(response) {
  // Clear previous results
  while (resultsList.firstChild) {
    resultsList.removeChild(resultsList.firstChild);
  }
  while (resultsErrors.firstChild) {
    resultsErrors.removeChild(resultsErrors.firstChild);
  }

  if (response.results && response.results.length > 0) {
    for (const r of response.results) {
      const item = document.createElement("div");
      item.className = "result-item";

      const nameSpan = document.createElement("span");
      nameSpan.className = "result-name";
      const fileName = r.outputPath.split("/").pop().split("\\").pop();
      nameSpan.textContent = fileName;
      nameSpan.title = r.outputPath;

      const statsDiv = document.createElement("div");
      statsDiv.className = "result-stats";

      const sizeDiv = document.createElement("div");
      sizeDiv.textContent = r.inputSize.toFixed(1) + " MB → " + r.outputSize.toFixed(1) + " MB";

      const reductionDiv = document.createElement("div");
      reductionDiv.className = "result-reduction";
      reductionDiv.textContent = r.reduction.toFixed(1) + "% smaller";

      statsDiv.appendChild(sizeDiv);
      statsDiv.appendChild(reductionDiv);
      item.appendChild(nameSpan);
      item.appendChild(statsDiv);
      resultsList.appendChild(item);
    }
  }

  if (response.errors && response.errors.length > 0) {
    const strong = document.createElement("strong");
    strong.textContent = "Errors:";
    resultsErrors.appendChild(strong);
    for (const err of response.errors) {
      resultsErrors.appendChild(document.createElement("br"));
      resultsErrors.appendChild(document.createTextNode(err));
    }
    resultsErrors.classList.remove("hidden");
  } else {
    resultsErrors.classList.add("hidden");
  }

  results.classList.remove("hidden");
}

function showError(msg) {
  while (resultsList.firstChild) {
    resultsList.removeChild(resultsList.firstChild);
  }
  while (resultsErrors.firstChild) {
    resultsErrors.removeChild(resultsErrors.firstChild);
  }
  resultsErrors.appendChild(document.createTextNode(msg));
  resultsErrors.classList.remove("hidden");
  results.classList.remove("hidden");
}

// Drag and drop (visual feedback only — Wails doesn't support drop events with file paths)
dropZone.addEventListener("dragover", (e) => {
  e.preventDefault();
  dropZone.classList.add("drag-over");
});
dropZone.addEventListener("dragleave", () => {
  dropZone.classList.remove("drag-over");
});
dropZone.addEventListener("drop", (e) => {
  e.preventDefault();
  dropZone.classList.remove("drag-over");
});

init();
