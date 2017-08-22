package netstringer

import (
	"testing"
)

func TestNewNetStringDecoder(t *testing.T) {
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

	go func(outputs []string, dataChannel chan []byte) {
		for _, output := range outputs {
			got := string(<-dataChannel)
			if got != output {
				t.Error("Got", got, "Expected", output)
			}
		}
	}(expectedOutputs, decoder.DataOutput)

	for _, testInput := range testInputs {
		decoder.FeedData([]byte(testInput))
	}

}

func TestEncode(t *testing.T) {
	type TestCase struct {
		Input    string
		Expected string
	}

	testCases := []TestCase{
		TestCase{Input: "hello world!", Expected: "12:hello world!,"},
		TestCase{Input: "5:hello,6:world!,", Expected: "17:5:hello,6:world!,,"},
		TestCase{Input: "hello", Expected: "5:hello,"},
		TestCase{Input: "world!", Expected: "6:world!,"},
		TestCase{Input: "How are you?", Expected: "12:How are you?,"},
		TestCase{Input: "I am fine", Expected: "9:I am fine,"},
	}
	for _, testCase := range testCases {
		output := Encode([]byte(testCase.Input))
		if string(output) != testCase.Expected {
			t.Error("Got", string(output), "Expected", testCase.Expected)
		}
	}

}
