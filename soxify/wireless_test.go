package main

import (
	"encoding/json"
	"testing"
)

func TestWireless(t *testing.T) {

	test_str := `[{"rssi": -82, "noise": -92, "sec": ["WPA2", "WPA"], "ssid": "symphony_corp", "bssid": "0:24:51:5:bb:f0"}, {"rssi": -79, "noise": -92, "sec": ["WPA2", "WPA"], "ssid": "testssid", "bssid": "0:24:51:5:bb:f2"}, {"rssi": -81, "noise": -92, "sec": [], "ssid": "harmony_guest", "bssid": "0:24:51:5:bb:f1"}, {"rssi": -57, "noise": -92, "sec": ["WPA2", "WPA"], "ssid": "nycmadap01", "bssid": "0:24:51:5:e3:6c"}, {"rssi": -57, "noise": -92, "sec": ["WPA2", "WPA"], "ssid": "testssid", "bssid": "0:24:51:5:e3:6d"}, {"rssi": -57, "noise": -92, "sec": [], "ssid": "harmony_guest", "bssid": "0:24:51:5:e3:6e"}, {"rssi": -57, "noise": -92, "sec": ["WPA2", "WPA"], "ssid": "symphony_corp", "bssid": "0:24:51:5:e3:6f"}, {"rssi": -44, "noise": -92, "sec": ["WPA2", "WPA"], "ssid": "nycmadap01", "bssid": "0:24:51:5:e3:63"}, {"rssi": -80, "noise": -92, "sec": ["WPA2", "WPA"], "ssid": "nycmadap01", "bssid": "0:24:51:5:bb:f3"}, {"rssi": -89, "noise": -91, "sec": ["WPA2", "WPA"], "ssid": "AdPeople Mansion", "bssid": "e8:8:8b:c9:c6:1"}, {"rssi": -43, "noise": -91, "sec": ["WPA2", "WPA"], "ssid": "AdPeople Teepee Ranch", "bssid": "e8:8:8b:c9:c2:3c"}, {"rssi": -51, "noise": -92, "sec": ["WPA2", "WPA"], "ssid": "testssid", "bssid": "0:24:51:5:e3:62"}, {"rssi": -51, "noise": -92, "sec": [], "ssid": "harmony_guest", "bssid": "0:24:51:5:e3:61"}, {"rssi": -51, "noise": -92, "sec": ["WPA2", "WPA"], "ssid": "symphony_corp", "bssid": "0:24:51:5:e3:60"}, {"rssi": -72, "noise": -92, "sec": ["WPA2", "WPA"], "ssid": "Geranium", "bssid": "6:27:22:d5:e0:96"}, {"rssi": -62, "noise": -92, "sec": ["WPA2", "WPA"], "ssid": "Geranium Guest", "bssid": "a:27:22:d5:e0:96"}]`
	var n []Network
	err := json.Unmarshal([]byte(test_str), &n)
	if err != nil {
		t.Error(err)
	}

}
