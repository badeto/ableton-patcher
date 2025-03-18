BUILD_CMD = go build
BUILD_SRC = ./cmd/ableton-patcher
BUILD_DIR = ./bin

all: darwin-arm64 darwin-amd64 windows-amd64

darwin-arm64:
	@echo "Building for macOS ARM64"
	GOOS=darwin \
	GOARCH=arm64 \
	${BUILD_CMD} -o=${BUILD_DIR}/ableton-patcher-darwin-arm64 ${BUILD_SRC}

darwin-amd64:
	@echo "Building for macOS AMD64"
	GOOS=darwin \
	GOARCH=amd64 \
	${BUILD_CMD} -o=${BUILD_DIR}/ableton-patcher-darwin-amd64 ${BUILD_SRC}

windows-amd64:
	@echo "Building for Windows AMD64"
	GOOS=windows \
	GOARCH=amd64 \
	${BUILD_CMD} -o=${BUILD_DIR}/ableton-patcher-windows-amd64.exe ${BUILD_SRC}