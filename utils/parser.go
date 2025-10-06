package utils

import (
	peparser "github.com/saferwall/pe"
)

type Parser struct {
	Path   string
	peInfo *peparser.File
}

func NewParser(path string) *Parser {
	return &Parser{Path: path}
}

func (p *Parser) Parse() error {
	pe, err := peparser.New(p.Path, &peparser.Options{})
	if err != nil {
		return err
	}
	err = pe.Parse()
	if err != nil {
		return err
	}
	p.peInfo = pe

	return nil
}

func (p *Parser) Close() {
	if p.peInfo != nil {
		p.peInfo.Close()
	}
}

func (p *Parser) GetExportInfo() *peparser.Export {
	if p.peInfo != nil {
		return &p.peInfo.Export
	}
	return nil
}
