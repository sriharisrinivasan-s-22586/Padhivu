# Setup Guide

This file explains how to:

1. run Padhivu locally from source (desktop)
2. build local installable packages on Windows/macOS/Linux
3. build mobile artifacts for Android/iOS/Harmony

## Prerequisites

- Node.js (LTS)
- pnpm
- Go (latest stable)
- CGO toolchain (`CGO_ENABLED=1` required)
- Git

Install pnpm globally:

```bash
npm install -g pnpm
```

Install app dependencies once:

```bash
cd app
pnpm install
pnpm install electron
```

---

## 1) Run locally from source (Desktop)

In development mode, start kernel manually first, then UI/Electron.

### Windows

Build kernel:

```bat
cd kernel
set CGO_ENABLED=1
go build --tags fts5 -o ../app/kernel/SiYuan-Kernel.exe .
```

Start kernel:

```bat
cd app\kernel
SiYuan-Kernel.exe --wd=.. --mode=dev
```

Start UI + Electron in two more terminals:

```bat
cd app
pnpm run dev
```

```bat
cd app
pnpm run start
```

### macOS

Build kernel:

```bash
cd kernel
CGO_ENABLED=1 go build --tags fts5 -o ../app/kernel/SiYuan-Kernel .
```

Start kernel:

```bash
cd app/kernel
./SiYuan-Kernel --wd=.. --mode=dev
```

Start UI + Electron in two more terminals:

```bash
cd app
pnpm run dev
```

```bash
cd app
pnpm run start
```

### Linux

Build kernel:

```bash
cd kernel
CGO_ENABLED=1 go build --tags fts5 -o ../app/kernel/SiYuan-Kernel .
```

Start kernel:

```bash
cd app/kernel
./SiYuan-Kernel --wd=.. --mode=dev
```

Start UI + Electron in two more terminals:

```bash
cd app
pnpm run dev
```

```bash
cd app
pnpm run start
```

---

## 2) Build local desktop packages

### Windows package (`.exe` installer)

Run on Windows machine.

Install `goversioninfo` (first time):

```bat
go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
```

Build installers:

```bat
scripts\win-build.bat --target amd64
```

```bat
scripts\win-build.bat --target arm64
```

```bat
scripts\win-build.bat --target all
```

Expected output:

- `app/build/` (installer artifacts, e.g. `padhivu-<version>-win.exe`)
- `app/build/win-unpacked/` (unpacked runnable app)

### macOS package (`.dmg`/mac targets)

```bash
./scripts/darwin-build.sh --target amd64
```

```bash
./scripts/darwin-build.sh --target arm64
```

```bash
./scripts/darwin-build.sh --target all
```

Expected output:

- `app/build/`

### Linux package

```bash
./scripts/linux-build.sh --target amd64
```

```bash
./scripts/linux-build.sh --target arm64
```

```bash
./scripts/linux-build.sh --target all
```

Expected output:

- `app/build/`

---

## 3) Android build

From `kernel/`:

```bash
gomobile bind --tags fts5 -ldflags "-s -w" -v -o kernel.aar -target=android/arm64 -androidapi 26 ./mobile/
```

On Windows, set encoding env var first:

```bat
set JAVA_TOOL_OPTIONS=-Dfile.encoding=UTF-8
```

Output artifact:

- `kernel/kernel.aar`

Reference app repo:

- https://github.com/siyuan-note/siyuan-android

---

## 4) iOS build

From `kernel/`:

```bash
gomobile bind --tags fts5 -ldflags '-s -w' -v -o ./ios/iosk.xcframework -target=ios ./mobile/
```

Output artifact:

- `kernel/ios/iosk.xcframework`

Reference app repo:

- https://github.com/siyuan-note/siyuan-ios

---

## 5) Harmony build

Harmony build is Linux-oriented and needs Harmony SDK plus required Go runtime patches.

```bash
cd kernel/harmony
./build.sh
```

For Windows emulator:

```bash
./build-win.sh
```

Reference app repo:

- https://github.com/siyuan-note/siyuan-harmony

---

## Troubleshooting

- If startup shows `no such module: fts5`, rebuild kernel with `--tags fts5`.
- If Electron download is slow/fails in some regions:
  - macOS/Linux: `ELECTRON_MIRROR=https://npmmirror.com/mirrors/electron/`
  - Windows: `SET ELECTRON_MIRROR=https://npmmirror.com/mirrors/electron/`
- If kernel does not start, verify binary location:
  - macOS/Linux: `app/kernel/SiYuan-Kernel`
  - Windows: `app/kernel/SiYuan-Kernel.exe`
