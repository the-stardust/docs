// Copyright 2017 FoxyUtils ehf. All rights reserved.
//
// DO NOT EDIT: generated by gooxml ECMA-376 generator
//
// Use of this source code is governed by the terms of the Affero GNU General
// Public License version 3.0 as published by the Free Software Foundation and
// appearing in the file LICENSE included in the packaging of this file. A
// commercial license can be purchased via https://unidoc.io website.

package relationships

import (
	"encoding/xml"
	"fmt"
)

type Relationship struct {
	CT_Relationship
}

func NewRelationship() *Relationship {
	ret := &Relationship{}
	ret.CT_Relationship = *NewCT_Relationship()
	return ret
}

func (m *Relationship) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return m.CT_Relationship.MarshalXML(e, start)
}

func (m *Relationship) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// initialize to default
	m.CT_Relationship = *NewCT_Relationship()
	for _, attr := range start.Attr {
		if attr.Name.Local == "TargetMode" {
			m.TargetModeAttr.UnmarshalXMLAttr(attr)
			continue
		}
		if attr.Name.Local == "Target" {
			parsed, err := attr.Value, error(nil)
			if err != nil {
				return err
			}
			m.TargetAttr = parsed
			continue
		}
		if attr.Name.Local == "Type" {
			parsed, err := attr.Value, error(nil)
			if err != nil {
				return err
			}
			m.TypeAttr = parsed
			continue
		}
		if attr.Name.Local == "Id" {
			parsed, err := attr.Value, error(nil)
			if err != nil {
				return err
			}
			m.IdAttr = parsed
			continue
		}
	}
	// skip any extensions we may find, but don't support
	for {
		tok, err := d.Token()
		if err != nil {
			return fmt.Errorf("parsing Relationship: %s", err)
		}
		if el, ok := tok.(xml.EndElement); ok && el.Name == start.Name {
			break
		}
	}
	return nil
}

// Validate validates the Relationship and its children
func (m *Relationship) Validate() error {
	return m.ValidateWithPath("Relationship")
}

// ValidateWithPath validates the Relationship and its children, prefixing error messages with path
func (m *Relationship) ValidateWithPath(path string) error {
	if err := m.CT_Relationship.ValidateWithPath(path); err != nil {
		return err
	}
	return nil
}