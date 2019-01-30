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

var _pluginsCodeampGraphqlSchemaGraphql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x59\xcd\x72\xe3\xb8\x11\xbe\xeb\x29\xa0\xda\x8b\xa6\xca\x4f\xa0\x5b\x66\x3c\x19\x3b\x99\x49\x1c\x6b\xe7\x90\x9a\xf2\x01\xa6\xda\x12\x62\x12\xe0\x02\xa0\x77\x55\xa9\xbc\x7b\x0a\xbf\xec\x06\x40\xd9\xf2\x6e\xaa\x72\xb1\x85\x8f\xc4\x07\x74\xa3\x7f\x41\xd3\xf1\x9e\x6b\xf6\xb3\x18\x60\x15\x7f\xff\x65\xf7\xf7\xbf\xad\x56\xa6\x3b\xc2\xc0\xd9\xbf\x57\x8c\xfd\x32\x81\x3e\x6d\xd9\x3f\xdc\xbf\x15\x63\xc3\x64\xb9\x15\x4a\x6e\xd9\xb7\xf8\x6b\xf5\x9f\xd5\xea\xa7\xf8\xdc\x9e\x46\x08\x3f\xfd\xdc\x9f\xd8\x77\x03\x7a\xc5\xd8\x64\x40\x6f\xc4\x7e\xcb\x6e\xaf\x3f\x6c\x13\x18\x9e\x9a\xf8\xd8\x6c\x3e\x6c\xd9\x0f\x87\x3c\xac\xfd\xc3\x3b\xad\xfe\x05\x9d\x5d\x31\x36\x86\x5f\x91\xe0\x8a\x99\x7e\x3a\x6c\xd9\xce\x6a\x21\x0f\x57\x4c\xf2\x01\xe6\x11\xc8\x17\xa1\x95\x1c\x40\xda\xdb\xeb\x04\x7f\xd8\x22\xb6\xcc\x6c\x66\x6a\xb3\x89\x3f\x76\xc0\x75\x77\xcc\xaf\x87\xe1\xad\x1c\x27\x7b\xc5\x46\xae\xf9\x60\xb6\xec\x8e\x1f\x84\xe4\x56\x69\x8f\xcf\xdc\x5f\x85\xb1\x61\xeb\x7f\x06\x6e\x27\x0d\x6e\x81\xa7\xf8\x73\xb3\x38\x3b\xbe\x3c\xcf\xde\x81\x7e\x11\x9d\x9f\x6d\xe2\xcf\xe5\xd9\xf1\xe5\x6a\x36\x33\x23\x74\x88\x62\xe7\x86\x5e\xc5\xbb\x19\x88\x9a\xbe\x87\x1e\xb8\xf1\x0b\xea\xf8\x73\x79\xc1\xf8\xf2\xbc\xe0\xe7\x59\xe3\x8e\x01\x1d\xc0\xac\x55\x74\x60\x6e\x0b\x68\xca\x43\x45\xc2\x5e\xb8\x16\xfc\xb1\x8f\x0a\xe8\x34\xd8\xb3\xf2\xbb\x17\x90\xf8\x42\x1e\x7a\x88\x70\x26\xf0\xa6\x93\xd7\xcf\x0f\xb3\x29\x7c\xfe\xcd\x82\x34\x42\x49\xaf\x36\x27\x45\x02\xcc\x66\xc9\xa2\x7e\xe4\x49\xd4\x60\x33\x8c\xec\x6b\xc6\xfc\x11\x94\x6f\xd2\x73\x40\x8b\xcf\x27\x52\x30\xdc\x17\x68\xda\x02\xe8\x41\x98\xbc\xf8\x3c\x72\x93\x9c\x6b\xaf\x83\xb7\x66\xdf\xf5\x0e\x9b\x46\xd1\x67\x3f\x69\xe0\x16\xd2\xd6\x57\x8c\x75\x1e\x88\x9b\x4e\x67\x9a\xcd\xbe\xf0\x82\xe0\xd8\xe3\x9e\x52\x4c\x1e\xb8\x84\x22\xee\x22\x8a\x9f\x77\x11\x05\xdf\x44\x3c\xdb\x63\x61\x9e\xc1\x16\xac\x1a\x11\x81\xb1\x6a\x4c\xd3\x43\x28\x59\x17\x13\xe2\x9a\xd1\x65\xf2\x9a\xd1\x63\x36\x11\xcf\x4e\x57\xf8\x20\x96\x7c\x86\x82\xe4\x97\x50\x5c\x43\x0f\x64\x17\x7b\x0f\x5c\x42\x41\x05\x49\x46\x4d\xa4\x71\xfe\xbf\x41\xc1\x21\x13\xb8\x41\xc1\xb9\x0b\xf3\xb3\x74\x05\x2f\x11\xf1\x5d\xbc\x54\xe4\xc4\x4b\xe4\x7e\x17\x6f\xd4\x03\x72\xe1\xac\x06\x14\x71\xb0\x8b\x6f\x71\x28\x4a\xb4\x9f\xc9\xfc\xac\x06\x4a\x1b\xb4\xf0\x7b\x68\xa3\x16\x28\x6d\x50\xc2\xef\xa1\xad\x95\x90\x43\x2c\x32\x0a\x1f\x26\x43\xb4\x4c\x11\x92\xc6\xd8\x05\xc9\x31\x57\x32\x84\xb7\x71\xd5\xe2\x66\x2e\x86\x4e\xff\x6d\x64\xb7\xc3\xa8\xb4\x65\x03\x97\xa7\x26\xa3\x61\xdc\x32\x25\xbd\x7f\x08\xff\xee\x2e\xa6\x96\x98\x62\xb6\x91\x22\xc2\x69\x89\x1f\x61\xbc\x8e\x01\x36\xe9\xb2\xcc\x17\xd1\xa6\x12\xbc\xc9\x2f\x6c\x59\x06\xf3\xf9\x24\x80\x68\xb4\x64\x8c\xe6\xf4\x0e\xc6\xa4\xd7\x92\x31\x5a\xd2\x3b\x18\x4b\xa9\xcb\x9c\x30\x73\x96\xf9\x6e\x5b\x65\xc5\x22\xd6\x9f\x57\x46\x99\x39\xfe\xb0\x85\x90\x8e\x22\x16\xb4\xf3\x3f\x12\xc8\x15\xb8\x38\x1d\x67\xb9\x5c\xbd\x8b\x92\xf6\x66\xa2\xe3\x50\x2b\x23\x60\x36\x4b\x5f\x86\x24\xb3\xa4\xf9\x96\xd4\x60\x95\x0a\x1b\xf5\x19\xc6\x66\x41\x10\x98\x97\x45\x60\x5a\xfb\xa3\x52\xcf\x03\xd7\xcf\x28\xdb\x3f\x46\xe8\x8e\x14\xee\x2e\xdb\x7e\x54\xaa\x07\x2e\xc3\xcc\x2f\x60\xd9\x17\x61\xd9\x27\x35\x0c\xc2\xef\xf4\x00\xf6\x8b\xb0\x71\x9c\x76\xe7\xaa\xae\xdb\xeb\x75\x55\xdb\x7b\x4c\xc2\xaf\x99\x15\xf3\xfb\x32\x27\x57\x8b\x2b\x21\x2d\xe8\x27\xde\xc1\x8c\xf9\x6a\xa7\x53\x93\x8b\xa2\xb7\xd2\xc6\x29\xa8\xbc\x0d\xc5\x11\x02\x5c\xe0\xe8\xc1\x2b\xe4\x0c\x8d\x2b\x81\xad\x16\x60\xe6\x32\xed\x21\x92\xcf\xc5\x6a\xe0\x9e\xc7\x97\x53\x87\xb9\x33\x73\x6e\x03\x12\x75\x06\xde\xc3\xed\x27\x27\x72\xd4\xa1\x04\x72\x04\x5c\x4e\x1e\x27\x27\x72\xd4\x3c\x05\x72\x04\x5c\x4e\x1e\x27\x27\x72\xdf\x69\x7a\x56\xf7\xcb\xcf\x8c\xb6\xe8\x66\x0d\x5c\xf4\xa9\xa0\x5f\xd3\x7a\xb9\xf0\xb0\x10\xed\xf6\x5b\xdf\x2a\x53\xad\x10\x8d\x14\x2b\x94\x75\xae\xc3\x06\x30\x86\x1f\x00\xaf\xeb\x9c\x1e\x8f\x8f\xdc\x1c\xc9\xbe\xb8\x06\x69\x6f\x0a\x54\xc3\x13\x1e\xb6\xb6\x98\x4a\x42\x6c\x11\x6f\xd8\x62\xa7\x86\x81\xcb\x3d\x66\xc7\x4d\xf6\x9a\x76\x95\xa4\xea\x5a\x57\x67\xe3\x72\xaa\x53\xa7\xeb\x3e\x1e\xd6\xb4\x41\x24\xb5\x8a\x7b\xe6\x36\x7a\x46\x28\x17\xa5\xc7\x5e\x9d\xdc\xeb\x3b\xab\xb9\x85\xc3\x29\xf4\x35\x2b\xc6\x7a\xf1\x02\x12\x8c\xb9\xd3\xea\x11\x32\xaa\x81\xef\x45\x0d\x8f\x1a\x5c\x77\x70\xa3\xd4\x73\x5a\x2f\xa8\x0c\x17\x4f\x5e\x6d\xb8\x33\xa5\xaa\x2b\x75\xf2\x0c\x27\x3c\x14\xe6\x1a\x9e\xf8\xd4\x5b\x12\xf5\x3a\xd5\x2b\x7d\x56\xc4\x74\x25\x51\x5b\x73\xab\x49\xc6\x81\xa4\xd8\x5f\xb1\x9f\x17\xde\x4f\xf4\x0c\x3b\x45\xb5\xdd\xb2\x85\x60\x9b\xce\x7b\x5a\xe7\xf3\x02\x3a\xbb\x4b\x8a\x47\xe7\x4f\xb8\x12\x57\x98\x5d\xac\xed\x68\xec\xc6\xf7\x18\xc4\x80\x9d\x99\xbd\x72\x12\xdd\x38\xdd\xc3\x2f\x13\x18\x5b\xa0\x5f\xc5\x20\x08\x36\xc0\xa0\xf4\xa9\xf1\x72\x78\x50\xbd\x6f\x5d\x88\x90\xbe\x5b\xfe\xa2\x79\x07\x77\xa0\x85\xda\x37\x3c\x23\x7b\xc5\x82\xd0\xb5\x6d\xe0\xf4\x43\x52\xcf\x1b\x3c\x96\x9e\x12\xd7\x56\x3c\x71\x6f\x42\xa1\xe7\x67\xec\x08\x7c\x1f\x63\x54\xbe\x71\xf2\xf2\x70\xd1\xb7\x70\x63\xb9\x05\x1a\x6d\x8a\x6b\x88\xa5\x4b\x08\x3f\xf3\x5b\x1d\xe4\x2e\x32\x0a\x63\xb9\x26\xc0\x93\x90\xc2\x1c\xa9\x0a\xef\x55\xdf\x3f\xf2\xee\xb9\xca\xfa\xb1\x0e\xc1\xd9\xe4\x15\x83\xc1\x37\x8a\x41\xda\x51\x19\x61\x95\x3e\xd1\xa3\x8d\x4d\x48\x46\x0e\xc2\x7e\xd7\x7d\x81\xdc\x69\x65\x55\xa7\x08\xac\x0d\xbf\xd3\xe2\x85\x5b\xf8\x2b\xf5\x4a\xf7\x60\x7a\xec\x45\x57\xe0\xf9\xde\xd0\x1c\xd5\xaf\xd7\x3e\xea\x39\xe9\xa3\xa4\x67\x2e\x23\x8b\xeb\xc4\x6e\xd2\x2e\x79\xdc\x17\xd7\x25\xef\xb9\xea\x7b\xe5\x32\xf2\x8a\x19\x7f\x5b\x8a\x04\xa9\xef\x27\xcf\xdf\xe7\x2d\x51\xe0\x2b\x3e\xc0\x16\xd8\xbc\x48\x3b\x08\xfb\x51\x73\xd9\x91\x6c\xd9\x29\x69\x85\x9c\xd4\x64\x82\x32\x49\x50\x06\x52\xff\xd6\x45\x6e\xaa\x67\xd1\x09\x2c\xa5\xdc\xe2\x26\x31\xa4\x90\x8c\xbd\x12\xb6\xd4\x30\x2a\xe9\x1d\x04\x45\x9c\x32\x25\xf2\xee\x08\x2e\xf2\x93\xad\x14\xb1\xfe\xac\xb3\x29\xf9\x24\x0e\x73\x68\x68\x49\x51\xf5\x32\xd8\x97\x96\xc4\x69\x05\xa6\x56\x8f\xb9\x10\xa2\xaa\x7d\x4d\xc6\xaa\xe1\x53\x81\x56\x81\xe9\x0f\x88\x37\x38\xf2\xa2\xde\x13\xc7\xe0\x25\x99\xcb\x6b\xc8\x42\xe6\x52\x63\xd8\x8d\x76\xe2\x20\x63\xe0\x2d\x43\xcc\xc2\xb3\x52\xf4\xd2\x34\x96\x54\xd1\x50\x36\x0d\xb0\x8c\x35\x43\xec\x19\xcb\x60\xc2\xf9\xeb\xca\xff\x25\xb7\xb8\x59\x3f\xb1\xa8\x5a\x8c\x88\xcd\xd0\x49\xfd\xb6\xed\x79\xf4\x6c\xe7\x2b\xf9\x73\x4e\x4e\xf6\x1e\x3e\xea\xb4\x24\x40\x9f\x7b\xbc\x1c\x75\x1a\x58\x0a\x06\xb8\xdd\x74\x56\x42\xc8\x49\x8c\xf3\xc4\xa3\x3f\xa1\x5b\x7f\x41\xd7\x87\x1a\xc3\x0d\x88\x21\x62\x0a\x7c\xc9\x5d\x6b\x18\x25\xf7\x59\x1d\xc8\x23\x29\xd8\x54\x9e\x4f\x39\x4a\x77\x70\x0f\x8f\x93\xe8\x2b\xd1\x52\x35\x86\x37\x85\xef\x9f\xeb\x4d\x35\xd7\x7e\x43\x67\xd1\xee\x1e\xe2\x5a\x77\x4a\x07\x33\x5b\x3f\x34\xec\x7f\x51\xb2\x56\xd7\x70\x5d\x61\x81\xb8\xee\x24\x6e\x80\xf7\xf6\xe8\x07\xfe\x95\x46\x57\xd1\x78\x65\xb1\xc3\xf8\xa4\xa4\xe5\x42\x82\xf6\xc2\xb5\x34\x9a\xa5\x0c\xb6\xa2\x34\xd2\x47\xf6\xa4\x74\x65\x11\x66\x2e\x48\xe3\x09\x4a\x35\x0d\xfc\xb7\xdd\xa4\xa3\x01\x46\xe0\xbb\xe4\x2f\x5c\xf4\x21\xaf\x50\xea\x52\xb6\x8a\xd3\x57\xcb\xf6\x48\xcb\xe0\xe2\xa4\xb1\x1c\xbe\xf5\x38\xc2\x80\x09\x46\x6e\xb1\xdb\x0b\x29\xac\xe0\xfd\x35\xf4\xfc\xb4\x83\x4e\xc9\xbd\x49\x53\x47\x5f\x71\x17\xa0\x15\x03\xa8\xc9\x16\xa8\x99\xba\x0e\x8c\xf9\xf9\xa8\xc1\x1c\x95\x33\xea\x80\x3f\x71\xd1\x4f\x1a\x2a\xfc\x68\xed\x78\x03\x7c\x0f\xda\x99\x1c\x92\xfb\x26\x3f\x48\xc6\xd7\xd2\x4e\xf1\x96\xd7\x53\x69\xdd\x45\x13\x56\x75\x3a\x2d\x73\xc8\x1f\x38\x6a\x27\xfb\x7f\x6a\x7c\x16\x7b\x1a\xdc\xb6\x62\xf1\xca\x4f\x17\xaf\x8b\xf7\x9e\x0e\x7b\xb1\x73\x2e\x54\x9d\xbf\x2b\xd4\xdb\x78\xad\x91\xae\xf2\xf0\x42\x63\x4d\x72\xd5\x62\xb0\x5a\x6a\x88\xdb\x32\x0c\x55\x10\xa9\x3f\x63\x78\x89\x62\x65\xf1\xcf\x3f\x7d\xfb\x1a\xd6\x7a\x35\x51\xb8\xa6\xf2\x4d\xa9\xa3\x51\xf8\xd2\x83\x26\x57\xe5\x17\x1f\x73\xb3\x2c\x5e\x54\x5f\xbb\x3e\x2e\x8b\x4b\x7a\x64\xcd\x9a\xb7\x55\x20\x9c\x95\xe4\xaa\xa9\xc8\x2b\x5c\x11\x16\xf8\xdb\x4a\xde\x73\x5a\xa7\x9f\x07\xc8\x96\x5b\x5f\x0e\xfc\x8e\xeb\x73\xa5\x17\x9f\x74\x62\x8c\x78\x97\xae\x37\x2f\x57\xb9\xcb\x41\x73\x59\x59\x77\xfd\xd1\xa1\xa9\xff\x56\xc8\x58\x34\x85\x37\x2e\x14\x54\xb3\xb8\xd0\xac\xb9\x96\x97\x50\xd5\x2d\x6c\x33\xa8\xef\xbf\x01\x00\x00\xff\xff\x36\xaa\x14\xa3\x6e\x25\x00\x00")

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

	info := bindataFileInfo{name: "plugins/codeamp/graphql/schema.graphql", size: 9582, mode: os.FileMode(420), modTime: time.Unix(1548814731, 0)}
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

	info := bindataFileInfo{name: "plugins/codeamp/graphql/static/index.html", size: 2810, mode: os.FileMode(420), modTime: time.Unix(1533663930, 0)}
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

