const express = require('express');
const axios = require('axios');
require('dotenv').config();
const app = express();
const line = require('@line/bot-sdk');

// LINE Messaging API credentials   
const config = {
    channelAccessToken: process.env.bottoken,  // Replace with your Channel Access Token
    channelSecret: process.env.channelsecret              // Replace with your Channel Secret
};

app.get('/', (req, res) => {
    res.send('Hello World! API WORKS!');
})

// Handle POST requests sent from LINE (events such as messages)
app.post('/api/webhook', line.middleware(config), (req, res) => {
    Promise
        .all([req.body.events.map(handleEvent)])
        .then((result) => res.json(result))
        .catch((err) => {
            console.error(err);
            res.status(500).send('Error');
        });
});

const client = new line.Client(config);

// Handle incoming events
async function handleEvent(event) {
    if (event.type === 'message' && event.message.type === 'text') {
        const userMessage = event.message.text;

        // Check if the message contains "PM2.5"
        const keywords = ["PM2.5", "pm2.5", "PM25", "pm25"];
        if (keywords.some(keyword => userMessage.includes(keyword))) {
            const { pm25, place, time } = await getPM25();

            let color;
            if (pm25 <= 50) {
                color = '#00E400'; // Good
            } else if (pm25 <= 100) {
                color = '#FFFF00'; // Moderate
            } else if (pm25 <= 150) {
                color = '#FF7E00'; // Unhealthy for Sensitive Groups
            } else if (pm25 <= 200) {
                color = '#FF0000'; // Unhealthy
            } else if (pm25 <= 300) {
                color = '#8F3F97'; // Very Unhealthy
            } else {
                color = '#7E0023'; // Hazardous
            }

            let description;
            if (pm25 <= 50) {
                description = 'อากาศดีมาก ไม่จำเป็นต้องใส่หน้ากาก 😷';
            } else if (pm25 <= 100) {
                description = 'อากาศพอใช้ได้ หากคุณมีอาการแพ้ ควรพิจารณาใส่หน้ากาก 😷';
            } else if (pm25 <= 150) {
                description = 'อากาศเริ่มมีผลกระทบต่อสุขภาพ ควรใส่หน้ากากเพื่อป้องกัน 😷';
            } else if (pm25 <= 200) {
                description = 'อากาศไม่ดีต่อสุขภาพ ควรใส่หน้ากากและหลีกเลี่ยงกิจกรรมกลางแจ้ง 😷';
            } else if (pm25 <= 300) {
                description = 'อากาศแย่มาก ควรใส่หน้ากาก N95 และหลีกเลี่ยงการออกนอกบ้าน 😷';
            } else {
                description = 'อากาศอันตรายมาก ควรอยู่ในบ้านและใช้เครื่องฟอกอากาศ 😷';
            }

            const message = {
                type: 'text',
                text: `📢 แจ้งเตือน PM2.5\n\n🌍 สถานที่: ${place}\n🌫 ค่า PM2.5: ${pm25} AQI\n📌 ระดับคุณภาพอากาศ: ${description}\n🕒 อัพเดตเมื่อ: ${time}`
            };
            
            // Send the message using LINE bot
        await client.replyMessage(event.replyToken, message);
        }else{
            await client.replyMessage(event.replyToken, { type: 'text', text: 'ไม่พบข้อมูล PM2.5' });
        }
    }
    return Promise.resolve(null);
}

async function getPM25() {
    const API_KEY = process.env.airapi; // Replace with your API Key
    const CITY = "San Sai";
    const STATE = "Chiang Mai";
    const COUNTRY = "Thailand";

    const url = `http://api.airvisual.com/v2/city?city=${CITY}&state=${STATE}&country=${COUNTRY}&key=${API_KEY}`;

    try {
        const response = await axios.get(url);
        const data = response.data;
        return {
            pm25: data.data.current.pollution.aqius,
            place: data.data.city,
            time: new Date(data.data.current.pollution.ts).toLocaleString('en-US', { timeZone: 'Asia/Bangkok' }),
        };
    } catch (error) {
        console.error("Error fetching PM2.5 data:", error);
        return { pm25: "N/A", place: "Unknown", time: "N/A" };
    }
}

module.exports = app;