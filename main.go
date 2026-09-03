package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/aamkam/photo-with-overlay/internal/config"
	"github.com/aamkam/photo-with-overlay/internal/photo"
)

//go:embed frontend/dist/appicon.png
var resources embed.FS

type fyneUI struct {
	backend *App
	window  fyne.Window
	status  *widget.Label
	preview *canvas.Image
	gallery *fyne.Container

	user, latitude, longitude, address, fontSize, output *widget.Entry
	camera, position, fontFamily                         *widget.Select
	manual, reverse                                      *widget.Check
	start, capture                                       *widget.Button

	cameras       map[string]string
	cameraSession *CameraSession
	frameMu       sync.RWMutex
	latestFrame   image.Image
}

func main() {
	application := app.NewWithID("ca.markham.field-photo")
	if icon, err := resources.ReadFile("frontend/dist/appicon.png"); err == nil {
		application.SetIcon(fyne.NewStaticResource("appicon.png", icon))
	}
	window := application.NewWindow("Field Photo Capture")
	window.Resize(fyne.NewSize(1200, 820))
	ui := newFyneUI(window)
	window.SetContent(ui.content())
	window.SetOnClosed(func() { ui.stopCamera() })
	ui.load()
	window.ShowAndRun()
}

func newFyneUI(window fyne.Window) *fyneUI {
	ui := &fyneUI{backend: NewApp(), window: window, cameras: map[string]string{}}
	ui.status = widget.NewLabel("Ready")
	ui.status.Wrapping = fyne.TextTruncate
	ui.preview = canvas.NewImageFromImage(nil)
	ui.preview.FillMode = canvas.ImageFillContain
	ui.preview.SetMinSize(fyne.NewSize(640, 430))
	ui.gallery = container.NewHBox()
	ui.user = widget.NewEntry()
	ui.latitude = widget.NewEntry()
	ui.longitude = widget.NewEntry()
	ui.address = widget.NewEntry()
	ui.fontSize = widget.NewEntry()
	ui.output = widget.NewEntry()
	ui.camera = widget.NewSelect(nil, nil)
	ui.position = widget.NewSelect([]string{"top-left", "top-right", "bottom-left", "bottom-right"}, nil)
	ui.fontFamily = widget.NewSelect([]string{"Arial", "Segoe UI", "Calibri"}, nil)
	ui.manual = widget.NewCheck("Use manual location (development)", func(bool) { ui.updateLocationMode() })
	ui.reverse = widget.NewCheck("Reverse-geocode GPS location", nil)
	ui.start = widget.NewButton("Start camera", ui.startCamera)
	ui.capture = widget.NewButton("CAPTURE PHOTO", ui.capturePhoto)
	ui.capture.Importance = widget.HighImportance
	ui.capture.Disable()
	return ui
}

func (ui *fyneUI) content() fyne.CanvasObject {
	brandData, _ := resources.ReadFile("frontend/dist/appicon.png")
	brand := canvas.NewImageFromReader(bytes.NewReader(brandData), "appicon.png")
	brand.SetMinSize(fyne.NewSize(42, 42))
	title := widget.NewLabelWithStyle("FIELD PHOTO", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	browse := widget.NewButton("Browse…", func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				ui.setStatus(err)
				return
			}
			if uri != nil {
				ui.output.SetText(uri.Path())
				ui.persist()
				ui.refreshGallery()
			}
		}, ui.window).Show()
	})
	form := container.NewVBox(
		container.NewHBox(brand, title),
		widget.NewLabel("Operator"), ui.user,
		widget.NewLabel("Camera"), ui.camera, ui.start,
		ui.manual,
		widget.NewLabel("Latitude"), ui.latitude,
		widget.NewLabel("Longitude"), ui.longitude,
		widget.NewLabel("Location label (optional)"), ui.address,
		ui.reverse,
		widget.NewLabel("Watermark position"), ui.position,
		widget.NewLabel("Font family"), ui.fontFamily,
		widget.NewLabel("Font size"), ui.fontSize,
		widget.NewLabel("Output folder"), ui.output, browse,
	)
	settingsPanel := container.NewVScroll(form)
	settingsPanel.SetMinSize(fyne.NewSize(330, 500))
	galleryScroll := container.NewHScroll(ui.gallery)
	galleryScroll.SetMinSize(fyne.NewSize(600, 160))
	left := container.NewBorder(nil, galleryScroll, nil, nil, ui.preview)
	body := container.NewHSplit(left, settingsPanel)
	body.Offset = .72
	footer := container.NewBorder(nil, nil, nil, container.NewGridWrap(fyne.NewSize(260, 58), ui.capture), ui.status)
	return container.NewBorder(nil, footer, nil, nil, body)
}

