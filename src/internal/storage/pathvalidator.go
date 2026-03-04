package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrPathNotAllowed = errors.New("path not allowed")
	ErrPathEmpty      = errors.New("path empty")
)

type PathValidator struct {
	bases []string
}

func NewPathValidator(bases []string) (*PathValidator, error) {
	norm := make([]string, 0, len(bases))
	for _, b := range bases {
		if b == "" {
			continue
		}
		ex, err := expandTildePath(b)
		if err != nil {
			return nil, err
		}
		clean := filepath.Clean(ex)
		real, err := filepath.EvalSymlinks(clean)
		if err != nil {
			// base 目录不存在时，先用 clean 兜底（初始化阶段允许）
			real = clean
		}
		norm = append(norm, filepath.Clean(real))
	}
	return &PathValidator{bases: norm}, nil
}

func (v *PathValidator) Validate(inputPath string) (string, error) {
	if strings.TrimSpace(inputPath) == "" {
		return "", ErrPathEmpty
	}
	if strings.ContainsRune(inputPath, '\x00') {
		return "", fmt.Errorf("%w: null byte", ErrPathNotAllowed)
	}

	expanded, err := expandTildePath(inputPath)
	if err != nil {
		return "", err
	}

	clean := filepath.Clean(expanded)
	real := clean
	if st, err := os.Lstat(clean); err == nil {
		if st.Mode()&os.ModeSymlink != 0 {
			s, err := filepath.EvalSymlinks(clean)
			if err != nil {
				return "", err
			}
			real = s
		} else {
			if s, err := filepath.EvalSymlinks(clean); err == nil {
				real = s
			}
		}
	}
	real = filepath.Clean(real)

	for _, base := range v.bases {
		if real == base || strings.HasPrefix(real+string(os.PathSeparator), base+string(os.PathSeparator)) || strings.HasPrefix(real, base+string(os.PathSeparator)) {
			return real, nil
		}
	}
	return "", fmt.Errorf("%w: %s", ErrPathNotAllowed, real)
}

func (v *PathValidator) JoinAndValidate(base, subPath string) (string, error) {
	if strings.TrimSpace(base) == "" {
		return "", ErrPathEmpty
	}
	joined := filepath.Join(base, subPath)
	return v.Validate(joined)
}

func expandTildePath(p string) (string, error) {
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if p == "~" {
			return home, nil
		}
		return filepath.Join(home, strings.TrimPrefix(p, "~/")), nil
	}
	return p, nil
}
