package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func dump(ast node) { // nolint:unused
	d := &dumper{w: os.Stderr}
	err := walk(d, ast)
	if err != nil {
		panic(err)
	}
}

type dumper struct { // nolint:unused
	w           io.Writer
	indentLevel int
}

func (d *dumper) visit(n node) (visitor, postVisit, error) { // nolint:unused
	_, err := fmt.Fprintf(d.w, strings.Repeat(". ", d.indentLevel)+"%d %T: %+v\n", n.id(), n, n)
	if err != nil {
		return nil, nil, fmt.Errorf("fprintf: %w", err)
	}
	d.indentLevel++

	return d, func() error {
		d.indentLevel--
		return nil
	}, nil
}
