🌦️ LINE Weather & PM2.5 Bot (Go + AWS Lambda)
A high-performance, serverless LINE Bot written in Go that provides real-time weather and air quality data from AirVisual API. It supports both on-demand requests via text/location and automated reports via scheduled events.

Features
On-Demand Weather: Triggered by text "pm25" or "pm2.5".

Location-Based Search: Send a location pin to get the nearest city's air quality.

Automated Reports: Integration with AWS EventBridge (Cron) to push daily weather updates to a specific Group ID.

Rich UI: Uses LINE Flex Messages with dynamic coloring based on AQI and Temperature levels.

Performance: Built with Go for near-instant cold starts on AWS Lambda.

![Alt text](images/demo.png)

🛠️ Tech Stack
- Language: Go (Golang)

- Hosting: AWS Lambda (Amazon Linux 2023)

- API: AirVisual (IQAir)

- SDK: LINE Messaging API SDK for Go v8

![Alt text](images/lambda.png)
for deploying on aws lambda use the zipScript.sh to build and zip then upload to aws lambda.
ps. thank you gemini for helping write this readme