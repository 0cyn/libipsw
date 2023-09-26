package download

//#cgo LDFLAGS:
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
import "C"
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const ipswMeAPI = "https://api.ipsw.me/v4/"

// Device struct
type Device struct {
	Name        string `json:"name,omitempty"`
	Identifier  string `json:"identifier,omitempty"`
	BoardConfig string `json:"boardconfig,omitempty"`
	Platform    string `json:"platform,omitempty"`
	CpID        int    `json:"cpid,omitempty"`
	BdID        int    `json:"bdid,omitempty"`
	Firmwares   []IPSW `json:"firmwares,omitempty"`
}

// IPSW struct
type IPSW struct {
	Identifier  string    `json:"identifier,omitempty"`
	Version     string    `json:"version,omitempty"`
	BuildID     string    `json:"buildid,omitempty"`
	SHA1        string    `json:"sha1sum,omitempty"`
	MD5         string    `json:"md5sum,omitempty"`
	FileSize    int       `json:"filesize,omitempty"`
	URL         string    `json:"url,omitempty"`
	ReleaseDate time.Time `json:"releasedate,omitempty"`
	UploadDate  time.Time `json:"uploaddate,omitempty"`
	Signed      bool      `json:"signed,omitempty"`
}

// GetAllDevices returns a list of all devices
func GetAllDevices() ([]Device, error) {
	devices := []Device{}

	res, err := http.Get(ipswMeAPI + "devices")
	if err != nil {
		return devices, err
	}
	if res.StatusCode != http.StatusOK {
		return devices, fmt.Errorf("api returned status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return devices, err
	}
	res.Body.Close()

	err = json.Unmarshal(body, &devices)
	if err != nil {
		return devices, err
	}

	return devices, nil
}

//export c_internal_download_ipsw_me_GetDevice
func c_internal_download_ipsw_me_GetDevice(identifier *C.char, identifierLen C.uint, outJson **C.char, outJsonLen *C.uint, err **C.char, errLen *C.uint) C.char {

	device, deviceError := GetDevice(C.GoStringN(identifier, C.int(identifierLen)))
	if deviceError != nil {
		outError := fmt.Sprintf("c_GetDevice: GetDevice failed with %w", deviceError)
		*err = C.CString(outError)
		*errLen = C.uint(len(outError))
		return C.char(0)
	}
	fret, jsonErr := json.Marshal(device)
	if jsonErr != nil {
		outError := fmt.Sprintf("c_GetDevice: Failed to serialize Device object: %w", jsonErr)
		*err = C.CString(outError)
		*errLen = C.uint(len(outError))
		return C.char(0)
	}
	cs := C.CString(string(fret))
	*outJson = cs
	*outJsonLen = C.uint(C.strlen(cs))

	return C.char(1)
}

// GetDevice returns a device from it's identifier
func GetDevice(identifier string) (Device, error) {
	d := Device{}

	res, err := http.Get(ipswMeAPI + "device" + "/" + identifier)
	if err != nil {
		return d, err
	}
	if res.StatusCode != http.StatusOK {
		return d, fmt.Errorf("api returned status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return d, err
	}
	res.Body.Close()

	err = json.Unmarshal(body, &d)
	if err != nil {
		return d, err
	}

	return d, nil
}

//export c_internal_download_ipsw_me_GetDeviceIPSWs
func c_internal_download_ipsw_me_GetDeviceIPSWs(identifier *C.char, identifierLen C.uint, outJson **C.char, outJsonLen *C.uint, err **C.char, errLen *C.uint) C.char {
	device, deviceError := GetDeviceIPSWs(C.GoStringN(identifier, C.int(identifierLen)))
	if deviceError != nil {
		outError := fmt.Sprintf("c_GetDeviceIPSWs: GetDeviceIPSWs failed with %w", deviceError)
		*err = C.CString(outError)
		*errLen = C.uint(len(outError))
		return C.char(0)
	}
	fret, jsonErr := json.Marshal(device)
	if jsonErr != nil {
		outError := fmt.Sprintf("c_GetDeviceIPSWs: Failed to serialize Device object: %w", jsonErr)
		*err = C.CString(outError)
		*errLen = C.uint(len(outError))
		return C.char(0)
	}
	cs := C.CString(string(fret))
	*outJson = cs
	*outJsonLen = C.uint(C.strlen(cs))

	return C.char(1)
}

// GetDeviceIPSWs returns a device's IPSWs from it's identifier
func GetDeviceIPSWs(identifier string) ([]IPSW, error) {
	d, err := GetDevice(identifier)
	if err != nil {
		return nil, err
	}
	return d.Firmwares, nil
}

// GetAllIPSW finds all IPSW files for a given iOS version
func GetAllIPSW(version string) ([]IPSW, error) {
	ipsws := []IPSW{}

	res, err := http.Get(ipswMeAPI + "ipsw/" + version)
	if err != nil {
		return ipsws, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return ipsws, err
	}
	res.Body.Close()

	err = json.Unmarshal(body, &ipsws)
	if err != nil {
		return ipsws, err
	}

	return ipsws, nil
}

// GetIPSW will get an IPSW when supplied an identifier and build ID
func GetIPSW(identifier, buildID string) (IPSW, error) {
	i := IPSW{}

	res, err := http.Get(ipswMeAPI + "ipsw/" + identifier + "/" + buildID)
	if err != nil {
		return i, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return i, fmt.Errorf("api returned status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return i, err
	}

	err = json.Unmarshal(body, &i)
	if err != nil {
		return i, err
	}

	return i, nil
}

// GetVersion returns the iOS version for a given build ID
func GetVersion(buildID string) (string, error) {

	devices, err := GetAllDevices()
	if err != nil {
		return "", fmt.Errorf("failed to get all devices from ipsw.me API: %v", err)
	}

	for i := len(devices) - 1; i >= 0; i-- {
		var dev Device
		res, err := http.Get(ipswMeAPI + "device/" + devices[i].Identifier)
		if err != nil {
			return "", err
		}
		if res.StatusCode != http.StatusOK {
			return "", fmt.Errorf("api returned status: %s", res.Status)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		res.Body.Close()

		err = json.Unmarshal(body, &dev)
		if err != nil {
			return "", err
		}

		for _, ipsw := range dev.Firmwares {
			if ipsw.BuildID == buildID {
				return ipsw.Version, nil
			}
		}
	}

	return "", fmt.Errorf("build did not a version")
}

// GetBuildID returns the BuildID for a given version and identifier
func GetBuildID(version, identifier string) (string, error) {
	var ipsws []IPSW

	res, err := http.Get(ipswMeAPI + "ipsw/" + version)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api returned status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	res.Body.Close()

	err = json.Unmarshal(body, &ipsws)
	if err != nil {
		return "", err
	}

	for _, i := range ipsws {
		if i.Identifier == identifier {
			return i.BuildID, nil
		}
	}
	return "", fmt.Errorf("no build found for version %s and device %s", version, identifier)
}

// https://api.ipsw.me/v4/releases
// func GetReleases() []Release {}
