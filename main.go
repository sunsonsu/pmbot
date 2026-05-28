package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/line/line-bot-sdk-go/v8/linebot"
)

func main() {
	lambda.Start(HandleRequest)

}

func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("HandleRequest start: method=%s path=%s requestPath=%s isBase64=%v", req.HTTPMethod, req.Path, req.RequestContext.Path, req.IsBase64Encoded)
	bot, err := linebot.New(os.Getenv("channelsecret"), os.Getenv("bottoken"))
	if err != nil {
		log.Println("linebot.New error:", err)
	}

	/* can not use req.HTTPMethod == "" as the AI said to check which request is from aws eventbridge or Line web hook because they always ""
	so, check Headers instead because if request from line will have signature */

	sig := req.Headers["x-line-signature"]
	log.Printf("x-line-signature present=%v", sig != "")
	if sig != "" {
		return handleLineWebhook(bot, req)
	}

	// other case is croJob from eventbridge
	weatherData, err := getAirVisualData("", "", http.DefaultClient)
	if err != nil || weatherData == nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	flexMessage := createFlexMessage(weatherData)
	altText := fmt.Sprintf("สภาพอากาศ %s - AQI %d - %d C", weatherData.City, weatherData.PM25, weatherData.Temp)
	targetID := os.Getenv("groupid")

	path := strings.ToLower(req.Path)
	if path == "" {
		path = strings.ToLower(req.RequestContext.Path)
	}
	if strings.Contains(path, "/dev") {
		targetID = os.Getenv("userid")
	}

	log.Printf("PushMessage target=%s altText=%s", targetID, altText)
	_, err = bot.PushMessage(targetID, linebot.NewFlexMessage(altText, flexMessage)).Do()
	if err != nil {
		log.Println("Push Message Error:", err)
	} else {
		log.Println("Push Message sent successfully")
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Push Message Success"}, nil
}

func handleLineWebhook(bot *linebot.Client, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("handleLineWebhook: bodyLen=%d isBase64=%v", len(req.Body), req.IsBase64Encoded)
	bodyReader := strings.NewReader(req.Body)
	httpReq, _ := http.NewRequest("POST", "", bodyReader)
	httpReq.Header.Set("X-Line-Signature", req.Headers["x-line-signature"])

	lineEvent, err := bot.ParseRequest(httpReq)
	if err != nil {
		log.Println("ParseRequest error:", err)
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}
	log.Printf("Parsed %d line events", len(lineEvent))

	for _, event := range lineEvent {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				msgText := strings.ToLower(message.Text)
				log.Printf("TextMessage received: replyToken=%s text=%s", event.ReplyToken, message.Text)
				if msgText == "pm25" || msgText == "pm2.5" {
					weatherData, _ := getAirVisualData("", "", http.DefaultClient)
					if weatherData != nil {
						flexMessage := createFlexMessage(weatherData)
						altText := fmt.Sprintf("สภาพอากาศ %s - AQI %d - %d C", weatherData.City, weatherData.PM25, weatherData.Temp)
						_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewFlexMessage(altText, flexMessage)).Do()
						if err != nil {
							log.Println("Reply Message From Text Error:", err)
						} else {
							log.Println("Reply Message From Text Success")
						}
					}
				}
			case *linebot.LocationMessage:
				log.Printf("LocationMessage received: lat=%f lon=%f", message.Latitude, message.Longitude)
				lat := fmt.Sprintf("%f", message.Latitude)
				lon := fmt.Sprintf("%f", message.Longitude)
				weatherData, _ := getAirVisualData(lat, lon, http.DefaultClient)
				if weatherData != nil {
					flexMessage := createFlexMessage(weatherData)
					altText := fmt.Sprintf("สภาพอากาศ %s - AQI %d - %d C", weatherData.City, weatherData.PM25, weatherData.Temp)
					_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewFlexMessage(altText, flexMessage)).Do()
					if err != nil {
						log.Println("Reply Message From Location  Error:", err)
					} else {
						log.Println("Reply Message From Location Success")
					}
				}
			}
		}
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Reply Message Success"}, nil
}

func getAirVisualData(lat, lon string, client *http.Client) (*WeatherResult, error) {
	baseURL := os.Getenv("AIRVISUAL_BASE_URL")
	if baseURL == "" {
		baseURL = "http://api.airvisual.com"
	}

	var apiKey string = os.Getenv("airapi")
	var url string = fmt.Sprintf("%s/v2/city?city=San%%20Sai&state=Chiang%%20Mai&country=Thailand&key=%s", baseURL, apiKey)

	if lat != "" && lon != "" {
		url = fmt.Sprintf("%s/v2/nearest_city?lat=%s&lon=%s&key=%s", baseURL, lat, lon, apiKey)
	}

	log.Printf("Requesting AirVisual URL=%s", url)
	resp, err := client.Get(url)
	if err != nil {
		log.Println("AirVisual request error:", err)
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("AirVisual response status=%s", resp.Status)

	var apiResponse AirVisualDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		log.Println("Decode AirVisual response error:", err)
		return nil, err
	}

	data := apiResponse.Data

	t, _ := time.Parse(time.RFC3339, data.Current.Pollution.Ts)
	loc, _ := time.LoadLocation("Asia/Bangkok")
	formattedTime := t.In(loc).Format("02 Jan 2006, 15:04:05")

	return &WeatherResult{
		PM25:      data.Current.Pollution.Aqius,
		Temp:      data.Current.Weather.Tp,
		Humidity:  data.Current.Weather.Hu,
		WindSpeed: data.Current.Weather.Ws,
		HeatIndex: data.Current.Weather.HeatIndex,
		City:      data.City,
		Country:   data.Country,
		Time:      formattedTime,
	}, nil

}

