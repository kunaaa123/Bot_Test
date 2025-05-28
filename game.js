// Thai Pop Song Lyrics
const songLyrics = {
    title: "แค่เธอที่ต้องการ",
    artist: "Calories Blah Blah",
    lyrics: [
        "เธอ...คนเดียวที่ฉันต้องการ",
        "ไม่ว่าเวลาจะผ่านไปนานเท่าไร",
        "ยังคงคิดถึงเธอทุกวัน",
        "อยากให้เธอรู้ว่าฉันรักเธอ"
    ],
    chorus: () => {
        console.log("🎵 ลา ลา ลา...");
        console.log("💕 เธอคือทุกอย่าง");
        console.log("🌟 ในใจของฉัน");
    }
};

// Test run the chorus
songLyrics.chorus();