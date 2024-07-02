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

type CT_OfficeArtExtensionList struct {
	Ext []*CT_OfficeArtExtension
}

func NewCT_OfficeArtExtensionList() *CT_OfficeArtExtensionList {
	ret := &CT_OfficeArtExtensionList{}
	return ret
}

func (m *CT_OfficeArtExtensionList) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeToken(start)
	if m.Ext != nil {
		seext := xml.StartElement{Name: xml.Name{Local: "a:ext"}}
		for _, c := range m.Ext {
			e.EncodeElement(c, seext)
		}
	}
	e.EncodeToken(xml.EndElement{Name: start.Name})
	return nil
}

func (m *CT_OfficeArtExtensionList) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// initialize to default
lCT_OfficeArtExtensionList:
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch el := tok.(type) {
		case xml.StartElement:
			switch el.Name {
			case xml.Name{Space: "http://schemas.openxmlformats.org/drawingml/2006/main", Local: "ext"},
				xml.Name{Space: "http://purl.oclc.org/ooxml/drawingml/main", Local: "ext"}:
				tmp := NewCT_OfficeArtExtension()
				if err := d.DecodeElement(tmp, &el); err != nil {
					return err
				}
				m.Ext = append(m.Ext, tmp)
			default:
				unioffice.Log("skipping unsupported element on CT_OfficeArtExtensionList %v", el.Name)
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			break lCT_OfficeArtExtensionList
		case xml.CharData:
		}
	}
	return nil
}

// Validate validates the CT_OfficeArtExtensionList and its children
func (m *CT_OfficeArtExtensionList) Validate() error {
	return m.ValidateWithPath("CT_OfficeArtExtensionList")
}

// ValidateWithPath validates the CT_OfficeArtExtensionList and its children, prefixing error messages with path
func (m *CT_OfficeArtExtensionList) ValidateWithPath(path string) error {
	for i, v := range m.Ext {
		if err := v.ValidateWithPath(fmt.Sprintf("%s/Ext[%d]", path, i)); err != nil {
			return err
		}
	}
	return nil
}