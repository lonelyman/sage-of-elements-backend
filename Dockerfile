FROM golang:1.25.1-alpine

# ติดตั้ง Air สำหรับ Hot Reload
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# Copy ไฟล์ dependency ก่อนเพื่อใช้ cache
COPY go.mod go.sum ./
RUN go mod download

# Copy โค้ดทั้งหมด
COPY . .

# คำสั่งเริ่มต้น Container คือให้ Air ทำงาน
CMD ["air", "-c", ".air.toml"]