func createFlexMessage(data *WeatherResult) linebot.FlexContainer {
	aqi := getAQIStatus(data.PM25)
	tempColor := getTempColor(data.Temp)

	container := map[string]any{
		"type": "bubble",
		"size": "kilo",
		"header": map[string]any{
			"type":            "box",
			"layout":          "vertical",
			"backgroundColor": tempColor,
			"paddingAll":      "16px",
			"contents": []any{
				textNode("สภาพอากาศ", "#ffffff", "lg", "bold", ""),
				textNode(fmt.Sprintf("%s, %s", data.City, data.Country), "#ffffff", "sm", "regular", "xs"),
			},
		},
		"body": map[string]any{
			"type":       "box",
			"layout":     "vertical",
			"paddingAll": "16px",
			"contents": []any{
				map[string]any{
					"type":           "box",
					"layout":         "vertical",
					"cornerRadius":   "10px",
					"paddingAll":     "10px",
					"justifyContent": "center",
					"alignItems":     "center",
					"contents": []any{
						textNode(fmt.Sprintf("%d°C", data.Temp), tempColor, "3xl", "bold", ""),
						textNode(fmt.Sprintf("รู้สึกเหมือน %d°C", data.HeatIndex), "#666666", "sm", "regular", "xs"),
					},
				},
				map[string]any{"type": "separator", "margin": "md", "color": "#E0E0E0"},
				map[string]any{
					"type":   "box",
					"layout": "vertical",
					"contents": []any{
						textNode("คุณภาพอากาศ (US AQI)", "#000000", "md", "bold", "md"),
						map[string]any{
							"type":            "box",
							"layout":          "vertical",
							"backgroundColor": aqi.Color,
							"cornerRadius":    "10px",
							"paddingAll":      "15px",
							"margin":          "sm",
							"contents": []any{
								textNode(aqi.Text, "#ffffff", "sm", "bold", ""),
								map[string]any{
									"type":   "text",
									"text":   fmt.Sprintf("%d", data.PM25),
									"size":   "3xl",
									"weight": "bold",
									"color":  "#ffffff",
									"align":  "center",
								},
							},
						},
					},
				},
				map[string]any{
					"type":            "box",
					"layout":          "horizontal",
					"backgroundColor": "#FFFBF0",
					"cornerRadius":    "8px",
					"paddingAll":      "12px",
					"margin":          "md",
					"borderWidth":     "1px",
					"borderColor":     aqi.Color,
					"contents": []any{
						textNode(aqi.Desc, "#666666", "xs", "regular", ""),
					},
				},
			},
		},
		"footer": map[string]any{
			"type":            "box",
			"layout":          "vertical",
			"backgroundColor": "#FAFAFA",
			"contents": []any{
				map[string]any{
					"type":  "text",
					"text":  fmt.Sprintf("ข้อมูลเมื่อ: %s", data.Time),
					"size":  "xxs",
					"color": "#999999",
					"align": "center",
				},
			},
		},
	}

	jsonBytes, _ := json.Marshal(container)
	flexContainer, err := linebot.UnmarshalFlexMessageJSON(jsonBytes)
	if err != nil {
		log.Println("UnmarshalFlexMessageJSON error:", err)
	} else {
		log.Printf("Created flex message for %s", data.City)
	}
	return flexContainer
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func textNode(text, color, size, weight, margin string) map[string]any {
	node := map[string]any{
		"type":   "text",
		"text":   text,
		"color":  color,
		"size":   size,
		"weight": weight,
	}
	if margin != "" {
		node["margin"] = margin
	}
	return node
}

func uriAction(label, uri string) map[string]any {
	return map[string]any{
		"type":  "uri",
		"label": label,
		"uri":   uri,
	}
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
	if pm25 <= 200 {
		return StatusDetail{"#c20000", "ไม่ดีต่อสุขภาพ", "ใส่หน้ากากและเปิดเครื่องฟอกอากาศ"}
	}
	if pm25 <= 300 {
		return StatusDetail{"#910ac2", "อันตรายต่อสุขภาพ", "งดออกนอกอาคารและเปิดเครื่องฟอกอากาศ"}
	} else {
		return StatusDetail{"#52291f", "อันตรายรุนแรง", "งดออกนอกอาคารและเปิดเครื่องฟอกอากาศ"}
	}
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
