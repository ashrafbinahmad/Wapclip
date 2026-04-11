package main

import (
	"bufio"
	"context"
	"fmt"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ashrafbinahmad/wapclip/internal/config"
	"github.com/ashrafbinahmad/wapclip/internal/startup"
	"github.com/ashrafbinahmad/wapclip/internal/whatsapp"
	
	_ "embed"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"golang.design/x/clipboard"
	"google.golang.org/protobuf/proto"
)

//go:embed assets/icon.png
var iconData []byte

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: wapclip <init|start|daemon|log|stop|info>\n")
		os.Exit(1)
	}

	cmd := os.Args[1]
	if cmd != "daemon" {
		attachConsole()
	}

	switch cmd {
	case "init":
		runInit()
	case "start":
		runStart()
	case "daemon":
		runDaemon()
	case "log":
		runLog()
	case "stop":
		runStop()
	case "cleanup":
		runCleanup()
	case "info":
		runInfo()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func runInit() {
	client, err := whatsapp.InitClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize client: %v\n", err)
		os.Exit(1)
	}

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
			os.Exit(1)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Clear screen to avoid terminal flooding from WhatsApp rotating codes
				fmt.Print("\033[H\033[2J")
				fmt.Println("\n📋 WapClip — Scan this QR with WhatsApp:\n")
				config := qrterminal.Config{
					Level:          qrterminal.L,
					Writer:         os.Stdout,
					HalfBlocks:     true,
					BlackChar:      qrterminal.BLACK_BLACK,
					WhiteBlackChar: qrterminal.WHITE_BLACK,
					WhiteChar:      qrterminal.WHITE_WHITE,
					BlackWhiteChar: qrterminal.BLACK_WHITE,
					QuietZone:      1,
				}
				qrterminal.GenerateWithConfig(evt.Code, config)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Connected to WhatsApp, checking for existing group...")
	groups, err := client.GetJoinedGroups(context.Background())
	var groupInfo *types.GroupInfo
	if err == nil {
		for _, g := range groups {
			if g.GroupName.Name == "WapClip by PeerHop" {
				groupInfo = g
				break
			}
		}
	}

	if groupInfo != nil {
		fmt.Printf("Found existing group: %s (%s)\n", groupInfo.GroupName.Name, groupInfo.JID)
	} else {
		fmt.Println("Group not found, creating new group...")
		req := whatsmeow.ReqCreateGroup{
			Name:         "WapClip by PeerHop",
			Participants: []types.JID{},
		}
		for i := 0; i < 5; i++ {
			groupInfo, err = client.CreateGroup(context.Background(), req)
			if err == nil {
				break
			}
			fmt.Printf("Retrying group creation... (%v)\n", err)
			time.Sleep(2 * time.Second)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create group: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Group created: %s\n", groupInfo.JID)
	}

	fmt.Print("Save media files (images/videos/docs) to Downloads? (y/n): ")
	var saveMediaStr string
	fmt.Scanln(&saveMediaStr)
	saveMedia := strings.ToLower(saveMediaStr) == "y"

	mediaLimit := 10
	if saveMedia {
		fmt.Print("Enter max number of media files to keep (default 10): ")
		var limitStr string
		fmt.Scanln(&limitStr)
		if limitStr != "" {
			fmt.Sscanf(limitStr, "%d", &mediaLimit)
		}
	}

	cfg := config.Config{
		GroupJID:   groupInfo.JID.String(),
		SaveMedia:  saveMedia,
		MediaLimit: mediaLimit,
	}

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Configured: Group=%s, SaveMedia=%v, Limit=%d\n", cfg.GroupJID, cfg.SaveMedia, cfg.MediaLimit)

	if err := startup.Register(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register OS startup: %v\n", err)
	}

	fmt.Println("Success! WapClip initialized. The daemon is registered for OS startup.")
	client.Disconnect()

	// Automatically start daemon so user doesn't have to wait for next login
	runStart()
}

func runStart() {
	exe, _ := os.Executable()
	var runCmd *exec.Cmd
	if strings.Contains(exe, "go-build") || strings.HasPrefix(filepath.Base(exe), "main") || strings.HasPrefix(filepath.Base(exe), "wapclip") == false {
		runCmd = exec.Command("go", "run", "./cmd/", "daemon")
	} else {
		runCmd = exec.Command(exe, "daemon")
	}

	setDaemonAttributes(runCmd)

	if err := runCmd.Start(); err != nil {
		fmt.Printf("Warning: failed to start daemon immediately: %v\n", err)
		fmt.Println("You can start it manually with: wapclip start")
	} else {
		fmt.Println("Daemon started successfully in the background!")
	}
}

func runDaemon() {
	home, _ := os.UserHomeDir()
	logPath := filepath.Join(home, ".wapclip", "wapclip.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		os.Stdout = logFile
		os.Stderr = logFile
	}
	pidPath := filepath.Join(home, ".wapclip", "daemon.pid")
	os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)

	systray.Run(onReady, onExit)
}

