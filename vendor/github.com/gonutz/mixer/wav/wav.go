// Package wav provides functions to load audio files in the WAV (wave) format.
// Only uncompressed PCM formats are supported.
// Unknown chunks in the WAV file are simply ignored when loading.
package wav

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// Wave contains uncompressed PCM data with samples interleaved, e.g. for 2
// channels the layout is:
//    channel1[0] channel2[0] channel1[1] channel2[1] channel1[2] channel2[2]...
type Wave struct {
	ChannelCount     int
	SamplesPerSecond int
	BitsPerSample    int
	Data             []byte
}

// LoadFromFile opens the given file and calls Read on it.
func LoadFromFile(path string) (*Wave, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Read(file)
}

// Read reads sound data in the WAV format. It assumes that the data is in
// uncompressed PCM format and has exactly one format chunk and one data chunk.
// Unknown chunks are ignored.
func Read(r io.Reader) (*Wave, error) {
	var header waveHeader
	if err := binary.Read(r, endiannes, &header); err != nil {
		return nil, loadErr("reading 'RIFF' header", err)
	}
	if header.ChunkID != riffChunkID {
		return nil, errors.New("load WAV: expected 'RIFF' as the ID but got '" +
			header.ChunkID.String() + "'")
	}
	if header.WaveID != waveChunkID {
		return nil, errors.New("load WAV: expected 'WAVE' ID in header but got '" +
			header.WaveID.String() + "'")
	}

	data := make([]byte, header.ChunkSize-4)
	_, err := io.ReadFull(r, data[:])
	if err != nil {
		return nil, loadErr("illegal chunk size in header", err)
	}

	wav := &Wave{}
	if err := wav.parse(bytes.NewReader(data), false); err != nil {
		return nil, err
	}
	return wav, nil
}

func loadErr(msg string, err error) error {
	return errors.New("load WAV: " + msg + ": " + err.Error())
}

var endiannes = binary.LittleEndian

func (wav *Wave) parse(r *bytes.Reader, formatWasRead bool) error {
	if r.Len() == 0 {
		return nil
	}

	var header chunkHeader
	if err := binary.Read(r, endiannes, &header); err != nil {
		return loadErr("unable to read chunk header", err)
	}

	if header.ChunkID == formatChunkID {
		if formatWasRead {
			return errors.New("load WAV: two format chunks detected")
		}

		var chunk formatChunkExtended
		if header.ChunkSize == 16 {
			if err := binary.Read(r, endiannes, &(chunk.formatChunkBase)); err != nil {
				return loadErr("reading format chunk", err)
			}
		} else if header.ChunkSize == 18 {
			err := binary.Read(r, endiannes, &(chunk.formatChunkWithExtension))
			if err != nil {
				return loadErr("reading format chunk", err)
			}
		} else if header.ChunkSize == 40 {
			if err := binary.Read(r, endiannes, &chunk); err != nil {
				return loadErr("reading format chunk", err)
			}
		} else {
			return fmt.Errorf("load WAV: illegal format chunk header size: %v",
				header.ChunkSize)
		}

		if chunk.FormatTag != pcmFormat {
			return fmt.Errorf(
				"load WAV: unsupported format: %v (only PCM is supported)",
				chunk.FormatTag)
		}

		wav.ChannelCount = int(chunk.Channels)
		wav.SamplesPerSecond = int(chunk.SamplesPerSec)
		wav.BitsPerSample = int(chunk.BitsPerSample)
		formatWasRead = true
	} else if header.ChunkID == dataChunkID {
		data := make([]byte, header.ChunkSize)
		if _, err := io.ReadFull(r, data); err != nil {
			return err
		}

		if len(wav.Data) > 0 {
			return errors.New("load WAV: multiple data chunks found")
		}
		if !formatWasRead {
			return errors.New("load WAV: found data chunk before format chunk")
		}
		wav.Data = data

		if header.ChunkSize%2 == 1 {
			// there is one byte padding if the chunk size is odd
			if _, err := r.ReadByte(); err != nil {
				return loadErr("reading data chunk padding", err)
			}
		}
	} else {
		// skip unknown chunks
		io.CopyN(ioutil.Discard, r, int64(header.ChunkSize))
	}

	if r.Len() == 0 {
		if !formatWasRead {
			return errors.New("load WAV: file does not contain format information")
		}
		return nil
	}

	return wav.parse(r, formatWasRead)
}

// String prints the format information and data size in a human readable
// manner.
func (wav *Wave) String() string {
	return fmt.Sprintf(
		"Wave{%v channels, %v bits/sample, %v samples/sec, %v samples (%v bytes)}",
		wav.ChannelCount, wav.BitsPerSample, wav.SamplesPerSecond,
		len(wav.Data)/(wav.ChannelCount*wav.BitsPerSample/8), len(wav.Data),
	)
}

type chunkHeader struct {
	ChunkID   chunkID
	ChunkSize uint32
}

type waveHeader struct {
	// ChunkID should be "RIFF"
	// ChunkSize is 4 + length of the wave chunks that follow after the header
	chunkHeader
	WaveID chunkID // should be "WAVE"
}

type chunkID [4]byte

func (c chunkID) String() string { return string(c[:]) }

var (
	riffChunkID   = chunkID{'R', 'I', 'F', 'F'}
	waveChunkID   = chunkID{'W', 'A', 'V', 'E'}
	formatChunkID = chunkID{'f', 'm', 't', ' '}
	dataChunkID   = chunkID{'d', 'a', 't', 'a'}
	factChunkID   = chunkID{'f', 'a', 'c', 't'}
	listChunkID   = chunkID{'L', 'I', 'S', 'T'}
)

type waveChunk struct {
	id   uint32
	size uint32
}

type formatCode uint16

const (
	pcmFormat        formatCode = 0x0001
	ieeeFormat                  = 0x0003
	aLawFormat                  = 0x0006
	muLawFormat                 = 0x0007
	extensibleFormat            = 0xFFFE
)

func (f formatCode) String() string {
	switch f {
	case pcmFormat:
		return "PCM"
	case ieeeFormat:
		return "IEEE float"
	case aLawFormat:
		return "8-bit ITU-T G.711 A-law"
	case muLawFormat:
		return "8-bit ITU-T G.711 µ-law"
	case extensibleFormat:
		return "Extensible"
	default:
		return "Unknown"
	}
}

type formatChunkBase struct {
	FormatTag      formatCode
	Channels       uint16
	SamplesPerSec  uint32
	AvgBytesPerSec uint32
	BlockAlignment uint16
	BitsPerSample  uint16
}

type formatChunkWithExtension struct {
	formatChunkBase
	ExtensionSize uint16
}

type formatChunkExtended struct {
	formatChunkWithExtension
	ValidBitsPerSample uint16
	ChannelMask        uint32
	SubFormat          [16]byte
}

func (c formatChunkExtended) String() string {
	return fmt.Sprintf(`WAV Format Chunk {
	%s format
	%v channel(s)
	%v samples/s
	%v bytes/s
	%v byte block alignment
	%v bits/sample
}`,
		c.FormatTag, c.Channels, c.SamplesPerSec, c.AvgBytesPerSec,
		c.BlockAlignment, c.BitsPerSample)
}
