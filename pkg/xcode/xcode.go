package xcode

//#cgo LDFLAGS:
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
import "C"
import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blacktop/ipsw/internal/utils"
)

//go:embed data/device_traits.gz
var traitsData []byte

// Device object
type Device struct {
	Target                   string      `gorm:"column:Target;primary_key" json:"target,omitempty"`
	TargetType               string      `gorm:"column:TargetType" json:"target_type,omitempty"`
	TargetVariant            string      `gorm:"column:TargetVariant" json:"target_variant,omitempty"`
	Platform                 string      `gorm:"column:Platform" json:"platform,omitempty"`
	ProductType              string      `gorm:"column:ProductType" json:"product_type,omitempty"`
	ProductDescription       string      `gorm:"column:ProductDescription" json:"product_description,omitempty"`
	CompatibleDeviceFallback string      `gorm:"column:CompatibleDeviceFallback" json:"compatible_device_fallback,omitempty"`
	DeviceTrait              DeviceTrait `gorm:"foreignkey:DeviceTraitSet" json:"traits,omitempty"`
	DeviceTraitSet           int         `gorm:"column:DeviceTraitSet" json:"-"`
}

// TableName returns the table name for the Device object
func (Device) TableName() string {
	return "Devices"
}

type Devices []Device

func (d Devices) Len() int      { return len(d) }
func (d Devices) Swap(i, j int) { d[i], d[j] = d[j], d[i] }

type ByProductType struct{ Devices }

func (s ByProductType) Less(i, j int) bool {
	devI := utils.DeconstructDevice(s.Devices[i].ProductType)
	devJ := utils.DeconstructDevice(s.Devices[j].ProductType)
	return fmt.Sprintf("%s%02d%02d", devI.Family, devI.Major, devI.Minor) < fmt.Sprintf("%s%02d%02d", devJ.Family, devJ.Major, devJ.Minor)
}

// DeviceTrait object
type DeviceTrait struct {
	DeviceTraitSetID             int    `gorm:"column:DeviceTraitSetID;primary_key" json:"-"`
	PreferredArchitecture        string `gorm:"column:PreferredArchitecture" json:"preferred_architecture,omitempty"`
	ArtworkDeviceIdiom           string `gorm:"column:ArtworkDeviceIdiom" json:"artwork_device_idiom,omitempty"`
	ArtworkHostedIdioms          string `gorm:"column:ArtworkHostedIdioms" json:"artwork_hosted_idioms,omitempty"`
	ArtworkScaleFactor           int    `gorm:"column:ArtworkScaleFactor" json:"artwork_scale_factor,omitempty"`
	ArtworkDeviceSubtype         int    `gorm:"column:ArtworkDeviceSubtype" json:"artwork_device_subtype,omitempty"`
	ArtworkDisplayGamut          string `gorm:"column:ArtworkDisplayGamut" json:"artwork_display_gamut,omitempty"`
	ArtworkDynamicDisplayMode    string `gorm:"column:ArtworkDynamicDisplayMode" json:"artwork_dynamic_display_mode,omitempty"`
	DevicePerformanceMemoryClass int    `gorm:"column:DevicePerformanceMemoryClass" json:"device_performance_memory_class,omitempty"`
	GraphicsFeatureSetClass      string `gorm:"column:GraphicsFeatureSetClass" json:"graphics_feature_set_class,omitempty"`
	GraphicsFeatureSetFallbacks  string `gorm:"column:GraphicsFeatureSetFallbacks" json:"graphics_feature_set_fallbacks,omitempty"`
}

// TableName returns the table name for the DeviceTrait object
func (DeviceTrait) TableName() string {
	return "DeviceTraits"
}

// WriteToJSON writes the data to JSON
func WriteToJSON(devices []Device, dest string) error {
	dJSON, err := json.Marshal(devices)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Clean(dest), dJSON, 0660)
}

//export c_pkg_xcode_xcode_GetDevices
func c_pkg_xcode_xcode_GetDevices(outJson **C.char, outJsonLen *C.uint, err **C.char, errLen *C.uint) C.char {
	devices, devicesError := GetDevices()
	if devicesError != nil {
		outError := fmt.Sprintf("c_pkg_xcode_xcode_GetDevices: GetDeviceIPSWs failed with %w", devicesError)
		*err = C.CString(outError)
		*errLen = C.uint(len(outError))
		return C.char(0)
	}
	fret, jsonErr := json.Marshal(devices)
	if jsonErr != nil {
		outError := fmt.Sprintf("c_pkg_xcode_xcode_GetDevices: Failed to serialize Device object: %w", jsonErr)
		*err = C.CString(outError)
		*errLen = C.uint(len(outError))
		return C.char(0)
	}
	cs := C.CString(string(fret))
	*outJson = cs
	*outJsonLen = C.uint(C.strlen(cs))

	return C.char(1)
}

// GetDevices reads the devices from embedded JSON
func GetDevices() ([]Device, error) {
	var devices []Device

	zr, err := gzip.NewReader(bytes.NewReader(traitsData))
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	if err := json.NewDecoder(zr).Decode(&devices); err != nil {
		return nil, fmt.Errorf("failed unmarshaling device_traits.gz data: %w", err)
	}

	return devices, nil
}

// GetDeviceForProd returns the device matching a given product type
func GetDeviceForProd(prod string) (*Device, error) {

	devices, err := GetDevices()
	if err != nil {
		return nil, err
	}

	for _, device := range devices {
		if device.ProductType == prod {
			return &device, nil
		}
	}

	return nil, fmt.Errorf("device not found")
}

// GetDeviceForModel returns the device matching a given model
func GetDeviceForModel(model string) (*Device, error) {

	devices, err := GetDevices()
	if err != nil {
		return nil, err
	}

	for _, device := range devices {
		if device.Target == model {
			return &device, nil
		}
	}

	return nil, fmt.Errorf("device not found")
}
