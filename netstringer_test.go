package netstringer

import (
	"reflect"
	"sync"
	"testing"
)

func TestNetStringDecoder(t *testing.T) {
	decoder := NewDecoder()
	//decoder.SetDebugMode(true)

	testInputs := []string{
		"12:hello world!,",
		"17:5:hello,6:world!,,",
		"5:hello,6:world!,",
		"12:How are you?,9:I am fine,12:this is cool,",
		"12:hello", // Partial messages
		" world!,",
	}
	expectedOutputs := []string{
		"hello world!",
		"5:hello,6:world!,",
		"hello",
		"world!",
		"How are you?",
		"I am fine",
		"this is cool",
		"hello world!",
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// This will verify outputs in background as the decoder emits complete messages
	verifyDataOutputsFromDecoder(t, &wg, expectedOutputs, decoder)

	for _, testInput := range testInputs {
		decoder.FeedData([]byte(testInput))
	}

	close(decoder.DataOutput)
	wg.Wait()
}

func TestRipMsgDecoder(t *testing.T) {
	decoder := NewDecoder()
	decoder.SetEndSymbol(';') //Rip messages end with ; character

	testInputs := []string{
		"12:hello world!;",
		"17:5:hello,6:world!,;",
		"5:hello;6:world!;",
		"12:How are you?;9:I am fine;12:this is cool;",
		"12:hello", // Partial messages
		" world!;",
	}
	expectedOutputs := []string{
		"hello world!",
		"5:hello,6:world!,",
		"hello",
		"world!",
		"How are you?",
		"I am fine",
		"this is cool",
		"hello world!",
	}

	var wg sync.WaitGroup
	wg.Add(1)

	//this will verify outputs in background as the decoder emits complete messages
	verifyDataOutputsFromDecoder(t, &wg, expectedOutputs, decoder)

	for _, testInput := range testInputs {
		decoder.FeedData([]byte(testInput))
	}

	close(decoder.DataOutput)
	wg.Wait()
}

func TestOgpTextBinaryMsgDecoder(t *testing.T) {
	decoder := NewDecoder()
	//decoder.SetDebugMode(true)

	strBytes := []byte{0x18, 0x2d, 0x44, 0x54}
	str := string(strBytes)
	testInputs := []string{
		"12:How are you?,",
		"12,4:hello world!",
		str,
		",12:hello", // Partial messages
		" world!,",
	}
	expectedOutputs := []string{
		"How are you?",
		"hello world!",
	}
	expectedTextBinOutputs := []TextBinaryMsg{
		{[]byte("hello world!" + str), 12},
	}

	var wg sync.WaitGroup
	wg.Add(2)

	//verify decoded messages match expectedOutputs
	verifyDataOutputsFromDecoder(t, &wg, expectedOutputs, decoder)

	//verify decoded text-binary output messages match textBinOutputs
	go func(textBinOutputs []TextBinaryMsg, textBinDataChannel <-chan TextBinaryMsg) {
		j := 0
		for {
			select {
			case msg := <-textBinDataChannel:
				if j < len(textBinOutputs) {
					if !reflect.DeepEqual(msg, textBinOutputs[j]) {
						t.Error("Got", msg, "Expected", textBinOutputs[j])
					}
					j++
				} else {
					wg.Done()
					return
				}
			}
		}
	}(expectedTextBinOutputs, decoder.TextBinDataOutput)

	for _, input := range testInputs {
		decoder.FeedData([]byte(input))
	}

	close(decoder.DataOutput)
	close(decoder.TextBinDataOutput)
	wg.Wait()
}

func verifyDataOutputsFromDecoder(t *testing.T, wg *sync.WaitGroup, expectedOutputs []string, decoder NetStringDecoder) {
	go func(outputs []string, dataChannel <-chan []byte) {
		i := 0
		for {
			select {
			case msg := <-dataChannel:
				got := string(msg)
				if i < len(outputs) {
					if got != outputs[i] {
						t.Error("Got", got, "Expected", outputs[i])
					}
					i++
				} else {
					wg.Done()
					return
				}
			}
		}
	}(expectedOutputs, decoder.DataOutput)
}

func TestEncode(t *testing.T) {
	type TestCase struct {
		Input    string
		Expected string
	}

	testCases := []TestCase{
		{Input: "hello world!", Expected: "12:hello world!,"},
		{Input: "5:hello,6:world!,", Expected: "17:5:hello,6:world!,,"},
		{Input: "hello", Expected: "5:hello,"},
		{Input: "world!", Expected: "6:world!,"},
		{Input: "How are you?", Expected: "12:How are you?,"},
		{Input: "I am fine", Expected: "9:I am fine,"},
	}
	for _, testCase := range testCases {
		output := Encode([]byte(testCase.Input))
		if string(output) != testCase.Expected {
			t.Error("Got", string(output), "Expected", testCase.Expected)
		}
	}
}