func (ui *fyneUI) load() {
	settings := ui.backend.LoadSettings()
	ui.user.SetText(settings.User)
	ui.output.SetText(settings.OutputFolder)
	ui.position.SetSelected(settings.WatermarkPosition)
	ui.fontFamily.SetSelected(settings.FontFamily)
	ui.fontSize.SetText(strconv.Itoa(settings.FontSize))
	ui.manual.SetChecked(settings.UseManualLocation)
	ui.latitude.SetText(strconv.FormatFloat(settings.ManualLatitude, 'f', 6, 64))
	ui.longitude.SetText(strconv.FormatFloat(settings.ManualLongitude, 'f', 6, 64))
	ui.address.SetText(settings.ManualAddress)
	ui.reverse.SetChecked(settings.ReverseGeocode)
	ui.updateLocationMode()
	ui.refreshGallery()
	go func() {
		cameras, err := ListCameras()
		fyne.Do(func() {
			if err != nil {
				ui.setStatus(fmt.Errorf("camera discovery: %w", err))
				return
			}
			names := make([]string, 0, len(cameras))
			for _, camera := range cameras {
				name := camera.Name
				if _, duplicate := ui.cameras[name]; duplicate {
					name += " (" + strconv.Itoa(len(names)+1) + ")"
				}
				ui.cameras[name] = camera.ID
				names = append(names, name)
				if camera.ID == settings.CameraID {
					ui.camera.SetSelected(name)
				}
			}
			ui.camera.Options = names
			ui.camera.Refresh()
			if ui.camera.Selected == "" && len(names) > 0 {
				ui.camera.SetSelected(names[0])
			}
		})
	}()
}

func (ui *fyneUI) settings() config.Settings {
	fontSize, _ := strconv.Atoi(ui.fontSize.Text)
	latitude, _ := strconv.ParseFloat(ui.latitude.Text, 64)
	longitude, _ := strconv.ParseFloat(ui.longitude.Text, 64)
	return config.Settings{User: ui.user.Text, OutputFolder: ui.output.Text, WatermarkPosition: ui.position.Selected,
		FontFamily: ui.fontFamily.Selected, FontSize: fontSize, UseManualLocation: ui.manual.Checked,
		ManualLatitude: latitude, ManualLongitude: longitude, ManualAddress: ui.address.Text,
		ReverseGeocode: ui.reverse.Checked, CameraID: ui.cameras[ui.camera.Selected]}
}

func (ui *fyneUI) persist() bool {
	if err := ui.backend.SaveSettings(ui.settings()); err != nil {
		ui.setStatus(err)
		return false
	}
	return true
}

func (ui *fyneUI) updateLocationMode() {
	if ui.manual.Checked {
		ui.latitude.Enable()
		ui.longitude.Enable()
		ui.address.Enable()
		ui.reverse.Disable()
	} else {
		ui.latitude.Disable()
		ui.longitude.Disable()
		ui.address.Disable()
		ui.reverse.Enable()
	}
}

func (ui *fyneUI) startCamera() {
	id := ui.cameras[ui.camera.Selected]
	if id == "" {
		ui.setStatus("Select an available camera")
		return
	}
	ui.stopCamera()
	ui.setStatus("Starting camera…")
	ui.start.Disable()
	ui.cameraSession = StartCamera(id, func(frame image.Image) {
		ui.frameMu.Lock()
		ui.latestFrame = frame
		ui.frameMu.Unlock()
		fyne.Do(func() {
			ui.preview.Image = frame
			ui.preview.Refresh()
			ui.capture.Enable()
			ui.start.Enable()
			ui.start.SetText("Restart camera")
			ui.setStatus("Camera ready")
		})
	}, func(err error) {
		fyne.Do(func() { ui.start.Enable(); ui.capture.Disable(); ui.setStatus(fmt.Errorf("camera: %w", err)) })
	})
	ui.persist()
}

func (ui *fyneUI) stopCamera() {
	if ui.cameraSession != nil {
		ui.cameraSession.Stop()
		ui.cameraSession = nil
	}
}

