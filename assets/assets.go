// Code generated by go-bindata.
// sources:
// plugins/codeamp/schema.graphql
// plugins/codeamp/static/index.html
// DO NOT EDIT!

package assets

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _pluginsCodeampSchemaGraphql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x58\xcd\x92\x1b\x27\x10\xbe\xcf\x53\xa0\xf2\x45\xae\xda\x27\x98\xa3\xbd\x4e\xe2\x24\x4e\x36\xab\xf8\xe4\xf2\x81\x1d\xb1\x12\xd9\x99\x41\x06\x46\x89\x2a\x95\x77\x4f\x01\x0d\xd3\xcd\x80\xb4\x5a\x3b\x55\xb9\xec\x0e\x2d\x68\xfa\xfb\xe8\x3f\x30\x1d\xef\xb9\x66\xbf\xcb\x41\x34\xf0\xfd\xe3\xe6\xd7\x5f\x9a\xc6\x74\x7b\x31\x70\xf6\x77\xc3\xd8\x97\x49\xe8\x53\xcb\x7e\x73\xff\x1a\xc6\x86\xc9\x72\x2b\xd5\xd8\xb2\x0f\xf0\xd5\xfc\xd3\x34\xaf\xe0\x77\x7b\x3a\x88\xf0\xe9\xd7\xbe\x62\x1f\x8d\xd0\x0d\x63\x93\x11\x7a\x2d\xb7\x2d\x7b\x7f\xfb\xba\x8d\xc2\xf0\xab\x81\x9f\xcd\xfa\x75\xcb\x3e\x39\xc9\xe7\x95\xff\xf1\x4e\xab\x3f\x44\x67\x1b\xc6\x0e\xe1\x0b\x14\xdc\x30\xd3\x4f\xbb\x96\x6d\xac\x96\xe3\xee\x86\x8d\x7c\x10\xf3\x48\x8c\x47\xa9\xd5\x38\x88\xd1\xbe\xbf\x8d\xe2\xd7\x2d\xd2\x96\x34\x9b\x59\xb5\x59\xc3\xc7\x46\x70\xdd\xed\xd3\xf4\x30\x7c\x3f\x1e\x26\xeb\xac\x03\x29\x18\xf8\x9d\xe0\x76\xd2\xc2\xa9\x79\x84\x4f\x8f\x01\xe4\x30\x6b\x23\xf4\x51\x76\x7e\x96\x81\x4f\x3f\x0b\xe4\x74\x16\x33\x07\xd1\xa1\xa9\x1b\x37\xc4\xd3\x9d\x00\x96\xdc\x8b\x5e\x70\xe3\x15\x6b\xf8\xf4\x33\x41\x0e\xb3\xde\xcd\x7c\xb8\x99\x88\x9e\x19\x33\xa2\xd3\x29\x40\x4b\x96\x4a\xd8\x91\x6b\xc9\x1f\x7a\x00\xd4\x69\x61\xa3\x81\xee\x9b\x9e\xdd\xbb\xbf\xac\x18\x8d\x54\xa3\xc7\xe5\xb6\x8f\x02\xb3\xae\x1d\xd4\xa7\xb4\xa8\xa2\x0b\x1d\xdb\x2c\x5b\xa3\xe3\xc9\x15\x00\x21\x68\xf3\x99\xb2\x4c\xc3\x7d\x26\x8d\x26\x08\x3d\x48\x93\x36\x9f\x47\x6e\x91\x8b\x98\x55\x08\x82\x14\x12\x3e\x0e\xe2\x08\x42\xe1\xad\x16\xdc\x8a\x68\x7a\xc3\x58\xe7\x05\x60\x74\x3c\x8c\xe4\x7a\xd1\xe9\xb0\xe3\x7e\x3c\x6c\xa9\x8a\xc9\x0b\xae\x51\x01\x56\x00\xfc\x64\x05\x00\x5f\x83\xbc\x8d\xa4\x45\x15\xf7\x69\xfe\x2b\xb6\xb1\xea\x80\x14\x18\xab\x0e\x71\x79\x88\xd0\x55\xb6\x00\xf6\x04\x9f\x4e\x7b\x82\x4b\xaf\x41\xde\xc6\x20\x88\x7b\x6e\xd2\xfc\x84\x7c\x16\x05\xe4\xd7\xa8\xb8\x15\xbd\x20\x56\x6c\xbd\xe0\x1a\x15\x14\x48\x74\x6a\x82\xc6\x05\xe8\x1a\x45\x6f\x52\xe0\x06\x99\xce\x4d\x58\x9f\xd0\x65\x7a\x09\xc4\x17\xe9\xa5\x90\xa3\x5e\x82\xfb\x45\x7a\x81\x07\x14\xc2\x89\x06\x94\x2a\x70\x88\xb7\x38\x87\x44\xb5\xef\xc8\xfa\x44\x03\x55\x1b\x58\xf8\x1a\xb5\xc0\x02\x55\x1b\x48\xf8\x1a\xb5\x4b\x12\x52\x6e\x44\x4e\xe1\x92\xe2\x3a\xe4\x49\xc7\xa3\xfb\x3f\xb3\xea\x46\x15\xe4\x58\x57\x74\x84\xe7\xe9\x5a\xc2\x4d\xba\x18\x3a\xfd\xe7\x29\x8b\x20\xf3\x44\x0e\x87\x1d\xc5\xeb\x34\xa1\x65\x49\x98\x88\x8b\x02\x02\x35\xd7\x08\xe7\xfc\x02\x8d\x11\x70\xae\x11\x8e\xf8\x05\x1a\x73\xd4\x79\xb2\x9e\x75\xe6\x85\xa8\x5d\x94\xab\x2c\x09\x9f\x27\x23\x4f\xe9\xdf\x6c\x23\xc4\x11\xc8\x02\x3b\xff\x11\x20\xd7\xd0\xe1\x3a\x99\x70\xb9\xfe\x0e\x55\xd3\xf5\x44\xc7\xa1\x37\x44\x82\xd4\x7a\x85\xfe\x60\x05\x05\x99\x16\x42\xd2\xd5\x2c\x28\x2c\x74\x3c\x58\x36\x03\x41\xc2\xb4\x2d\x12\xc6\xbd\xdf\x28\xf5\x34\x70\xfd\x84\xca\xf0\x03\x88\xee\x48\xa3\xea\xca\xe0\x1b\xa5\x7a\xc1\x47\xe8\x0f\x7c\xe3\xeb\x7b\x03\xf7\xe5\xfb\x02\x98\xea\x7a\xa3\x81\xcb\x3e\x36\x42\x2b\xda\x67\x64\x04\x04\x67\xdc\xb6\xbe\x73\x07\xdd\xd0\x75\x06\xf5\x30\xc8\x76\xc8\xfb\x03\x27\x1b\x84\x31\x7c\x27\xf0\xbe\xee\x4c\xf0\x78\xcf\xcd\x9e\xd8\xc5\xb5\x18\xed\x0f\x99\x54\x8b\x47\x3c\x2c\x99\x18\x4b\xa9\x37\x31\x76\xbc\x97\x4d\xec\xd4\x30\xf0\x71\x8b\xb5\xe3\x9e\x7f\x45\xdb\x65\x52\xad\xc2\xf2\xc9\x65\x74\x64\xbf\xd2\xee\xd8\x3f\xb9\xc6\xcd\xd3\x59\xcb\xfc\xee\x37\x67\xeb\x25\x5c\xb8\x32\x78\x6c\xb8\x5f\xa6\xf8\x72\xc3\x9f\xc4\x09\x0f\xa5\xb9\x15\x8f\x7c\xea\x2d\xf2\x1c\x87\xa0\x57\xfa\x8c\x11\xf3\x35\x86\xdc\x51\x72\xdb\xe6\x92\x02\x07\xe0\xb2\x7c\x66\x5f\x66\xcf\x91\xf7\x13\x25\xba\x53\x94\x8f\xd2\x81\x05\x07\x72\x2e\x5e\x62\xf0\x28\x74\xf2\xe9\xf9\xce\x70\xee\x0c\x16\x70\xa5\xd9\x40\xe1\xa2\xf1\x85\x6f\x51\xc4\xcb\x9c\x2f\x5c\x38\x89\xee\x30\xdd\x8b\x2f\x93\x30\x36\x93\xfe\x2c\x07\x49\x64\x83\x18\x94\x3e\x15\x26\x87\x1f\x16\xf3\xad\x8b\xe3\xd1\x5f\x05\xbe\xd7\xbc\x13\x77\x42\x4b\xb5\xbd\xe4\x54\xb1\x81\xf6\x30\xe2\x15\xe6\x72\xb0\x50\xee\xb9\xb6\xf2\x91\x7b\xc7\x08\xd7\x14\xc6\xf6\x82\x6f\x21\x3d\xb4\x31\x4f\x78\x2b\xb9\xec\x4b\x72\x63\xb9\x15\x34\xd0\xb3\x9b\x53\xed\xde\xe4\x57\x7e\x58\xe6\x97\x2b\x8e\xda\x33\x11\xaf\x2f\x9e\x09\x18\x5c\x38\x4c\xfc\x42\x10\x6c\x3e\x28\x23\xad\xd2\x27\x9a\x35\xa0\xfb\x49\x92\x9d\xb4\x1f\x75\x9f\x49\xee\xb4\xb2\xaa\x53\x44\xac\x0d\xbf\xd3\xf2\xc8\xad\xf8\x89\x46\x8c\xfb\x61\x7a\xe8\x65\x97\xc9\xd3\x0b\x81\xd9\xab\x3f\x6f\xc5\xa1\x57\x27\x07\x13\xfc\x37\x7f\x36\xe8\x26\xed\x72\xec\x7d\x76\x1b\x43\x57\xfd\xec\xa2\x1f\x9f\x16\xb2\x87\x05\xb8\xa0\x67\xa1\x86\x4f\xae\x78\x67\xde\x49\xfb\x46\xf3\xb1\x23\x09\xbe\x53\xa3\x95\xe3\xa4\x26\x13\xcc\x27\x29\x4a\x90\x8a\xba\x2c\x9b\xb1\x42\x22\xcc\xd5\x6c\x4a\xfb\xb8\x90\x50\x93\xec\x42\x10\xab\xe1\xa0\x46\x41\xf3\x7d\x9e\x80\xb2\x1c\x77\xd6\x1d\xd5\xf8\x28\x77\x73\xf0\x9c\x71\xcf\xb9\x0f\xc2\x7e\x5a\x33\xbc\x14\xba\xa5\xfe\xb4\x12\xc4\x0b\xbb\x26\x63\xd5\xf0\x36\x93\x2e\x42\xf7\x5b\x45\xe4\xe2\x45\x85\x64\xa9\x1a\xe6\xfc\x6d\x21\xc3\x9c\x33\x86\xdd\x7a\x23\x77\x23\xa4\xa6\x3c\x7c\x2b\xbf\xe5\xd0\x73\x27\xa8\x51\x51\x20\xfb\x51\x8e\xd2\xec\x71\xfd\x39\xe3\x06\x4c\xba\xd6\xb1\xf1\x7f\xc9\x3b\x4c\x22\x23\xec\x55\x4f\x2d\xc5\x1c\x44\xc3\xb1\x1c\x50\xf4\x20\xe7\x47\xb5\x73\xb1\x4b\x6c\x0f\xaf\x9d\x25\x04\xe8\x1d\xd4\xe3\x58\xe6\xd3\x5a\x8c\x13\x7f\xc1\x9a\xf1\x03\xd3\x92\x1b\x54\xa5\x66\x20\x28\x70\xa8\xb0\x08\xdb\x1f\x9d\xd2\x9d\xb8\x17\x0f\x93\xec\x17\x46\xc5\x66\x01\x1b\x85\xdf\x7e\x96\x46\x15\xf7\xbe\xae\x3b\xcd\xd7\x56\x5a\x53\xb0\xe3\x4e\xe9\xe0\x3c\xab\xcf\x05\x17\xae\xa0\xf6\xe0\xde\xaa\xd1\x72\x39\xba\x5b\x98\xd2\xb6\x84\x31\xe9\xf6\x38\xdd\xac\x8c\xe6\xcc\x31\x17\x0d\x56\x49\x69\x7a\x34\x5a\x92\xf7\xff\xe9\xb7\x16\x7d\x31\x06\x92\x3f\xfc\x5c\x06\xf2\x92\x16\xbe\xda\x9a\x67\xa4\xa6\x57\x99\xa5\x19\x97\x3a\xf5\x45\xc2\xab\x74\xee\x24\x4f\x54\x03\xa9\xd6\x71\x17\x1f\xfa\x29\xa1\xe4\xdd\xe0\x6a\x3a\x8b\x15\xbd\x6a\x66\x5e\x17\x29\x09\x65\x8b\x0b\xe9\xee\xac\xcd\x37\xc5\x4c\x70\x83\x8b\x59\x26\x7f\x5e\xb5\x3e\x17\xce\xf4\x55\x84\x98\x5c\x7a\x30\xf1\x16\xbb\x9b\x40\x96\x3b\xc9\x83\x02\x5d\x08\x69\xe6\xda\xfd\xe6\xed\x16\x0e\xb8\xd3\x7c\xac\xfa\x4b\x25\xfa\x96\x3f\xcf\x1b\x54\x0f\xfd\x99\x1b\x05\x6a\xaa\x1b\xcd\xcc\x95\xd2\x3c\xa5\xae\x62\x66\xa0\xef\xdf\x00\x00\x00\xff\xff\xa7\x1a\x91\xa5\x55\x1d\x00\x00")

