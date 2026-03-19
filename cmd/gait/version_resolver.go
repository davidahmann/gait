package main

import (
	"regexp"
	"runtime/debug"
	"strings"
)

const localDevVersion = "0.0.0-dev"

var readBuildInfo = debug.ReadBuildInfo

var (
	semverLikeVersionPattern = regexp.MustCompile(`^v?\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$`)
	pseudoVersionPattern     = regexp.MustCompile(`^v\d+\.\d+\.\d+-0\.\d{14}-[0-9a-f]{12,}(?:\+[0-9A-Za-z.-]+)?$`)
)

func currentVersion() string {
	return resolveCLIVersion(version, readBuildInfo)
}

func resolveCLIVersion(stamped string, reader func() (*debug.BuildInfo, bool)) string {
	if candidate := normalizeExplicitVersion(stamped); candidate != "" && candidate != localDevVersion {
		return candidate
	}
	if reader != nil {
		if info, ok := reader(); ok {
			if candidate := trustedBuildInfoVersion(info); candidate != "" {
				return candidate
			}
		}
	}
	if candidate := normalizeExplicitVersion(stamped); candidate != "" {
		return candidate
	}
	return localDevVersion
}

func normalizeExplicitVersion(candidate string) string {
	trimmed := strings.TrimSpace(candidate)
	if trimmed == "" {
		return ""
	}
	if normalized, ok := normalizeReleaseVersion(trimmed); ok {
		return normalized
	}
	return trimmed
}

func trustedBuildInfoVersion(info *debug.BuildInfo) string {
	if info == nil {
		return ""
	}
	candidate := strings.TrimSpace(info.Main.Version)
	if candidate == "" || candidate == "(devel)" || strings.Contains(candidate, "+dirty") {
		return ""
	}
	if !semverLikeVersionPattern.MatchString(candidate) || pseudoVersionPattern.MatchString(candidate) {
		return ""
	}
	normalized, ok := normalizeReleaseVersion(candidate)
	if !ok {
		return ""
	}
	return normalized
}

func normalizeReleaseVersion(candidate string) (string, bool) {
	trimmed := strings.TrimSpace(candidate)
	if trimmed == localDevVersion {
		return "", false
	}
	if !semverLikeVersionPattern.MatchString(trimmed) {
		return "", false
	}
	return strings.TrimPrefix(trimmed, "v"), true
}