func (ui *fyneUI) capturePhoto() {
	if !ui.persist() {
		return
	}
	ui.frameMu.RLock()
	frame := ui.latestFrame
	ui.frameMu.RUnlock()
	if frame == nil {
		ui.setStatus("Start the camera before capturing")
		return
	}
	settings := ui.settings()
	ui.capture.Disable()
	go func() {
		location, err := ui.getLocation(settings)
		if err != nil {
			fyne.Do(func() { ui.capture.Enable(); ui.setStatus(err) })
			return
		}
		when := time.Now()
		overlaid, err := drawOverlay(frame, settings, when, location)
		if err == nil {
			var encoded bytes.Buffer
			err = jpeg.Encode(&encoded, overlaid, &jpeg.Options{Quality: 92})
			if err == nil {
				_, err = ui.backend.photos.Save(encoded.Bytes(), photo.Metadata{CapturedAt: when, User: settings.User,
					Latitude: location.latitude, Longitude: location.longitude, Accuracy: location.accuracy,
					Location: strings.TrimSpace(location.address + " " + location.roadClue), LocationSource: location.source}, settings.OutputFolder)
			}
		}
		fyne.Do(func() {
			ui.capture.Enable()
			if err != nil {
				ui.setStatus(fmt.Errorf("capture failed: %w", err))
				return
			}
			ui.setStatus("Photo saved")
			ui.refreshGallery()
		})
	}()
}

func (ui *fyneUI) getLocation(settings config.Settings) (captureLocation, error) {
	if settings.UseManualLocation {
		return captureLocation{latitude: settings.ManualLatitude, longitude: settings.ManualLongitude, address: settings.ManualAddress, source: "Manual"}, nil
	}
	fyne.Do(func() { ui.setStatus("Getting Windows location…") })
	latitude, longitude, accuracy, err := windowsLocation()
	if err != nil {
		return captureLocation{}, err
	}
	result := captureLocation{latitude: latitude, longitude: longitude, accuracy: accuracy, source: "Windows Location"}
	if settings.ReverseGeocode {
		fyne.Do(func() { ui.setStatus("Looking up address and nearby roads…") })
		details, lookupErr := ui.backend.ReverseGeocode(latitude, longitude)
		if lookupErr != nil {
			return captureLocation{}, lookupErr
		}
		result.address, result.roadClue = details.Address, details.RoadClue
	}
	return result, nil
}

func (ui *fyneUI) refreshGallery() {
	items, err := ui.backend.ListPhotos(ui.output.Text)
	if err != nil {
		ui.setStatus(err)
		return
	}
	ui.gallery.RemoveAll()
	if len(items) == 0 {
		ui.gallery.Add(widget.NewLabel("Captured photos will appear here"))
		return
	}
	for _, item := range items {
		item := item
		thumbnail := canvas.NewImageFromImage(loadImage(item.Path))
		thumbnail.FillMode = canvas.ImageFillContain
		thumbnail.SetMinSize(fyne.NewSize(180, 105))
		name := widget.NewLabel(item.Name)
		name.Truncation = fyne.TextTruncate
		show := widget.NewButton("Show", func() {
			if err := ui.backend.ShowPhoto(item.Path, ui.output.Text); err != nil {
				ui.setStatus(err)
			}
		})
		remove := widget.NewButton("Delete", func() {
			dialog.NewConfirm("Delete photo", "Permanently delete "+item.Name+"?", func(ok bool) {
				if ok {
					if err := ui.backend.DeletePhoto(item.Path, ui.output.Text); err != nil {
						ui.setStatus(err)
						return
					}
					ui.setStatus("Deleted: " + item.Name)
					ui.refreshGallery()
				}
			}, ui.window).Show()
		})
		buttons := container.NewGridWithColumns(2, show, remove)
		card := widget.NewCard("", "", container.NewVBox(thumbnail, name, buttons))
		card.Resize(fyne.NewSize(195, 155))
		ui.gallery.Add(container.NewGridWrap(fyne.NewSize(200, 160), card))
	}
}

func loadImage(path string) image.Image {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	decoded, _, err := image.Decode(file)
	if err != nil && err != io.EOF {
		log.Print(err)
		return nil
	}
	return decoded
}

func (ui *fyneUI) setStatus(value any) {
	ui.status.SetText(fmt.Sprint(value))
}

var _ = layout.NewSpacer
