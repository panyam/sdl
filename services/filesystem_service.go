//go:build !wasm
// +build !wasm

package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	protos "github.com/panyam/sdl/gen/go/sdl/v1/models"
	"github.com/panyam/sdl/lib/loader"
)

// FilesystemService wraps a loader.FileSystem to provide filesystem operations
type FilesystemService struct {
	// The underlying composite filesystem from the loader
	fs *loader.CompositeFS

	// Filesystem metadata for listing purposes
	filesystems map[string]*FilesystemMeta
}

// FilesystemMeta stores metadata about a mounted filesystem
type FilesystemMeta struct {
	ID         string
	Prefix     string
	Type       string
	ReadOnly   bool
	BasePath   string
	Extensions []string
}

// NewFilesystemService creates a new FilesystemService with default filesystems mounted
func NewFilesystemService() *FilesystemService {
	svc := &FilesystemService{
		fs:          loader.NewCompositeFS(),
		filesystems: make(map[string]*FilesystemMeta),
	}

	// Mount default filesystems
	svc.MountLocal("examples", "./examples", false, []string{".sdl", ".recipe"})
	svc.MountLocal("demos", "./demos", true, []string{".sdl", ".recipe"})

	// Mount GitHub and HTTP filesystems
	svc.MountGitHub()
	svc.MountHTTP()

	return svc
}

// NewFilesystemServiceWithLoader creates a FilesystemService that wraps an existing loader's filesystem
func NewFilesystemServiceWithLoader(l *loader.Loader) *FilesystemService {
	svc := &FilesystemService{
		fs:          loader.NewCompositeFS(),
		filesystems: make(map[string]*FilesystemMeta),
	}
	// The loader's filesystem would be set from the loader's resolver
	// For now, we mount defaults - this can be enhanced to read from the loader
	svc.MountLocal("examples", "./examples", false, []string{".sdl", ".recipe"})
	svc.MountLocal("demos", "./demos", true, []string{".sdl", ".recipe"})
	svc.MountGitHub()
	svc.MountHTTP()
	return svc
}

// MountLocal mounts a local filesystem at the given ID
func (s *FilesystemService) MountLocal(id, basePath string, readOnly bool, extensions []string) {
	var fs loader.FileSystem
	if readOnly {
		fs = loader.NewReadOnlyLocalFS(basePath)
	} else {
		fs = loader.NewLocalFS(basePath)
	}
	s.fs.Mount(id+"/", fs)
	s.filesystems[id] = &FilesystemMeta{
		ID:         id,
		Prefix:     "/" + id + "/",
		Type:       "local",
		ReadOnly:   readOnly,
		BasePath:   basePath,
		Extensions: extensions,
	}
}

// MountGitHub mounts the GitHub filesystem
func (s *FilesystemService) MountGitHub() {
	s.fs.Mount("github.com/", loader.NewGitHubFS())
	s.filesystems["github"] = &FilesystemMeta{
		ID:       "github",
		Prefix:   "github.com/",
		Type:     "github",
		ReadOnly: true,
	}
}

// MountHTTP mounts the HTTP filesystem
func (s *FilesystemService) MountHTTP() {
	s.fs.Mount("https://", loader.NewHTTPFileSystem(""))
	s.filesystems["https"] = &FilesystemMeta{
		ID:       "https",
		Prefix:   "https://",
		Type:     "http",
		ReadOnly: true,
	}
}

// GetCompositeFS returns the underlying composite filesystem
func (s *FilesystemService) GetCompositeFS() *loader.CompositeFS {
	return s.fs
}

// ListFilesystems returns all available filesystems
func (s *FilesystemService) ListFilesystems(ctx context.Context, req *protos.ListFilesystemsRequest) (*protos.ListFilesystemsResponse, error) {
	resp := &protos.ListFilesystemsResponse{
		Filesystems: make([]*protos.FilesystemInfo, 0, len(s.filesystems)),
	}

	for _, meta := range s.filesystems {
		resp.Filesystems = append(resp.Filesystems, &protos.FilesystemInfo{
			Id:         meta.ID,
			Prefix:     meta.Prefix,
			Type:       meta.Type,
			ReadOnly:   meta.ReadOnly,
			BasePath:   meta.BasePath,
			Extensions: meta.Extensions,
		})
	}

	return resp, nil
}

