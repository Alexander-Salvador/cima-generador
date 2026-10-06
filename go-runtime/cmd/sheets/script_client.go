// script_client.go — Apps Script API v1 client (stub for CIMA).
//
// CIMA does not use Apps Script operations (ADR-001: Sheets API only).
// This file provides the minimal type definitions and stubs so the
// sheets bridge compiles. Scripts commands will return ErrNotImplemented.
package main

import (
	"errors"
	"fmt"
)

// ErrNotImplemented is returned by all ScriptClient methods since
// CIMA uses Sheets API exclusively and does not interact with
// Apps Script projects.
var ErrNotImplemented = errors.New("script: Apps Script not used in CIMA (sheets API only)")

// ScriptClient wraps an OAuth2 token source for Apps Script API calls.
type ScriptClient struct {
	creds *Creds
}

// NewScriptClient returns a ScriptClient bound to the given credentials.
func NewScriptClient(creds *Creds) *ScriptClient {
	return &ScriptClient{creds: creds}
}

// Project describes an Apps Script project.
type Project struct {
	Title     string `json:"title"`
	ScriptID  string `json:"scriptId"`
	ParentID  string `json:"parentId"`
	UpdateTime string `json:"updateTime"`
	Creator   struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"creator"`
}

// ListProjects lists all Apps Script projects accessible to the user.
func (sc *ScriptClient) ListProjects() ([]Project, error) {
	return nil, ErrNotImplemented
}

// FindByContainer returns the Apps Script project bound to a container (sheet/folder ID).
func (sc *ScriptClient) FindByContainer(container string) (*Project, error) {
	return nil, ErrNotImplemented
}

// ManifestFile describes one file inside an Apps Script project.
type ManifestFile struct {
	Path      string `json:"path"`
	Type      string `json:"type"`
	SizeBytes int64  `json:"sizeBytes"`
	SHA256    string `json:"sha256"`
}

// PullManifest is returned by PullToDir.
type PullManifest struct {
	Files []ManifestFile `json:"files"`
}

// PullToDir downloads all files from an Apps Script project to a local directory.
func (sc *ScriptClient) PullToDir(id, outDir string) (*PullManifest, error) {
	return nil, ErrNotImplemented
}

// PushManifest is returned by PushFromDir.
type PushManifest struct {
	Files []ManifestFile `json:"files"`
}

// PushFromDir uploads all files from a local directory to an Apps Script project.
func (sc *ScriptClient) PushFromDir(id, inDir string, confirm bool) (*PushManifest, error) {
	return nil, ErrNotImplemented
}

// ContentSnapshot holds pre-push script content for rollback.
type ContentSnapshot struct {
	ScriptID   string          `json:"scriptId"`
	Version    int             `json:"versionNumber"`
	Description string         `json:"description"`
	Files      []ManifestFile  `json:"files"`
}

// GetContent returns the current content of all files in an Apps Script project.
func (sc *ScriptClient) GetContent(id string) (*ContentSnapshot, error) {
	return nil, ErrNotImplemented
}

// CaptureForScript stores a pre-operation snapshot for rollback purposes.
func (sc *ScriptClient) CaptureForScript(repoRoot, id, desc string, preBytes []byte) (int, error) {
	return 0, ErrNotImplemented
}

// RunResult holds the return value from an Apps Script function execution.
type RunResult struct {
	Result any `json:"result"`
}

// RunFunction executes a function in an Apps Script project.
func (sc *ScriptClient) RunFunction(id, fn string, args any) (*RunResult, error) {
	return nil, ErrNotImplemented
}

// Version describes a versioned snapshot of an Apps Script project.
type Version struct {
	VersionNumber int    `json:"versionNumber"`
	CreateTime   string `json:"createTime"`
	Description  string `json:"description"`
}

// ListVersions returns all versions of an Apps Script project.
func (sc *ScriptClient) ListVersions(id string) ([]Version, error) {
	return nil, ErrNotImplemented
}

// CreateVersion creates a new immutable version snapshot.
func (sc *ScriptClient) CreateVersion(id, desc string) (int, error) {
	return 0, ErrNotImplemented
}

// Stub implementations for allowlist subcommands and MCP server.
func cmdAllowlistList(args []string) error {
	return fmt.Errorf("allowlist list: %w", ErrNotImplemented)
}

func cmdAllowlistAdd(args []string) error {
	return fmt.Errorf("allowlist add: %w", ErrNotImplemented)
}

func cmdAllowlistRemove(args []string) error {
	return fmt.Errorf("allowlist remove: %w", ErrNotImplemented)
}

func runMCPServer(repoRoot string) error {
	return fmt.Errorf("mcp: %w", ErrNotImplemented)
}