func pluginsCodeampSchemaGraphqlBytes() ([]byte, error) {
	return bindataRead(
		_pluginsCodeampSchemaGraphql,
		"plugins/codeamp/schema.graphql",
	)
}

func pluginsCodeampSchemaGraphql() (*asset, error) {
	bytes, err := pluginsCodeampSchemaGraphqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "plugins/codeamp/schema.graphql", size: 7509, mode: os.FileMode(420), modTime: time.Unix(1525737872, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _pluginsCodeampStaticIndexHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x56\xeb\x6f\xdb\x36\x10\xff\xde\xbf\xe2\xe6\x6c\x90\x5d\xd8\x94\xd3\xf5\x01\xa8\x76\x86\xb6\x49\xbb\x16\x69\xd3\xc6\x19\x8a\x7d\x2b\x4d\x9e\x2c\xa6\x14\xa9\x1e\x29\x3b\x6a\x91\xff\x7d\xa0\x64\xc9\x8a\x91\x0c\xd9\xb0\x41\x1f\x4c\xde\xe3\x77\x0f\xde\xc3\xb3\x9f\x8e\xcf\x5e\x5d\xfc\xf9\xf1\x04\x32\x9f\xeb\xa3\x07\xb3\xe6\x07\x60\x96\x21\x97\xe1\x00\x30\x73\xbe\xd2\xd8\x9c\x01\x96\x56\x56\xf0\x63\x7b\x01\xc8\x50\xad\x32\x9f\xc0\xe1\x74\xfa\xcb\xf3\x8e\x9a\x73\x5a\x29\x93\xc0\x74\x47\xda\x28\xe9\xb3\x7d\x39\xbb\x46\x4a\xb5\xdd\x24\x90\x29\x29\xd1\xb4\x9c\xeb\xed\xef\xc1\x8a\x78\x91\xa9\x6f\xfa\x76\x8b\xeb\x6c\x5f\x81\x5d\x6e\xfc\xc4\xdb\xaf\x68\x7a\x1a\x4b\x2e\xbe\xae\xc8\x96\x46\x26\xa0\x95\x41\x4e\x93\x15\x71\xa9\xd0\xf8\xe1\x41\xfa\x2c\x7c\x63\x38\xc0\x47\xe1\x1b\xed\x9c\x5b\x5a\x92\x48\x93\xa5\xf5\xde\xe6\x09\x1c\x16\x57\xe0\xac\x56\x12\x0e\xe4\x34\x7c\x3b\xc9\xd4\x1a\x3f\x49\x79\xae\x74\x95\x80\xab\x9c\xc7\x7c\x0c\x13\x5e\x14\x1a\x27\xed\x35\x5a\x70\x03\xaf\x89\x1b\xa1\x9c\xb0\xd1\x18\x22\xb6\x78\xfd\x61\x71\xac\x5c\xa1\x79\x35\x39\xc7\x55\xa9\x39\x05\xfa\x02\x57\x16\xe1\x8f\xb7\xd1\x18\xea\x63\x47\xfa\xfc\x31\xb0\x7f\x47\xbd\x46\xaf\x04\x87\x0f\x58\x62\x34\x86\xac\x25\x8c\x21\x3a\x2d\x85\x92\x1c\xde\x10\x37\x32\xf0\x38\x29\xae\xc7\xe0\xb8\x71\x13\x87\xa4\xd2\x9d\xd3\x05\x97\x52\x99\x55\x02\xcf\x8a\x2b\x38\x7c\x5c\x5c\xc1\xd3\xe2\x6a\x2f\x26\xa7\xbe\x63\x52\x33\x6f\x26\x7a\x16\xf7\x6a\x62\xa6\x95\xf9\x0a\x84\x7a\x3e\xa8\xa9\x2e\x43\xf4\x03\xc8\x08\xd3\xf9\x20\xf3\xbe\x70\x49\x1c\x0b\x69\x2e\x1d\x13\xda\x96\x32\xd5\x9c\x90\x09\x9b\xc7\xfc\x92\x5f\xc5\x5a\x2d\x5d\xdc\xbe\x73\x3c\x65\x87\x53\xf6\xa8\xbb\x33\xe1\xdc\x00\xe2\xdb\x0a\x31\x7e\x08\x67\x6b\x24\x52\x12\x1d\x3c\x8c\xdb\x02\x68\x35\x27\xc2\x1a\xcf\x95\x41\x02\xb6\x0e\x69\x58\x6a\x9c\xa0\x54\xde\xd2\x2d\xc5\xf4\xf4\xe9\xdf\x87\xe8\x04\xa9\xc2\x83\x23\x71\xef\x90\x52\xf4\x22\x8b\x1f\xb1\x29\xfb\xb5\x39\xb3\x5c\x19\x76\xe9\x06\x47\xb3\xb8\x81\xfb\xf7\xd8\x84\x5c\xf8\xf8\xf0\x09\x7b\xc2\x1e\x37\x97\xff\x15\x7c\x22\x6d\xfe\x1f\x1a\xb8\xf3\xb1\xf7\xe1\x67\x71\x3b\x85\x66\x61\xec\x6c\x2d\x4a\xb5\x06\xa1\xb9\x73\xf3\x41\xd7\xed\x83\xa3\x77\x9f\x2f\xe0\xa2\x6e\xfc\x99\x32\x45\xe9\x41\xc9\x3e\x1f\x0a\xcd\x05\x66\x56\x4b\xa4\xf9\x60\x27\xbc\xb2\xe8\x20\x43\xc2\x60\x59\xaa\x75\xcf\x46\x00\x68\x5d\x1b\x1c\x9d\x5a\x1e\xda\x85\x31\xd6\x97\xeb\xa7\x62\xcd\x09\x1c\x72\x12\x19\xcc\x61\xa3\x8c\xb4\x1b\xa6\xad\xe0\x5e\x59\xc3\x1a\xc6\xf3\x4e\xb0\xe0\xc4\x73\xf4\x48\x0e\xe6\xf0\xe3\xba\x61\x48\x2b\xca\x1c\x8d\x67\x2b\xf4\x27\x1a\xc3\xf1\x65\xf5\x56\x0e\xa3\x2e\x8e\x68\xc4\xd6\x5c\x97\x08\x73\x08\xd0\x7a\xe1\x2d\xf1\x15\x06\x85\xb7\x1e\xf3\x61\xd4\x3a\x9c\x5c\x6e\xfc\x45\xa3\xf1\xfc\x41\x0d\x9e\x96\x46\x04\x57\xa0\x16\xf9\x74\xfa\x3a\xd4\x24\xd2\x70\x7b\xfd\x18\x1c\x72\xa3\xae\x37\x84\x35\xce\x43\x8b\x02\xf3\x7f\xe0\x5c\xdb\x48\x1a\x3d\x84\x07\xdc\x06\xd9\x35\x5d\xf4\x42\x08\x2c\x7c\x94\x40\x14\x66\xa4\x6a\x52\x14\x5f\x3a\x6b\xa2\xf1\x4e\xea\x95\x35\x1e\x8d\x9f\x5c\x54\x05\xde\x2a\xdb\xf6\x6b\x6b\x4f\xa5\x30\x6c\x1d\x1e\xf5\xec\xdd\xc8\x94\xbb\x3b\x53\xe3\x2e\xdc\xde\x16\xb8\x2d\x80\xfb\x86\x70\x9f\x20\x6e\x48\xbf\x28\x7d\x66\x49\x7d\xaf\xf9\x51\xb2\x4b\xff\x6f\xf0\xe5\x25\x72\x42\x82\x9f\x7f\xb4\xc4\xeb\x2f\x90\x80\x29\xb5\xee\x10\xae\xf7\xd7\x21\xa1\x2f\xc9\x40\x3d\x7e\x86\x51\xfc\xad\x44\xaa\xa2\x71\x2f\x92\x1c\x7d\x66\x65\x02\x51\x61\x9d\xef\xf9\xb2\x8d\x7a\xdc\x5b\x86\xb2\x4a\xe0\xdd\xe2\xec\x03\x73\x9e\x94\x59\xa9\xb4\xda\x2b\x9d\x9d\xb0\x20\x94\x68\xbc\xe2\xda\x25\x10\x29\x23\x74\x19\xb6\x51\xeb\xdc\x88\xf9\x0c\xcd\xb0\x2b\xc8\x21\xa1\x2b\xac\x71\xd8\x7f\xb4\xad\xeb\x2d\x8b\x79\xbc\xf2\xc3\xee\x5d\xee\xc6\x78\x69\x65\xd5\xc7\xf1\x54\xdd\x78\xb9\x2d\x6e\x1d\x49\xc1\xc9\xe1\x4d\xcd\xdd\xc3\x5f\x83\xe0\x5e\x64\x30\x44\x22\x4b\xa3\xdb\x40\xfa\x9a\x3d\xc5\xce\xc7\x86\xd6\xdc\xe3\x18\xce\xd1\x48\x24\x98\xbd\xa9\x2b\xef\xd3\x29\xc4\x47\xa0\x8c\xb7\xe0\x33\xac\x13\xcc\x5a\xc9\x05\x62\x4d\x3c\x3f\x79\x71\xfc\xfe\x04\x94\xa9\x6f\xde\x16\xa0\x71\x8d\x1a\x6c\x0a\x3e\x53\x0e\x72\x2b\x4b\x1d\x18\xa0\x91\x93\x81\xdc\x12\x02\x5f\xda\xd2\xb7\x48\x99\xdd\x40\x65\x4b\x10\xdc\x80\x28\x9d\xb7\xb9\xfa\x8e\xd0\x79\xb0\xac\xa0\x20\xbb\x56\x61\xb4\x81\x54\x69\x8a\x84\xc6\x43\xdd\xc6\x0e\x2c\xb5\x30\xe1\xbf\x42\xc8\x33\xd7\x20\x32\xa5\x25\x60\x33\x01\x5c\xe3\xf2\x79\xd8\x12\xc7\x67\xef\x19\xd5\x21\x0e\xb7\x19\xa8\xc9\x4c\x10\x72\x8f\xdb\x91\x31\x6c\x4d\xf7\xab\x30\x6d\x46\x51\xb2\x37\x9a\x76\x05\xd3\x9e\xee\x9c\x41\x6d\x33\x47\xa3\x5a\x72\x9b\xf9\x9b\xbb\xa4\x59\x21\xb3\xb8\xf9\x8b\xfb\x57\x00\x00\x00\xff\xff\x45\xab\x31\x54\xfa\x0a\x00\x00")

func pluginsCodeampStaticIndexHtmlBytes() ([]byte, error) {
	return bindataRead(
		_pluginsCodeampStaticIndexHtml,
		"plugins/codeamp/static/index.html",
	)
}

func pluginsCodeampStaticIndexHtml() (*asset, error) {
	bytes, err := pluginsCodeampStaticIndexHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "plugins/codeamp/static/index.html", size: 2810, mode: os.FileMode(420), modTime: time.Unix(1510690037, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"plugins/codeamp/schema.graphql": pluginsCodeampSchemaGraphql,
	"plugins/codeamp/static/index.html": pluginsCodeampStaticIndexHtml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"plugins": &bintree{nil, map[string]*bintree{
		"codeamp": &bintree{nil, map[string]*bintree{
			"schema.graphql": &bintree{pluginsCodeampSchemaGraphql, map[string]*bintree{}},
			"static": &bintree{nil, map[string]*bintree{
				"index.html": &bintree{pluginsCodeampStaticIndexHtml, map[string]*bintree{}},
			}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

