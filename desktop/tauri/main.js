import { invoke } from "@tauri-apps/api/core";

const statusEl = document.getElementById("status");
const panel = document.getElementById("panel");
const startBtn = document.getElementById("startBtn");
const reloadBtn = document.getElementById("reloadBtn");

function setStatus(text) {
  statusEl.textContent = `Status: ${text}`;
}

async function refreshStatus() {
  try {
    const state = await invoke("manager_status");
    setStatus(state.running ? `running (pid=${state.pid || "-"})` : "stopped");
  } catch (err) {
    setStatus(`status error: ${String(err)}`);
  }
}

startBtn.addEventListener("click", async () => {
  try {
    await invoke("start_manager");
    setStatus("start requested");
    setTimeout(refreshStatus, 500);
    setTimeout(() => {
      panel.src = panel.src;
    }, 1000);
  } catch (err) {
    setStatus(`start failed: ${String(err)}`);
  }
});

reloadBtn.addEventListener("click", () => {
  panel.src = panel.src;
});

refreshStatus();
