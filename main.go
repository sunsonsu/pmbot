package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/line/line-bot-sdk-go/v8/linebot"
)

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

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	bot, _ := linebot.New(os.Getenv("channelsecret"), os.Getenv("bottoken"))

	/* can not use req.HTTPMethod == "" as the AI said to check which request is from aws eventbridge or Line web hook because they always ""
	so, check Headers instead because if request from line will have signature */

	if req.Headers["x-line-signature"] != "" {
		return handleLineWebhook(bot, req)
	}

	// other case is croJob from eventbridge
	weatherData, err := getAirVisualData("", "")
	if err != nil || weatherData == nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	flexMessage := createFlexMessage(weatherData)
	altText := fmt.Sprintf("สภาพอากาศ %s - AQI %d - %d C", weatherData.City, weatherData.PM25, weatherData.Temp)

	_, err = bot.PushMessage(os.Getenv("groupid"), linebot.NewFlexMessage(altText, flexMessage)).Do()
	if err != nil {
		fmt.Println("Push Message Error:", err)
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Push Message Success"}, nil
}

func handleLineWebhook(bot *linebot.Client, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	bodyReader := strings.NewReader(req.Body)
	httpReq, _ := http.NewRequest("POST", "", bodyReader)
	httpReq.Header.Set("X-Line-Signature", req.Headers["x-line-signature"])

	lineEvent, err := bot.ParseRequest(httpReq)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	for _, event := range lineEvent {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				msgText := strings.ToLower(message.Text)
				if msgText == "pm25" || msgText == "pm2.5" {
					weatherData, _ := getAirVisualData("", "")
					if weatherData != nil {
						flexMessage := createFlexMessage(weatherData)
						altText := fmt.Sprintf("สภาพอากาศ %s - AQI %d - %d C", weatherData.City, weatherData.PM25, weatherData.Temp)
						_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewFlexMessage(altText, flexMessage)).Do()
						if err != nil {
							fmt.Println("Reply Message From Text Error:", err)
						}
					}
				}
			case *linebot.LocationMessage:
				lat := fmt.Sprintf("%f", message.Latitude)
				lon := fmt.Sprintf("%f", message.Longitude)
				weatherData, _ := getAirVisualData(lat, lon)
				if weatherData != nil {
					flexMessage := createFlexMessage(weatherData)
					altText := fmt.Sprintf("สภาพอากาศ %s - AQI %d - %d C", weatherData.City, weatherData.PM25, weatherData.Temp)
					_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewFlexMessage(altText, flexMessage)).Do()
					if err != nil {
						fmt.Println("Reply Message From Location  Error:", err)
					}
				}
			}
		}
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Reply Message Success"}, nil
}

