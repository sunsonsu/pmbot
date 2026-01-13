function getAQIStatus(pm25) {
    if (pm25 <= 50) return { 
        color: "#66BB6A", 
        text: "ดี - Good", 
        icon: "🟢",
        desc: "อากาศดีมาก เหมาะแก่การทำกิจกรรมกลางแจ้ง" 
    };
    if (pm25 <= 100) return { 
        color: "#FFA726", 
        text: "ปานกลาง - Moderate",
        icon: "🟡", 
        desc: "คุณภาพอากาศปานกลาง กลุ่มเสี่ยงควรลดกิจกรรมหนัก" 
    };
    if (pm25 <= 150) return { 
        color: "#FF7043", 
        text: "ไม่ดีต่อกลุ่มเสี่ยง - Unhealthy for Sensitive",
        icon: "🟠",
        desc: "กลุ่มเสี่ยงควรใส่หน้ากากและลดเวลากิจกรรมกลางแจ้ง" 
    };
    return { 
        color: "#EF5350", 
        text: "ไม่ดี - Unhealthy",
        icon: "🔴", 
        desc: "ควรใส่หน้ากากและหลีกเลี่ยงกิจกรรมกลางแจ้งเป็นเวลานาน" 
    };
}

function getTemperatureDetail(temp) {
    if (temp >= 35) {
        return { color: "#d03718", feeling: "ร้อนมาก" };
    }
    if (temp >= 30) {
        return { color: "#cb7b12", feeling: "แดดจัด" };
    }
    if (temp > 20) {
        return { color: "#39c154", feeling: "อากาศดี" };
    }
    return { color: "#14a4cc", feeling: "เย็นสบาย" };
}

const createWeatherFlex = (data) => {
    const aqiStatus = getAQIStatus(data.pm25);
    const tempDetail = getTemperatureDetail(data.temp)

    return {
        type: "flex",
        altText: `สภาพอากาศ ${data.place} - อุณหภูมิ ${data.temp}°C, AQI ${data.pm25}`,
        contents: {
            type: "bubble",
            size: "kilo",
            header: {
                type: "box",
                layout: "vertical",
                contents: [
                    { 
                        type: "text", 
                        text: "สภาพอากาศ", 
                        color: "#ffffff", 
                        size: "lg", 
                        weight: "bold" 
                    },
                    { 
                        type: "text", 
                        text: data.place + ","+ data.country, 
                        color: "#ffffff", 
                        size: "sm", 
                        margin: "xs" 
                    },
                ],
                backgroundColor: tempDetail.color,
                paddingAll: "16px"
            },
            body: {
                type: "box",
                layout: "vertical",
                contents: [
                    // Weather Section
                    {
                        type: "box",
                        layout: "horizontal",
                        contents: [                       
                            {
                                type: "box",
                                layout: "vertical",
                                contents: [
                                    { 
                                        type: "text", 
                                        text: `${data.temp}°C`, 
                                        size: "3xl", 
                                        weight: "bold", 
                                        color: `${tempDetail.color}`
                                    },
                                    { 
                                        type: "text", 
                                        text: `${tempDetail.feeling}`, 
                                        size: "xs", 
                                        color: "#95A5A6",
                                        margin: "sm"
                                    }
                                ],
                                flex: 1,
                                justifyContent: "center",
                                alignItems: "center"
                            
                            }
                        ],
                        cornerRadius: "10px",
                        paddingAll: "10px",
                        spacing: "md"
                    },
                    { 
                        type: "separator", 
                        margin: "md",
                        color: "#E0E0E0"
                    },
                    // AQI Section
                    {
                        type: "box",
                        layout: "vertical",
                        contents: [
                            { 
                                type: "text", 
                                text: "คุณภาพอากาศ (US AQI)", 
                                size: "md", 
                                weight: "bold", 
                                color: "#333333",
                                margin: "md"
                            },
                            {
                                type: "box",
                                layout: "horizontal",
                                contents: [
                                    {
                                        type: "box",
                                        layout: "vertical",
                                        contents: [
                                            { 
                                                type: "text", 
                                                text: aqiStatus.text, 
                                                size: "sm", 
                                                color: "#ffffff",
                                                weight: "bold",
                                                align: "center"
                                            },
                                            { 
                                                type: "text", 
                                                text: `${data.pm25}`, 
                                                size: "3xl", 
                                                weight: "bold", 
                                                color: "#ffffff",
                                                align: "center"

                                            }
                                        ],
                                        flex: 0,
                                        spacing: "xs"
                                    },
                                
                                ],
                                alignItems: "center",
                                justifyContent: "center",
                                backgroundColor: aqiStatus.color,
                                cornerRadius: "10px",
                                paddingAll: "15px",
                                margin: "sm",
                                spacing: "md"
                            }
                        ]
                    },
                    // Advice Section
                    {
                        type: "box",
                        layout: "horizontal",
                        contents: [
                            {
                                type: "box",
                                layout: "vertical",
                                contents: [
                                    { 
                                        type: "text", 
                                        text: "💡", 
                                        size: "lg"
                                    }
                                ],
                                flex: 0,
                                paddingEnd: "8px"
                            },
                            {
                                type: "box",
                                layout: "vertical",
                                contents: [
                                    { 
                                        type: "text", 
                                        text: aqiStatus.desc, 
                                        size: "xs", 
                                        color: "#666666", 
                                        wrap: true,
                                        flex: 1
                                    }
                                ],
                                flex: 1
                            }
                        ],
                        backgroundColor: "#FFFBF0",
                        cornerRadius: "8px",
                        paddingAll: "12px",
                        margin: "md",
                        borderWidth: "1px",
                        borderColor: aqiStatus.color
                    }
                ],
                paddingAll: "16px",
                spacing: "none"
            },
            footer: {
                type: "box",
                layout: "vertical",
                contents: [
                    { 
                        type: "text", 
                        text: `ข้อมูลเมื่อ: ${data.time}`, 
                        size: "xxs", 
                        color: "#999999", 
                        align: "center" 
                    }
                ],
                paddingAll: "10px",
                backgroundColor: "#FAFAFA"
            }
        }
    };
};

module.exports = { createWeatherFlex, getAQIStatus };