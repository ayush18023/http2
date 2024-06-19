package main

import (
	"encoding/binary"
	"io"
)

type FrameType uint8

const (
	DataFrameType         FrameType = 0x0
	HeadersFrameType                = 0x1
	PriorityFrameType               = 0x2
	RstStreamFrameType              = 0x3
	SettingsFrameType               = 0x4
	PushPromiseFrameType            = 0x5
	PingFrameType                   = 0x6
	GoAwayFrameType                 = 0x7
	WindowUpdateFrameType           = 0x8
	ContinuationFrameType           = 0x9
)

type FlagType uint8

const (
	END_STREAM FlagType = 0x01
	PADDED              = 0x08
)

type Frame interface {
	Read(r io.Reader) error
	Write(w io.Writer) error
}

type Headers struct {
	Length uint32
	Type   FrameType
	Flags  uint8
	// Unused_Flags1  4
	// Padded_Flags   1
	// Unused_Flags2  2
	// END_STREAM  1
	StreamID uint32
}

func (h *Headers) Read(r io.Reader) error {
	var LengthType uint32
	if err := binary.Read(r, binary.BigEndian, &LengthType); err != nil {
		return err
	}
	h.Length = LengthType >> 8
	h.Type = FrameType(LengthType & 0xFF)
	var flags uint8
	if err := binary.Read(r, binary.BigEndian, &flags); err != nil {
		return err
	}
	return nil
}

func (h *Headers) Write(w io.Writer) error {
	var LengthType uint32 = h.Length<<8 + uint32(h.Type)
	if err := binary.Write(w, binary.BigEndian, LengthType); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, h.Flags); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, h.StreamID); err != nil {
		return err
	}

	return nil
}

// DATA Frame {
// 	Length (24),
// 	Type (8) = 0x00,

// 	Unused Flags (4),
// 	PADDED Flag (1),
// 	Unused Flags (2),
// 	END_STREAM Flag (1),

// 	Reserved (1),
// 	Stream Identifier (31),

//		[Pad Length (8)],
//		Data (..),
//		Padding (..2040),
//	  }

type PaddingType uint8

type DataFrame struct {
	*Headers
	PaddingLength PaddingType
	Data          []byte
	Padding       []byte
}

func (df *DataFrame) Read(r io.Reader) error {
	var isPadded bool = (df.Flags&PADDED == PADDED)
	if isPadded {
		if err := binary.Read(r, binary.BigEndian, &df.PaddingLength); err != nil {
			return err
		}
	}
	data := make([]byte, df.Length)
	if err := binary.Read(r, binary.BigEndian, &data); err != nil {
		return err
	}
	if isPadded {
		dataLen := len(data) - int(df.PaddingLength)
		df.Data = data[:dataLen]
		df.Padding = data[dataLen:]
	} else {
		df.Data = data
	}
	return nil
}

func (df *DataFrame) Write(w io.Writer) error {
	var err error
	var isPadded bool = (df.Flags&PADDED == PADDED)
	if err = df.Headers.Write(w); err != nil {
		return err
	}
	if isPadded {
		if err = binary.Write(w, binary.BigEndian, df.PaddingLength); err != nil {
			return err
		}
		if err = binary.Write(w, binary.BigEndian, df.Data); err != nil {
			return err
		}
		if err = binary.Write(w, binary.BigEndian, df.Padding); err != nil {
			return err
		}
	} else {
		if err = binary.Write(w, binary.BigEndian, df.Data); err != nil {
			return err
		}
	}
	return nil
}

func FrameFactory(headers *Headers) Frame {
	switch headers.Type {
	case DataFrameType:
		return &DataFrame{
			Headers: headers,
		}
	}
	return nil
}

func GetFrame(r io.Reader) (frame Frame, err error) {
	headers := new(Headers)
	err = headers.Read(r)
	if err != nil {
		return nil, err
	}
	frame = FrameFactory(headers)
	err = frame.Read(r)
	if err != nil {
		return nil, err
	}
	return frame, nil
}
