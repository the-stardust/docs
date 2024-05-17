// Copyright 2017 FoxyUtils ehf. All rights reserved.
//
// DO NOT EDIT: generated by gooxml ECMA-376 generator
//
// Use of this source code is governed by the terms of the Affero GNU General
// Public License version 3.0 as published by the Free Software Foundation and
// appearing in the file LICENSE included in the packaging of this file. A
// commercial license can be purchased via https://unidoc.io website.

package dml

import (
	"encoding/xml"
	"fmt"

	"gitee.com/xintujing/unioffice"
)

type CT_Path2DCubicBezierTo struct {
	Pt []*CT_AdjPoint2D
}

func NewCT_Path2DCubicBezierTo() *CT_Path2DCubicBezierTo {
	ret := &CT_Path2DCubicBezierTo{}
	return ret
}

func (m *CT_Path2DCubicBezierTo) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeToken(start)
	sept := xml.StartElement{Name: xml.Name{Local: "a:pt"}}
	for _, c := range m.Pt {
		e.EncodeElement(c, sept)
	}
	e.EncodeToken(xml.EndElement{Name: start.Name})
	return nil
}

func (m *CT_Path2DCubicBezierTo) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// initialize to default
lCT_Path2DCubicBezierTo:
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch el := tok.(type) {
		case xml.StartElement:
			switch el.Name {
			case xml.Name{Space: "http://schemas.openxmlformats.org/drawingml/2006/main", Local: "pt"},
				xml.Name{Space: "http://purl.oclc.org/ooxml/drawingml/main", Local: "pt"}:
				tmp := NewCT_AdjPoint2D()
				if err := d.DecodeElement(tmp, &el); err != nil {
					return err
				}
				m.Pt = append(m.Pt, tmp)
			default:
				unioffice.Log("skipping unsupported element on CT_Path2DCubicBezierTo %v", el.Name)
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			break lCT_Path2DCubicBezierTo
		case xml.CharData:
		}
	}
	return nil
}

// Validate validates the CT_Path2DCubicBezierTo and its children
func (m *CT_Path2DCubicBezierTo) Validate() error {
	return m.ValidateWithPath("CT_Path2DCubicBezierTo")
}

// ValidateWithPath validates the CT_Path2DCubicBezierTo and its children, prefixing error messages with path
func (m *CT_Path2DCubicBezierTo) ValidateWithPath(path string) error {
	for i, v := range m.Pt {
		if err := v.ValidateWithPath(fmt.Sprintf("%s/Pt[%d]", path, i)); err != nil {
			return err
		}
	}
	return nil
}
