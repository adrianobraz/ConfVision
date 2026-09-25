use std::time::Instant;

use crate::capacity::types::ResourceScope;

#[derive(Debug, Clone, Default)]
pub struct RawHostSample {
    pub cpu_usage_percent: Option<f64>,
    pub cpu_scope: ResourceScope,
    pub memory_usage_percent: Option<f64>,
    pub memory_scope: ResourceScope,
    pub load_1m: Option<f64>,
    pub network_rx_bytes: Option<u64>,
    pub network_tx_bytes: Option<u64>,
    pub network_scope: ResourceScope,
    pub disk_total: Option<u64>,
    pub disk_used: Option<u64>,
    pub disk_free: Option<u64>,
    pub disk_scope: ResourceScope,
}

#[derive(Debug, Default)]
struct CpuCounterState {
    last_idle: u64,
    last_total: u64,
    initialized: bool,
}

#[derive(Debug, Default)]
struct NetworkCounterState {
    last_rx: u64,
    last_tx: u64,
    last_at: Option<Instant>,
}

pub struct ResourceCollector {
    cpu_state: CpuCounterState,
    net_state: NetworkCounterState,
}

impl ResourceCollector {
    pub fn new() -> Self {
        Self {
            cpu_state: CpuCounterState::default(),
            net_state: NetworkCounterState::default(),
        }
    }

    pub fn sample(&mut self) -> RawHostSample {
        let mut out = RawHostSample::default();
        self.sample_cpu(&mut out);
        self.sample_memory(&mut out);
        self.sample_load(&mut out);
        self.sample_network(&mut out);
        self.sample_disk(&mut out);
        out
    }

    fn sample_cpu(&mut self, out: &mut RawHostSample) {
        #[cfg(target_os = "linux")]
        {
            if let Some((idle, total)) = read_proc_stat_cpu() {
                out.cpu_scope = detect_linux_scope();
                if self.cpu_state.initialized && total > self.cpu_state.last_total {
                    let idle_delta = idle.saturating_sub(self.cpu_state.last_idle);
                    let total_delta = total.saturating_sub(self.cpu_state.last_total);
                    if total_delta > 0 {
                        let used = total_delta.saturating_sub(idle_delta) as f64;
                        out.cpu_usage_percent = Some((used / total_delta as f64) * 100.0);
                    }
                }
                self.cpu_state.last_idle = idle;
                self.cpu_state.last_total = total;
                self.cpu_state.initialized = true;
                return;
            }
        }
        out.cpu_scope = ResourceScope::Unavailable;
    }

    fn sample_memory(&mut self, out: &mut RawHostSample) {
        #[cfg(target_os = "linux")]
        {
            if let Some((used_pct, scope)) = read_proc_meminfo() {
                out.memory_usage_percent = Some(used_pct);
                out.memory_scope = scope;
                return;
            }
        }
        out.memory_scope = ResourceScope::Unavailable;
    }

    fn sample_load(&mut self, out: &mut RawHostSample) {
        #[cfg(target_os = "linux")]
        {
            out.load_1m = read_proc_loadavg();
        }
    }

    fn sample_network(&mut self, out: &mut RawHostSample) {
        #[cfg(target_os = "linux")]
        {
            if let Some((rx, tx)) = read_proc_net_dev_totals() {
                out.network_scope = detect_linux_scope();
                out.network_rx_bytes = Some(rx);
                out.network_tx_bytes = Some(tx);
                return;
            }
        }
        out.network_scope = ResourceScope::Unavailable;
    }

    fn sample_disk(&mut self, out: &mut RawHostSample) {
        let _ = self;
        out.disk_scope = ResourceScope::Unavailable;
    }

