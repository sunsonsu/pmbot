package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetAQIStatus(t *testing.T) {
	testCases := []struct {
		name     string
		aqiValue int
		expected StatusDetail
	}{
		{"pm25 less equal 50", 50, StatusDetail{"#66BB6A", "ดี", "อากาศดีมาก เหมาะแก่การทำกิจกรรมกลางแจ้ง"}},
		{"pm25 less equal 50", 1, StatusDetail{"#66BB6A", "ดี", "อากาศดีมาก เหมาะแก่การทำกิจกรรมกลางแจ้ง"}},
		{"pm25 less equal 100", 100, StatusDetail{"#FFA726", "ปานกลาง", "คุณภาพอากาศเริ่มแย่ ควรใส่หน้ากากกลุ่มเสี่ยงควรลดกิจกรรมหนัก"}},
		{"pm25 less equal 100", 51, StatusDetail{"#FFA726", "ปานกลาง", "คุณภาพอากาศเริ่มแย่ ควรใส่หน้ากากกลุ่มเสี่ยงควรลดกิจกรรมหนัก"}},
		{"pm25 less equal 150", 150, StatusDetail{"#FF7043", "ไม่ดีต่อกลุ่มเสี่ยง", "ควรใส่หน้ากากและเลี่ยงการออกนอกอาคาร"}},
		{"pm25 less equal 150", 101, StatusDetail{"#FF7043", "ไม่ดีต่อกลุ่มเสี่ยง", "ควรใส่หน้ากากและเลี่ยงการออกนอกอาคาร"}},
		{"pm25 over 150", 151, StatusDetail{"#612698", "อันตราย", "ใส่หน้ากากและงดการออกนอกอาคาร"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getAQIStatus(tc.aqiValue)
			if result != tc.expected {
				t.Errorf("Test %s Failed. Expected %s got %s", tc.name, tc.expected, result)
			}
		})
	}
}

func TestGetTempColor(t *testing.T) {
	testCases := []struct {
		name      string
		tempValue int
		expected  string
	}{
		{"temp greater equal 35", 35, "#d03718"},
		{"temp greater equal 35", 40, "#d03718"},
		{"temp greater equal 30", 34, "#cb7b12"},
		{"temp greater equal 30", 30, "#cb7b12"},
		{"temp greater 20", 21, "#39c154"},
		{"temp less than 20", 19, "#14a4cc"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getTempColor(tc.tempValue)
			if result != tc.expected {
				t.Errorf("Test %s Failed. Expected %s got %s", tc.name, tc.expected, result)
			}
		})
	}
}

func TestGetAirVisualData(t *testing.T) {

	mockAqiApiResponse := `{
        "data": {
            "city": "San Sai",
            "country": "Thailand",
            "current": {
                "pollution": { "ts": "2024-01-15T10:00:00Z", "aqius": 75 },
                "weather": { "tp": 32, "hu": 65 }
            }
        }
    }`

	testCases := []struct {
		name       string
		lat        string
		lon        string
		expectCity string
		expectPM25 int
		expectTemp int
	}{
		{"default city (no lat/lon)", "", "", "San Sai", 75, 32},
		{"with lat/lon", "18.7", "98.9", "San Sai", 75, 32},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.lat != "" {
					if !strings.Contains(r.URL.Path, "nearest_city") {
						t.Errorf("Expected nearest_city endpoint, got %s", r.URL.Path)
					}
				} else {
					if !strings.Contains(r.URL.Path, "city") {
						t.Errorf("Expected city endpoint, got %s", r.URL.Path)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, mockAqiApiResponse)
			}))
			defer server.Close()

			t.Setenv("AIRVISUAL_BASE_URL", server.URL)

			result, err := getAirVisualData(tc.lat, tc.lon, server.Client())

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if result.City != tc.expectCity {
				t.Errorf("Expected city %s, got %s", tc.expectCity, result.City)
			}
			if result.PM25 != tc.expectPM25 {
				t.Errorf("Expected PM25 %d, got %d", tc.expectPM25, result.PM25)
			}
			if result.Temp != tc.expectTemp {
				t.Errorf("Expected Temp %d, got %d", tc.expectTemp, result.Temp)
			}
		})
	}
}

func TestGetAirVisualData_Error(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		body       string
	}{
		{"server error 500", http.StatusInternalServerError, ""},
		{"invalid json", http.StatusOK, `{ invalid json `},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()

			t.Setenv("AIRVISUAL_BASE_URL", server.URL)

			result, err := getAirVisualData("", "", server.Client())

			if err == nil {
				t.Error("Expected error, got nil")
			}
			if result != nil {
				t.Error("Expected nil result, got value")
			}
		})
	}
}

func TestCreateFlexMessage(t *testing.T) {
	testCases := []struct {
		name string
		data *WeatherResult
	}{
		{"good AQI", &WeatherResult{PM25: 30, Temp: 25, Humidity: 60, City: "San Sai", Country: "Thailand", Time: "15 Jan 2024, 10:00:00"}},
		{"moderate AQI", &WeatherResult{PM25: 75, Temp: 30, Humidity: 65, City: "San Sai", Country: "Thailand", Time: "15 Jan 2024, 10:00:00"}},
		{"unhealthy AQI", &WeatherResult{PM25: 120, Temp: 35, Humidity: 70, City: "San Sai", Country: "Thailand", Time: "15 Jan 2024, 10:00:00"}},
		{"dangerous AQI", &WeatherResult{PM25: 200, Temp: 40, Humidity: 80, City: "San Sai", Country: "Thailand", Time: "15 Jan 2024, 10:00:00"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := createFlexMessage(tc.data)
			if result == nil {
				t.Errorf("Test %s: Expected FlexContainer, got nil", tc.name)
			}
		})
	}
}