// ListFiles lists files in a directory
func (s *FilesystemService) ListFiles(ctx context.Context, req *protos.ListFilesRequest) (*protos.ListFilesResponse, error) {
	meta, ok := s.filesystems[req.FilesystemId]
	if !ok {
		return nil, fmt.Errorf("filesystem not found: %s", req.FilesystemId)
	}

	// For local filesystems, use the basePath + requestPath
	if meta.Type == "local" {
		return s.listLocalFiles(meta, req.Path)
	}

	// For remote filesystems, try to list via the FileSystem interface
	path := meta.Prefix + strings.TrimPrefix(req.Path, "/")
	files, err := s.fs.ListFiles(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	resp := &protos.ListFilesResponse{
		Files: make([]*protos.FileInfo, 0),
	}
	for _, file := range files {
		resp.Files = append(resp.Files, &protos.FileInfo{
			Name: filepath.Base(file),
			Path: file,
		})
	}

	return resp, nil
}

// listLocalFiles lists files in a local filesystem directory with full metadata
func (s *FilesystemService) listLocalFiles(meta *FilesystemMeta, requestPath string) (*protos.ListFilesResponse, error) {
	fullPath := filepath.Join(meta.BasePath, requestPath)
	fullPath = filepath.Clean(fullPath)

	// Security check: ensure path stays within basePath
	absBase, _ := filepath.Abs(meta.BasePath)
	absFull, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFull, absBase) {
		return nil, fmt.Errorf("access denied: path outside filesystem root")
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	resp := &protos.ListFilesResponse{
		Files: make([]*protos.FileInfo, 0),
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// For directories, always include them
		if entry.IsDir() {
			resp.Files = append(resp.Files, &protos.FileInfo{
				Name:        entry.Name(),
				Path:        filepath.Join(requestPath, entry.Name()) + "/",
				IsDirectory: true,
				ModTime:     info.ModTime().Format("2006-01-02T15:04:05Z"),
			})
		} else if s.isAllowedFile(entry.Name(), meta) {
			resp.Files = append(resp.Files, &protos.FileInfo{
				Name:        entry.Name(),
				Path:        filepath.Join(requestPath, entry.Name()),
				IsDirectory: false,
				Size:        info.Size(),
				ModTime:     info.ModTime().Format("2006-01-02T15:04:05Z"),
			})
		}
	}

	return resp, nil
}

// GetFileInfo returns information about a file
func (s *FilesystemService) GetFileInfo(ctx context.Context, req *protos.GetFileInfoRequest) (*protos.GetFileInfoResponse, error) {
	meta, ok := s.filesystems[req.FilesystemId]
	if !ok {
		return nil, fmt.Errorf("filesystem not found: %s", req.FilesystemId)
	}

	if meta.Type == "local" {
		fullPath := filepath.Join(meta.BasePath, req.Path)
		info, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("file not found: %s", req.Path)
			}
			return nil, fmt.Errorf("failed to get file info: %w", err)
		}

		return &protos.GetFileInfoResponse{
			FileInfo: &protos.FileInfo{
				Name:        info.Name(),
				Path:        req.Path,
				IsDirectory: info.IsDir(),
				Size:        info.Size(),
				ModTime:     info.ModTime().Format("2006-01-02T15:04:05Z"),
			},
		}, nil
	}

	// For remote filesystems, just check existence
	path := meta.Prefix + strings.TrimPrefix(req.Path, "/")
	if !s.fs.Exists(path) {
		return nil, fmt.Errorf("file not found: %s", req.Path)
	}

	return &protos.GetFileInfoResponse{
		FileInfo: &protos.FileInfo{
			Name: filepath.Base(req.Path),
			Path: req.Path,
		},
	}, nil
}

// ReadFile reads file content
func (s *FilesystemService) ReadFile(ctx context.Context, req *protos.ReadFileRequest) (*protos.ReadFileResponse, error) {
	meta, ok := s.filesystems[req.FilesystemId]
	if !ok {
		return nil, fmt.Errorf("filesystem not found: %s", req.FilesystemId)
	}

	// Check file extension for local filesystems
	if meta.Type == "local" && !s.isAllowedFile(filepath.Base(req.Path), meta) {
		return nil, fmt.Errorf("file type not allowed")
	}

	path := s.resolvePath(meta, req.Path)
	content, err := s.fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return &protos.ReadFileResponse{
		Content: content,
	}, nil
}

