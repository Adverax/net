package net

import "fmt"

type DefaultCodec struct{}

func (that *DefaultCodec) Headers() map[string]string {
	return nil
}

func (that *DefaultCodec) Encode(request interface{}) ([]byte, error) {
	if request == nil {
		return nil, nil
	}

	s := fmt.Sprintf("%v", request)
	return []byte(s), nil
}

func (that *DefaultCodec) Decode(data []byte, response interface{}) error {
	if len(data) == 0 {
		return nil
	}

	switch resp := response.(type) {
	case *[]byte:
		*resp = data
	case *string:
		*resp = string(data)
	}
	return nil
}

var CodecDefault Codec = &DefaultCodec{}
