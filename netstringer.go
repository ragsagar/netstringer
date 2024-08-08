package netstringer

import (
	"bytes"
	"log"
	"strconv"
)

const (
	parseLength = iota
	parseBinLength
	parseSeparator
	parseData
	parseEnd
)

const bufferCount = 10 //output buffered channels size

type TextBinaryMsg struct {
	Data       []byte
	TextLength int
}

type NetStringDecoder struct {
	parsedData            []byte
	length                int
	binLength             int // extra property to support mixed text and binary messages
	state                 int
	DataOutput            chan []byte
	TextBinDataOutput     chan TextBinaryMsg
	separatorSymbol       byte
	separatorLengthSymbol byte
	endSymbol             byte
	debugMode             bool
}

// Caller receives the parsed parsedData through the output channel.
func NewDecoder() NetStringDecoder {
	return NetStringDecoder{
		length:                0,
		state:                 parseLength,
		DataOutput:            make(chan []byte, bufferCount),
		TextBinDataOutput:     make(chan TextBinaryMsg, bufferCount),
		separatorSymbol:       byte(':'),
		separatorLengthSymbol: byte(','),
		endSymbol:             byte(','),
		debugMode:             false,
	}
}

func (decoder *NetStringDecoder) SetEndSymbol(symbol byte) {
	decoder.endSymbol = symbol
}

func (decoder *NetStringDecoder) SetDebugMode(mode bool) {
	decoder.debugMode = mode
}

func (decoder NetStringDecoder) DebugLog(v ...any) {
	if decoder.debugMode {
		log.Println(v...)
	}
}

func (decoder *NetStringDecoder) reset() {
	decoder.length = 0
	decoder.binLength = 0
	decoder.parsedData = []byte{}
	decoder.state = parseLength
}

func (decoder *NetStringDecoder) FeedData(data []byte) {
	// New incoming parsedData packets are fed into the decoder using this method.
	// Call this method every time we have a new set of parsedData.
	i := 0
	for i < len(data) {
		i = decoder.parse(i, data)
	}
}

func (decoder *NetStringDecoder) parse(i int, data []byte) int {
	switch decoder.state {
	case parseLength:
		i = decoder.parseLength(i, data)
	case parseBinLength:
		i = decoder.parseBinLength(i, data)
	case parseSeparator:
		i = decoder.parseSeparator(i, data)
	case parseData:
		i = decoder.parseData(i, data)
	case parseEnd:
		i = decoder.parseEnd(i, data)
	}
	return i
}

func (decoder *NetStringDecoder) parseLength(i int, data []byte) int {
	symbol := data[i]
	decoder.DebugLog("Parsing length, symbol =", string(symbol))
	if symbol >= '0' && symbol <= '9' {
		decoder.length = (decoder.length * 10) + (int(symbol) - 48)
		i++
	} else {
		decoder.state = parseSeparator
	}
	return i
}

func (decoder *NetStringDecoder) parseBinLength(i int, data []byte) int {
	symbol := data[i]
	decoder.DebugLog("Parsing bin length, symbol =", string(symbol))
	if symbol >= '0' && symbol <= '9' {
		decoder.binLength = (decoder.binLength * 10) + (int(symbol) - 48)
		i++
	} else {
		decoder.state = parseSeparator
	}
	return i
}

func (decoder *NetStringDecoder) parseSeparator(i int, data []byte) int {
	decoder.DebugLog("Parsing separator, symbol =", string(data[i]))
	switch data[i] {
	case decoder.separatorSymbol:
		decoder.length = decoder.length + decoder.binLength
		decoder.state = parseData
	case decoder.separatorLengthSymbol:
		decoder.state = parseBinLength
	default:
		// Something is wrong with the parsedData. let's reset everything to start looking for next valid parsedData
		decoder.reset()
	}
	return i + 1
}

func (decoder *NetStringDecoder) parseData(i int, data []byte) int {
	decoder.DebugLog("Parsing data, symbol =", string(data[i]))
	dataSize := len(data) - i
	dataLength := min(decoder.length, dataSize)
	decoder.parsedData = append(decoder.parsedData, data[i:i+dataLength]...)
	decoder.length = decoder.length - dataLength
	if decoder.length == 0 {
		decoder.state = parseEnd
	}
	// We already parsed till i + dataLength
	return i + dataLength
}

func (decoder *NetStringDecoder) parseEnd(i int, data []byte) int {
	decoder.DebugLog("Parsing end.")
	symbol := data[i]
	// Irrespective of what symbol we got we have to reset.
	// Since we are looking for new data from now onwards.
	defer decoder.reset()
	if symbol == decoder.endSymbol {
		// Symbol matches, that means this is valid data
		decoder.sendData(decoder.parsedData)
		return i + 1
	}
	return i
}

func (decoder *NetStringDecoder) sendData(parsedData []byte) {
	decoder.DebugLog("Successfully parsed data: ", string(parsedData))
	if decoder.binLength == 0 { // netstring messages emits on DataOutput channel
		decoder.DataOutput <- parsedData
	} else { // text binary message emits on TextBinDataOutput channel
		decoder.TextBinDataOutput <- TextBinaryMsg{parsedData, len(parsedData) - decoder.binLength}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Encode(data []byte) []byte {
	var buffer bytes.Buffer
	length := strconv.FormatInt(int64(len(data)), 10)
	buffer.WriteString(length)
	buffer.WriteByte(':')
	buffer.Write(data)
	buffer.WriteByte(',')
	return buffer.Bytes()
}
