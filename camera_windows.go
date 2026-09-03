//go:build windows

package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"runtime"
	"sort"
	"sync"

	camera "github.com/Kirizu-Official/windows-camera-go/camera/v1"
	"github.com/Kirizu-Official/windows-camera-go/windows/guid"
)

type CameraInfo struct {
	Name string
	ID   string
}

type CameraSession struct {
	stop chan struct{}
	done chan struct{}
	once sync.Once
}

func ListCameras() ([]CameraInfo, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := camera.Init(); err != nil {
		return nil, err
	}
	defer camera.Shutdown()
	devices, err := camera.EnumDevice()
	if err != nil {
		return nil, err
	}
	result := make([]CameraInfo, 0, len(devices))
	for _, device := range devices {
		result = append(result, CameraInfo{Name: device.Name, ID: device.SymbolLink})
	}
	return result, nil
}

func StartCamera(id string, onFrame func(image.Image), onError func(error)) *CameraSession {
	session := &CameraSession{stop: make(chan struct{}), done: make(chan struct{})}
	go func() {
		defer close(session.done)
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		if err := camera.Init(); err != nil {
			onError(err)
			return
		}
		defer camera.Shutdown()
		device, err := camera.OpenDevice(id)
		if err != nil {
			onError(err)
			return
		}
		defer device.CloseDevice()
		formats, err := device.EnumerateCaptureFormats()
		if err != nil {
			onError(err)
			return
		}
		format := bestMJPEGFormat(formats)
		if format == nil {
			onError(fmt.Errorf("camera does not offer an MJPEG capture format"))
			return
		}
		capture, err := device.StartCapture(format)
		if err != nil {
			onError(err)
			return
		}
		for {
			select {
			case <-session.stop:
				return
			default:
			}
			frame, frameErr := capture.GetFrame()
			if frameErr != nil {
				onError(frameErr)
				return
			}
			buffer, bufferErr := device.ParseSampleToBuffer(frame.PpSample)
			if bufferErr != nil {
				_ = frame.Release()
				onError(bufferErr)
				return
			}
			data := append([]byte(nil), buffer.Buffer[:buffer.Length]...)
			buffer.Release()
			_ = frame.Release()
			decoded, decodeErr := image.Decode(bytes.NewReader(data))
			if decodeErr == nil {
				onFrame(decoded)
			}
		}
	}()
	return session
}

func (s *CameraSession) Stop() {
	if s == nil {
		return
	}
	s.once.Do(func() { close(s.stop) })
}

func bestMJPEGFormat(formats []*camera.CaptureFormats) *camera.CaptureFormats {
	compatible := make([]*camera.CaptureFormats, 0)
	for _, format := range formats {
		if format.SubType != nil && format.SubType.IsMatch(&guid.SubTypeMediaSubTypeMJPG) {
			compatible = append(compatible, format)
		}
	}
	sort.Slice(compatible, func(i, j int) bool {
		iArea := compatible[i].Width * compatible[i].Height
		jArea := compatible[j].Width * compatible[j].Height
		iPreferred := iArea <= 1920*1080
		jPreferred := jArea <= 1920*1080
		if iPreferred != jPreferred {
			return iPreferred
		}
		return iArea > jArea
	})
	if len(compatible) == 0 {
		return nil
	}
	return compatible[0]
}
