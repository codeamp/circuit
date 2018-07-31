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

var _pluginsCodeampGraphqlSchemaGraphql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x59\x4d\x73\xdb\xbc\x11\xbe\xeb\x57\x40\xf3\x5e\x94\x19\xff\x02\x1d\x63\xa5\xb6\xdb\x37\xad\x6a\x25\x27\x8f\x0f\x30\xb5\x92\x50\x93\x00\x03\x80\x8a\x35\x9d\xfc\xf7\x0e\x3e\x08\x2c\x3e\x28\x5b\x76\x3a\xf3\x5e\x6c\x62\x49\x3c\xfb\x81\xdd\xc5\xee\x4a\x35\xb4\xa5\x92\x7c\x63\x1d\xcc\xfc\xf3\xdf\x37\xff\xfa\xe7\x6c\xa6\x9a\x03\x74\x94\xfc\x77\x46\xc8\x8f\x01\xe4\x69\x49\xfe\x6d\xfe\xcd\x08\xe9\x06\x4d\x35\x13\x7c\x49\xbe\xfa\xa7\xd9\xaf\xd9\xec\x0f\xff\x5e\x9f\x7a\x70\x8f\x76\xef\x1f\xe4\xbb\x02\x39\x23\x64\x50\x20\x17\x6c\xbb\x24\x77\xab\x4f\xcb\x91\xe8\xde\x2a\xff\x5a\x2d\x3e\x2d\xc9\x83\xa1\x3c\xce\xed\xcb\xb5\x14\xff\x81\x46\xcf\x08\xe9\xdd\x93\x07\xb8\x22\xaa\x1d\xf6\x4b\xb2\xd1\x92\xf1\xfd\x15\xe1\xb4\x83\xb8\x02\x7e\x64\x52\xf0\x0e\xb8\xbe\x5b\x8d\xe4\x4f\x4b\x84\x16\x90\x55\x84\x56\x0b\xff\xb0\x01\x2a\x9b\x43\xf8\xdc\x2d\xef\x78\x3f\xe8\x2b\xd2\x53\x49\x3b\xb5\x24\x6b\xba\x67\x9c\x6a\x21\x2d\x3d\x62\xff\xc9\x94\x76\xa2\xff\x0d\xa8\x1e\x24\x18\x06\x3b\xff\xb8\x98\xdc\xed\x3f\x8e\xbb\x37\x20\x8f\xac\xb1\xbb\x95\x7f\x9c\xde\xed\x3f\x2e\x76\x13\xd5\x43\x83\x20\x36\x66\x69\x4d\xbc\x89\x04\x6f\xe9\x7b\x68\x81\x2a\xcb\x50\xfa\xc7\x69\x86\xfe\xe3\xc8\xf0\x4b\xb4\xb8\x41\x40\x07\x10\xad\x8a\x0e\xcc\x88\x80\xb6\x3c\x16\x20\xe4\x48\x25\xa3\x4f\xad\x37\x40\x23\x41\x9f\xd5\xdf\x7c\x10\xa5\xf1\x67\xf1\xe5\x45\x03\x57\x4c\x70\x6b\x07\x23\xd6\x48\x50\x8b\x29\x17\x79\x08\x9b\x1e\xeb\x58\xc8\x61\x22\xcd\xda\x34\xff\x32\x35\x2c\x62\x1e\x4d\x9c\x21\xdc\x67\xd4\x51\x04\x90\x1d\x53\x81\x79\x5c\x99\x4d\x26\x56\xe7\x2e\xfc\x42\x30\xda\x08\x1c\x57\x3e\x08\xaf\x25\x50\x0d\xa3\xe8\x33\x42\x1a\x4b\xf0\x42\x8f\x87\x14\xfc\x38\x73\x6b\x17\xa9\xfd\x36\x85\x18\x2c\xe1\x12\x08\x2f\x85\x57\x3f\x48\xe1\x15\x5f\x78\x7a\x70\xb0\xcc\xdf\x9c\x6f\x6b\xd1\x23\x00\xa5\x45\x3f\x6e\x77\xb9\x61\x9e\x6d\xf0\x3c\x7d\x0c\x04\x9e\x3e\x04\x16\x9e\x1e\xa2\x28\x0b\x2a\xac\x79\x24\x39\xcd\x2f\x81\x58\x41\x0b\x89\x14\x5b\x4b\xb8\x04\x22\x55\x64\x74\xea\x44\x1b\x13\xd0\x0b\x14\xed\x01\xc0\x2c\x32\xcc\x8d\xdb\x1f\xb4\xcb\x70\x13\x15\xdf\x85\x9b\xaa\x3c\xe2\x26\x7a\xbf\x0b\xd7\xdb\x01\x85\x70\x30\x03\x4a\x21\x38\xc4\x97\x38\xb7\x8c\xb0\x5f\x92\xfd\xc1\x0c\x29\xac\xb3\xc2\x47\x60\xbd\x15\x52\x58\x67\x84\x8f\xc0\x96\x46\x08\x39\x13\x39\x85\xc9\x8b\x0b\x97\x3f\xc7\x34\x99\x26\xcd\x09\xcd\x31\xd6\xe8\x08\x6f\xc3\x2a\xd5\x0d\x58\x04\x9d\xfe\xdb\xc0\x46\x25\xf3\x44\xee\x0f\x7b\x24\x2f\xc2\x07\x4b\x12\x88\xc1\x70\x23\x21\x51\x35\x47\xf4\xe7\xfc\x0e\xc4\x51\xe1\x1c\xd1\x1f\xf1\x3b\x10\x73\xad\xf3\x64\x1d\x31\xf3\x8b\x68\x59\x5c\x57\x59\x12\x3e\x6f\x8c\x3c\xa5\xff\x36\x46\xc8\x46\x9e\xe6\xac\xf3\x7f\x52\xc8\x94\x92\xf8\x9e\x0c\x7a\x99\xca\x12\xdd\xa6\x8b\x21\x5d\xbb\xaa\x14\x11\x46\x66\x0f\xae\x3e\x98\xfb\x0b\x39\xbd\x08\x93\x6a\xa7\x30\x61\xa5\x12\xc2\xb4\xa8\x08\x22\x06\xb6\x88\x38\xf2\xfe\x2c\xc4\x73\x47\xe5\x33\xba\x86\x9f\x3c\x69\x9d\x94\xc8\xe6\x1a\xfc\x2c\x44\x0b\x94\xbb\x9d\x37\xa0\xc9\x0d\xd3\xe4\x5a\x74\x1d\xb3\x92\xee\x41\xdf\x30\xed\xd7\xa3\x74\xa6\x1c\xba\x5b\xcd\x8b\x2a\xda\xd2\x38\xfc\x0c\xa8\x18\xdf\xd6\x1f\xa1\x2e\x9b\x31\xae\x41\xee\x68\x03\x91\x66\xcb\x90\x9e\xee\x61\x49\xee\xb8\xad\xd3\x1a\x31\x98\x5c\xe7\x57\x1c\x5e\xf4\xf5\x20\x95\x90\x63\x35\xe6\x51\x51\xad\xe9\x0a\x1b\x44\x20\xac\xeb\x5b\xb0\x36\x7b\x3b\x27\xe0\x5a\x32\x50\xb1\xde\x7a\x3c\xcb\x3f\x16\x97\x8e\x7d\x5c\x7f\x90\xbb\x03\x7a\x8d\x79\xa8\xec\x47\xee\x81\xf0\x61\xf6\x16\xe9\x3c\x7f\xd4\x97\x38\xfe\x88\xf0\x41\xfe\x1e\xe9\x3c\x7f\xd4\x55\x39\xfe\x88\xf0\x41\xfe\x1e\xe9\x3c\x7f\xdb\xa5\x5a\xc6\xe6\xc9\x82\xfb\xe8\x32\x58\x1d\x65\x6d\xdc\x90\x94\xe6\x59\xce\x70\xf9\x7b\xbb\xb4\x6d\x76\x6a\xdb\xc4\xae\x19\x87\xbc\xa4\x36\xb4\x0e\x94\xb2\xfa\x45\xbe\x26\x8d\xe1\xf5\x81\xaa\x43\x22\x17\x95\xc0\xf5\x6d\x46\x95\xb0\xc3\xcb\x9a\x88\x63\xf5\x89\x5d\xef\x0d\x22\x36\xa2\xeb\x28\xdf\x62\x74\xdc\xa0\xcf\xd3\x8e\x34\x29\xf0\xca\x13\xeb\x85\x34\x69\xf2\xc1\x34\x3a\x8f\xf3\xb4\xb9\x4c\xca\x22\xf3\xce\x08\x7a\x46\x29\x73\xef\xf4\xad\x38\x99\xcf\x37\x5a\x52\x0d\xfb\x93\x6b\xa1\x66\x84\xb4\xec\x08\x1c\x94\x5a\x4b\xf1\x04\x81\x2a\x81\x6e\x59\x49\xee\x25\x98\x46\xe4\x56\x88\xe7\x91\x9f\x33\x19\xae\xd3\xac\xd9\x70\x57\x9b\x9a\x2e\xb7\xc9\x33\x9c\xf0\x92\xa9\x15\xec\xe8\xd0\xea\x24\x8f\x37\xa2\x15\xf2\xac\x8a\xe3\x38\x23\xf1\xf1\x5c\xb6\x58\xe0\xa1\xa4\x96\xc9\x97\xc9\x73\xa4\xed\x90\x9e\x61\x23\x52\x6b\xd7\x7c\xc1\xf9\xa6\x89\x9e\xda\xf9\x1c\x41\x86\x70\x89\xe9\xf0\xdc\x09\x17\xea\x32\xb5\xf1\x65\x64\x7a\x1b\xe1\x19\x48\xe2\xc0\xc6\xcd\x5e\x39\x89\xa6\x1f\xee\xe1\xc7\x00\x4a\x67\xd4\x3f\x59\xc7\x12\x5a\x07\x9d\x90\xa7\xca\xc7\xee\x45\xf1\xbd\x36\x29\x82\xdb\xc6\xfc\x46\xd2\x06\xd6\x20\x99\xd8\xbe\x16\x87\x63\x3b\x8b\xef\xbf\x37\xc4\x61\x6a\x7b\x2a\x35\xdb\x51\xeb\x18\x6e\x68\x40\xc8\x01\xe8\xd6\x67\x9e\x30\x83\xb2\x52\x52\xd6\xd6\xe8\x4a\x53\x0d\x69\x0e\xc9\xe6\x18\x53\x53\x0c\xbb\xf3\x6b\x99\xba\x2e\x3a\xea\x1d\xe3\x4c\x1d\xd2\xc3\xbf\x17\x6d\xfb\x44\x9b\xe7\xa2\x18\xf1\xe5\x11\xbe\x35\x5e\x39\x75\x3c\x52\x74\xca\xf5\x42\x31\x2d\xe4\x29\xcd\x5c\xbe\x69\x09\x94\x3d\xd3\xdf\x65\x9b\x51\xd6\x52\x68\xd1\x88\x84\x2c\x15\x5d\x4b\x76\xa4\x1a\xfe\x91\x86\x96\x79\x31\x3c\xb5\xac\xc9\xe8\x61\x70\xa8\x0e\xe2\xe7\xca\xa6\x2e\xa3\xbd\xd7\xf4\xcc\x34\x32\x9b\x27\x36\x83\x34\x37\xc0\x7d\x36\x5e\x79\xcf\xac\xef\xe2\x69\xe4\x85\xd3\x3b\xc0\xae\x54\x1d\xa9\xed\x99\xfe\x2c\x29\x6f\x92\xcb\xac\x11\x5c\x33\x3e\x88\x41\x39\x33\x25\x39\x13\x92\x82\xbb\xac\xaa\xc7\x02\x1a\xd9\x76\x2a\x12\xb3\x99\xa2\xcb\xf0\x81\xf6\x4a\x56\x11\x5d\x2f\xb8\xf5\x74\x94\x10\xb2\x8c\x98\x25\xdd\xb3\xf1\x21\xf8\x8e\xed\x63\x34\xd7\xe4\x2d\xda\x24\x1c\x0f\x53\x82\xd7\x72\x49\xad\x7d\x9d\xc8\x2a\x85\x5c\x83\xd2\xa2\xbb\xce\xa8\x45\x2e\xf9\x0d\x29\x02\x27\x4b\xd4\xd6\xe2\xb4\x39\xa5\x73\x3e\x7a\xcc\x74\xce\x2d\x86\x43\x61\xc3\xf6\xdc\xe7\xca\x3c\x4d\x4c\xbc\xcb\x55\xcf\x9d\x60\xca\x14\x15\x63\x97\x39\xf1\x8c\x1b\x10\x66\x62\x6e\x66\xff\x26\x63\xda\x60\x0c\x5f\xca\x4c\xa6\xb0\x6a\xae\x4b\xc3\xb1\x1e\x50\xe9\x41\xc6\x99\xfb\xb9\xd8\x4d\x64\x77\x3f\xc3\xd4\x34\x40\x3f\xd0\x58\x3d\xca\xbc\x3d\x15\xe3\xb8\x6d\x35\x2e\x91\x80\x27\x79\xca\x02\x37\x49\x8b\x60\x4b\x46\x7b\xbd\xdf\x71\x9d\xba\x1e\xc6\xc1\xa3\xec\xd2\xcc\xe8\x06\x8e\x36\x41\x31\x98\x12\xab\x16\xb4\x5e\x20\x64\x03\xf7\xf0\x34\xb0\xb6\xd0\x6f\x2c\x84\xb0\x50\x78\xca\x5c\x0a\x55\xe5\x7d\x59\x51\x9f\xef\xad\x55\xf4\x5e\x88\xb5\x90\xce\x09\xe7\x8f\x95\x50\x98\x54\xb9\x56\xc9\xaf\x0a\x9a\x03\x2e\xab\xfb\x5b\xa0\xad\x3e\xd8\x85\xfd\xa4\x52\xe9\x57\x3e\x99\xac\xfa\xaf\x05\xd7\x94\x71\x90\x56\xb9\x9a\xa9\x83\x96\xae\x43\x15\x12\xd9\x23\xc4\x99\x25\xfc\x9a\xf9\x9d\x13\xda\x58\x80\xdc\x4c\x1d\x7d\xd9\x0c\x12\xb5\xbd\x1d\x7d\xf9\xce\xe9\x91\xb2\xd6\x54\xf9\x39\x74\xae\x5b\x81\x69\x2b\x58\x7d\xc8\x4a\xd3\xd4\x05\xb0\x1e\xb6\x1d\x38\x40\x87\x01\x7a\xaa\x71\x52\x60\x9c\x69\x46\xdb\x15\xb4\xf4\xb4\x81\x46\xf0\xad\x1a\xb7\xf6\xb6\x0a\xce\x88\x9a\x75\x20\x06\x9d\x51\xd5\xd0\x34\xa0\xd4\xb7\x83\x04\x75\x10\xc6\xdb\x1d\x7d\x47\x59\x3b\x48\x28\xe8\x07\xad\xfb\x5b\xa0\x5b\x90\xc6\xe5\x90\xde\xb7\xe1\xc5\xe8\x7c\x35\xeb\x64\x5f\x59\x3b\xe5\x6e\x9f\x35\x46\x45\xf7\x51\x73\x87\xf0\xfb\x46\x19\x7d\x7f\x9d\x66\xa4\x68\x1a\xb1\x22\xf9\x6f\x14\xaf\x2b\xf2\x9e\xfe\x76\xb2\x6f\xcd\x8c\x1a\x7e\x40\x28\xc5\x78\xad\x8d\x2d\x2e\xdf\x89\xb6\x36\xb9\xb3\x26\xd3\xd2\x54\x3b\x5a\xfd\x4d\x3a\x35\x68\x32\xe2\xbe\xd8\x9c\xd5\xea\x72\x52\xcc\xbc\x46\x4b\x8d\x50\x97\xb8\x72\xf5\x9e\x95\xf9\xaa\x7a\x95\x5c\xe1\xc2\x2a\xa3\xbf\xad\x72\x9c\x50\x2a\x0c\xec\xe2\x00\x3f\x11\xb9\x36\xdb\xb7\x12\x9b\x36\x39\xbb\x7c\x93\x41\x5e\xba\xd1\x67\x8b\x4b\xf9\x45\x76\x85\x03\xee\x25\xe5\x93\xfe\x32\x11\x7d\xe5\xeb\xc8\x60\xf2\xd0\xdf\xc8\xc8\x99\x66\x92\x51\xb4\x5c\xad\x4e\x48\x4d\x37\x21\xa6\x33\xdf\xff\x02\x00\x00\xff\xff\x02\x86\xcc\x9d\x7a\x24\x00\x00")

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

	info := bindataFileInfo{name: "plugins/codeamp/graphql/schema.graphql", size: 9338, mode: os.FileMode(420), modTime: time.Unix(1533323165, 0)}
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

	info := bindataFileInfo{name: "plugins/codeamp/graphql/static/index.html", size: 2810, mode: os.FileMode(420), modTime: time.Unix(1532982390, 0)}
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

