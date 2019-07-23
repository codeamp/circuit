// Code generated by go-bindata.
// sources:
// plugins/codeamp/graphql/schema.graphql
// plugins/codeamp/graphql/static/index.html
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

var _pluginsCodeampGraphqlSchemaGraphql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x59\xcd\x72\x1b\xb9\x11\xbe\xf3\x29\xa0\xda\x0b\x5d\xa5\x27\xe0\x2d\xb6\xbc\x96\x12\x3b\x51\xc4\xf5\x21\xe5\xd2\x01\x1a\xb6\x48\x44\x33\xc0\x2c\x80\x91\xad\x4a\xe5\xdd\x53\xf8\x9d\x6e\x00\x43\x89\x5c\xa7\x6a\x2f\x12\xa7\x67\xd0\xe8\x6e\x74\x7f\xfd\x03\xd3\xf1\x9e\x6b\xf6\x9b\x18\x60\x15\x7f\xff\x75\xfb\x8f\xbf\xaf\x56\xa6\x3b\xc0\xc0\xd9\x7f\x56\x8c\xfd\x3e\x81\x7e\xd9\xb0\x7f\xba\x7f\x2b\xc6\x86\xc9\x72\x2b\x94\xdc\xb0\x2f\xf1\xd7\xea\xbf\xab\xd5\x2f\xf1\xbd\x7d\x19\x21\xfc\xf4\x6b\x7f\x61\x5f\x0d\xe8\x15\x63\x93\x01\xbd\x16\xbb\x0d\xbb\xb9\x7a\xb7\x49\xc4\xf0\xd6\xc4\xd7\x66\xfd\x6e\xc3\xbe\x39\xca\xfd\x85\x7f\x79\xab\xd5\xbf\xa1\xb3\x2b\xc6\xc6\xf0\x2b\x32\xb8\x64\xa6\x9f\xf6\x1b\xb6\xb5\x5a\xc8\xfd\x25\x93\x7c\x80\xf9\x09\xe4\xb3\xd0\x4a\x0e\x20\xed\xcd\x55\x22\xbf\xdb\x20\x6e\x99\xb3\x99\x59\x9b\x75\xfc\xb1\x05\xae\xbb\x43\xfe\x3c\x3c\xde\xc8\x71\xb2\x97\x6c\xe4\x9a\x0f\x66\xc3\x6e\xf9\x5e\x48\x6e\x95\xf6\xf4\x99\xf7\x67\x61\x6c\x10\xfd\x57\xe0\x76\xd2\xe0\x36\x78\x8c\x3f\xd7\x8b\xab\xe3\xc7\xf3\xea\x2d\xe8\x67\xd1\xf9\xd5\x26\xfe\x5c\x5e\x1d\x3f\x9e\x57\x7f\xfc\x31\x2a\x6d\x19\xef\xfb\xbc\x9a\x7d\x17\xf6\x20\x24\xdb\x8b\x67\x90\x49\xe5\x9b\x2b\xc6\xe5\x8e\xda\x6b\xc5\x18\xf8\xe5\xdb\x72\xdf\x8f\x84\x9c\x37\xf7\xd6\xbd\x60\x0c\xcb\xcd\xcc\x08\x1d\x12\x7e\xeb\x1e\xfd\xe1\x6e\x67\x42\x3c\xe3\x3b\xe8\x81\x1b\xaf\xaa\x8e\x3f\x97\x55\x8d\x1f\x23\x55\x67\xd9\x1d\x07\xa4\xca\x7c\x9e\xc8\x55\x9c\x08\x68\xc9\x7d\xc5\x84\x3d\x73\x2d\xf8\x43\x1f\x4d\xdf\x69\xb0\x47\x2d\xef\x3e\x68\x1a\x1e\x5a\x3c\xcf\x39\x05\x2a\xc2\x47\x4c\x5d\x38\x03\x21\xf7\x3d\x44\xd9\xb2\x16\x3e\x72\xb2\x11\xf2\xcb\x1c\x09\x1f\x7f\x58\x90\x46\x28\xe9\xcf\xce\xef\x1f\x09\x66\xbd\x14\x50\xdf\xf2\x22\x1a\xaf\x99\x8c\xc2\x6b\xa6\x79\x3f\x28\xbf\xa4\xce\x80\x36\x9f\xdd\xa2\xe0\x70\x57\x50\x93\x08\xa0\x07\x61\xf2\xe6\xf3\x93\x5b\xe4\x90\xed\x22\x80\x55\x86\x2e\x8f\x57\xe9\x29\x42\xd6\x07\x0d\xdc\x42\x12\x7d\xc5\x58\xe7\x09\x51\xe8\xe4\x58\x39\xea\x0b\x10\x08\xb8\x36\xee\x28\x8b\xc9\x13\x4e\x61\x11\xa5\x88\xea\x67\x29\xa2\xe2\xeb\x48\xcf\x41\x51\xc4\x48\xf0\x05\xab\x46\xc4\xc0\x58\x35\xa6\xe5\x01\x49\x2f\x8a\x05\x71\xcf\x18\xb7\x79\xcf\x18\xb6\xeb\x48\xcf\x98\x53\x40\x10\xd6\x7c\x26\x05\xcd\x4f\x61\x71\x05\x3d\x10\x29\x76\x9e\x70\x0a\x8b\x9b\xc1\x87\xe2\xc0\xe5\xcb\x0c\x82\xdc\x32\x25\xfd\x07\x62\x20\x18\x97\xbe\xd8\xc4\x75\x25\xca\x25\xdc\xba\x88\x5e\x46\xcd\x94\x42\x86\xd8\xca\x41\xdc\x1a\xe1\x5f\x16\xcf\x3d\x14\x12\x6f\xc3\xfa\x6c\xbb\x82\x2f\x31\xe0\x59\x7c\xa9\x41\x13\x5f\x62\xd5\xb3\xf8\x46\x3b\x20\x80\xc8\x66\x40\xa0\x8a\x01\x64\x83\xd1\x36\xb1\xfd\x48\xd6\x67\x33\x50\xb6\xc1\x0a\x7f\x84\x6d\xb4\x02\x65\x1b\x8c\xf0\x47\xd8\xd6\x46\xc8\x88\x8f\x9c\xc2\x83\x70\xc0\xe2\x84\xbf\x34\x8d\x2c\x68\x8e\x79\x25\x47\x78\x1b\xaf\x5a\xdd\xcc\x8b\xa1\xd3\x7f\x1b\x33\x1c\x4f\xed\xdc\x56\x07\x57\x48\x5d\x31\x8b\xce\xa1\x45\x72\xd7\xb7\xf0\x5c\x04\x16\x94\xd9\x28\xfa\x54\x22\xaf\xf3\x07\x2e\x25\xc6\x9f\xf9\x7c\x12\x81\x58\xb4\xe4\x18\xdd\xe9\x0c\x8e\xc9\xae\x25\xc7\xe8\x49\x67\x70\x2c\xb5\x2e\x33\xce\xcc\xb3\xcc\xa6\x9b\x2a\xe7\x16\x99\xe4\xb8\x31\xca\xbc\xf4\xd3\x36\x42\x36\x8a\xb4\x60\x9d\xff\x93\x42\xae\x7b\xc0\xc9\x3e\xeb\xe5\x9a\x09\x54\x12\xac\x27\xfa\x1c\x1a\x11\x44\x98\xdd\x32\xd4\x54\xd1\x2d\x69\x36\x27\x65\x66\x65\xc2\x46\x09\x8a\x69\xb3\x22\x88\x98\xb7\x45\xc4\xb4\xf7\x7b\xa5\x9e\x06\xae\x9f\x50\x2d\xf1\x10\x49\xb7\xa4\x2b\x72\xb9\xfc\xbd\x52\x3d\x70\x19\x56\x7e\x02\xcb\x3e\x09\xcb\x3e\xa8\x61\x10\x5e\xd2\x3d\xd8\x4f\xc2\xc6\xe7\x75\x2e\x40\xfd\xea\xaa\x71\xf2\x34\x09\xdf\x33\x57\xcc\xdf\x17\x51\xb9\x20\x5e\x09\x69\x41\x3f\xf2\x0e\x66\x9a\xaf\xa5\x3a\x35\x39\x14\xbd\x91\x36\x2e\x41\x15\x7c\x28\xbd\x10\xc1\x01\x47\x0f\xde\x20\x47\xd8\xb8\x2a\xdf\x6a\xe1\x92\x75\x2a\x02\xef\x23\xf3\xb9\x1e\x0f\xbc\xe7\xe7\xd3\x59\x87\xb5\x33\xe7\xdc\x63\x25\xd6\x99\x70\x0e\x6f\xbf\x38\x31\x47\xed\x5f\x60\x8e\x08\xa7\x33\x8f\x8b\x13\x73\xd4\x99\x06\xe6\x88\x70\x3a\xf3\xb8\x38\x31\xf7\x6d\xbc\xe7\xea\x7e\xf9\x95\xd1\x17\xdd\xaa\x81\x8b\x3e\x77\x27\xb4\x1a\x2f\x22\x2c\xa0\xdd\x6e\xe3\xe7\x10\xd4\x2a\xc4\x22\xc5\x0e\x65\x15\xed\x68\x03\x18\xc3\xf7\x80\xf7\x75\x41\x8f\x9f\x0f\xdc\x1c\x88\x5c\x5c\x83\xb4\xd7\x05\x55\xc3\x23\x7e\x2c\x44\x64\x4c\x2a\xfb\xab\x9a\xe4\x6e\x2b\xa4\xab\x45\x91\xe4\xa9\x0e\xc5\x8e\xf2\x06\xc9\x3b\x35\x0c\x5c\xee\xf0\xa6\x78\xb0\x71\x41\xfb\x69\x52\x8c\x5d\x54\x47\xe6\x52\xad\xb3\xb2\x6b\x79\xee\x2f\x68\x6b\x4c\x4a\x18\xf7\xce\x09\x7a\x54\xd7\x1d\x8c\xbd\x7a\x71\x9f\x6f\xad\xe6\x16\xf6\x2f\xa1\x99\x5a\x31\xd6\xbb\x5e\x16\x8c\xb9\xd5\xea\x01\x32\x55\x03\xdf\x89\x9a\x3c\x6a\x70\x2d\xc9\xb5\x52\x4f\x69\xbf\x60\x32\x5c\x53\x79\xb3\xe1\x9e\x9c\x9a\xae\xb4\xc9\x13\xbc\xe0\x47\x61\xae\xe0\x91\x4f\xbd\x25\x60\xd8\xa9\x5e\xe9\xa3\x2a\xa6\x31\x50\xed\xe4\xad\xf1\x00\xc6\x97\x42\xbe\x42\x9e\x67\xde\x4f\xf4\x0c\x3b\x45\xad\xdd\xf2\x85\xe0\xb2\x2e\xa8\x5a\xe7\xf3\x0c\x3a\x47\x51\x82\xa9\xe3\x27\x5c\xa9\x2b\xcc\x36\x96\x7c\x14\xd2\xf1\x04\x87\x38\xb0\x73\xb3\x57\x4e\xa2\x1b\xa7\x3b\xf8\x7d\x02\x63\x0b\xea\x67\x31\x08\x42\x1b\x60\x50\xfa\xa5\xf1\x71\x78\x51\x7d\x6f\x1d\x72\x48\xdf\xa2\x7f\xd2\xbc\x83\x5b\xd0\x42\xed\x1a\x91\x91\xa3\x62\x41\xe9\xda\x37\x70\x56\x22\x19\xe9\x0d\x11\x4b\x4f\x89\x6b\x2b\x1e\xb9\x77\xa1\x30\x68\x60\xec\x00\x7c\x17\xa1\x2b\x4f\xf9\xbc\x3e\x5c\xf4\x2d\xba\xb1\xdc\x02\x05\xa1\x62\xf6\xb1\x34\xf9\xf0\x2b\xbf\xd4\xd8\x77\x92\x53\x18\xcb\x35\x21\x3c\x0a\x29\xcc\x81\x9a\xf0\x4e\xf5\xfd\x03\xef\x9e\x48\x7c\x69\x08\x10\xe1\x62\x63\xf1\x45\x25\x1e\x4e\x50\x24\x39\xbd\xe2\x68\x78\xfa\x1b\x36\x19\x95\x11\x56\xe9\x17\xea\x12\xb1\xa7\xc9\x94\xbd\xb0\x5f\x75\x5f\x50\x6e\xb5\xb2\xaa\x53\x84\xac\x0d\xbf\xd5\xe2\x99\x5b\xf8\x1b\x8d\x66\xf7\x62\x7a\xe8\x45\x57\xd0\xf3\x8c\xd7\x1c\xd4\xf7\x2b\xaf\xb1\xb3\x5a\x34\xc4\x91\xc1\x71\x31\xfa\xed\x26\xed\x72\xd1\x5d\x31\xdb\x39\x67\x38\xfa\xca\xe0\xf8\x92\x19\x3f\xd9\x46\x8a\xd4\xb3\xe4\xe3\x13\xd0\x25\x16\x78\x28\x0a\xd8\x73\x9b\x53\xbf\xbd\xb0\xef\x35\x97\x1d\x49\xbe\x9d\x92\x56\xc8\x49\x4d\x26\x18\x93\xf8\x14\x90\x72\xba\xae\x99\x53\x79\x8c\x4e\x60\xa9\xc8\x28\xc6\x9e\x21\xf5\x64\xda\x2b\x70\xa7\x86\x51\x49\x1f\x58\x08\xa9\xca\x54\xca\xbb\x03\x54\x51\x51\xe4\x88\xa3\x41\xaa\xe4\xa3\xd8\xcf\x90\xd2\xd2\xa2\x6a\x8d\x70\x2c\x2d\xa9\xd3\x02\xb4\x56\xcb\xba\x00\x6d\x95\x5c\x93\xb1\x6a\xf8\x50\x50\x2b\x40\xfb\x09\x38\x85\x11\x1b\xb5\xb2\x18\xbb\x97\x74\x2e\x67\xa6\x85\xce\xa5\xc5\x70\x18\x6d\xc5\x5e\x46\xc0\x2e\x21\x66\xe1\x5d\xa9\x7a\xe9\x1a\x4b\xa6\x68\x18\x9b\x02\xb3\x1f\xf4\xd7\xd0\x7c\xc4\x33\x98\x70\xf1\xba\xf2\x7f\xc9\xc8\x39\xdb\x27\x16\x63\x8b\x88\xd8\x84\x4e\x1a\xb7\xed\xc8\xa3\x67\x3b\xdf\x1f\x1c\x0b\x72\x22\x7b\xb8\x80\x6b\x69\x80\xae\xe6\xbc\x1e\x75\x1a\x58\x02\x03\xdc\xbd\x3a\x2f\x21\xcc\x09\xc6\x79\xc6\xa3\x3f\xa1\x1b\x3f\xef\xeb\x43\x6d\xe2\x1e\x88\x23\x62\x16\x78\x22\x5f\x5b\x18\x15\x05\xb3\x39\x50\x44\x52\x62\xd3\x78\x3e\xe5\x28\xdd\xc1\x1d\x3c\x4c\xa2\xaf\x54\x4b\x55\x1c\x16\x0a\x0f\xcb\x6b\xa1\x9a\x7b\xbf\xa1\x23\x69\x77\x1d\x71\xaf\x5b\xa5\x83\x9b\x5d\xdc\x37\xfc\x7f\x51\xb3\x56\xb7\x71\x55\xd1\x02\xe3\xba\x03\xb9\x06\xde\xdb\x83\x7f\xf0\x9f\x34\xba\x91\xc6\x27\x8b\x9d\x49\xba\x0a\x88\x53\x4d\x62\xd2\xc6\x65\x81\xb7\x6c\xc2\x8c\x7f\xfd\xe5\xcb\xe7\xc0\xeb\xdc\x63\xa6\x22\x84\x3b\x41\x22\x42\xe3\x56\x36\xb8\xec\xa9\x9b\x7c\x50\xd2\x72\x21\x41\xb3\x6a\x8f\xf2\x34\xc3\x06\x4a\xa3\x73\xcf\x88\x91\x26\x3d\x61\xe5\xc2\xa9\x79\x06\xa5\x3b\x0c\xfc\xc7\x76\xd2\x31\xd0\x22\xe1\xab\xe4\xcf\x5c\xf4\x21\x7f\x52\xd6\xe5\x19\x56\x3c\x7d\x37\x61\x0f\xb4\x4d\x28\x3c\x1a\xeb\xe1\x5b\xb3\x03\x0c\x98\xc1\xc8\x2d\x86\x37\x21\x85\x15\xbc\xbf\x82\x9e\xbf\x6c\xa1\x53\x72\x67\xd2\xd2\xd1\x77\x24\x05\xd1\x8a\x01\xd4\x64\x0b\xaa\x99\xba\x0e\x8c\xf9\xed\xa0\xc1\x1c\x94\x0b\xde\x40\x7f\xe4\xa2\x9f\x34\x54\xf4\x83\xb5\xe3\x35\xf0\x1d\x68\x17\x5a\x48\xef\xeb\xfc\x22\x05\x59\xcb\x3a\xc5\x57\xde\x4e\x65\x14\x17\x4d\x6a\xd5\x09\xb6\xdc\x21\xdf\x0b\xd5\x60\xf2\x67\x6a\x0c\x17\x7b\x3e\xdc\xd6\x93\x88\x2a\x6e\x7c\x5e\x57\xef\x9c\x09\xc4\xe2\x64\xa1\x30\x75\xbe\x8e\xa9\xc5\x78\x6d\xd0\x50\xd5\x1b\x0b\x83\x07\x92\x93\x17\x41\x79\x69\x60\xd0\xd6\x61\x19\x2b\xe7\xdb\x9f\x08\x95\x9e\xf0\x93\x90\xb2\x29\xcc\x11\xd4\x2c\x84\x39\x7d\x3f\xd2\x3e\x14\x5b\xe0\xfb\x8b\x93\x9d\xa8\xd9\x5c\x2c\x1e\x4e\xbb\xcb\x28\x4b\x74\xea\x10\xcd\xce\xa1\x55\x66\x1d\xd5\xe4\xb2\x69\xb6\x4b\x5c\x57\x17\xf4\xb7\x35\x0e\xc7\xac\x4e\xef\x6c\x88\xc8\xad\xeb\x1c\x2f\xf1\x64\x40\x17\xd5\x16\x99\x46\xd3\x85\x11\x4f\x4f\xdd\x6f\xde\xae\x0a\xc6\xbd\xe6\xb2\x8a\x9d\xfa\x26\xa8\x69\xff\x16\x20\x2d\xba\xc2\x1b\x37\x0a\xa6\x59\xdc\x68\xb6\x5c\x2b\x26\xa8\xe9\x16\xc4\x0c\xe6\xfb\x5f\x00\x00\x00\xff\xff\x0d\x36\x3c\xd9\x60\x28\x00\x00")

