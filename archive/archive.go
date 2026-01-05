/*
 Copyright 2023 NanaFS Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/basenana/plugin/api"
	"github.com/basenana/plugin/types"
)

const (
	pluginName    = "archive"
	pluginVersion = "1.0"
)

var PluginSpec = types.PluginSpec{
	Name:    pluginName,
	Version: pluginVersion,
	Type:    types.TypeProcess,
}

type ArchivePlugin struct{}

func (p *ArchivePlugin) Name() string {
	return pluginName
}

func (p *ArchivePlugin) Type() types.PluginType {
	return types.TypeProcess
}

func (p *ArchivePlugin) Version() string {
	return pluginVersion
}

func (p *ArchivePlugin) Run(ctx context.Context, request *api.Request) (*api.Response, error) {
	filePath := api.GetParameter("file_path", request, "")
	format := api.GetParameter("format", request, "")
	destPath := api.GetParameter("dest_path", request, "")

	if filePath == "" {
		return api.NewFailedResponse("file_path is required"), nil
	}

	if format == "" {
		return api.NewFailedResponse("format is required"), nil
	}

	if destPath == "" {
		destPath = "."
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return api.NewFailedResponse(fmt.Sprintf("create dest directory failed: %w", err)), nil
	}

	var err error
	switch format {
	case "zip":
		err = extractZip(filePath, destPath)
	case "tar":
		err = extractTar(filePath, destPath)
	case "gzip":
		err = extractGzip(filePath, destPath)
	default:
		return api.NewFailedResponse(fmt.Sprintf("unsupported format: %s (supported: zip, tar, gzip)", format)), nil
	}

	if err != nil {
		return api.NewFailedResponse(err.Error()), nil
	}

	return api.NewResponse(), nil
}

func extractZip(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("open zip file failed: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(dest, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.Mode()); err != nil {
				return fmt.Errorf("create directory failed: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("create parent directory failed: %w", err)
		}

		destFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("create file failed: %w", err)
		}

		srcFile, err := file.Open()
		if err != nil {
			destFile.Close()
			return fmt.Errorf("open zip entry failed: %w", err)
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()

		if err != nil {
			return fmt.Errorf("extract file failed: %w", err)
		}
	}

	return nil
}

func extractTar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open tar file failed: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("create gzip reader failed: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar header failed: %w", err)
		}

		path := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("create directory failed: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return fmt.Errorf("create parent directory failed: %w", err)
			}

			destFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("create file failed: %w", err)
			}

			_, err = io.Copy(destFile, tarReader)
			destFile.Close()

			if err != nil {
				return fmt.Errorf("extract file failed: %w", err)
			}
		}
	}

	return nil
}

func extractGzip(src, dest string) error {
	// For gzip, we extract to the same directory with the .gz extension removed
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open gzip file failed: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("create gzip reader failed: %w", err)
	}
	defer gzipReader.Close()

	// Determine output filename (remove .gz extension)
	outputName := src
	if len(outputName) > 3 && outputName[len(outputName)-3:] == ".gz" {
		outputName = outputName[:len(outputName)-3]
	} else if len(outputName) > 7 && outputName[len(outputName)-7:] == ".tgz" {
		outputName = outputName[:len(outputName)-3] + "tar"
	}

	outputPath := filepath.Join(dest, filepath.Base(outputName))

	destFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("create output file failed: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, gzipReader)
	if err != nil {
		return fmt.Errorf("extract gzip failed: %w", err)
	}

	return nil
}

func NewArchivePlugin() *ArchivePlugin {
	return &ArchivePlugin{}
}