func getAirVisualData(lat, lon string) (*WeatherResult, error) {
	var apiKey string = os.Getenv("airapi")
	var url string = fmt.Sprintf("http://api.airvisual.com/v2/city?city=San%%20Sai&state=Chiang%%20Mai&country=Thailand&key=%s", apiKey)

	if lat != "" && lon != "" {
		url = fmt.Sprintf("http://api.airvisual.com/v2/nearest_city?lat=%s&lon=%s&key=%s", lat, lon, apiKey)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResponse AirVisualDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	data := apiResponse.Data

	t, _ := time.Parse(time.RFC3339, data.Current.Pollution.Ts)
	loc, _ := time.LoadLocation("Asia/Bangkok")
	formattedTime := t.In(loc).Format("02 Jan 2006, 15:04:05")

	return &WeatherResult{
		PM25:     data.Current.Pollution.Aqius,
		Temp:     data.Current.Weather.Tp,
		Humidity: data.Current.Weather.Hu,
		City:     data.City,
		Country:  data.Country,
		Time:     formattedTime,
	}, nil

}

func createFlexMessage(data *WeatherResult) linebot.FlexContainer {
	aqi := getAQIStatus(data.PM25)
	tempColor := getTempColor(data.Temp)

	flexJSON := fmt.Sprintf(`{
		"type": "bubble",
		"size": "kilo",
		"header": {
			"type": "box",
			"layout": "vertical",
			"contents": [
				{ "type": "text", "text": "สภาพอากาศ", "color": "#ffffff", "size": "lg", "weight": "bold" },
				{ "type": "text", "text": "%s, %s", "color": "#ffffff", "size": "sm", "margin": "xs" }
			],
			"backgroundColor": "%s",
			"paddingAll": "16px"
		},
		"body": {
			"type": "box",
			"layout": "vertical",
			"contents": [
				{
					"type": "box",
					"layout": "horizontal",
					"contents": [
						{
							"type": "box",
							"layout": "vertical",
							"contents": [
								{ "type": "text", "text": "%d°C", "size": "3xl", "weight": "bold", "color": "%s" }
							],
							"flex": 1,
							"justifyContent": "center",
							"alignItems": "center"
						}
					],
					"cornerRadius": "10px", "paddingAll": "10px"
				},
				{ "type": "separator", "margin": "md", "color": "#E0E0E0" },
				{
					"type": "box",
					"layout": "vertical",
					"contents": [
						{ "type": "text", "text": "คุณภาพอากาศ (US AQI)", "size": "md", "weight": "bold", "margin": "md" },
						{
							"type": "box",
							"layout": "vertical",
							"contents": [
								{ "type": "text", "text": "%s", "size": "sm", "color": "#ffffff", "weight": "bold", "align": "center" },
								{ "type": "text", "text": "%d", "size": "3xl", "weight": "bold", "color": "#ffffff", "align": "center" }
							],
							"backgroundColor": "%s",
							"cornerRadius": "10px", "paddingAll": "15px", "margin": "sm"
						}
					]
				},
				{
					"type": "box",
					"layout": "horizontal",
					"contents": [
						{ "type": "text", "text": "%s", "size": "xs", "color": "#666666", "wrap": true }
					],
					"backgroundColor": "#FFFBF0", "cornerRadius": "8px", "paddingAll": "12px", "margin": "md", "borderWidth": "1px", "borderColor": "%s"
				}
			],
			"paddingAll": "16px"
		},
		"footer": {
			"type": "box",
			"layout": "vertical",
			"contents": [
				{ "type": "text", "text": "ข้อมูลเมื่อ: %s", "size": "xxs", "color": "#999999", "align": "center" }
			],
			"backgroundColor": "#FAFAFA"
		}
	}`, data.City, data.Country, tempColor, data.Temp, tempColor, aqi.Text, data.PM25, aqi.Color, aqi.Desc, aqi.Color, data.Time)

	container, _ := linebot.UnmarshalFlexMessageJSON([]byte(flexJSON))
	return container
}

func getAQIStatus(pm25 int) StatusDetail {
	if pm25 <= 50 {
		return StatusDetail{"#66BB6A", "ดี", "อากาศดีมาก เหมาะแก่การทำกิจกรรมกลางแจ้ง"}
	}
	if pm25 <= 100 {
		return StatusDetail{"#FFA726", "ปานกลาง", "คุณภาพอากาศเริ่มแย่ ควรใส่หน้ากากกลุ่มเสี่ยงควรลดกิจกรรมหนัก"}
	}
	if pm25 <= 150 {
		return StatusDetail{"#FF7043", "ไม่ดีต่อกลุ่มเสี่ยง", "ควรใส่หน้ากากและเลี่ยงการออกนอกอาคาร"}
	}
	return StatusDetail{"#EF5350", "ไม่ดี", "ควรใส่หน้ากากและหลีกเลี่ยงกิจกรรมกลางแจ้ง"}
}

func getTempColor(temp int) string {
	if temp >= 35 {
		return "#d03718"
	}
	if temp >= 30 {
		return "#cb7b12"
	}
	if temp > 20 {
		return "#39c154"
	}
	return "#14a4cc"
}