// WriteFile writes content to a file
func (s *FilesystemService) WriteFile(ctx context.Context, req *protos.WriteFileRequest) (*protos.WriteFileResponse, error) {
	meta, ok := s.filesystems[req.FilesystemId]
	if !ok {
		return nil, fmt.Errorf("filesystem not found: %s", req.FilesystemId)
	}

	if meta.ReadOnly {
		return nil, fmt.Errorf("filesystem is read-only")
	}

	// Check file extension
	if !s.isAllowedFile(filepath.Base(req.Path), meta) {
		return nil, fmt.Errorf("file type not allowed")
	}

	path := s.resolvePath(meta, req.Path)
	if err := s.fs.WriteFile(path, req.Content); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &protos.WriteFileResponse{}, nil
}

// DeleteFile deletes a file
func (s *FilesystemService) DeleteFile(ctx context.Context, req *protos.DeleteFileRequest) (*protos.DeleteFileResponse, error) {
	meta, ok := s.filesystems[req.FilesystemId]
	if !ok {
		return nil, fmt.Errorf("filesystem not found: %s", req.FilesystemId)
	}

	if meta.ReadOnly {
		return nil, fmt.Errorf("filesystem is read-only")
	}

	if meta.Type != "local" {
		return nil, fmt.Errorf("delete not supported for this filesystem type")
	}

	fullPath := filepath.Join(meta.BasePath, req.Path)
	fullPath = filepath.Clean(fullPath)

	// Security check
	absBase, _ := filepath.Abs(meta.BasePath)
	absFull, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFull, absBase) {
		return nil, fmt.Errorf("access denied: path outside filesystem root")
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", req.Path)
		}
		return nil, fmt.Errorf("failed to access file: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("cannot delete directories")
	}

	if err := os.Remove(fullPath); err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	return &protos.DeleteFileResponse{
		Success: true,
	}, nil
}

// CreateDirectory creates a directory
func (s *FilesystemService) CreateDirectory(ctx context.Context, req *protos.CreateDirectoryRequest) (*protos.CreateDirectoryResponse, error) {
	meta, ok := s.filesystems[req.FilesystemId]
	if !ok {
		return nil, fmt.Errorf("filesystem not found: %s", req.FilesystemId)
	}

	if meta.ReadOnly {
		return nil, fmt.Errorf("filesystem is read-only")
	}

	if meta.Type != "local" {
		return nil, fmt.Errorf("directory creation not supported for this filesystem type")
	}

	fullPath := filepath.Join(meta.BasePath, req.Path)
	fullPath = filepath.Clean(fullPath)

	// Security check
	absBase, _ := filepath.Abs(meta.BasePath)
	absFull, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFull, absBase) {
		return nil, fmt.Errorf("access denied: path outside filesystem root")
	}

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	info, _ := os.Stat(fullPath)
	var dirInfo *protos.FileInfo
	if info != nil {
		dirInfo = &protos.FileInfo{
			Name:        info.Name(),
			Path:        req.Path,
			IsDirectory: true,
			ModTime:     info.ModTime().Format("2006-01-02T15:04:05Z"),
		}
	}

	return &protos.CreateDirectoryResponse{
		DirectoryInfo: dirInfo,
	}, nil
}

// resolvePath resolves a path for the given filesystem
func (s *FilesystemService) resolvePath(meta *FilesystemMeta, requestPath string) string {
	if meta.Type == "local" {
		return meta.ID + "/" + strings.TrimPrefix(requestPath, "/")
	}
	return meta.Prefix + strings.TrimPrefix(requestPath, "/")
}

// isAllowedFile checks if a file matches the allowed extensions
func (s *FilesystemService) isAllowedFile(filename string, meta *FilesystemMeta) bool {
	if len(meta.Extensions) == 0 {
		return true
	}
	ext := filepath.Ext(filename)
	for _, allowed := range meta.Extensions {
		if ext == allowed {
			return true
		}
	}
	return false
}
