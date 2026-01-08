package codeowners

import (
	"io/fs"
	"time"
)

type mockfsinfo bool

func (f mockfsinfo) IsDir() bool        { return bool(f) }
func (f mockfsinfo) Name() string       { panic("not implemented") }
func (f mockfsinfo) Size() int64        { panic("not implemented") }
func (f mockfsinfo) Mode() fs.FileMode  { panic("not implemented") }
func (f mockfsinfo) ModTime() time.Time { panic("not implemented") }
func (f mockfsinfo) Sys() any           { panic("not implemented") }

func statMock(m ...map[string]bool) FsStat {
	return func(file string) (fs.FileInfo, error) {
		if m == nil || len(m) == 0 {
			return mockfsinfo(false), nil
		}
		isdir := m[0][file]
		return mockfsinfo(isdir), nil
	}
}
