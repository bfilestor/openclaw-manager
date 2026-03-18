#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use once_cell::sync::Lazy;
use serde::Serialize;
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Stdio};
use std::sync::Mutex;
use tauri::Manager;

static MANAGER_CHILD: Lazy<Mutex<Option<Child>>> = Lazy::new(|| Mutex::new(None));

#[derive(Serialize)]
struct ManagerState {
    running: bool,
    pid: Option<u32>,
}

#[derive(Serialize)]
struct DependencyState {
    openclaw_cli: bool,
    message: String,
}

fn manager_binary_path(app: &tauri::AppHandle) -> PathBuf {
    if let Ok(v) = std::env::var("MANAGERD_PATH") {
        if !v.trim().is_empty() {
            return PathBuf::from(v);
        }
    }

    let resolver = app.path();
    if cfg!(target_os = "windows") {
        if let Ok(p) = resolver.resolve("managerd.exe", tauri::path::BaseDirectory::Resource) {
            return p;
        }
    }
    if let Ok(p) = resolver.resolve("managerd", tauri::path::BaseDirectory::Resource) {
        return p;
    }

    let mut fallback = dirs_home();
    fallback.push(".openclaw-manager");
    if cfg!(target_os = "windows") {
        fallback.push("managerd.exe");
    } else {
        fallback.push("managerd");
    }
    fallback
}

fn dirs_home() -> PathBuf {
    std::env::var_os("HOME")
        .or_else(|| std::env::var_os("USERPROFILE"))
        .map(PathBuf::from)
        .unwrap_or_else(|| PathBuf::from("."))
}

fn manager_config_path() -> PathBuf {
    let mut p = dirs_home();
    p.push(".openclaw-manager");
    p.push("config.toml");
    p
}

fn manager_static_dir(app: &tauri::AppHandle) -> PathBuf {
    if let Ok(v) = std::env::var("MANAGER_STATIC_DIR") {
        if !v.trim().is_empty() {
            return PathBuf::from(v);
        }
    }
    let resolver = app.path();
    resolver
        .resolve("frontend-dist", tauri::path::BaseDirectory::Resource)
        .unwrap_or_else(|_| PathBuf::from("."))
}

#[tauri::command]
fn manager_status() -> ManagerState {
    let guard = MANAGER_CHILD.lock().expect("manager mutex poisoned");
    if let Some(child) = guard.as_ref() {
        ManagerState {
            running: true,
            pid: Some(child.id()),
        }
    } else {
        ManagerState {
            running: false,
            pid: None,
        }
    }
}

#[tauri::command]
fn start_manager(app: tauri::AppHandle) -> Result<ManagerState, String> {
    let mut guard = MANAGER_CHILD.lock().map_err(|_| "manager lock poisoned")?;
    if let Some(child) = guard.as_ref() {
        return Ok(ManagerState {
            running: true,
            pid: Some(child.id()),
        });
    }

    let bin = manager_binary_path(&app);
    if !Path::new(&bin).exists() {
        return Err(format!("manager binary not found: {}", bin.display()));
    }

    let cfg = manager_config_path();
    if !Path::new(&cfg).exists() {
        return Err(format!("manager config not found: {}", cfg.display()));
    }

    let static_dir = manager_static_dir(&app);
    let child = Command::new(&bin)
        .arg("--config")
        .arg(cfg)
        .arg("--static-dir")
        .arg(static_dir)
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .spawn()
        .map_err(|e| format!("start manager failed: {e}"))?;

    let pid = child.id();
    *guard = Some(child);

    Ok(ManagerState {
        running: true,
        pid: Some(pid),
    })
}

#[tauri::command]
fn stop_manager() -> Result<ManagerState, String> {
    let mut guard = MANAGER_CHILD.lock().map_err(|_| "manager lock poisoned")?;
    if let Some(mut child) = guard.take() {
        child.kill().map_err(|e| format!("stop manager failed: {e}"))?;
    }
    Ok(ManagerState {
        running: false,
        pid: None,
    })
}

fn main() {
    tauri::Builder::default()
        .setup(|app| {
            let _ = start_manager(app.handle().clone());
            Ok(())
        })
        .invoke_handler(tauri::generate_handler![manager_status, start_manager, stop_manager])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
           .map(|s| s.success())
            .unwrap_or(false)
    }
}

#[tauri::command]
fn check_dependencies() -> DependencyState {
    let has_openclaw = has_command("openclaw");
    let message = if has_openclaw {
        "openclaw CLI detected".to_string()
    } else {
        "openclaw CLI not detected. Some actions (agents/gateway controls) may fail.".to_string()
    };
    DependencyState {
        openclaw_cli: has_openclaw,
        message,
    }
}

fn main() {
    tauri::Builder::default()
        .setup(|app| {
            let _ = start_manager(app.handle().clone());
            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            manager_status,
            start_manager,
            stop_manager,
            check_dependencies
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