func onExit() {
	// Cleanup if needed
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("WapClip")
	systray.SetTooltip("WapClip - WhatsApp Clipboard Sync")

	mInfo := systray.AddMenuItem("Show Info", "Show configuration")
	mToggleMedia := systray.AddMenuItem("Save Media: OFF", "Toggle saving media files")
	systray.AddSeparator()
	mStop := systray.AddMenuItem("Stop", "Stop WapClip")

	cfg, err := config.Load()
	if err != nil {
		beeep.Notify("WapClip 📋", "Error: Run 'wapclip init' first", "")
		fmt.Fprintln(os.Stderr, "Error loading config")
		systray.Quit()
		return
	}
	if cfg.SaveMedia {
		mToggleMedia.SetTitle("Save Media: ON")
		mToggleMedia.Check()
	} else {
		mToggleMedia.SetTitle("Save Media: OFF")
		mToggleMedia.Uncheck()
	}

	err = clipboard.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Clipboard init error: %v\n", err)
	}

	client, err := whatsapp.InitClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize client: %v\n", err)
		systray.Quit()
		return
	}

	if client.Store.ID == nil {
		fmt.Fprintln(os.Stderr, "Not logged in. Run 'wapclip init' first")
		systray.Quit()
		return
	}

	var notifyMut sync.Mutex
	var notifyCount int
	var notifyTimer *time.Timer
	var lastText string
	var isReady atomic.Bool

	client.AddEventHandler(func(evt interface{}) {
		if !isReady.Load() {
			return
		}
		switch v := evt.(type) {
		case *events.Message:
			if v.Info.Chat.String() == cfg.GroupJID {
				if v.Info.IsFromMe && v.Info.Sender.Device == client.Store.ID.Device {
					return
				}
				text := v.Message.GetConversation()
				if text == "" && v.Message.GetExtendedTextMessage() != nil {
					text = v.Message.GetExtendedTextMessage().GetText()
				}
				if text != "" {
					clipboard.Write(clipboard.FmtText, []byte(text))
					
					notifyMut.Lock()
					notifyCount++
					lastText = text
					if notifyTimer != nil {
						notifyTimer.Stop()
					}
					notifyTimer = time.AfterFunc(1500*time.Millisecond, func() {
						notifyMut.Lock()
						count := notifyCount
						textToShow := lastText
						notifyCount = 0
						notifyMut.Unlock()

						if count == 1 {
							beeep.Notify("WapClip 📋", "Copied: "+truncate(textToShow, 60), "")
						} else if count > 1 {
							beeep.Notify("WapClip 📋", fmt.Sprintf("Copied %d new items. Latest: %s", count, truncate(textToShow, 40)), "")
						}
					})
					notifyMut.Unlock()
				}

				// Media Handling
				var downloadable whatsmeow.DownloadableMessage
				fileName := ""
				if v.Message.GetImageMessage() != nil {
					downloadable = v.Message.GetImageMessage()
					exts, _ := mime.ExtensionsByType(v.Message.GetImageMessage().GetMimetype())
					ext := ".jpg"
					if len(exts) > 0 {
						ext = exts[0]
					}
					fileName = fmt.Sprintf("image_%s%s", v.Info.ID, ext)
				} else if v.Message.GetVideoMessage() != nil {
					downloadable = v.Message.GetVideoMessage()
					exts, _ := mime.ExtensionsByType(v.Message.GetVideoMessage().GetMimetype())
					ext := ".mp4"
					if len(exts) > 0 {
						ext = exts[0]
					}
					fileName = fmt.Sprintf("video_%s%s", v.Info.ID, ext)
				} else if v.Message.GetDocumentMessage() != nil {
					downloadable = v.Message.GetDocumentMessage()
					fileName = v.Message.GetDocumentMessage().GetFileName()
					if fileName == "" {
						exts, _ := mime.ExtensionsByType(v.Message.GetDocumentMessage().GetMimetype())
						ext := ".dat"
						if len(exts) > 0 {
							ext = exts[0]
						}
						fileName = fmt.Sprintf("doc_%s%s", v.Info.ID, ext)
					}
				}

				if downloadable != nil && cfg.SaveMedia {
					go func(msg whatsmeow.DownloadableMessage, name string) {
						data, err := client.Download(context.Background(), msg)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Failed to download media: %v\n", err)
							return
						}

						home, _ := os.UserHomeDir()
						outputPath := filepath.Join(home, "Downloads", "WapClip")
						os.MkdirAll(outputPath, 0755)

						finalPath := filepath.Join(outputPath, name)
						err = os.WriteFile(finalPath, data, 0644)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Failed to save media: %v\n", err)
							return
						}

						beeep.Notify("WapClip 📂", "Downloaded: "+name, "")
						manageMediaLimit(outputPath, cfg.MediaLimit)
					}(downloadable, fileName)
				}
			}
		case *events.LoggedOut:
			fmt.Fprintln(os.Stderr, "WhatsApp session logged out. Run 'wapclip init' to re-authenticate.")
			systray.Quit()
		}
	})

	err = client.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to WhatsApp: %v\n", err)
		systray.Quit()
		return
	}

	fmt.Println("Daemon started and connected to WhatsApp.")
	beeep.Notify("WapClip 📋", "Daemon is active and syncing clipboard", "")

	targetJID, _ := types.ParseJID(cfg.GroupJID)
	client.SendMessage(context.Background(), targetJID, &waE2E.Message{
		Conversation: proto.String("🚀 WapClip: Daemon is now active and syncing! 📋✨"),
	})

	isReady.Store(true)

	go func() {
		for {
			select {
			case <-mInfo.ClickedCh:
				beeep.Notify("WapClip 📋", "Active Group: "+cfg.GroupJID, "")
				// Also send info message to group
				go client.SendMessage(context.Background(), targetJID, &waE2E.Message{
					Conversation: proto.String("ℹ️ WapClip Status: Service is active. ✅📡"),
				})
			case <-mToggleMedia.ClickedCh:
				cfg.SaveMedia = !cfg.SaveMedia
				if cfg.SaveMedia {
					mToggleMedia.SetTitle("Save Media: ON")
					mToggleMedia.Check()
					beeep.Notify("WapClip 📂", "Media saving enabled", "")
				} else {
					mToggleMedia.SetTitle("Save Media: OFF")
					mToggleMedia.Uncheck()
					beeep.Notify("WapClip 📂", "Media saving disabled", "")
				}
				config.Save(*cfg)
			case <-mStop.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func runLog() {
	home, _ := os.UserHomeDir()
	logPath := filepath.Join(home, ".wapclip", "wapclip.log")
	f, err := os.Open(logPath)
	if err != nil {
		fmt.Printf("App not running or no logs found at %s\n", logPath)
		return
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		fmt.Print(line)
	}
}

func runStop() {
	fmt.Println("Stopping WapClip daemon...")
	home, _ := os.UserHomeDir()
	pidPath := filepath.Join(home, ".wapclip", "daemon.pid")
	pidStr, err := os.ReadFile(pidPath)
	if err == nil {
		var pid int
		fmt.Sscanf(string(pidStr), "%d", &pid)
		if pid > 0 {
			proc, _ := os.FindProcess(pid)
			if proc != nil {
				proc.Kill()
				fmt.Println("Daemon stopped.")
			}
		}
		os.Remove(pidPath)
	}
}

func runCleanup() {
	runStop()
	fmt.Println("Deregistering from OS startup...")
	err := startup.Deregister()
	if err != nil {
		fmt.Printf("Warning: failed to deregister startup: %v\n", err)
	}

	fmt.Println("Removing ~/.wapclip configuration folder...")
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".wapclip")
	os.RemoveAll(configPath)

	fmt.Println("WapClip has been completely removed.")
}

func runInfo() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v (Run 'wapclip init' first)\n", err)
		os.Exit(1)
	}

	fmt.Printf("WapClip Configuration:\n")
	fmt.Printf("- Using Group JID: %s\n", cfg.GroupJID)

	client, err := whatsapp.InitClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize client: %v\n", err)
		os.Exit(1)
	}

	if client.Store.ID == nil {
		fmt.Fprintln(os.Stderr, "Not logged in. Run 'wapclip init' first")
		os.Exit(1)
	}

	err = client.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to WhatsApp: %v\n", err)
		os.Exit(1)
	}
	defer client.Disconnect()

	targetJID, _ := types.ParseJID(cfg.GroupJID)
	_, err = client.SendMessage(context.Background(), targetJID, &waE2E.Message{
		Conversation: proto.String("ℹ️ WapClip Status: This group is actively configured for clipboard synchronization. ✅📡"),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send info message: %v\n", err)
	} else {
		fmt.Println("Successfully sent info message to the group.")
	}
}
func manageMediaLimit(dir string, limit int) {
	if limit <= 0 {
		return
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	if len(files) <= limit {
		return
	}

	// Sort by mod time
	type fileStat struct {
		name string
		time time.Time
	}
	stats := make([]fileStat, 0, len(files))
	for _, f := range files {
		info, err := f.Info()
		if err == nil {
			stats = append(stats, fileStat{f.Name(), info.ModTime()})
		}
	}

	// Simple selection sort or sort package
	for i := 0; i < len(stats); i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[i].time.After(stats[j].time) {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	// Delete oldest
	toDelete := len(stats) - limit
	for i := 0; i < toDelete; i++ {
		os.Remove(filepath.Join(dir, stats[i].name))
	}
}
