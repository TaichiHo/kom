package kom

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/weibaohui/kom/utils"
	"k8s.io/klog/v2"
)

// FileInfo represents a file node structure
type FileInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // file or directory
	Permissions string `json:"permissions"`
	Owner       string `json:"owner"`
	Group       string `json:"group"`
	Size        int64  `json:"size"`
	ModTime     string `json:"modTime"`
	Path        string `json:"path"`  // Storage path
	IsDir       bool   `json:"isDir"` // Indicates whether it's a directory
}

// ListFiles gets the list of files and directories at the specified path in the container
func (p *pod) ListFiles(path string) ([]*FileInfo, error) {
	klog.V(6).Infof("ListFiles %s from [%s/%s:%s]\n", path, p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, p.kubectl.Statement.ContainerName)

	var result []byte
	err := p.Command("ls", "-l", path).Execute(&result).Error
	if err != nil {
		return nil, fmt.Errorf("error executing ListFiles: %v", err)
	}

	return parseFileList(path, string(result)), nil
}
func (p *pod) ListAllFiles(path string) ([]*FileInfo, error) {
	klog.V(6).Infof("ListFiles %s from [%s/%s:%s]\n", path, p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, p.kubectl.Statement.ContainerName)

	var result []byte
	err := p.Command("ls", "-l", "-a", path).Execute(&result).Error
	if err != nil {
		return nil, fmt.Errorf("error executing ListFiles: %v", err)
	}

	return parseFileList(path, string(result)), nil
}
func (p *pod) DownloadFile(filePath string) ([]byte, error) {
	klog.V(6).Infof("DownloadFile %s from [%s/%s:%s]\n", filePath, p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, p.kubectl.Statement.ContainerName)

	result, err := p.DownloadTarFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error executing DownloadTarFile: %v", err)
	}

	tr := tar.NewReader(bytes.NewReader(result))
	var fileContent []byte
	found := false

	// Iterate through each file in the tar
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tar header: %v", err)
		}

		if header.Name == strings.TrimPrefix(filePath, "/") {
			found = true
			// Use size-limited reading method
			if header.Size > 500*1024*1024 { // 500MB limit
				return nil, fmt.Errorf("file size %d exceeds maximum allowed size", header.Size)
			}

			buf := bytes.NewBuffer(make([]byte, 0, header.Size))
			if _, err := io.Copy(buf, tr); err != nil {
				return nil, fmt.Errorf("error reading file content: %v", err)
			}
			fileContent = buf.Bytes()
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("file %s not found in container", filePath)
	}

	return fileContent, nil
}
func (p *pod) DownloadTarFile(filePath string) ([]byte, error) {
	klog.V(6).Infof("DownloadTarFile %s from [%s/%s:%s]\n", filePath, p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, p.kubectl.Statement.ContainerName)

	var result []byte
	err := p.Command("tar", "cf", "-", filePath).Execute(&result).Error
	if err != nil {
		return nil, fmt.Errorf("error executing DownloadTarFile: %v", err)
	}

	return result, nil
}
func (p *pod) DeleteFile(filePath string) ([]byte, error) {
	klog.V(6).Infof("DeleteFile %s from [%s/%s:%s]\n", filePath, p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, p.kubectl.Statement.ContainerName)
	var result []byte
	err := p.Command("rm", "-rf", filePath).Execute(&result).Error
	if err != nil {
		return nil, fmt.Errorf("error executing DeleteFile : %v", err)
	}

	return result, nil
}

// UploadFile uploads a file to the specified container
func (p *pod) UploadFile(destPath string, file *os.File) error {
	klog.V(6).Infof("UploadFile %s to [%s/%s:%s] \n", destPath, p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, p.kubectl.Statement.ContainerName)

	// Read and compress file content
	var buf bytes.Buffer
	if err := createTar(file, &buf); err != nil {
		panic(err.Error())
	}
	var result []byte
	err := p.
		Stdin(&buf).
		Command("tar", "-xmf", "-", "-C", destPath).
		Execute(&result).Error
	if err != nil {
		return fmt.Errorf("error executing UploadFile: %v", err)
	}
	return nil
}

// createTar creates a tar format compressed file
func createTar(file *os.File, buf *bytes.Buffer) error {
	// Create tar writer
	tw := tar.NewWriter(buf)
	defer tw.Close()

	// Get file information
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	// Add file header information
	hdr := &tar.Header{
		Name: stat.Name(),
		Mode: int64(stat.Mode()),
		Size: stat.Size(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	// Write file content to tar
	_, err = io.Copy(tw, file)
	return err
}

// SaveFile
// TODO Byte data to be written to file
//
//	data := []byte("This is some byte data.")
//
//	// Create or open file
//	file, err := os.Create("output.txt")
//	if err != nil {
//	    fmt.Println("Cannot create file:", err)
//	    return
//	}
//	defer file.Close() // Ensure file is closed when function ends
//
//	// Write []byte to file
//	_, err = file.Write(data)
//	if err != nil {
//	    fmt.Println("Failed to write to file:", err)
//	    return
//	}
//
//	fmt.Println("Byte data has been successfully written to file.")
func (p *pod) SaveFile(destPath string, context string) error {
	klog.V(6).Infof("SaveFile %s to [%s/%s:%s]\n", destPath, p.kubectl.Statement.Namespace, p.kubectl.Statement.Name, p.kubectl.Statement.ContainerName)
	klog.V(8).Infof("SaveFile %s \n", context)

	var result []byte
	err := p.
		Stdin(strings.NewReader(context)).
		Command("sh", "-c", fmt.Sprintf("cat > %s", destPath)).
		Execute(&result).Error
	if err != nil {
		return fmt.Errorf("error executing command: %v", err)
	}
	return nil
}

// getFileType gets the file type based on file permissions
//
// l represents symbolic link
// - represents regular file
// d represents directory
// b represents block device
// c represents character device
// p represents named pipe
// s represents socket
func getFileType(permissions string) string {
	// Get file type flag
	p := permissions[0]
	var fileType string

	switch p {
	case 'd':
		fileType = "directory" // Directory
	case '-':
		fileType = "file" // Regular file
	case 'l':
		fileType = "link" // Symbolic link
	case 'b':
		fileType = "block" // Block device
	case 'c':
		fileType = "character" // Character device
	case 'p':
		fileType = "pipe" // Named pipe
	case 's':
		fileType = "socket" // Socket
	default:
		fileType = "unknown" // Unknown type
	}

	return fileType
}

// parseFileList parses output and generates FileInfo list
func parseFileList(path, output string) []*FileInfo {
	var nodes []*FileInfo
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		klog.V(6).Infof("parseFileList path %s %s\n", path, line)
		if len(parts) < 9 {
			continue // Incomplete line
		}

		permissions := parts[0]
		name := parts[8]
		size := parts[4]
		owner := parts[2]
		group := parts[3]
		modTime := strings.Join(parts[5:8], " ")

		// Determine file type

		fileType := getFileType(permissions)

		// Package into FileInfo
		node := FileInfo{
			Path:        fmt.Sprintf("/%s", name),
			Name:        name,
			Type:        fileType,
			Owner:       owner,
			Group:       group,
			Permissions: permissions,
			Size:        utils.ToInt64(size),
			ModTime:     modTime,
			IsDir:       fileType == "directory",
		}
		if strings.HasPrefix(name, "/") {
			node.Path = name
		} else if path != "/" && path != name {
			node.Path = fmt.Sprintf("%s/%s", path, name)
		} else {
			node.Path = fmt.Sprintf("/%s", name)
		}

		nodes = append(nodes, &node)
	}

	return nodes
}
