package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"tinygo.org/x/bluetooth"

	"DoorLockMon/messages"
	"DoorLockMon/realm"
)

var (
	adapter        = bluetooth.DefaultAdapter
	serviceUUID, _ = bluetooth.ParseUUID("a0b40001-926d-4d61-98df-8c5c62ee53b3")
	requestUUID, _ = bluetooth.ParseUUID("a0b40002-926d-4d61-98df-8c5c62ee53b4")
	reportUUID, _  = bluetooth.ParseUUID("a0b40002-926d-4d61-98df-8c5c62ee53b5")
)

type Notify struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body"`
	Data  string `json:"data,omitempty"`
}

func PushNotify(ctx context.Context, info *Notify) error {
	log.Println("push:", info)
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(info); err != nil {
		return err
	}
	args := []string{
		"push", "-title", info.Title,
	}
	if info.Data != "" {
		args = append(args, "-data", info.Data)
	}
	args = append(args, info.Body)
	cmd := exec.CommandContext(ctx, "notify-tool", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func ScanAndConnect(ctx context.Context) error {
	result := make(chan bluetooth.ScanResult)
	errorCh := make(chan error)
	go func() {
		log.Println("scan start")
		defer log.Println("scan stop")
		if err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
			if device.LocalName() == "Door-Lock" {
				adapter.StopScan()
				result <- device
			}
		}); err != nil {
			errorCh <- fmt.Errorf("scan filed: %w", err)
			adapter.StopScan()
		}
	}()
	lastRandom := make([]byte, 4)
	select {
	case <-ctx.Done():
		return nil
	case res := <-result:
		time.Sleep(time.Second)
		log.Println("connect:", res.Address)
		if len(res.ManufacturerData()) > 0 {
			copy(lastRandom, res.ManufacturerData()[0].Data)
		}
		dev, err := adapter.Connect(res.Address, bluetooth.ConnectionParams{
			ConnectionTimeout: bluetooth.NewDuration(10 * time.Second),
		})
		if err != nil {
			return fmt.Errorf("connect filed: %w", err)
		}
		defer log.Println("disconnect:", res.Address)
		svcs, err := dev.DiscoverServices([]bluetooth.UUID{serviceUUID})
		if err != nil {
			return fmt.Errorf("discover services filed: %w", err)
		}
		if len(svcs) == 0 {
			return fmt.Errorf("not found service")
		}
		svc := svcs[0]
		chars, err := svc.DiscoverCharacteristics([]bluetooth.UUID{
			requestUUID, reportUUID,
		})
		if err != nil {
			return fmt.Errorf("discover characteristics filed: %w", err)
		}
		if len(chars) != 2 {
			return fmt.Errorf("not enough characteristics")
		}
		request, report := chars[0], chars[1]
		done := make(chan []byte)
		msg := messages.New()
		report.EnableNotifications(func(b []byte) {
			done <- b
		})
		tick := time.NewTicker(15 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return nil
			case b := <-done:
				if err := msg.Unmarshal(b); err != nil {
					return err
				}
				log.Println(msg)
				title := "アンロック"
				if msg.Rotation == messages.RotationF {
					title = "ロック"
				}
				if err := PushNotify(context.Background(), &Notify{
					Title: title,
					Body:  fmt.Sprintf("バッテリー残量: %4.2fV", msg.Battery),
				}); err != nil {
					return err
				}
				return nil
			case <-tick.C:
				println("write request")
				h := md5.New()
				h.Write([]byte(realm.Value))
				h.Write(lastRandom)
				b := h.Sum(nil)
				if _, err := request.WriteWithoutResponse(b); err != nil {
					return fmt.Errorf("request failed: %w", err)
				}
			}
		}
	}
}

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Enable BLE interface.
	if err := adapter.Enable(); err != nil {
		log.Fatal(err)
	}
	for {
		if err := ScanAndConnect(ctx); err != nil {
			log.Println(err)
		}
	}
}
