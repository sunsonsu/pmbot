package main

type AirVisualDataResponse struct {
	Data struct {
		City    string `json:"city"`
		Country string `json:"country"`
		Current struct {
			Pollution struct {
				Aqius int    `json:"aqius"`
				Ts    string `json:"ts"`
			} `json:"pollution"`
			Weather struct {
				Tp int     `json:"tp"`
				Hu int     `json:"hu"`
				Ws float64 `json:"ws"`
			} `json:"weather"`
		} `json:"current"`
	} `json:"data"`
}

type WeatherResult struct {
	PM25      int
	Temp      int
	Humidity  int
	WindSpeed float64
	City      string
	Country   string
	Time      string
}

type StatusDetail struct {
	Color, Text, Desc string
}
