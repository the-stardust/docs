// Copyright 2019 FoxyUtils ehf. All rights reserved.
//
// DO NOT EDIT: generated by gooxml ECMA-376 generator
//
// Use of this source code is governed by the terms of the Affero GNU General
// Public License version 3.0 as published by the Free Software Foundation and
// appearing in the file LICENSE included in the packaging of this file. A
// commercial license can be purchased via https://unidoc.io website.

package custom_properties

import (
	"encoding/xml"
	"fmt"

	"gitee.com/xintujing/unioffice"
)

type CT_Properties struct {
	Property []*CT_Property
}

func NewCT_Properties() *CT_Properties {
	ret := &CT_Properties{}
	return ret
}

func (m *CT_Properties) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeToken(start)
	if m.Property != nil {
		seproperty := xml.StartElement{Name: xml.Name{Local: "property"}}
		for _, c := range m.Property {
			e.EncodeElement(c, seproperty)
		}
	}
	e.EncodeToken(xml.EndElement{Name: start.Name})
	return nil
}

func (m *CT_Properties) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// initialize to default
lCT_Properties:
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch el := tok.(type) {
		case xml.StartElement:
			switch el.Name {
			case xml.Name{Space: "http://schemas.openxmlformats.org/officeDocument/2006/custom-properties", Local: "property"}:
				tmp := NewCT_Property()
				if err := d.DecodeElement(tmp, &el); err != nil {
					return err
				}
				m.Property = append(m.Property, tmp)
			default:
				unioffice.Log("skipping unsupported element on CT_Properties %v", el.Name)
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			break lCT_Properties
		case xml.CharData:
		}
	}
	return nil
}

// Validate validates the CT_Properties and its children
func (m *CT_Properties) Validate() error {
	return m.ValidateWithPath("CT_Properties")
}

// ValidateWithPath validates the CT_Properties and its children, prefixing error messages with path
func (m *CT_Properties) ValidateWithPath(path string) error {
	for i, v := range m.Property {
		if err := v.ValidateWithPath(fmt.Sprintf("%s/Property[%d]", path, i)); err != nil {
			return err
		}
	}
	return nil
}
