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

var _pluginsCodeampGraphqlSchemaGraphql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x59\xcd\x72\x23\xb9\x0d\xbe\xeb\x29\xa8\xda\x8b\xa6\xca\x4f\xa0\xe3\x8c\x27\x63\x27\xbb\x89\x63\xed\x9c\x5c\x3e\xd0\x2d\x58\x62\xdc\x4d\xf6\x92\x6c\xef\xaa\x52\x79\xf7\x14\x7f\x1b\x20\xd9\xb2\xe5\x9d\x54\xe5\x62\x37\xbf\x6e\x7e\x24\x40\x00\x04\x20\xd3\xf1\x9e\x6b\xf6\xab\x18\x60\x15\x9f\xff\xba\xfb\xc7\xdf\x57\x2b\xd3\x1d\x61\xe0\xec\xdf\x2b\xc6\x7e\x9b\x40\x9f\xb6\xec\x9f\xee\xdf\x8a\xb1\x61\xb2\xdc\x0a\x25\xb7\xec\x97\xf8\xb4\xfa\xcf\x6a\xf5\x53\x7c\x6f\x4f\x23\x84\x47\x3f\xf7\x27\xf6\xdd\x80\x5e\x31\x36\x19\xd0\x1b\xb1\xdf\xb2\xdb\xeb\x4f\xdb\x04\x86\xb7\x26\xbe\x36\x9b\x4f\x5b\xf6\xe0\x90\xc7\xb5\x7f\x79\xa7\xd5\xbf\xa0\xb3\x2b\xc6\xc6\xf0\x14\x09\xae\x98\xe9\xa7\xc3\x96\xed\xac\x16\xf2\x70\xc5\x24\x1f\x60\x1e\x81\x7c\x15\x5a\xc9\x01\xa4\xbd\xbd\x4e\xf0\xa7\x2d\x62\xcb\xcc\x66\xa6\x36\x9b\xf8\xb0\x03\xae\xbb\x63\xfe\x3c\x0c\x6f\xe5\x38\xd9\x2b\x36\x72\xcd\x07\xb3\x65\x77\xfc\x20\x24\xb7\x4a\x7b\x7c\xe6\xfe\x59\x18\x1b\xb6\xfe\x17\xe0\x76\xd2\xe0\x16\x78\x8e\x8f\x9b\xc5\xd9\xf1\xe3\x79\xf6\x0e\xf4\xab\xe8\xfc\x6c\x13\x1f\x97\x67\xc7\x8f\xab\xd9\xcc\x8c\xd0\x21\x8a\x9d\x1b\x7a\x15\xef\x66\x20\x6a\xfa\x1e\x7a\xe0\xc6\x2f\xa8\xe3\xe3\xf2\x82\xf1\xe3\x79\xc1\xaf\xb3\xc6\x1d\x03\x3a\x80\x59\xab\xe8\xc0\xdc\x16\xd0\x94\xc7\x8a\x84\xbd\x72\x2d\xf8\x53\x1f\x15\xd0\x69\xb0\x67\xe5\x77\x1f\x20\xf1\x85\x3c\xf4\x10\xe1\x4c\xe0\x4d\x27\xaf\x9f\x5f\x66\x53\xf8\xfa\x87\x05\x69\x84\x92\x5e\x6d\x4e\x8a\x04\x98\xcd\x92\x45\x3d\xe4\x49\xd4\x60\x33\x8c\xec\x6b\xc6\xfc\x11\x94\x5f\xd2\x73\x40\x8b\xcf\x27\x52\x30\xdc\x17\x68\xda\x02\xe8\x41\x98\xbc\xf8\x3c\x72\x93\x9c\x6b\xaf\x83\xb7\x66\xdf\xf5\x0e\x9b\x46\xd1\x67\xbf\x68\xe0\x16\xd2\xd6\x57\x8c\x75\x1e\x88\x9b\x4e\x67\x9a\xcd\xbe\xf0\x82\xe0\xd8\xe3\x9e\x52\x4c\x1e\xb8\x84\x22\xee\x22\x8a\x9f\x77\x11\x05\xdf\x44\x3c\xdb\x63\x61\x9e\xc1\x16\xac\x1a\x11\x81\xb1\x6a\x4c\xd3\x43\x28\x59\x17\x13\xe2\x9a\xd1\x65\xf2\x9a\xd1\x63\x36\x11\xcf\x4e\x57\xf8\x20\x96\x7c\x86\x82\xe4\x97\x50\x5c\x43\x0f\x64\x17\x7b\x0f\x5c\x42\x41\x05\x49\x46\x4d\xa4\x71\xfe\xbf\x41\xc1\x21\x13\xb8\x41\xc1\xb9\x0b\xf3\xb3\x74\x05\x2f\x11\xf1\x43\xbc\x54\xe4\xc4\x4b\xe4\xfe\x10\x6f\xd4\x03\x72\xe1\xac\x06\x14\x71\xb0\x8b\x6f\x71\x28\x4a\xb4\x5f\xc9\xfc\xac\x06\x4a\x1b\xb4\xf0\x67\x68\xa3\x16\x28\x6d\x50\xc2\x9f\xa1\xad\x95\x90\x43\x2c\x32\x0a\x1f\x26\x43\xb4\x4c\x11\x92\xc6\xd8\x05\xc9\x31\x57\x32\x84\xf7\x71\xd5\xe2\x66\x2e\x86\x4e\xff\x7d\x64\x49\xc8\x32\x90\xc7\xc3\x4e\xf0\x26\x7f\xb0\x65\x19\xcc\x8a\x4b\x00\x11\xb5\x64\x8c\xe7\xfc\x01\xc6\x24\x70\xc9\x18\x8f\xf8\x03\x8c\xa5\xd4\x65\xb0\x9e\x39\xcb\x8b\x68\x5b\x5d\x57\x45\x10\x3e\xaf\x8c\x32\xa4\xff\xb0\x85\x90\x8e\x22\x16\xb4\xf3\x3f\x12\xc8\x65\x9e\xf8\x9e\xcc\x72\xb9\x44\x14\xdd\xa6\x9b\x89\x8e\x43\x12\x8b\x80\xb4\xd8\x43\xc8\x0f\xd6\xf1\x42\xa6\x17\x21\x49\x8e\x2a\x15\x36\x12\x27\x8c\xcd\x82\x20\x30\x2f\x8b\xc0\xb4\xf6\x67\xa5\x5e\x06\xae\x5f\xd0\x35\xfc\x14\xa1\x3b\x92\x51\xbb\x6b\xf0\xb3\x52\x3d\x70\x19\x66\x7e\x03\xcb\xbe\x09\xcb\xbe\xa8\x61\x10\x7e\xa7\x07\xb0\xdf\x84\x8d\xe3\xb4\x3b\x97\x0e\xdd\x5e\xaf\xab\xa4\xdb\x63\x12\x7e\xcf\xac\x98\xdf\xe7\x1f\x39\x8d\x5b\x09\x69\x41\x3f\xf3\x0e\x66\xcc\xa7\x21\x9d\x9a\x5c\x78\xbb\x95\x36\x4e\x41\x79\x67\xc8\x5a\x10\xc0\xc4\x30\xf6\xe0\x15\x72\x86\xc6\xe5\xa6\x56\x0b\x30\x73\xfe\xf4\x18\xc9\xe7\x2c\x32\x70\xcf\xe3\xcb\xa9\xc3\xdc\x99\x39\xe7\xe7\x89\x3a\x03\x1f\xe1\xf6\x93\x13\x39\x2a\x1d\x02\x39\x02\x2e\x27\x8f\x93\x13\x39\xaa\x6a\x02\x39\x02\x2e\x27\x8f\x93\x13\xb9\x2f\x01\x3d\xab\x7b\xf2\x33\xa3\x2d\xba\x59\x03\x17\x7d\xca\xb4\xd7\x34\x91\x2d\x3c\x2c\x44\xbb\xfd\xd6\xd7\xb0\x54\x2b\x44\x23\xc5\x0a\x65\x02\xea\xb0\x01\x8c\xe1\x07\xc0\xeb\x3a\xa7\xc7\xe3\x23\x37\x47\xb2\x2f\xae\x41\xda\x9b\x02\xd5\xf0\x8c\x87\xad\x2d\xa6\x5c\x0d\x5b\xc4\x3b\xb6\xd8\xa9\x61\xe0\x72\x8f\xd9\x71\xf5\xbb\xa6\xe5\x1e\x49\x87\xd6\xd5\xd9\x8c\x4a\xbb\xa0\xf2\xe0\xca\x82\xc7\x35\xad\xdc\x48\x12\xe1\xde\xb9\x8d\x9e\x11\xca\x45\xe9\xb1\x57\x27\xf7\xf9\xce\x6a\x6e\xe1\x70\x0a\x05\xc7\x8a\xb1\x5e\xbc\x82\x04\x63\xee\xb4\x7a\x82\x8c\x6a\xe0\x7b\x51\xc3\xa3\x06\x97\xb6\xdf\x28\xf5\x92\xd6\x0b\x2a\xc3\x59\x8d\x57\x1b\x2e\x19\xa9\xea\x4a\x9d\xbc\xc0\x09\x0f\x85\xb9\x86\x67\x3e\xf5\x96\x44\xbd\x4e\xf5\x4a\x9f\x15\x31\xf5\x0a\x6a\x6b\x6e\x55\xaf\x38\x90\x14\xfb\x2b\xf6\xf3\xca\xfb\x89\x9e\x61\xa7\xa8\xb6\x5b\xb6\x10\x6c\xd3\x79\x4f\xeb\x7c\x5e\x41\x67\x77\x49\xf1\xe8\xfc\x09\x57\xe2\x0a\xb3\x8b\x49\x17\x8d\xdd\xb8\xc1\x40\x0c\xd8\x99\xd9\x1b\x27\xd1\x8d\xd3\x3d\xfc\x36\x81\xb1\x05\xfa\xb3\x18\x04\xc1\x06\x18\x94\x3e\x35\x3e\x0e\x2f\xaa\xef\xad\x0b\x11\xd2\x97\xb1\xdf\x34\xef\xe0\x0e\xb4\x50\xfb\x86\x67\x64\xaf\x58\x10\xba\xb6\x0d\x7c\xfd\x90\xab\xe7\x1d\x1e\x4b\x4f\x89\x6b\x2b\x9e\xb9\x37\xa1\x50\x8c\x33\x76\x04\xbe\x8f\x31\x2a\xb7\x82\xbc\x3c\x5c\xf4\x2d\xdc\x58\x6e\x81\x46\x9b\xa2\x3f\xb0\xd4\x1d\xf0\x33\x7f\xa9\x83\xdc\x45\x46\x61\x2c\xd7\x04\x78\x16\x52\x98\x23\x55\xe1\xbd\xea\xfb\x27\xde\xbd\x54\xb7\x7e\xcc\x43\xf0\x6d\xf2\x86\xc1\xe0\x56\x5f\x90\x76\x54\x46\x58\xa5\x4f\xf4\x68\x63\x75\x90\x91\x83\xb0\xdf\x75\x5f\x20\x77\x5a\x59\xd5\x29\x02\x6b\xc3\xef\xb4\x78\xe5\x16\xfe\x46\xbd\xd2\xbd\x98\x9e\x7a\xd1\x15\x78\x6e\xe8\x99\xa3\xfa\xfd\xda\x47\x3d\x27\x7d\x94\xf4\x4c\x97\xb0\xe8\xf3\x75\x93\x76\x97\xc7\x7d\xd1\xc7\xf8\x48\x0f\xee\xe2\x2e\xe1\x85\x5d\x35\xc0\xb6\xd5\xec\x5d\x1d\x84\xfd\xac\xb9\xec\xc8\x3d\xd8\x29\x69\x85\x9c\xd4\x64\x82\x9a\x48\xb8\x05\x92\xd9\xd6\xe9\x6b\xca\x54\x91\x6e\x97\x2e\xd3\xa2\x79\x17\x2e\x87\x8c\xbd\x11\x90\xd4\x30\x2a\xe9\x4d\x1f\xc5\x92\xf2\xb2\xe3\xdd\x11\x5c\x4c\x27\x5b\x29\xa2\xf8\x59\x37\x52\xf2\x59\x1c\x66\xa7\x6f\x49\x51\x55\x29\xd8\x4b\x96\xc4\x69\x85\x9c\x56\xf5\xb8\x10\x7c\xaa\x7d\x4d\xc6\xaa\xe1\x4b\x81\x56\x21\xe7\x07\x44\x12\x1c\x53\x51\x55\x89\xa3\xeb\x92\xcc\x65\xe7\xaf\x90\xb9\xd4\x18\x76\x90\x9d\x38\xc8\x18\x52\xcb\xe0\xb1\xf0\xae\x14\xbd\x34\x8d\x25\x55\x34\x94\x4d\x43\x27\x63\xcd\xe0\x79\xc6\x32\x98\x70\xce\xb9\xf2\x7f\x49\xe3\x34\xeb\x27\xa6\x4b\x8b\xb1\xae\x19\x14\xa9\xdf\xb6\x3d\x8f\x9e\xed\xdc\x05\x3f\xe7\xe4\x64\xef\xe1\x77\x94\x96\x04\xe8\x17\x16\x2f\x47\x1d\xe0\x97\x82\x01\x2e\x24\x9d\x95\x10\x72\x12\xd0\x3c\xf1\xe8\x4f\xe8\xd6\xf7\xc4\xfa\x90\x3d\xb8\x01\x31\x44\x4c\x81\xfb\xca\xb5\x86\xd1\xb5\x3d\xab\x03\x79\x24\x05\x9b\xca\xf3\x97\x89\xd2\x1d\xdc\xc3\xd3\x24\xfa\x4a\xb4\x94\x67\xe1\x4d\xe1\x96\x6f\xbd\xa9\xe6\xda\xef\xa8\x19\xda\x75\x41\x5c\xeb\x4e\xe9\x60\x66\xeb\xc7\x86\xfd\x2f\x4a\xd6\xaa\x07\xae\x2b\x2c\x10\xd7\x35\xc2\x0d\xf0\xde\x1e\xfd\xc0\x7f\xd2\xa8\x17\x1a\x9f\x2c\xd6\x0e\x5f\x94\xb4\x5c\x48\xd0\x5e\xb8\x96\x46\xb3\x94\xc1\x56\x94\x46\xfa\xc8\x9e\x94\x9a\x11\x61\xe6\x82\x34\x9e\xa0\x54\xd3\xc0\xff\xd8\x4d\x3a\x1a\x60\x04\xbe\x4b\xfe\xca\x45\x1f\xee\x15\x4a\x5d\xca\x56\x71\xfa\x3c\xd8\x1e\x69\x82\x5b\x9c\x34\x96\xc3\x17\x15\x47\x18\x30\xc1\xc8\x2d\x76\x7b\x21\x85\x15\xbc\xbf\x86\x9e\x9f\x76\xd0\x29\xb9\x37\x69\xea\xe8\x73\xe9\x02\xb4\x62\x00\x35\xd9\x02\x35\x53\xd7\x81\x31\xbf\x1e\x35\x98\xa3\x72\x46\x1d\xf0\x67\x2e\xfa\x49\x43\x85\x1f\xad\x1d\x6f\x80\xef\x41\x3b\x93\x43\x72\xdf\xe4\x17\xc9\xf8\x5a\xda\x29\xbe\xf2\x7a\x2a\xad\xbb\x28\xaf\xaa\x1a\xa6\x65\x0e\xf9\x37\x85\xda\xc9\xfe\x9f\x4a\x9a\xc5\x6a\x05\x17\xa4\x58\xbc\xf2\xd7\x82\xb7\xc5\xfb\x48\xed\xbc\x58\x13\x17\xaa\xce\xad\xfc\x7a\x1b\x6f\x95\xc8\xd5\x3d\xbc\x50\x32\x93\xbb\x6a\x31\x58\x2d\x95\xba\x34\xb3\xa4\x9a\x24\x5d\xe6\x8b\xf5\xd8\xcc\x3b\x17\xf7\xd7\x4e\x40\xcb\xec\x8d\xea\xa4\x99\x54\xb6\x6e\xe0\xb3\x92\x5c\x35\xaf\x95\x2b\x9c\x72\x15\xf8\xfb\x72\xca\x05\x51\x73\x6f\x70\xee\xac\x93\x2d\xb7\x9a\xee\x7e\xc7\xae\xce\x2e\x2e\x62\xd2\x33\xa4\x13\x63\x48\xb9\x74\xbd\x79\xb9\xca\x1e\x0f\x9a\xcb\xca\x7c\xea\x7e\x7d\x53\xff\x2d\x9f\x5c\x34\x85\x77\x2e\x14\x54\xb3\xb8\xd0\xac\xb9\x56\xce\x40\x55\xb7\xb0\xcd\xa0\xbe\xff\x06\x00\x00\xff\xff\x11\x04\xb4\x3f\x42\x24\x00\x00")

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

	info := bindataFileInfo{name: "plugins/codeamp/graphql/schema.graphql", size: 9282, mode: os.FileMode(420), modTime: time.Unix(1543962407, 0)}
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

