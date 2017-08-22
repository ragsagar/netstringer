package netstringer

import (
	"bytes"
	"log"
	"strconv"
)

const (
	PARSE_LENGTH = iota
	PARSE_SEPARATOR
	PARSE_DATA
	PARSE_END
)

const BUFFER_COUNT = 10

type NetStringDecoder struct {
	parsedData                 []byte
	length                     int
	state                      int
	DataOutput                 chan []byte
	separatorSymbol, endSymbol byte
	debugMode                  bool
}

// Caller receives the parsed parsedData through the output channel.
func NewDecoder() NetStringDecoder {
	return NetStringDecoder{
		length:          0,
		state:           PARSE_LENGTH,
		DataOutput:      make(chan []byte, BUFFER_COUNT),
		separatorSymbol: byte(':'),
		endSymbol:       byte(','),
		debugMode:       false,
	}

}

func (decoder *NetStringDecoder) SetDebugMode(mode bool) {
	decoder.debugMode = mode
}

func (decoder NetStringDecoder) DebugLog(v ...interface{}) {
	if decoder.debugMode {
		log.Println(v...)
	}
}

func (decoder *NetStringDecoder) reset() {
	decoder.length = 0
	decoder.parsedData = []byte{}
	decoder.state = PARSE_LENGTH
}

func (decoder *NetStringDecoder) FeedData(data []byte) {
	// New incoming parsedData packets are feeded into the decoder using this method.
	// Call this method every time we have a new set of parsedData.
	i := 0
	for i < len(data) {
		i = decoder.parse(i, data)
	}
}

func (decoder *NetStringDecoder) parse(i int, data []byte) int {
	switch decoder.state {
	case PARSE_LENGTH:
		i = decoder.parseLength(i, data)
	case PARSE_SEPARATOR:
		i = decoder.parseSeparator(i, data)
	case PARSE_DATA:
		i = decoder.parseData(i, data)
	case PARSE_END:
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
		decoder.state = PARSE_SEPARATOR
	}

	return i
}

func (decoder *NetStringDecoder) parseSeparator(i int, data []byte) int {
	decoder.DebugLog("Parsing separator, symbol =", string(data[i]))
	if data[i] != decoder.separatorSymbol {
		// Something is wrong with the parsedData.
		// let's reset everything to start looking for next valid parsedData
		decoder.reset()
	} else {
		decoder.state = PARSE_DATA
	}
	i++
	return i
}

func (decoder *NetStringDecoder) parseData(i int, data []byte) int {
	decoder.DebugLog("Parsing data, symbol =", string(data[i]))
	dataSize := len(data) - i
	dataLength := min(decoder.length, dataSize)
	decoder.parsedData = append(decoder.parsedData, data[i:i+dataLength]...)
	decoder.length = decoder.length - dataLength
	if decoder.length == 0 {
		decoder.state = PARSE_END
	}
	// We already parsed till i + dataLength
	return i + dataLength
}

func (decoder *NetStringDecoder) parseEnd(i int, data []byte) int {
	decoder.DebugLog("Parsing end.")
	symbol := data[i]
	if symbol == decoder.endSymbol {
		// Symbol matches, that means this is valid data
		decoder.sendData(decoder.parsedData)
	}
	// Irrespective of what symbol we got we have to reset.
	// Since we are looking for new data from now onwards.
	decoder.reset()
	return i
}

func (decoder *NetStringDecoder) sendData(parsedData []byte) {
	decoder.DebugLog("Successfully parsed data: ", string(parsedData))
	decoder.DataOutput <- parsedData
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
