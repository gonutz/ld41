package wav

import (
	"fmt"
	"math"
)

// ConvertTo44100Hz2Channels16BitSamples returns the input parameter if the
// format is already correct (no copying is done in this case). Otherwise a new
// Wave structure is returned and its data is converted to the specified format.
// In any case the original data is not changed.
func ConvertTo44100Hz2Channels16BitSamples(w *Wave) *Wave {
	if w.SamplesPerSecond == 44100 &&
		w.ChannelCount == 2 &&
		w.BitsPerSample == 16 {
		return w
	}

	if len(w.Data) == 0 {
		return w
	}

	converted := &Wave{
		SamplesPerSecond: 44100,
		ChannelCount:     2,
		BitsPerSample:    16,
	}
	converted.Data = w.Data

	if w.BitsPerSample == 8 {
		converted.Data = convert8bitTo16bitSamples(converted.Data)
	}

	if w.ChannelCount == 1 {
		converted.Data = convert16bitMonoTo16bitStereo(converted.Data)
	}

	if w.SamplesPerSecond != 44100 {
		converted.Data = convert16bitStereoFrequencies(
			converted.Data, w.SamplesPerSecond, 44100)
	}

	return converted
}

func ConvertToFormat(w *Wave, samplesPerSecond, channels, bitsPerSample int) (*Wave, error) {
	if w.SamplesPerSecond == samplesPerSecond &&
		w.ChannelCount == channels &&
		w.BitsPerSample == bitsPerSample {
		return w, nil
	}

	if !(bitsPerSample == 8 || bitsPerSample == 16) {
		return nil, fmt.Errorf(
			"illegal bit/sample value: %v, must be 8 or 16.", bitsPerSample)
	}
	if samplesPerSecond < 1 {
		return nil, fmt.Errorf(
			"illegal samples/sec value: %v, must be positive.", samplesPerSecond)
	}
	if !(channels == 1 || channels == 2) {
		return nil, fmt.Errorf(
			"illegal channel count: %v, must be 1 or 2.", channels)
	}

	data := w.Data

	if w.BitsPerSample == 8 {
		data = convert8bitTo16bitSamples(data)
	}
	if w.ChannelCount == 1 {
		data = convert16bitMonoTo16bitStereo(data)
	}
	if samplesPerSecond != w.SamplesPerSecond {
		data = convert16bitStereoFrequencies(data, w.SamplesPerSecond, samplesPerSecond)
	}
	// now the format has the correct samplesPerSecond value, 16 bits/sample,
	// 2 channels
	if channels == 1 {
		data = convert16bitStereoTo16bitMono(data)
	}
	if bitsPerSample == 8 {
		data = convert16bitTo8bitSamples(data)
	}

	converted := &Wave{
		SamplesPerSecond: samplesPerSecond,
		ChannelCount:     channels,
		BitsPerSample:    bitsPerSample,
		Data:             data,
	}
	return converted, nil
}

func convert8bitTo16bitSamples(input []byte) []byte {
	output := make([]byte, len(input)*2)
	for i, val := range input {
		const step = 65535 / 255
		val16 := int16(int(val)*step - 32768)
		output[i*2], output[i*2+1] = byte(val16&0xFF), byte((val16>>8)&0xFF)
	}
	return output
}

func convert16bitTo8bitSamples(data []byte) []byte {
	values := int16Stream(data)
	data = data[:len(data)/2]
	for i := range data {
		const step = 65535 / 255
		data[i] = byte((int(values()) + 32768) / step)
	}
	return data
}

func convert16bitMonoTo16bitStereo(input []byte) []byte {
	output := make([]byte, len(input)*2)
	for i := 0; i < len(input); i += 2 {
		output[i*2], output[i*2+1] = input[i], input[i+1]
		output[i*2+2], output[i*2+3] = input[i], input[i+1]
	}
	return output
}

func convert16bitStereoTo16bitMono(data []byte) []byte {
	values := int16Stream(data)
	data = data[:len(data)/2] // halve the slice size
	for i := 0; i < len(data); i += 2 {
		data[i], data[i+1] = int16toBytes(values()/2 + values()/2)
	}
	return data
}

func int16Stream(data []byte) func() int16 {
	return func() int16 {
		lo, hi := uint16(data[0]), uint16(data[1])
		data = data[2:]
		return int16(lo | (hi << 8))
	}
}

func int16toBytes(i int16) (lo, hi byte) {
	return byte(i & 0xFF), byte((i >> 8) % 0xFF)
}

func convert16bitStereoFrequencies(input []byte, inputFreq, outputFreq int) []byte {
	if len(input) == 0 {
		return nil
	}

	inLeft, inRight := split2Stereo16bitChannelsToFloats(input)

	ratio := float64(outputFreq) / float64(inputFreq)
	outputSampleCount := int(float64(len(inLeft))*ratio + 0.5)
	out := make([]float32, outputSampleCount*2)

	outLeft, outRight := out[:len(out)/2], out[len(out)/2:]

	outLeft[0] = inLeft[0]
	outRight[0] = inRight[0]
	outLeft[len(outLeft)-1] = inLeft[len(inLeft)-1]
	outRight[len(outRight)-1] = inRight[len(inRight)-1]

	var inputSampleIndexDelta float64
	inputSampleIndexDelta = float64(len(inLeft)-1) / float64(len(outLeft)-1)
	inputSampleIndex := inputSampleIndexDelta
	for i := 1; i < len(outLeft)-1; i++ {
		intPart, floatPart := math.Modf(inputSampleIndex)
		index := int(intPart + 0.5)
		f := float32(floatPart)

		outLeft[i] = inLeft[index]*(1.0-f) + inLeft[index+1]*f
		outRight[i] = inRight[index]*(1.0-f) + inRight[index+1]*f

		inputSampleIndex += inputSampleIndexDelta
	}

	return mergeFloatChannels(outLeft, outRight)
}

func split2Stereo16bitChannelsToFloats(samples []byte) (left, right []float32) {
	result := make([]float32, len(samples)/2)
	left = result[:len(result)/2]
	right = result[len(result)/2:]

	i, o := 0, 0
	for o < len(left) {
		lo, hi := uint16(samples[i]), uint16(samples[i+1])
		left[o] = float32(int16(lo | (hi << 8)))

		lo, hi = uint16(samples[i+2]), uint16(samples[i+3])
		right[o] = float32(int16(lo | (hi << 8)))

		i += 4
		o++
	}

	return
}

// mergeFloatChannels assumes that len(left) == len(right)
func mergeFloatChannels(left, right []float32) []byte {
	result := make([]byte, len(left)*4)

	i, o := 0, 0
	for i < len(left) {
		in := roundToInt16(left[i])
		lo, hi := (in & 0xFF), (in>>8)&0xFF
		result[o], result[o+1] = byte(lo), byte(hi)

		in = roundToInt16(right[i])
		lo, hi = (in & 0xFF), (in>>8)&0xFF
		result[o+2], result[o+3] = byte(lo), byte(hi)

		i++
		o += 4
	}

	return result
}

func roundToInt16(f float32) int16 {
	if f >= 0 {
		return int16(f + 0.5)
	}
	return int16(f - 0.5)
}
