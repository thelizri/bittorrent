package bencode

import (
	"fmt"
	"strconv"
)

func decodeString(bencode string, start int) (result string, index int, err error) {
	var firstColonIndex int
	for i := start; i < len(bencode); i++ {
		if bencode[i] == ':' {
			firstColonIndex = i
			break
		}
	}

	lengthStr := bencode[start:firstColonIndex]
	length, err := strconv.Atoi(lengthStr)

	if err != nil {
		return "", 0, err
	}

	index = firstColonIndex + 1 + length

	if index > len(bencode) {
		return "", 0, fmt.Errorf("invalid bencoded string: length exceeds available data")
	}

	if index < len(bencode) {
		nextChar := bencode[index]

		if nextChar != 'i' && nextChar != 'l' && nextChar != 'd' && nextChar != 'e' && (nextChar < '0' || nextChar > '9') {
			return "", 0, fmt.Errorf("invalid bencoded string: length mismatch")
		}

		if nextChar >= '0' && nextChar <= '9' {
			colonFound := false
			for i := index + 1; i < len(bencode); i++ {
				if bencode[i] == ':' {
					colonFound = true
					break
				}
				if bencode[i] < '0' || bencode[i] > '9' {
					return "", 0, fmt.Errorf("invalid bencoded string: length mismatch")
				}
			}
			if !colonFound {
				return "", 0, fmt.Errorf("invalid bencoded string: length mismatch")
			}
		}
	}

	return bencode[firstColonIndex+1 : index], index, nil
}

func decodeInt(bencode string, start int) (number int, index int, err error) {
	length := len(bencode)
	for index = start; index < length; index++ {
		if rune(bencode[index]) == 'e' {
			break
		}
	}
	numberStr := bencode[start+1 : index]
	number, err = strconv.Atoi(numberStr)
	if err != nil {
		return 0, 0, err
	}
	return number, index + 1, nil
}

func decodeList(bencode string, start int) (list []interface{}, index int, err error) {
	if start >= len(bencode) {
		return nil, start, fmt.Errorf("bad list")
	}

	list = make([]interface{}, 0)
	index = start + 1

	for {
		if index >= len(bencode) {
			return nil, index, fmt.Errorf("unexpected end of list")
		}
		if rune(bencode[index]) == 'e' {
			break
		}

		var result interface{}
		result, index, err = Decode(bencode, index)
		if err != nil {
			return nil, index, err
		}
		list = append(list, result)
	}

	return list, index + 1, nil
}

func decodeDict(bencode string, start int) (dict map[string]interface{}, index int, err error) {
	if start >= len(bencode) {
		return nil, start, fmt.Errorf("bad dict")
	}

	dict = make(map[string]interface{})
	index = start + 1
	for {
		if index >= len(bencode) {
			return nil, index, fmt.Errorf("unexpected end of dictionary")
		}
		if rune(bencode[index]) == 'e' {
			break
		}

		// Get key
		var key, value interface{}
		key, index, err = Decode(bencode, index)
		if err != nil {
			return nil, index, err
		}

		// Cast to string
		key_str, ok := key.(string)
		if !ok {
			return nil, index, fmt.Errorf("dict key is not a string")
		}

		// Get value
		value, index, err = Decode(bencode, index)
		if err != nil {
			return nil, index, err
		}

		dict[key_str] = value
	}

	return dict, index + 1, nil
}

func Decode(bencode string, start int) (result interface{}, index int, err error) {
	switch bencode[start] {
	case 'i':
		return decodeInt(bencode, start)
	case 'l':
		return decodeList(bencode, start)
	case 'd':
		return decodeDict(bencode, start)
	default:
		return decodeString(bencode, start)
	}
}
