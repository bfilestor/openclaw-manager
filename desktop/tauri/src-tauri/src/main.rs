#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use once_cell::sync::Lazy;
use serde::Serialize;
use std::fs;
use std::path::{Path, PathBuf};
use std::process::{Child, Command, Stdio};
use std::sync::Mutex;
use tauri::image::Image;
use tauri::menu::{Menu, MenuItem};
use tauri::tray::{MouseButton, MouseButtonState, TrayIconBuilder, TrayIconEvent};
use tauri::{Manager, WindowEvent};

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

fn dirs_home() -> PathBuf {
    std::env::var_os("HOME")
        .or_else(|| std::env::var_os("USERPROFILE"))
        .map(PathBuf::from)
        .unwrap_or_else(|| PathBuf::from("."))
}

fn manager_home_dir() -> PathBuf {
    let mut p = dirs_home();
    p.push(".openclaw-manager");
    p
}

fn manager_binary_name() -> &'static str {
    if cfg!(target_os = "windows") {
        "managerd.exe"
    } else {
        "managerd"
    }
}

fn manager_binary_path() -> PathBuf {
    if let Ok(v) = std::env::var("MANAGERD_PATH") {
        if !v.trim().is_empty() {
            return PathBuf::from(v);
        }
    }
    let mut p = manager_home_dir();
    p.push(manager_binary_name());
    p
}

fn manager_config_path() -> PathBuf {
    let mut p = manager_home_dir();
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

fn ensure_runtime_files(app: &tauri::AppHandle) -> Result<(), String> {
    let home = manager_home_dir();
    fs::create_dir_all(&home).map_err(|e| format!("create manager home failed: {e}"))?;

    let resolver = app.path();
    let target_bin = manager_binary_path();
    if !target_bin.exists() {
        let resource_name = manager_binary_name();
        if let Ok(src) = resolver.resolve(resource_name, tauri::path::BaseDirectory::Resource) {
            if src.exists() {
                fs::copy(&src, &target_bin).map_err(|e| format!("copy manager binary failed: {e}"))?;
                #[cfg(unix)]
                {
                    use std::os::unix::fs::PermissionsExt;
                    let mut perms = fs::metadata(&target_bin)
                        .map_err(|e| format!("stat copied binary failed: {e}"))?
                        .permissions();
                    perms.set_mode(0o755);
                    fs::set_permissions(&target_bin, perms)
                        .map_err(|e| format!("chmod copied binary failed: {e}"))?;
                }
            }
        }
    }

    let cfg = manager_config_path();
    if !cfg.exists() {
        if let Ok(src) = resolver.resolve("config.template.toml", tauri::path::BaseDirectory::Resource) {
            if src.exists() {
                fs::copy(src, &cfg).map_err(|e| format!("copy config template failed: {e}"))?;
            }
        }
    }

    Ok(())
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

    ensure_runtime_files(&app)?;

    let bin = manager_binary_path();
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

fn has_command(name: &str) -> bool {
    if cfg!(target_os = "windows") {
        Command::new("where")
            .arg(name)
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .status()
            .map(|s| s.success())
            .unwrap_or(false)
    } else {
        Command::new("which")
            .arg(name)
            .stdout(Stdio::null())
            .stderr(Stdio::null())
            .status()
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

fn resolve_tray_icon(app: &tauri::AppHandle) -> Option<Image<'static>> {
    if let Ok(path) = std::env::var("MANAGER_TRAY_ICON_PATH") {
        let p = PathBuf::from(path);
        if p.exists() {
            if let Ok(icon) = Image::from_path(&p) {
                return Some(icon);
            }
        }
    }

    if let Ok(p) = app.path().resolve("tray-icon.png", tauri::path::BaseDirectory::Resource) {
        if p.exists() {
            if let Ok(icon) = Image::from_path(&p) {
                return Some(icon);
            }
        }
    }

    // Fallback to app icon from tauri context to avoid invisible tray icon.
    app.default_window_icon().cloned().map(|icon| icon.to_owned())
}

fn stop_manager_child() {
    if let Ok(mut guard) = MANAGER_CHILD.lock() {
        if let Some(mut child) = guard.take() {
            let _ = child.kill();
        }
    }
}

fn main() {
    tauri::Builder::default()
        .setup(|app| {
            let _ = ensure_runtime_files(app.handle());
            let _ = start_manager(app.handle().clone());

            let show_item = MenuItem::with_id(app, "show", "Show", true, None::<&str>)?;
            let quit_item = MenuItem::with_id(app, "quit", "Quit", true, None::<&str>)?;
            let tray_menu = Menu::with_items(app, &[&show_item, &quit_item])?;

            let app_handle = app.handle().clone();
            let mut tray_builder = TrayIconBuilder::with_id("main");
            if let Some(icon) = resolve_tray_icon(&app_handle) {
                tray_builder = tray_builder.icon(icon);
            }
            tray_builder
                .menu(&tray_menu)
                .tooltip("OpenClaw Manager")
                .on_menu_event(move |app, event| match event.id().as_ref() {
                    "show" => {
                        if let Some(win) = app.get_webview_window("main") {
                            let _ = win.show();
                            let _ = win.set_focus();
                        }
                    }
                    "quit" => {
                        stop_manager_child();
                        app.exit(0);
                    }
                    _ => {}
                })
                .on_tray_icon_event(move |_tray, event| {
                    if let TrayIconEvent::Click {
                        button: MouseButton::Left,
                        button_state: MouseButtonState::Up,
                        ..
                    } = event
                    {
                        if let Some(win) = app_handle.get_webview_window("main") {
                            let _ = win.show();
                            let _ = win.set_focus();
                        }
                    }
                })
                .build(app)?;

            Ok(())
        })
        .on_window_event(|window, event| {
            if let WindowEvent::CloseRequested { api, .. } = event {
                let tray_ready = window.app_handle().tray_by_id("main").is_some();
                if tray_ready {
                    let _ = window.hide();
                    api.prevent_close();
                } else {
                    stop_manager_child();
                }
            }
            if let WindowEvent::Destroyed = event {
                stop_manager_child();
            }
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