func pluginsCodeampGraphqlSchemaGraphqlBytes() ([]byte, error) {
	return bindataRead(
		_pluginsCodeampGraphqlSchemaGraphql,
		"plugins/codeamp/graphql/schema.graphql",
	)
}

func pluginsCodeampGraphqlSchemaGraphql() (*asset, error) {
	bytes, err := pluginsCodeampGraphqlSchemaGraphqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "plugins/codeamp/graphql/schema.graphql", size: 10336, mode: os.FileMode(420), modTime: time.Unix(1563837970, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _pluginsCodeampGraphqlStaticIndexHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x56\xeb\x6f\xdb\x36\x10\xff\xde\xbf\xe2\xe6\x6c\x90\x5d\xd8\x94\xd3\xf5\x01\xa8\x76\x86\xb6\x49\xbb\x16\x69\xd3\xc6\x19\x8a\x7d\x2b\x4d\x9e\x2c\xa6\x14\xa9\x1e\x29\x3b\x6a\x91\xff\x7d\xa0\x64\xc9\x8a\x91\x0c\xd9\xb0\x41\x1f\x4c\xde\xe3\x77\x0f\xde\xc3\xb3\x9f\x8e\xcf\x5e\x5d\xfc\xf9\xf1\x04\x32\x9f\xeb\xa3\x07\xb3\xe6\x07\x60\x96\x21\x97\xe1\x00\x30\x73\xbe\xd2\xd8\x9c\x01\x96\x56\x56\xf0\x63\x7b\x01\xc8\x50\xad\x32\x9f\xc0\xe1\x74\xfa\xcb\xf3\x8e\x9a\x73\x5a\x29\x93\xc0\x74\x47\xda\x28\xe9\xb3\x7d\x39\xbb\x46\x4a\xb5\xdd\x24\x90\x29\x29\xd1\xb4\x9c\xeb\xed\xef\xc1\x8a\x78\x91\xa9\x6f\xfa\x76\x8b\xeb\x6c\x5f\x81\x5d\x6e\xfc\xc4\xdb\xaf\x68\x7a\x1a\x4b\x2e\xbe\xae\xc8\x96\x46\x26\xa0\x95\x41\x4e\x93\x15\x71\xa9\xd0\xf8\xe1\x41\xfa\x2c\x7c\x63\x38\xc0\x47\xe1\x1b\xed\x9c\x5b\x5a\x92\x48\x93\xa5\xf5\xde\xe6\x09\x1c\x16\x57\xe0\xac\x56\x12\x0e\xe4\x34\x7c\x3b\xc9\xd4\x1a\x3f\x49\x79\xae\x74\x95\x80\xab\x9c\xc7\x7c\x0c\x13\x5e\x14\x1a\x27\xed\x35\x5a\x70\x03\xaf\x89\x1b\xa1\x9c\xb0\xd1\x18\x22\xb6\x78\xfd\x61\x71\xac\x5c\xa1\x79\x35\x39\xc7\x55\xa9\x39\x05\xfa\x02\x57\x16\xe1\x8f\xb7\xd1\x18\xea\x63\x47\xfa\xfc\x31\xb0\x7f\x47\xbd\x46\xaf\x04\x87\x0f\x58\x62\x34\x86\xac\x25\x8c\x21\x3a\x2d\x85\x92\x1c\xde\x10\x37\x32\xf0\x38\x29\xae\xc7\xe0\xb8\x71\x13\x87\xa4\xd2\x9d\xd3\x05\x97\x52\x99\x55\x02\xcf\x8a\x2b\x38\x7c\x5c\x5c\xc1\xd3\xe2\x6a\x2f\x26\xa7\xbe\x63\x52\x33\x6f\x26\x7a\x16\xf7\x6a\x62\xa6\x95\xf9\x0a\x84\x7a\x3e\xa8\xa9\x2e\x43\xf4\x03\xc8\x08\xd3\xf9\x20\xf3\xbe\x70\x49\x1c\x0b\x69\x2e\x1d\x13\xda\x96\x32\xd5\x9c\x90\x09\x9b\xc7\xfc\x92\x5f\xc5\x5a\x2d\x5d\xdc\xbe\x73\x3c\x65\x87\x53\xf6\xa8\xbb\x33\xe1\xdc\x00\xe2\xdb\x0a\x31\x7e\x08\x67\x6b\x24\x52\x12\x1d\x3c\x8c\xdb\x02\x68\x35\x27\xc2\x1a\xcf\x95\x41\x02\xb6\x0e\x69\x58\x6a\x9c\xa0\x54\xde\xd2\x2d\xc5\xf4\xf4\xe9\xdf\x87\xe8\x04\xa9\xc2\x83\x23\x71\xef\x90\x52\xf4\x22\x8b\x1f\xb1\x29\xfb\xb5\x39\xb3\x5c\x19\x76\xe9\x06\x47\xb3\xb8\x81\xfb\xf7\xd8\x84\x5c\xf8\xf8\xf0\x09\x7b\xc2\x1e\x37\x97\xff\x15\x7c\x22\x6d\xfe\x1f\x1a\xb8\xf3\xb1\xf7\xe1\x67\x71\x3b\x85\x66\x61\xec\x6c\x2d\x4a\xb5\x06\xa1\xb9\x73\xf3\x41\xd7\xed\x83\xa3\x77\x9f\x2f\xe0\xa2\x6e\xfc\x99\x32\x45\xe9\x41\xc9\x3e\x1f\x0a\xcd\x05\x66\x56\x4b\xa4\xf9\x60\x27\xbc\xb2\xe8\x20\x43\xc2\x60\x59\xaa\x75\xcf\x46\x00\x68\x5d\x1b\x1c\x9d\x5a\x1e\xda\x85\x31\xd6\x97\xeb\xa7\x62\xcd\x09\x1c\x72\x12\x19\xcc\x61\xa3\x8c\xb4\x1b\xa6\xad\xe0\x5e\x59\xc3\x1a\xc6\xf3\x4e\xb0\xe0\xc4\x73\xf4\x48\x0e\xe6\xf0\xe3\xba\x61\x48\x2b\xca\x1c\x8d\x67\x2b\xf4\x27\x1a\xc3\xf1\x65\xf5\x56\x0e\xa3\x2e\x8e\x68\xc4\xd6\x5c\x97\x08\x73\x08\xd0\x7a\xe1\x2d\xf1\x15\x06\x85\xb7\x1e\xf3\x61\xd4\x3a\x9c\x5c\x6e\xfc\x45\xa3\xf1\xfc\x41\x0d\x9e\x96\x46\x04\x57\xa0\x16\xf9\x74\xfa\x3a\xd4\x24\xd2\x70\x7b\xfd\x18\x1c\x72\xa3\xae\x37\x84\x35\xce\x43\x8b\x02\xf3\x7f\xe0\x5c\xdb\x48\x1a\x3d\x84\x07\xdc\x06\xd9\x35\x5d\xf4\x42\x08\x2c\x7c\x94\x40\x14\x66\xa4\x6a\x52\x14\x5f\x3a\x6b\xa2\xf1\x4e\xea\x95\x35\x1e\x8d\x9f\x5c\x54\x05\xde\x2a\xdb\xf6\x6b\x6b\x4f\xa5\x30\x6c\x1d\x1e\xf5\xec\xdd\xc8\x94\xbb\x3b\x53\xe3\x2e\xdc\xde\x16\xb8\x2d\x80\xfb\x86\x70\x9f\x20\x6e\x48\xbf\x28\x7d\x66\x49\x7d\xaf\xf9\x51\xb2\x4b\xff\x6f\xf0\xe5\x25\x72\x42\x82\x9f\x7f\xb4\xc4\xeb\x2f\x90\x80\x29\xb5\xee\x10\xae\xf7\xd7\x21\xa1\x2f\xc9\x40\x3d\x7e\x86\x51\xfc\xad\x44\xaa\xa2\x71\x2f\x92\x1c\x7d\x66\x65\x02\x51\x61\x9d\xef\xf9\xb2\x8d\x7a\xdc\x5b\x86\xb2\x4a\xe0\xdd\xe2\xec\x03\x73\x9e\x94\x59\xa9\xb4\xda\x2b\x9d\x9d\xb0\x20\x94\x68\xbc\xe2\xda\x25\x10\x29\x23\x74\x19\xb6\x51\xeb\xdc\x88\xf9\x0c\xcd\xb0\x2b\xc8\x21\xa1\x2b\xac\x71\xd8\x7f\xb4\xad\xeb\x2d\x8b\x79\xbc\xf2\xc3\xee\x5d\xee\xc6\x78\x69\x65\xd5\xc7\xf1\x54\xdd\x78\xb9\x2d\x6e\x1d\x49\xc1\xc9\xe1\x4d\xcd\xdd\xc3\x5f\x83\xe0\x5e\x64\x30\x44\x22\x4b\xa3\xdb\x40\xfa\x9a\x3d\xc5\xce\xc7\x86\xd6\xdc\xe3\x18\xce\xd1\x48\x24\x98\xbd\xa9\x2b\xef\xd3\x29\xc4\x47\xa0\x8c\xb7\xe0\x33\xac\x13\xcc\x5a\xc9\x05\x62\x4d\x3c\x3f\x79\x71\xfc\xfe\x04\x94\xa9\x6f\xde\x16\xa0\x71\x8d\x1a\x6c\x0a\x3e\x53\x0e\x72\x2b\x4b\x1d\x18\xa0\x91\x93\x81\xdc\x12\x02\x5f\xda\xd2\xb7\x48\x99\xdd\x40\x65\x4b\x10\xdc\x80\x28\x9d\xb7\xb9\xfa\x8e\xd0\x79\xb0\xac\xa0\x20\xbb\x56\x61\xb4\x81\x54\x69\x8a\x84\xc6\x43\xdd\xc6\x0e\x2c\xb5\x30\xe1\xbf\x42\xc8\x33\xd7\x20\x32\xa5\x25\x60\x33\x01\x5c\xe3\xf2\x79\xd8\x12\xc7\x67\xef\x19\xd5\x21\x0e\xb7\x19\xa8\xc9\x4c\x10\x72\x8f\xdb\x91\x31\x6c\x4d\xf7\xab\x30\x6d\x46\x51\xb2\x37\x9a\x76\x05\xd3\x9e\xee\x9c\x41\x6d\x33\x47\xa3\x5a\x72\x9b\xf9\x9b\xbb\xa4\x59\x21\xb3\xb8\xf9\x8b\xfb\x57\x00\x00\x00\xff\xff\x45\xab\x31\x54\xfa\x0a\x00\x00")

func pluginsCodeampGraphqlStaticIndexHtmlBytes() ([]byte, error) {
	return bindataRead(
		_pluginsCodeampGraphqlStaticIndexHtml,
		"plugins/codeamp/graphql/static/index.html",
	)
}

func pluginsCodeampGraphqlStaticIndexHtml() (*asset, error) {
	bytes, err := pluginsCodeampGraphqlStaticIndexHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "plugins/codeamp/graphql/static/index.html", size: 2810, mode: os.FileMode(420), modTime: time.Unix(1562954265, 0)}
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
	"plugins/codeamp/graphql/schema.graphql": pluginsCodeampGraphqlSchemaGraphql,
	"plugins/codeamp/graphql/static/index.html": pluginsCodeampGraphqlStaticIndexHtml,
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
			"graphql": &bintree{nil, map[string]*bintree{
				"schema.graphql": &bintree{pluginsCodeampGraphqlSchemaGraphql, map[string]*bintree{}},
				"static": &bintree{nil, map[string]*bintree{
					"index.html": &bintree{pluginsCodeampGraphqlStaticIndexHtml, map[string]*bintree{}},
				}},
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

