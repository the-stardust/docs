// Copyright 2017 FoxyUtils ehf. All rights reserved.
//
// DO NOT EDIT: generated by gooxml ECMA-376 generator
//
// Use of this source code is governed by the terms of the Affero GNU General
// Public License version 3.0 as published by the Free Software Foundation and
// appearing in the file LICENSE included in the packaging of this file. A
// commercial license can be purchased via https://unidoc.io website.

package content_types

import (
	"encoding/xml"
	"fmt"
)

type Default struct {
	CT_Default
}

func NewDefault() *Default {
	ret := &Default{}
	ret.CT_Default = *NewCT_Default()
	return ret
}

func (m *Default) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return m.CT_Default.MarshalXML(e, start)
}

func (m *Default) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// initialize to default
	m.CT_Default = *NewCT_Default()
	for _, attr := range start.Attr {
		if attr.Name.Local == "Extension" {
			parsed, err := attr.Value, error(nil)
			if err != nil {
				return err
			}
			m.ExtensionAttr = parsed
			continue
		}
		if attr.Name.Local == "ContentType" {
			parsed, err := attr.Value, error(nil)
			if err != nil {
				return err
			}
			m.ContentTypeAttr = parsed
			continue
		}
	}
	// skip any extensions we may find, but don't support
	for {
		tok, err := d.Token()
		if err != nil {
			return fmt.Errorf("parsing Default: %s", err)
		}
		if el, ok := tok.(xml.EndElement); ok && el.Name == start.Name {
			break
		}
	}
	return nil
}

// Validate validates the Default and its children
func (m *Default) Validate() error {
	return m.ValidateWithPath("Default")
}

// ValidateWithPath validates the Default and its children, prefixing error messages with path
func (m *Default) ValidateWithPath(path string) error {
	if err := m.CT_Default.ValidateWithPath(path); err != nil {
		return err
	}
	return nil
}