package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ashrafbinahmad/wapclip/internal/config"
	"github.com/ashrafbinahmad/wapclip/internal/startup"
	"github.com/ashrafbinahmad/wapclip/internal/whatsapp"
	
	"github.com/gen2brain/beeep"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"golang.design/x/clipboard"
)

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: wapclip <init|daemon>\n")
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "init":
		runInit()
	case "daemon":
		runDaemon()
	case "log":
		runLog()
	case "stop":
		runStop()
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

	fmt.Println("Connected to WhatsApp, creating group...")
	
	req := whatsmeow.ReqCreateGroup{
		Name:         "WapClip by PeerHop",
		Participants: []types.JID{},
	}
	var groupInfo *types.GroupInfo
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

	cfg := config.Config{
		GroupJID: groupInfo.JID.String(),
	}

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Group created and configured: %s\n", cfg.GroupJID)

	if err := startup.Register(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register OS startup: %v\n", err)
	}

	fmt.Println("Success! WapClip initialized. The daemon is registered for OS startup.")
	client.Disconnect()

	// Automatically start daemon so user doesn't have to wait for next login
	exe, _ := os.Executable()
	var runCmd *exec.Cmd
	if strings.Contains(exe, "go-build") || strings.HasPrefix(filepath.Base(exe), "main") || strings.HasPrefix(filepath.Base(exe), "wapclip") == false {
		runCmd = exec.Command("go", "run", "./cmd/", "daemon")
	} else {
		runCmd = exec.Command(exe, "daemon")
	}

	if err := runCmd.Start(); err != nil {
		fmt.Printf("Warning: failed to start daemon immediately: %v\n", err)
		fmt.Println("You can start it manually with: npm run daemon")
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

	cfg, err := config.Load()
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			fmt.Fprintln(os.Stderr, "Run 'wapclip init' first")
		} else {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		}
		os.Exit(1)
	}

	err = clipboard.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Clipboard init error: %v\n", err)
	}

	client, err := whatsapp.InitClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize client: %v\n", err)
		os.Exit(1)
	}

	if client.Store.ID == nil {
		fmt.Fprintln(os.Stderr, "Not logged in. Run 'wapclip init' first")
		os.Exit(1)
	}

	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			if v.Info.Chat.String() == cfg.GroupJID {
				text := v.Message.GetConversation()
				if text == "" && v.Message.GetExtendedTextMessage() != nil {
					text = v.Message.GetExtendedTextMessage().GetText()
				}
				if text != "" {
					clipboard.Write(clipboard.FmtText, []byte(text))
					beeep.Notify("WapClip 📋", "Copied: "+truncate(text, 60), "")
				}
			}
		case *events.LoggedOut:
			fmt.Fprintln(os.Stderr, "WhatsApp session logged out. Run 'wapclip init' to re-authenticate.")
			os.Exit(1)
		}
	})

	err = client.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to WhatsApp: %v\n", err)
		os.Exit(1)
	}

	select {}
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
			}
		}
	}

	fmt.Println("Deregistering from OS startup...")
	err = startup.Deregister()
	if err != nil {
		fmt.Printf("Warning: failed to deregister startup: %v\n", err)
	}

	fmt.Println("Removing ~/.wapclip configuration folder...")
	configPath := filepath.Join(home, ".wapclip")
	os.RemoveAll(configPath)

	fmt.Println("WapClip has been completely removed.")
}