    pub fn network_rates(
        &mut self,
        rx_bytes: Option<u64>,
        tx_bytes: Option<u64>,
    ) -> (Option<f64>, Option<f64>) {
        let (Some(rx), Some(tx)) = (rx_bytes, tx_bytes) else {
            return (None, None);
        };
        let now = Instant::now();
        let (Some(last_rx), Some(last_tx), Some(last_at)) = (
            Some(self.net_state.last_rx),
            Some(self.net_state.last_tx),
            self.net_state.last_at,
        ) else {
            self.net_state.last_rx = rx;
            self.net_state.last_tx = tx;
            self.net_state.last_at = Some(now);
            return (None, None);
        };

        if rx < last_rx || tx < last_tx {
            self.net_state.last_rx = rx;
            self.net_state.last_tx = tx;
            self.net_state.last_at = Some(now);
            return (None, None);
        }

        let secs = now.duration_since(last_at).as_secs_f64();
        if secs <= 0.0 {
            return (None, None);
        }

        let rx_bps = (rx - last_rx) as f64 / secs;
        let tx_bps = (tx - last_tx) as f64 / secs;
        self.net_state.last_rx = rx;
        self.net_state.last_tx = tx;
        self.net_state.last_at = Some(now);
        (Some(rx_bps), Some(tx_bps))
    }
}

#[cfg(target_os = "linux")]
fn detect_linux_scope() -> ResourceScope {
    if std::path::Path::new("/.dockerenv").exists()
        || std::env::var("KUBERNETES_SERVICE_HOST").is_ok()
    {
        if std::path::Path::new("/sys/fs/cgroup").exists() {
            return ResourceScope::Container;
        }
    }
    ResourceScope::Host
}

#[cfg(target_os = "linux")]
fn read_proc_stat_cpu() -> Option<(u64, u64)> {
    let content = std::fs::read_to_string("/proc/stat").ok()?;
    let first = content.lines().next()?;
    if !first.starts_with("cpu ") {
        return None;
    }
    let parts: Vec<u64> = first
        .split_whitespace()
        .skip(1)
        .filter_map(|p| p.parse().ok())
        .collect();
    if parts.len() < 4 {
        return None;
    }
    let idle = parts[3] + parts.get(4).copied().unwrap_or(0);
    let total: u64 = parts.iter().sum();
    Some((idle, total))
}

#[cfg(target_os = "linux")]
fn read_proc_meminfo() -> Option<(f64, ResourceScope)> {
    let content = std::fs::read_to_string("/proc/meminfo").ok()?;
    let mut total = 0u64;
    let mut available = 0u64;
    for line in content.lines() {
        if line.starts_with("MemTotal:") {
            total = parse_kb(line)?;
        } else if line.starts_with("MemAvailable:") {
            available = parse_kb(line)?;
        }
    }
    if total == 0 {
        return None;
    }
    let used_pct = ((total - available) as f64 / total as f64) * 100.0;
    Some((used_pct.clamp(0.0, 100.0), detect_linux_scope()))
}

#[cfg(target_os = "linux")]
fn parse_kb(line: &str) -> Option<u64> {
    let kb: u64 = line.split_whitespace().nth(1)?.parse().ok()?;
    Some(kb * 1024)
}

#[cfg(target_os = "linux")]
fn read_proc_loadavg() -> Option<f64> {
    let content = std::fs::read_to_string("/proc/loadavg").ok()?;
    content.split_whitespace().next()?.parse().ok()
}

#[cfg(target_os = "linux")]
fn read_proc_net_dev_totals() -> Option<(u64, u64)> {
    let content = std::fs::read_to_string("/proc/net/dev").ok()?;
    let mut rx = 0u64;
    let mut tx = 0u64;
    for line in content.lines().skip(2) {
        let mut parts = line.split_whitespace();
        let iface = parts.next()?.trim_end_matches(':');
        if iface == "lo" {
            continue;
        }
        let r: u64 = parts.next()?.parse().ok()?;
        parts.next();
        parts.next();
        parts.next();
        parts.next();
        parts.next();
        parts.next();
        let t: u64 = parts.next()?.parse().ok()?;
        rx = rx.saturating_add(r);
        tx = tx.saturating_add(t);
    }
    Some((rx, tx))
}
