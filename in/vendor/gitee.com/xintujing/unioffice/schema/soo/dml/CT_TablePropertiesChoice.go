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
	"gitee.com/xintujing/unioffice/schema/soo/ofc/sharedTypes"
)

type CT_TablePropertiesChoice struct {
	TableStyle   *CT_TableStyle
	TableStyleId *string
}

func NewCT_TablePropertiesChoice() *CT_TablePropertiesChoice {
	ret := &CT_TablePropertiesChoice{}
	return ret
}

func (m *CT_TablePropertiesChoice) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if m.TableStyle != nil {
		setableStyle := xml.StartElement{Name: xml.Name{Local: "a:tableStyle"}}
		e.EncodeElement(m.TableStyle, setableStyle)
	}
	if m.TableStyleId != nil {
		setableStyleId := xml.StartElement{Name: xml.Name{Local: "a:tableStyleId"}}
		unioffice.AddPreserveSpaceAttr(&setableStyleId, *m.TableStyleId)
		e.EncodeElement(m.TableStyleId, setableStyleId)
	}
	return nil
}

func (m *CT_TablePropertiesChoice) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// initialize to default
lCT_TablePropertiesChoice:
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch el := tok.(type) {
		case xml.StartElement:
			switch el.Name {
			case xml.Name{Space: "http://schemas.openxmlformats.org/drawingml/2006/main", Local: "tableStyle"},
				xml.Name{Space: "http://purl.oclc.org/ooxml/drawingml/main", Local: "tableStyle"}:
				m.TableStyle = NewCT_TableStyle()
				if err := d.DecodeElement(m.TableStyle, &el); err != nil {
					return err
				}
			case xml.Name{Space: "http://schemas.openxmlformats.org/drawingml/2006/main", Local: "tableStyleId"},
				xml.Name{Space: "http://purl.oclc.org/ooxml/drawingml/main", Local: "tableStyleId"}:
				m.TableStyleId = new(string)
				if err := d.DecodeElement(m.TableStyleId, &el); err != nil {
					return err
				}
			default:
				unioffice.Log("skipping unsupported element on CT_TablePropertiesChoice %v", el.Name)
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			break lCT_TablePropertiesChoice
		case xml.CharData:
		}
	}
	return nil
}

// Validate validates the CT_TablePropertiesChoice and its children
func (m *CT_TablePropertiesChoice) Validate() error {
	return m.ValidateWithPath("CT_TablePropertiesChoice")
}

// ValidateWithPath validates the CT_TablePropertiesChoice and its children, prefixing error messages with path
func (m *CT_TablePropertiesChoice) ValidateWithPath(path string) error {
	if m.TableStyle != nil {
		if err := m.TableStyle.ValidateWithPath(path + "/TableStyle"); err != nil {
			return err
		}
	}
	if m.TableStyleId != nil {
		if !sharedTypes.ST_GuidPatternRe.MatchString(*m.TableStyleId) {
			return fmt.Errorf(`%s/m.TableStyleId must match '%s' (have %v)`, path, sharedTypes.ST_GuidPatternRe, *m.TableStyleId)
		}
	}
	return nil
}
