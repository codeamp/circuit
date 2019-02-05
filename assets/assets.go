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

var _pluginsCodeampGraphqlSchemaGraphql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x59\x4f\x73\x1b\xbb\x0d\xbf\xeb\x53\xd0\xf3\x2e\xca\x8c\x3f\x81\x6e\x4d\xec\xc6\x6e\x93\xd6\xb5\x5e\x0e\x9d\x37\x3e\xd0\x2b\x58\x62\xbd\x4b\xee\x23\xb9\x4e\x34\x9d\x7e\xf7\x0e\xff\x2e\x40\x72\x65\xcb\x2f\x9d\xe9\xc5\x16\xb1\xe4\x8f\x00\x08\x80\x00\x68\x3a\xde\x73\xcd\x7e\x15\x03\xac\xe2\xef\xbf\x6c\xff\xfe\xb7\xd5\xca\x74\x07\x18\x38\xfb\xf7\x8a\xb1\xdf\x27\xd0\xc7\x0d\xfb\x87\xfb\xb7\x62\x6c\x98\x2c\xb7\x42\xc9\x0d\xfb\x1a\x7f\xad\xfe\xb3\x5a\xfd\x12\xbf\xdb\xe3\x08\xe1\xa7\x5f\xfb\x0b\xfb\x66\x40\xaf\x18\x9b\x0c\xe8\xb5\xd8\x6d\xd8\xed\xd5\x87\x4d\x22\x86\xaf\x26\x7e\x36\xeb\x0f\x1b\xf6\x9b\xa3\x3c\x5c\xf8\x8f\x77\x5a\xfd\x0b\x3a\xbb\x62\x6c\x0c\xbf\x22\xc0\x25\x33\xfd\xb4\xdf\xb0\xad\xd5\x42\xee\x2f\x99\xe4\x03\xcc\x23\x90\x2f\x42\x2b\x39\x80\xb4\xb7\x57\x89\xfc\x61\x83\xd0\x32\xb2\x99\xa1\xcd\x3a\xfe\xd8\x02\xd7\xdd\x21\x4f\x0f\xc3\x5b\x39\x4e\xf6\x92\x8d\x5c\xf3\xc1\x6c\xd8\x1d\xdf\x0b\xc9\xad\xd2\x9e\x3e\x63\x7f\x11\xc6\x06\xd6\xff\x0c\xdc\x4e\x1a\xdc\x06\x4f\xf1\xe7\x7a\x71\x75\x9c\x3c\xaf\xde\x82\x7e\x11\x9d\x5f\x6d\xe2\xcf\xe5\xd5\x71\x72\xb5\x9a\x99\x11\x3a\x04\xb1\x75\x43\xaf\xe2\xed\x4c\x88\x9a\xbe\x87\x1e\xb8\xf1\x1b\xea\xf8\x73\x79\xc3\x38\x79\xde\xf0\x7a\xd6\xb8\x43\x40\x07\x30\x6b\x15\x1d\x98\x63\x01\x2d\x79\xa8\x40\xd8\x0b\xd7\x82\x3f\xf6\x51\x01\x9d\x06\x7b\x52\x7e\x37\x01\x89\x2f\xe4\xbe\x87\x48\xce\x00\xde\x74\xf2\xfe\xf9\x63\x36\x85\xeb\x1f\x16\xa4\x11\x4a\x7a\xb5\x39\x29\x12\xc1\xac\x97\x2c\xea\xb7\xbc\x88\x1a\x6c\x26\x23\xfb\x9a\x69\xfe\x08\xca\x99\xf4\x1c\xd0\xe6\xf3\x89\x14\x08\xf7\x05\x35\xb1\x00\x7a\x10\x26\x6f\x3e\x8f\xdc\x22\xe7\xda\x17\xc1\x5b\xb3\xef\x7a\x87\x4d\xa3\xe8\xb3\x9f\x34\x70\x0b\x89\xf5\x15\x63\x9d\x27\x44\xa6\xd3\x99\x66\xb3\x2f\xbc\x20\x38\xf6\xb8\xa3\x10\x93\x27\x9c\x03\x11\xb9\x88\xe2\x67\x2e\xa2\xe0\xeb\x48\xcf\xf6\x58\x98\x67\xb0\x05\xab\x46\x04\x60\xac\x1a\xd3\xf2\x10\x4a\x2e\x8a\x05\x71\xcf\xe8\x32\x79\xcf\xe8\x31\xeb\x48\xcf\x4e\x57\xf8\x20\x96\x7c\x26\x05\xc9\xcf\x81\xb8\x82\x1e\x08\x17\x3b\x4f\x38\x07\xe2\x76\x18\x95\xb6\x6c\xe0\xf2\x98\x63\x08\xe3\x96\x29\xe9\x27\x08\xff\x39\xc5\x99\x84\x68\x36\x71\x5d\xfa\x90\x90\x53\xc8\xb8\x88\x56\x46\xd5\x94\x5c\x86\xe8\xca\x45\x97\x35\x0a\x3d\x99\x3d\x37\x28\x38\xde\x86\xf5\x59\x77\x05\x2e\x51\xe0\xbb\x70\xa9\x42\x13\x2e\xd1\xea\xbb\x70\xa3\x1e\x50\x80\xc8\x6a\x40\xf1\x0c\x07\x90\x0d\x0e\x74\x09\xf6\x9a\xac\xcf\x6a\xa0\xb0\x41\x0b\x7f\x04\x36\x6a\x81\xc2\x06\x25\xfc\x11\xd8\x5a\x09\x39\x80\x23\xa3\xf0\x41\x38\xc4\xe2\x14\x7f\x69\x04\x5f\x90\x1c\x63\x25\x43\x78\x1b\x56\x2d\x6e\xc6\x62\xe8\xf4\xdf\x06\x86\xfd\xa9\x85\xd8\x72\xae\x70\x71\xc5\x0b\x6c\x76\x2d\x3f\x44\x9e\xe5\xc6\xc9\xb1\xae\x7f\xf8\x6d\x78\xdf\x2f\xec\xf2\x5d\xd8\x83\x90\x6c\x2f\x5e\x40\xa6\xe8\x7a\x7b\xc5\xb8\xdc\xd1\xc4\xc7\xdf\x60\x98\x8d\x74\x7f\x5e\xff\x68\x70\x11\xae\x34\xe2\xdb\x50\x5e\x88\xd1\xac\x13\x79\x9d\x27\x38\xcc\xf8\x33\x9b\x48\x22\x90\x43\x2d\x11\xa3\x45\xbf\x03\x31\x1d\x6d\x89\x18\x8d\xf9\x1d\x88\xa5\xd4\xe5\xa5\x37\x63\x96\x17\xfa\xa6\xba\xf6\x8b\xcb\xec\xb4\x32\xca\xab\xf1\xa7\x6d\x84\x74\x14\x69\x41\x3b\xff\x23\x81\x5c\x06\x8f\xf3\x8d\x2c\x97\x4b\xe8\x51\x56\xb2\x9e\xe8\x38\x14\x03\x88\x30\x7b\x46\x30\xca\xe8\x19\x34\xa1\x20\x49\x66\xa5\xc2\x46\x02\x8a\x69\xb3\x20\x88\x98\xb7\x45\xc4\xb4\xf7\x47\xa5\x9e\x07\xae\x9f\x51\x3a\xf3\x18\x49\x77\xa4\x32\x71\xe9\xc4\x47\xa5\x7a\xe0\x32\xac\xfc\x0c\x96\x7d\x16\x96\x7d\x52\xc3\x20\x3c\xa7\x7b\xb0\x9f\x85\x8d\xe3\x75\xf6\x60\xbf\xba\x2a\x5e\x3c\x4d\xc2\xf7\x8c\x8a\xf1\x7d\x1e\x97\xd3\xe1\x95\x90\x16\xf4\x13\xef\x60\xa6\xf9\x74\xae\x53\x93\x0b\xe4\xb7\xd2\xc6\x25\x28\x7f\x0f\xd9\x1f\x22\xb8\xd8\xd5\x83\x57\xc8\x09\x18\x97\xe3\x5b\x2d\x5c\xbe\x90\xf2\xd0\x87\x08\x3e\x67\xe3\x01\x7b\x1e\x9f\x0f\x1d\xd6\xce\xc8\xb9\xce\x49\xd0\x99\xf0\x1e\x6c\xbf\x38\x81\xa3\x12\x2c\x80\x23\xc2\xf9\xe0\x71\x71\x02\x47\xd5\x61\x00\x47\x84\xf3\xc1\xe3\xe2\x04\xee\x4b\x69\x8f\xea\x7e\xf9\x95\xd1\x16\xdd\xaa\x81\x8b\x1e\x87\xf7\x11\x3b\x1e\xf1\xb0\x10\xed\x76\x1b\xdf\x0b\xa0\x5a\x21\x1a\x29\x76\x28\x13\x79\x47\x1b\xc0\x18\xbe\x07\xbc\xaf\x73\x7a\x3c\x3e\x70\x73\x20\x7c\x71\x0d\xd2\xde\x14\x54\x0d\x4f\x78\xd8\x62\x31\xe5\xbc\xd8\x22\xde\xc0\x62\xa7\x86\x81\xcb\x1d\x46\xc7\x5d\x84\x0b\x5a\x36\x93\xc4\xef\xa2\x3a\x1b\x77\xa1\x3a\x75\xba\xf2\xea\xe1\x82\x56\xc0\x24\x5d\x72\xdf\x1c\xa3\x27\x84\x72\x51\x7a\xec\xd5\xd1\x4d\xdf\x5a\xcd\x2d\xec\x8f\xa1\x70\x5b\x31\xd6\xbb\x5b\x1f\x8c\xb9\xd3\xea\x11\x32\x55\x03\xdf\x89\x9a\x3c\x6a\x70\xe5\xcf\x8d\x52\xcf\x69\xbf\xa0\x32\x9c\xbf\x79\xb5\xe1\xd2\x9b\xaa\xae\xd4\xc9\x33\x1c\xf1\x50\x98\x2b\x78\xe2\x53\x6f\x49\xd4\xeb\x54\xaf\xf4\x49\x11\x53\xcf\xa5\xb6\xe6\x56\x17\x00\x07\x92\x82\xbf\x82\x9f\x17\xde\x4f\xf4\x0c\x3b\x45\xb5\xdd\xb2\x85\x60\x9b\xce\x7b\x5a\xe7\xf3\x02\x3a\xbb\x4b\x8a\x47\xa7\x4f\xb8\x12\x57\x98\x6d\x4c\x2f\x69\xec\xc6\x8d\x1a\x62\xc0\xce\xcc\x5e\x39\x89\x6e\x9c\xee\xe1\xf7\x09\x8c\x2d\xa8\x5f\xc4\x20\x08\x6d\x80\x41\xe9\x63\x63\x72\xf8\x50\xcd\xb7\x2e\x44\x48\xdf\x0e\xf8\xac\x79\x07\x77\xa0\x85\xda\x35\x3c\x23\x7b\xc5\x82\xd0\xb5\x6d\xe0\xeb\x87\x5c\x3d\x6f\xf0\x58\x7a\x4a\x5c\x5b\xf1\xc4\xbd\x09\x85\xa6\x06\x63\x07\xe0\xbb\x18\xa3\x72\x4b\xcd\xcb\xc3\x45\xdf\xa2\x1b\xcb\x2d\xd0\x68\x53\xf4\x59\x96\xba\x2c\x7e\xe5\xd7\x3a\xc8\x9d\x65\x14\xc6\x72\x4d\x08\x4f\x42\x0a\x73\xa0\x2a\xbc\x57\x7d\xff\xc8\xbb\xe7\xea\xd6\x8f\x79\x08\xbe\x4d\x5e\x31\x18\xdc\x32\x0d\xd2\x8e\xca\x08\xab\xf4\x91\x1e\x6d\xac\x83\x32\x65\x2f\xec\x37\xdd\x17\x94\x3b\xad\xac\xea\x14\x21\x6b\xc3\xef\xb4\x78\xe1\x16\xfe\x4a\xbd\xd2\x7d\x98\x1e\x7b\xd1\x15\xf4\xdc\x18\x35\x07\xf5\xfd\xca\x47\x3d\x27\x7d\x94\xf4\x44\xb7\xb5\xe8\x97\x76\x93\x76\x97\xc7\x7d\xd1\x0f\x7a\x4f\x2f\xf3\x95\x6e\xeb\x25\x33\xbe\x1d\x8c\x04\xa9\x1b\xb0\xa7\x1b\x96\x4b\x10\xb8\x87\x09\xd8\x02\x9b\x9d\xc2\xbd\xb0\x1f\x35\x97\x1d\xb9\x2d\x3b\x25\xad\x90\x93\x9a\x4c\x50\x26\x09\xca\x40\xf2\xdf\x3a\xc9\x4d\xf9\x2c\x3a\x81\xa5\x2b\xb7\x68\x95\x86\x2b\x24\xd3\x5e\x09\x5b\x6a\x18\x95\xf4\x0e\x82\x22\x4e\x79\x25\xf2\xee\x00\x2e\xf2\x13\x56\x8a\x58\x7f\xd2\xd9\x94\x7c\x12\xfb\x39\x34\xb4\xa4\xa8\x6a\x19\xec\x4b\x4b\xe2\xb4\x02\x53\xab\xc6\x5c\x08\x51\x15\x5f\x93\xb1\x6a\xf8\x54\x50\xab\xc0\xf4\x13\xe2\x0d\x8e\xbc\xa8\xf6\xc4\x31\x78\x49\xe6\xb2\xcf\x5a\xc8\x5c\x6a\x0c\xbb\xd1\x56\xec\x65\x0c\xbc\x65\x88\x59\xf8\x56\x8a\x5e\x9a\xc6\x92\x2a\x1a\xca\xa6\x01\x96\xb1\x66\x88\x3d\x61\x19\x4c\x38\x7f\x5d\xf9\xbf\xa4\x4d\x9d\xf5\x13\x93\xaa\xc5\x88\xd8\x0c\x9d\xd4\x6f\xdb\x9e\x47\xcf\x76\x7e\x73\x38\xe5\xe4\x84\xf7\xf0\x6a\xd5\x92\x00\xbd\x67\x79\x39\xea\x6b\x60\x29\x18\xe0\x72\xd3\x59\x09\x01\x27\x31\xce\x03\x8f\xfe\x84\x6e\x7d\x8f\xb0\x0f\x39\x86\x1b\x10\x43\xc4\x10\xb8\x8b\x5f\x6b\x18\x5d\xee\xb3\x3a\x90\x47\x52\x62\x53\x79\xfe\xca\x51\xba\x83\x7b\x78\x9c\x44\x5f\x89\x96\xb2\x31\xcc\x14\x6e\xb0\xd7\x4c\x35\xf7\x7e\x43\x65\xd1\xae\x1e\xe2\x5e\x77\x4a\x07\x33\xbb\x78\x68\xd8\xff\xa2\x64\xad\xaa\xe1\xaa\xa2\x05\xe0\xba\x92\xb8\x01\xde\xdb\x83\x1f\xf8\x29\x8d\xaa\xa2\x31\x65\xb1\xc2\x48\xcf\x07\xb1\x13\x4a\x54\xda\x78\x60\xf0\x9a\x4d\x31\xe3\x9f\x7f\xfa\xfa\x25\x60\xbd\xf7\x98\x3d\x0b\x9f\x94\xb4\x5c\x48\xd0\xac\xe2\xa0\x54\x74\x30\x57\xa5\xd1\x91\x64\x67\x4e\x5d\x93\xb0\x72\x41\xa1\x1e\xa0\x3c\xa9\x81\xff\xd8\x4e\x3a\xfa\x40\x24\x7c\x93\xfc\x85\x8b\x3e\x5c\x6d\x14\xba\x54\x6f\x85\xe9\x13\x76\x7b\xa0\x99\x78\x61\x6c\x58\x0e\x5f\xfd\x1c\x60\xc0\x00\x23\xb7\x38\xf2\x08\x29\xac\xe0\xfd\x15\xf4\xfc\xb8\x85\x4e\xc9\x9d\x49\x4b\x47\x9f\xf4\x17\x44\x2b\x06\x50\x93\x2d\xa8\x66\xea\x3a\x30\xe6\xd7\x83\x06\x73\x50\xce\xaf\x02\xfd\x89\x8b\x7e\xd2\x50\xd1\x0f\xd6\x8e\x37\xc0\x77\xa0\x9d\xd5\x23\xb9\x6f\xf2\x87\x64\xff\x2d\xed\x14\xb3\xbc\x9e\x4a\x07\x2b\xea\xc0\xaa\xd8\x6a\x99\x43\x7e\xe6\xa9\xfd\xfc\xff\xa9\xf6\x5a\x2c\xab\x70\xe5\x8c\xc5\x2b\x1f\x70\x5e\x17\xef\x3d\x45\xfe\x62\xf1\x5e\xa8\x3a\xbf\xae\xd4\x6c\xbc\x56\xcb\x57\xa9\xc0\x42\x6d\x4f\xae\xcb\xc5\x78\xb9\x54\x93\xb7\x65\x58\x0e\x63\xf3\x33\x4a\x8c\x62\x9e\xf0\x93\x82\x58\x93\x99\xf0\xac\x43\xcf\xb8\x7a\xd3\x09\x31\xed\xec\xfd\x48\x66\x5f\x6c\x81\xdf\x02\xce\x36\xa2\x66\xde\xbf\x78\x38\xed\x02\xa0\xcc\x9e\xa9\x41\x34\x93\xfa\x56\x06\x74\x52\x92\xcb\xa6\xda\x2e\x71\xca\x5b\xd0\xdf\x96\xd3\x9f\xd2\x3a\x7d\xff\x20\x2c\xb7\x9e\x46\x3c\xc7\x93\x01\x5d\x24\x42\xa4\xb3\x4b\x17\xc6\x78\x7a\xee\x7e\xf3\x76\x95\x33\xee\x35\x97\x95\xef\xd4\xaf\x2a\x4d\xfd\xb7\x02\xd2\xa2\x29\xbc\x71\xa3\xa0\x9a\xc5\x8d\x66\xcd\xb5\x7c\x82\xaa\x6e\x81\xcd\xa0\xbe\xff\x06\x00\x00\xff\xff\x8a\xde\x85\x8f\x30\x27\x00\x00")

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

	info := bindataFileInfo{name: "plugins/codeamp/graphql/schema.graphql", size: 10032, mode: os.FileMode(420), modTime: time.Unix(1549347628, 0)}
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

