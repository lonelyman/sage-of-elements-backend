// pkg/logger/logger.go
package applogger

// Logger คือ "สัญญาใจ" หรือ Interface ที่นักข่าวทุกคนต้องทำตาม
type Logger interface {
	// --- Structured Logging ---
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Success(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, err error, args ...any)

	// --- Simple Dumping ---
	Dump(a ...any)

	// --- Highlight ---
	Highlight(color string, msg string, data ...any)
}

// Color Constants - รหัสสี ANSI สำหรับตกแต่ง Log ให้น่าอ่าน
const (
	ColorReset = "\033[0m" // รีเซ็ตสีทั้งหมด กลับเป็นสีปกติ

	// --- สีหลัก (ฉบับสว่าง/นีออน) ---
	ColorRed    = "\033[91m" // สีแดงสว่าง (สำหรับ Error)
	ColorGreen  = "\033[92m" // สีเขียวสว่าง (สำหรับ Success)
	ColorYellow = "\033[93m" // สีเหลืองสว่าง (สำหรับ Warning)
	ColorBlue   = "\033[94m" // สีน้ำเงินสว่าง (สำหรับ Debug)
	ColorPurple = "\033[95m" // สีม่วงสว่าง (สำหรับ Dump ข้อมูล)
	ColorCyan   = "\033[96m" // สีฟ้า/ไซแอนสว่าง (สำหรับ Info)

	// --- สีเพิ่มเติม ---
	ColorGray   = "\033[90m"       // สีเทาสว่าง
	ColorWhite  = "\033[97m"       // สีขาวสว่าง
	ColorOrange = "\033[38;5;208m" // สีส้ม (256-color)
	ColorPink   = "\033[38;5;205m" // สีชมพู (256-color)
	ColorLime   = "\033[38;5;154m" // สีเขียวมะนาว / นีออน (256-color)
	ColorGold   = "\033[38;5;220m" // สีทอง (256-color)
)
