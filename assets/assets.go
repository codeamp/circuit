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

var _pluginsCodeampGraphqlSchemaGraphql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x59\xcf\x72\x1b\xbd\x0d\xbf\xeb\x29\xe8\xc9\x45\x99\xf1\x13\xe8\xd6\xc4\x69\xec\x36\x69\x5d\xeb\xcb\xa1\x93\xf1\x81\x5e\xc1\x12\xeb\x5d\x72\x3f\x92\xeb\xc4\xd3\xe9\xbb\x77\xf8\x77\x01\x92\x2b\x5b\xfa\xd2\x99\x5e\xec\x25\xb4\x04\x01\x10\xf8\xe1\xcf\x9a\x8e\xf7\x5c\xb3\xdf\xc4\x00\xab\xf8\xfc\x97\xed\xdf\xff\xb6\x5a\x99\xee\x00\x03\x67\xff\x5e\x31\xf6\xfb\x04\xfa\x65\xc3\xfe\xe1\xfe\xad\x18\x1b\x26\xcb\xad\x50\x72\xc3\xbe\xc6\xa7\xd5\x7f\x56\xab\x77\xf1\x77\xfb\x32\x42\x78\xf4\x7b\xdf\xb1\x6f\x06\xf4\x8a\xb1\xc9\x80\x5e\x8b\xdd\x86\xdd\x5c\xbd\xdf\x24\x62\xf8\xd5\xc4\x9f\xcd\xfa\xfd\x86\x7d\x77\x94\xfb\x0b\xff\xe3\xad\x56\xff\x82\xce\xae\x18\x1b\xc3\x53\x64\x70\xc9\x4c\x3f\xed\x37\x6c\x6b\xb5\x90\xfb\x4b\x26\xf9\x00\xf3\x0a\xe4\xb3\xd0\x4a\x0e\x20\xed\xcd\x55\x22\xbf\xdf\x20\x6e\xef\xd8\xa7\x9f\xa3\xd2\x76\xa6\x00\x5e\xe7\x53\x96\x38\x85\x07\x22\xa3\x99\x85\x34\xeb\xf8\xb0\x05\xae\xbb\x43\x3e\x38\x2c\x6f\xe4\x38\xd9\x4b\x36\x72\xcd\x07\xb3\x61\xb7\x7c\x2f\x24\xb7\x4a\x7b\xfa\x2c\xe5\x17\x61\x6c\x38\xe0\xcf\xc0\xed\xa4\xc1\x1d\xf0\x18\x1f\xd7\x8b\xbb\xe3\xcb\xf3\xee\x2d\xe8\x67\xd1\xf9\xdd\x26\x3e\x2e\xef\x8e\x2f\xcf\xbb\x83\x99\x18\xef\xfb\xbc\x9b\xfd\x10\xf6\x20\x24\xdb\x8b\x67\x90\x49\xe5\x9b\x2b\xc6\xe5\x8e\xda\x2b\x5b\x75\x5b\x9e\xfb\x89\x90\xf3\xe1\xc1\xa8\x8c\x61\xb9\x99\x19\xa1\x43\xc2\x6f\xdd\xd2\xbb\xc9\x76\x26\x44\x6f\xb9\x83\x1e\xb8\xf1\xaa\xea\xf8\xb8\xac\x6a\x7c\x19\xa9\x3a\xcb\xee\x38\x20\x55\xe6\xfb\x44\x4e\xe7\x44\x40\x5b\xee\x2b\x26\xec\x99\x6b\xc1\x1f\xfa\x68\xfa\x4e\x83\x3d\x6a\x79\xf7\x42\xd3\xf0\xd0\xe2\x79\xce\x2d\x50\x11\x3e\x61\xea\xc2\x1d\x08\xb9\xef\x21\xca\x96\xb5\xf0\xd1\x31\x87\x42\xfa\x31\x47\xc2\xa7\x9f\x16\xa4\x11\x4a\xfa\xbb\xf3\xe7\x47\x82\x59\x2f\x05\xd4\xf7\xbc\x89\x46\x7e\x26\xa3\xf0\x9a\x69\xde\x0f\xca\x37\xa9\x33\xa0\xc3\x67\xb7\x28\x38\xdc\x15\xd4\x24\x02\xe8\x41\x98\x7c\xf8\xbc\x72\x9b\x1c\x46\x5e\x04\xd8\xcb\x20\xe8\x91\x2f\xad\x22\xf8\x7d\xd4\xc0\x2d\x24\xd1\x57\x8c\x75\x9e\x90\x80\x26\xd2\x73\xd4\x17\x20\x10\x10\x72\xdc\x51\x16\x93\x27\x9c\xc2\x22\x4a\x11\xd5\xcf\x52\x44\xc5\xd7\x91\x9e\x83\xa2\x88\x91\xe0\x0b\x56\x8d\x88\x81\xb1\x6a\x4c\xdb\x03\x5a\x5e\x14\x1b\xe2\x99\x31\x6e\xf3\x99\x31\x6c\xd7\x91\x9e\x31\xa7\x80\x20\xac\xf9\x4c\x0a\x9a\x9f\xc2\xe2\x0a\x7a\x20\x52\xec\x3c\xe1\x14\x16\x37\x83\x0f\xc5\x81\xcb\x97\x19\x04\xb9\x65\x4a\xfa\x17\xc4\x40\x30\x2e\xbd\xb1\x89\xfb\x4a\x94\x4b\xb8\x75\x11\xbd\x8c\x9a\x29\x85\x0c\xb1\x95\x83\xb8\x35\xc2\xbf\x2c\x9e\x5b\x14\x12\x6f\xc3\xfe\x6c\xbb\x82\x2f\x31\xe0\x59\x7c\xa9\x41\x13\x5f\x62\xd5\xb3\xf8\x46\x3b\x20\x80\xc8\x66\x40\xa0\x8a\x01\x64\x83\xd1\x36\xb1\xfd\x44\xf6\x67\x33\x50\xb6\xc1\x0a\x7f\x84\x6d\xb4\x02\x65\x1b\x8c\xf0\x47\xd8\xd6\x46\xc8\x88\x8f\x9c\xc2\x83\x70\xc0\xe2\x84\xbf\x34\x8d\x2c\x68\x8e\x79\x25\x47\x78\x1b\xaf\x5a\xdd\xcc\x8b\xa1\xdb\x7f\x1b\x33\x1c\x4f\xed\xdc\x56\x07\x57\x48\x5d\x31\x8b\xce\xa1\x45\x72\xd7\xf7\xb0\x2e\x02\x0b\xca\x6c\x14\x7d\x2a\x91\xd7\xf9\x05\x97\x12\xe3\x63\xbe\x9f\x44\x20\x16\x2d\x39\x46\x77\x3a\x83\x63\xb2\x6b\xc9\x31\x7a\xd2\x19\x1c\x4b\xad\xcb\x8c\x33\xf3\x2c\xb3\xe9\xa6\xca\xb9\x45\x26\x39\x6e\x8c\x32\x2f\xfd\xb2\x83\x90\x8d\x22\x2d\x58\xe7\x7f\xa4\x90\xeb\x43\x70\xb2\xcf\x7a\xb9\xb6\x04\x95\x04\xeb\x89\xae\x43\x4b\x83\x08\xb3\x5b\x86\x9a\x2a\xba\x25\xcd\xe6\xa4\xcc\xac\x4c\xd8\x28\x41\x31\x6d\x56\x04\x11\xf3\xb1\x88\x98\xce\xfe\xa0\xd4\xd3\xc0\xf5\x13\xaa\x25\x1e\x22\x89\x76\x3e\x2e\x97\x7f\x50\xaa\x07\x2e\xc3\xce\xcf\x60\xd9\x67\x61\xd9\x47\x35\x0c\xc2\x4b\xba\x07\xfb\x59\xd8\xb8\x5e\xe7\x02\xd4\xef\xae\x1a\x27\x4f\x93\xf0\x23\x73\xc5\xfc\x7d\x11\x95\x0b\xe2\x95\x90\x16\xf4\x23\xef\x60\xa6\xf9\x5a\xaa\x53\x93\x43\xd1\x1b\x69\xe3\x16\x54\xc1\x87\xd2\x0b\x11\x1c\x70\xf4\xe0\x0d\x72\x84\x8d\xab\xf2\xad\x16\x2e\x59\xa7\x22\xf0\x3e\x32\x9f\xeb\xf1\xc0\x7b\x5e\x9f\xce\x3a\xec\x9d\x39\xe7\x1e\x2b\xb1\xce\x84\x73\x78\xfb\xcd\x89\x39\x6a\xff\x02\x73\x44\x38\x9d\x79\xdc\x9c\x98\xa3\xce\x34\x30\x47\x84\xd3\x99\xc7\xcd\x89\xb9\x1f\x08\x78\xae\xee\xc9\xef\x8c\xbe\xe8\x76\x0d\x5c\xf4\xb8\xed\x1e\x71\xe0\x91\x08\x0b\x68\xb7\xdb\xf8\x89\x06\xb5\x0a\xb1\x48\x71\x42\x59\x45\x3b\xda\x00\xc6\xf0\x3d\xe0\x73\x5d\xd0\xe3\xf5\x81\x9b\x03\x91\x8b\x6b\x90\xf6\xba\xa0\x6a\x78\xc4\xcb\x96\x88\xa9\xe0\xc4\x1e\xf1\x06\x11\x3b\x35\x0c\x5c\xee\x30\x77\x3c\x0b\xb9\xa0\x8d\x33\xa9\xba\x2e\xaa\xbb\x71\x39\xd5\x99\xd3\xf5\x36\xf7\x17\xb4\x07\x26\xb5\x8a\xfb\xcd\x09\x7a\x44\x29\x87\xd2\x63\xaf\x5e\xdc\xeb\x5b\xab\xb9\x85\xfd\x4b\xe8\x9a\x56\x8c\xf5\xae\x69\x05\x63\x6e\xb5\x7a\x80\x4c\xd5\xc0\x77\xa2\x26\x8f\x1a\x5c\xef\x71\xad\xd4\x53\x3a\x2f\x98\x0c\x17\x4f\xde\x6c\xb8\xf9\xa6\xa6\x2b\x6d\xf2\x04\x2f\x78\x29\xcc\x15\x3c\xf2\xa9\xb7\x04\xf5\x3a\xd5\x2b\x7d\x54\xc5\x34\xef\xa9\xbd\xb9\x35\x07\xc0\x40\x52\xc8\x57\xc8\xf3\xcc\xfb\x89\xde\x61\xa7\xa8\xb5\x5b\xbe\x10\x7c\xd3\x45\x4f\xeb\x7e\x9e\x41\xe7\x70\x49\x78\x74\xfc\x86\x2b\x75\x85\xd9\xc6\xda\x8e\x62\x37\x1e\xd5\x10\x07\x76\x6e\xf6\xca\x4d\x74\xe3\x74\x07\xbf\x4f\x60\x6c\x41\xfd\x22\x06\x41\x68\x03\x0c\x4a\xbf\x34\x5e\x0e\x3f\x54\xef\x5b\x07\x11\xd2\xf7\xe2\x9f\x35\xef\xe0\x16\xb4\x50\xbb\x46\x64\xe4\xa8\x58\x50\xba\xf6\x0d\x9c\x7e\x48\xea\x79\x43\xc4\xd2\x5b\xe2\xda\x8a\x47\xee\x5d\x28\x4c\x14\x18\x3b\x00\xdf\x45\x8c\xca\xe3\x3c\xaf\x0f\x17\x7d\x8b\x6e\x2c\xb7\x40\xd1\xa6\x18\x72\x2c\x8d\x38\xfc\xce\xaf\x35\xc8\x9d\xe4\x14\xc6\x72\x4d\x08\x8f\x42\x0a\x73\xa0\x26\xbc\x53\x7d\xff\xc0\xbb\xa7\x2a\xeb\xc7\x3a\x04\x67\x93\x57\x1c\x06\x0f\x7e\x83\xb6\xa3\x32\xc2\x2a\xfd\x42\xaf\x36\x36\x21\x99\xb2\x17\xf6\x9b\xee\x0b\xca\xad\x56\x56\x75\x8a\x90\xb5\xe1\xb7\x5a\x3c\x73\x0b\x7f\xa5\x51\xe9\x7e\x98\x1e\x7a\xd1\x15\xf4\x3c\x94\x35\x07\xf5\xe3\xca\xa3\x9e\xd3\x3e\x6a\x7a\x64\xd2\x5b\xcc\x6a\xbb\x49\xbb\xe4\x71\x57\x0c\x63\xce\x99\x66\xbe\x32\xe9\xbd\x64\xc6\x8f\xa2\x91\x22\xf5\xf0\xf7\xf8\xc8\x72\x89\x05\x9e\x62\x02\xf6\xc0\xe6\x98\x6e\x2f\xec\x07\xcd\x65\x47\xb2\x65\xa7\xa4\x15\x72\x52\x93\x09\xc6\x24\xa0\x0c\xa4\xfe\xad\x8b\xdc\x54\xcf\xa2\x1b\x58\x4a\xb9\xc5\x9c\x32\xa4\x90\x4c\x7b\x05\xb6\xd4\x30\x2a\xe9\x03\x04\x21\x4e\x99\x12\x79\x77\x00\x87\xfc\x44\x94\x02\xeb\x8f\x06\x9b\x92\x8f\x62\x3f\x43\x43\x4b\x8b\xaa\x97\xc1\xb1\xb4\xa4\x4e\x0b\x98\x5a\x3d\xe6\x02\x44\x55\x72\x4d\xc6\xaa\xe1\x63\x41\xad\x80\xe9\x17\xe0\x0d\x46\x5e\xd4\x7b\x62\x0c\x5e\xd2\xb9\x1c\x72\x16\x3a\x97\x16\xc3\x61\xb4\x15\x7b\x19\x81\xb7\x84\x98\x85\xdf\x4a\xd5\x4b\xd7\x58\x32\x45\xc3\xd8\x14\x60\xfd\x64\xbe\x86\xd8\x23\x9e\xc1\x84\x8b\xd7\x95\xff\x4b\x66\xc4\xd9\x3e\xb1\xa8\x5a\x44\xc4\x26\x74\xd2\xb8\x6d\x47\x1e\xbd\xdb\x79\xe0\x7f\x2c\xc8\x89\xec\xe1\x8b\x59\x4b\x03\xf4\x2d\xcd\xeb\x51\xa7\x81\x25\x30\xc0\xed\xa6\xf3\x12\xc2\x9c\x60\x9c\x67\x3c\xfa\x1b\xba\xf1\x03\xba\x3e\xd4\x18\x6e\x41\x1c\x11\xb3\xc0\x23\xf4\xda\xc2\x28\xb9\xcf\xe6\x40\x11\x49\x89\x4d\xe3\xf9\x94\xa3\x74\x07\x77\xf0\x30\x89\xbe\x52\x2d\x55\x63\x58\x28\x3c\xdd\xae\x85\x6a\x9e\xfd\x86\xce\xa2\xdd\x3d\xc4\xb3\x6e\x95\x0e\x6e\x76\x71\xdf\xf0\xff\x45\xcd\x5a\x5d\xc3\x55\x45\x0b\x8c\xeb\x4e\xe2\x1a\x78\x6f\x0f\x7e\xe1\x5f\x69\x74\x15\x8d\x57\x16\x3b\x8c\x34\xbb\x8f\x63\x48\x62\xd2\xc6\x74\xdf\x5b\x36\x61\xc6\x3f\xff\xf4\xf5\x4b\xe0\x75\xee\x35\x53\x11\xc2\x47\x3c\x22\x42\xe3\x33\x6a\x70\xd9\x53\x0f\xf9\xa8\xa4\xe5\x42\x82\x66\xd5\x19\xe5\x6d\x86\x03\x94\x46\xf7\x9e\x11\x23\x8d\x66\xc2\xce\x85\x5b\xf3\x0c\x4a\x77\x18\xf8\xcf\xed\xa4\x63\xa0\x45\xc2\x37\xc9\x9f\xb9\xe8\x43\xfe\xa4\xac\xcb\x3b\xac\x78\xfa\xae\xc0\x1e\x68\xb9\x5f\x78\x34\xd6\xc3\xb7\x58\x07\x18\x30\x83\x91\x5b\x0c\x6f\x42\x0a\x2b\x78\x7f\x05\x3d\x7f\xd9\x42\xa7\xe4\xce\xa4\xad\xa3\xef\x2c\x0a\xa2\x15\x03\xa8\xc9\x16\x54\x33\x75\x1d\x18\xf3\xdb\x41\x83\x39\x28\x17\xbc\x81\xfe\xc8\x45\x3f\x69\xa8\xe8\x07\x6b\xc7\x6b\xe0\x3b\xd0\x2e\xb4\x90\xde\xd7\xf9\x87\x14\x64\x2d\xeb\x14\x6f\x79\x3b\x95\x51\x5c\x34\x9b\x55\x47\xd7\x72\x87\xfc\x21\xa7\x06\x93\xff\xa7\x06\x6f\xb1\x77\xc3\xed\x39\x89\xa8\xe2\x13\xcd\xeb\xea\x9d\x33\x49\x58\x9c\x10\x14\xa6\xce\xdf\x4f\x6a\x31\x5e\x1b\x18\x54\xf5\xc6\xc2\x00\x81\xe4\xe4\x45\x50\x5e\x6a\xfc\xdb\x3a\x2c\x63\xe5\xfc\xb9\x26\x42\xa5\x27\xfc\x22\xa4\x6c\x0a\x73\x04\x35\x0b\x61\x4e\x3f\x8f\xb4\x0f\xc5\x11\xf8\x83\xc3\xc9\x4e\xd4\x6c\x2e\x16\x2f\xa7\xdd\x65\x94\x25\x3a\x75\x88\x66\xe7\xd0\x2a\xb3\x8e\x6a\x72\xd9\x34\xdb\x25\xae\xab\x0b\xfa\xdb\x1a\x87\x63\x56\xa7\x1f\x59\x88\xc8\xad\xef\x2f\x5e\xe2\xc9\x80\x2e\xaa\x2d\x32\x3e\xa6\x1b\x23\x9e\x9e\x7a\xde\x7c\x5c\x15\x8c\x7b\xcd\x65\x15\x3b\xf5\xa7\x9b\xa6\xfd\x5b\x80\xb4\xe8\x0a\x6f\x3c\x28\x98\x66\xf1\xa0\xd9\x72\xad\x98\xa0\xa6\x5b\x10\x33\x98\xef\xbf\x01\x00\x00\xff\xff\x87\xf6\x51\x44\x5b\x28\x00\x00")

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

	info := bindataFileInfo{name: "plugins/codeamp/graphql/schema.graphql", size: 10331, mode: os.FileMode(420), modTime: time.Unix(1551136783, 0)}
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

