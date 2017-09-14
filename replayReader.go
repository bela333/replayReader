package replayReader

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
)

func NewReplay(r io.ReadCloser) *Replay {
	replay := Replay{r, nil}
	return &replay
}

type Replay struct {
	replayFile io.ReadCloser
	error      error
}

//Set p to the next element in the Replay file.
//If reading the packet was successful, it returns true.
//If it wasn't successful, run p.Error(), to get the error Next() returned.
//If Next() got to EOF, it return false and p.Error() returns nil.
func (r *Replay) Next(p *Packet) (success bool) {
	var time uint32
	err := binary.Read(r.replayFile, binary.BigEndian, &time)
	if err != nil {
		if err == io.EOF {
			return false
		}
		r.error = err
		return false
	}

	var len uint32
	err = binary.Read(r.replayFile, binary.BigEndian, &len)
	if err != nil {
		r.error = err
		return false
	}

	data := make([]byte, len)
	_, err = io.ReadAtLeast(r.replayFile, data, int(len))
	if err != nil {
		r.error = err
		return false
	}
	dataReader := bytes.NewReader(data)

	*p = Packet{Time: int(time), Len: int(len), Data: dataReader}
	return true
}

//Returns the error that happened after the latest Next()
func (r Replay) Error() (err error) {
	return r.error
}

//Packet can be used to read from packets in the Replay.
//Time is the milliseconds elapsed since the beginning of the Replay.
//Len is the length of the packet.
//Data is an io.ReadSeeker containing all the information of the packet.
type Packet struct {
	Time int
	Len  int
	Data io.ReadSeeker
}

//Reads an unsigned byte from the packet. Len: 1 byte
func (p *Packet) ReaduByte() (byte, error) {
	byteArray := make([]byte, 1)
	_, err := p.Data.Read(byteArray)
	return byteArray[0], err
}

//Reads a signed byte from the packet. Len: 1 byte
func (p *Packet) ReadByte() (int8, error) {
	unsignedByte, err := p.ReaduByte()
	return int8(unsignedByte), err
}

//Reads a short from the packet. Len: 2 bytes
func (p *Packet) ReadShort() (int16, error) {
	var output int16
	err := binary.Read(p.Data, binary.BigEndian, &output)
	return output, err
}

//Reads an unsigned short from the packet. Len: 2 bytes
func (p *Packet) ReaduShort() (uint16, error) {
	var output uint16
	err := binary.Read(p.Data, binary.BigEndian, &output)
	return output, err
}

//Reads an Integer from the packet. Len: 4 bytes
func (p *Packet) ReadInt() (int32, error) {
	var output int32
	err := binary.Read(p.Data, binary.BigEndian, &output)
	return output, err
}

//Reads a Long from the packet. Len: 8 bytes
func (p *Packet) ReadLong() (int64, error) {
	var output int64
	err := binary.Read(p.Data, binary.BigEndian, &output)
	return output, err
}

//Reads a Float from the packet. Len: 4 bytes
func (p *Packet) ReadFloat() (float32, error) {
	var uint32Form uint32
	err := binary.Read(p.Data, binary.BigEndian, &uint32Form)
	float32Form := math.Float32frombits(uint32Form)
	return float32Form, err
}

//Reads a Double-precision Float from the packet. Len: 8 bytes
func (p *Packet) ReadDouble() (float64, error) {
	var uint64Form uint64
	err := binary.Read(p.Data, binary.BigEndian, &uint64Form)
	float64Form := math.Float64frombits(uint64Form)
	return float64Form, err
}

//Reads a Boolean from the packet. Len: 1 byte
func (p *Packet) ReadBool() (bool, error) {
	boolbyte, err := p.ReaduByte()
	return boolbyte != 0, err
}

//Reads a Variable-length Integer from the packet. Len: len bytes
func (p *Packet) ReadVarInt() (n int, len int, err error) {
	count := 0
	result := int(0)
	last := byte(128)

	for (last & 128) != 0 {

		if count >= 5 {
			return result, count, VarIntTooBigError
		}

		last, err = p.ReaduByte()
		if err != nil {
			return result, count, err
		}

		value := last & 127

		result = result | int(value)<<uint(7*count)
		count++
	}
	return result, count, nil
}

//Reads a Variable-length Long from the packet. Len: len bytes
func (p *Packet) ReadVarLong() (n int64, len int, err error) {
	count := 0
	result := int64(0)
	last := byte(128)

	for (last & 128) != 0 {

		if count >= 10 {
			return result, count, VarIntTooBigError
		}

		last, err = p.ReaduByte()
		if err != nil {
			return result, count, err
		}

		value := last & 127

		result = result | int64(value)<<uint(7*count)
		count++
	}
	return result, count, nil
}

//Reads a byte array from the packet. Len: len bytes
func (p *Packet) ReaduByteArray(n int) (bytes []byte, len int, error error) {
	outputByteArray := make([]byte, n)
	n, err := io.ReadAtLeast(p.Data, outputByteArray, n)
	return outputByteArray, n, err
}

//Reads a string from the packet. Len: len bytes
func (p *Packet) ReadString() (result string, len int, error error) {
	stringLen, stringLenLen, err := p.ReadVarInt()
	if error != nil {
		return "", stringLenLen, err
	}
	outputString, byteArrayLen, err := p.ReaduByteArray(stringLen)
	return string(outputString), stringLenLen + byteArrayLen, err

}

//Same as io.Seeker.Seek
func (p *Packet) Seek(offset int64, whence int) (int64, error) {
	return p.Data.Seek(offset, whence)
}
