package session

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type Session struct {
	ID         string
	Agent      string
	Name       string
	ProjectDir string
	LastActive time.Time
	FilePath   string
	CustomName string
}

func (s Session) DisplayName() string {
	if strings.TrimSpace(s.CustomName) != "" {
		return s.CustomName
	}
	if strings.TrimSpace(s.Name) != "" {
		return s.Name
	}
	if strings.TrimSpace(s.ID) != "" {
		return s.ID
	}
	return "untitled session"
}

func (s Session) MetadataKey() string {
	h := sha1.Sum([]byte(fmt.Sprintf("%s|%s|%s", s.Agent, s.ID, s.FilePath)))
	return hex.EncodeToString(h[:])
}
