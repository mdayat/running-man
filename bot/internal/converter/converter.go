package converter

import (
	"bytes"
	"encoding/binary"
)

func Int32SliceToBytes(data []int32) ([]byte, error) {
	buf := new(bytes.Buffer)
	for _, v := range data {
		if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func BytesToInt32Slice(data []byte) ([]int32, error) {
	var result []int32
	buf := bytes.NewReader(data)

	for buf.Len() > 0 {
		var v int32
		if err := binary.Read(buf, binary.LittleEndian, &v); err != nil {
			return nil, err
		}
		result = append(result, v)
	}

	return result, nil
}
