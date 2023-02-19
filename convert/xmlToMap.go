package convert

import (
	"encoding/xml"
	"io"
)

type XMLMap map[string]string

type xmlMapEntry struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

func (m XMLMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(m) == 0 {
		return nil
	}

	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for k, v := range m {
		err = e.Encode(xmlMapEntry{XMLName: xml.Name{Local: k}, Value: v})
		if err != nil {
			break
		}
	}
	if err != nil {
		return err
	}

	return e.EncodeToken(start.End())
}

// 这种写法只能转换单层
func (m *XMLMap) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*m = XMLMap{}
	for {
		var e xmlMapEntry

		err := d.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		(*m)[e.XMLName.Local] = e.Value
	}
	return nil
}

// 这个方法只能转换单层XML
func XMLToMap(buf []byte) (stringMap map[string]string, err error) {
	stringMap = make(map[string]string)

	err = xml.Unmarshal(buf, (*XMLMap)(&stringMap))
	if err != nil {
		return nil, err
	}
	return stringMap, err
}
