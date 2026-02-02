# SRPS → Arch Linux Feasibility Report

Date: 2026-02-02
Host: Arch Linux (local package queries)

## Summary
**Medium** – core logic is portable, but install.sh needs a package-manager abstraction (pacman support), plus a few Arch-specific package name changes and a safer update/upgrade strategy.

## Package Availability (Arch)
- `ananicy-cpp` **is available** in the official repos: `extra/ananicy-cpp`.
- AUR also provides variants (e.g., `ananicy-cpp-git`, non-systemd service variants).
- `base-devel` **is installed** on this machine; this is the Arch equivalent of Debian’s `build-essential`.

Commands run:
- `pacman -Ss ananicy` → `extra/ananicy-cpp 1.1.1-9`
- `yay -Ss ananicy-cpp` → AUR variants + `extra/ananicy-cpp`
- `pacman -Ss base-devel` → `core/base-devel 1-2 [installed]`

## Inventory: apt/apt-get Calls in `install.sh`
1. Detect apt in `detect_system()` (line ~185):
   - `command -v apt-get` check
   - Hard fail if missing.
2. `apt_install()` function (lines ~260-285):
   - `sudo apt-get update -qq`
   - `sudo apt-get install -y -qq <pkgs>`
   - `DEBIAN_FRONTEND=noninteractive` used.
3. `install_sysmon_go()` (lines ~1096-1100):
   - `sudo apt-get update -qq`
   - `sudo apt-get install -y -qq golang-go`
4. Final summary banner (line ~1920):
   - `Package: apt-get`

## Debian-Specific Assumptions / Paths
- Package manager detection hard-coded to `apt-get`.
- `DEBIAN_FRONTEND=noninteractive` env is Debian/Ubuntu specific.
- Package names are Debian-specific:
  - `build-essential`
  - `libsystemd-dev`
  - `libfmt-dev`
  - `libspdlog-dev`
  - `nlohmann-json3-dev`
  - `pkg-config`
  - `golang-go`
- Summary output assumes `apt-get`.

Paths like `/etc/ananicy.d`, `/etc/sysctl.d`, `/etc/systemd/system.conf.d`, and `/etc/bash_completion.d` are **valid on Arch** (present on this host).

## Package Mapping (Debian → Arch)
| Debian/Ubuntu package | Arch package |
|---|---|
| `build-essential` | `base-devel` (group) |
| `libsystemd-dev` | `systemd` (headers included) |
| `libfmt-dev` | `fmt` |
| `libspdlog-dev` | `spdlog` |
| `nlohmann-json3-dev` | `nlohmann-json` |
| `pkg-config` | `pkgconf` |
| `golang-go` | `go` |
| `git` | `git` |
| `cmake` | `cmake` |
| `util-linux` | `util-linux` |

## Required Changes (High-Level)
- Replace `apt-get` detection with multi-PM detection (at least `pacman`).
- Replace `apt_install()` with a `pacman_install()` (or generic `pkg_install`) that uses:
  - `pacman -S --needed --noconfirm <pkgs>`
  - Optionally `pacman -Sy` once per run (avoid partial upgrades).
- Add Arch-specific package name mapping for all build deps.
- Update sysmoni Go fallback to install Go via `pacman -S --needed go`.
- Update final summary line to report `pacman` when on Arch.

## Recommended Approach
1. **Introduce a package-manager abstraction**:
   - Detect `pacman` and set `PKG_MANAGER=pacman` (fallback to `apt` for Debian).
   - Replace `apt_install` with `pkg_install` that routes to `pacman` or `apt`.
2. **Prefer repo `ananicy-cpp` on Arch**:
   - If `pacman -Qi ananicy-cpp` not installed, install it via pacman instead of building from source.
   - Keep the source build path as a fallback for other distros.
3. **Add Arch package map** for build dependencies and Go.
4. **Keep systemd/sysctl logic unchanged**; it’s portable on Arch.

## Risks / Unknowns
- **Pacman update strategy**: using `pacman -Sy` without full upgrade can cause partial-upgrade issues. Decide whether SRPS should do `pacman -Syu --noconfirm` (more intrusive) or avoid refreshing entirely.
- **Non-systemd Arch setups**: AUR variants exist, but SRPS assumes systemd for service control. If the target Arch user is non-systemd, extra branching is needed.
- **Kernel sysctl support**: `bbr` and `fq` may be unavailable depending on kernel config; existing script already warns on sysctl failures.

