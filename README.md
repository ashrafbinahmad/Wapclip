# 📋 WapClip — Ultra-lightweight WhatsApp Clipboard Daemon

WapClip is a high-performance cross-device clipboard synchronization tool built purely natively in Go. It turns a dedicated WhatsApp group into a zero-latency payload tunnel, delivering text and media from your phone to your computer instantly.

---

### 💡 Why WapClip?

We built WapClip because **WhatsApp Web is too slow** for simple tasks.

If you just need to copy a verification code, a URL, or a snippet of text from your phone to your laptop, waiting 10–15 seconds for WhatsApp Web to "Organize Messages" is frustrating. WapClip runs as a silent, invisible daemon that listens for messages in the background. **The moment you send it to the group on your phone, it's on your computer's clipboard.** No browser tab required.

---

### ⚡ Performance & Resource Usage

WapClip is engineered for users who care about system resources:

- **Zero-Latency:** Uses pure WebSockets (via `whatsmeow`). Messages are delivered to your clipboard in milliseconds.
- **Ultra-Leant RAM:** Typically uses only **15MB – 30MB** of memory. (Compare this to 500MB+ for a chrome tab).
- **CPU Friendly:** 0% CPU usage when idle. Built natively in Go with no heavy runtimes (Electron/Java/Node).
- **Rolling Storage:** Automatically manages your media downloads. You can set a limit (e.g., keep only the last 10 images) to prevent disk bloat.

---

### 🚀 Getting Started

#### 1. Setup & Authentication

Run the initialization command. It will show a QR code in your terminal. Scan it with your WhatsApp (Linked Devices).

```bash
npm run init
```

_During init, you can configure if you want to save media files and set your rolling file limit._

#### 2. Start the Daemon

Once initialized, the daemon usually starts automatically. You can ensure it is running or restart it with:

```bash
npm run start
```

_The app will appear in your system tray (bottom-right on Windows, top-right on macOS)._

---

### 🛠️ How to Use

1.  **On your Phone:** Open WhatsApp and find the group named `WapClip by PeerHop`.
2.  **Send Text:** Type or paste anything into the group.
3.  **On your Computer:** The text is already on your clipboard! Just press `Ctrl+V` (or `Cmd+V`).
4.  **Send Media:** Send an image or document to the group. If enabled, it will appear in `~/Downloads/WapClip` instantly.

---

### 🖱️ Tray Menu Features

Right-click the **📋** icon in your system tray to:

- **Show Info:** See your current session details.
- **Toggle Media:** Instantly enable/disable saving images/videos to your disk.
- **Stop:** Exit the daemon completely.

---

### 🗒️ Advanced Commands

- **`npm run log`**: View real-time telemetry and sync activity logs.
- **`npm run build`**: Compile a platform-specific binary to `dist/` (strips debug symbols for even smaller size).
- **`npm run stop`**: Kill the background process.
- **`npm run cleanup`**: Unregister from OS startup and wipe all configuration safely.

---

### 🛠️ Core Tech Stack

- **Engine**: `go.mau.fi/whatsmeow` (Native WhatsApp logic)
- **Database**: `modernc.org/sqlite` (Zero-config local storage)
- **System**: `golang.design/x/clipboard` & `getlantern/systray`

_Built with speed in mind. ⚡_
