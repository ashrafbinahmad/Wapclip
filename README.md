# WapClip — WhatsApp Clipboard Daemon

WapClip is a cross-device clipboard synchronization tool built purely natively in Go. It turns an isolated WhatsApp group (`WapClip by PeerHop`) into a high-speed payload tunnel to seamlessly transport your clipboard data instantly to any computer running the daemon.

### ⚠️ Prerequisite
Since the backend runs entirely via [whatsmeow](https://github.com/tulir/whatsmeow), this package ships directly as Go source code. You will need [Go installed](https://go.dev/doc/install) to execute and build it seamlessly. 

## 🚀 Getting Started 

Execute the commands using the pre-defined NPM scripts:

### 1. Initialize
Sets up the WhatsApp connection, automatically creates the secure relay group mapping, registers the OS background daemon sequentially, and boots it silently.
```bash
npm run init
```

### 2. View Telemetry
Since the native daemon operates completely silently in the background, access real-time logging telemetry using:
```bash
npm run log
```

### 3. Uninstall
Want to cleanly eliminate WapClip? This unbinds it securely from your OS (Windows Registry / Apple LaunchAgents / Linux autostarts), shuts down currently spinning background components natively, and strictly strips its memory parameters.
```bash
npm run stop
```

### Compilation (Optional)
If you wish to compile fixed platform binaries rather than relying purely on script engines:
```bash
npm run build
```

---

**Core Dependencies:**
* `go.mau.fi/whatsmeow`: Transport protocol engine
* `modernc.org/sqlite`: Native Go SQL processor
* `golang.design/x/clipboard` / `github.com/gen2brain/beeep`: UX abstractions

*Built purely natively in Go. ⚡*
