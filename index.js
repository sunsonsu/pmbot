const express = require('express');
const axios = require('axios');
const line = require('@line/bot-sdk');
const serverless = require('serverless-http');
const { createWeatherFlex } = require('./flexTemplate');

const app = express();
const config = {
    channelAccessToken: process.env.bottoken,
    channelSecret: process.env.channelsecret
};
const client = new line.Client(config);

async function getWeatherData(lat, lon) {
    const API_KEY = process.env.airapi;
    const CITY = "San Sai", STATE = "Chiang Mai", COUNTRY = "Thailand";

    let url = `http://api.airvisual.com/v2/city?city=${CITY}&state=${STATE}&country=${COUNTRY}&key=${API_KEY}`;
    if (lat && lon) {
        url = `http://api.airvisual.com/v2/nearest_city?lat=${lat}&lon=${lon}&key=${API_KEY}`;
    }

    try {
        const response = await axios.get(url);
        const d = response.data.data;
        return {
            pm25: d.current.pollution.aqius,
            temp: d.current.weather.tp,          // Temperature
            humidity: d.current.weather.hu,      // Humidity
            windSpeed: d.current.weather.ws,     // Wind Speed
            place: d.city,
            country: d.country,
            time: new Date(d.current.pollution.ts).toLocaleString('th-TH', {
    timeZone: 'Asia/Bangkok',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false 
}),
        };
    } catch (error) {
        console.error("Error fetching data:", error);
        return null;
    }
}

async function handleEvent(event) {
    if (event.type !== 'message') return null;
    const { message } = event;

    let weatherData = null;
    if (message.type === 'location') {
        weatherData = await getWeatherData(message.latitude, message.longitude);
    } else if (message.type === 'text' && (message.text.toLowerCase().includes("pm2.5") || message.text.toLowerCase().includes("pm25"))) {
        weatherData = await getWeatherData();
    }

    if (weatherData) {
        const flex = createWeatherFlex(weatherData);
        await client.replyMessage(event.replyToken, flex);
    }
}

async function runCronJob() {
    try {
        const weatherData = await getWeatherData();
        if (weatherData) {
            const flex = createWeatherFlex(weatherData);
            await client.pushMessage(process.env.groupid, flex);
            console.log("Cron Job: Push Success");
        }
    } catch (error) {
        console.error("Cron Job Error:", error);
    }
}

// --- Lambda Handler ---
const httpHandler = serverless(app);
module.exports.handler = async (event, context) => {
    // ถ้าไม่มี httpMethod แสดงว่ามาจาก EventBridge
    if (!event.httpMethod && !event.requestContext) {
        await runCronJob();
        return { statusCode: 200, body: "OK" };
    }
    return await httpHandler(event, context);
};

app.post('/api', line.middleware(config), async (req, res) => {
    await Promise.all(req.body.events.map(handleEvent));
    res.json({ status: 'success' });
});