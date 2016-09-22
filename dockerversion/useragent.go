package dockerversion

import (
	"fmt"
	"runtime"

	"github.com/docker/docker/pkg/parsers/kernel"
	"github.com/docker/docker/pkg/useragent"
)

// DockerUserAgent is the User-Agent the Docker client uses to identify itself.
// In accordance with RFC 7231 (5.5.3) is of the form:
//    [docker client's UA] UpstreamClient([upstream client's UA])
func DockerUserAgent(upstreamUA string) string {
	httpVersion := make([]useragent.VersionInfo, 0, 6)
	httpVersion = append(httpVersion, useragent.VersionInfo{Name: "docker", Version: Version})
	httpVersion = append(httpVersion, useragent.VersionInfo{Name: "go", Version: runtime.Version()})
	httpVersion = append(httpVersion, useragent.VersionInfo{Name: "git-commit", Version: GitCommit})
	if kernelVersion, err := kernel.GetKernelVersion(); err == nil {
		httpVersion = append(httpVersion, useragent.VersionInfo{Name: "kernel", Version: kernelVersion.String()})
	}
	httpVersion = append(httpVersion, useragent.VersionInfo{Name: "os", Version: runtime.GOOS})
	httpVersion = append(httpVersion, useragent.VersionInfo{Name: "arch", Version: runtime.GOARCH})

	dockerUA := useragent.AppendVersions("", httpVersion...)

	if len(upstreamUA) > 0 {
		dockerUA = insertUpstreamUserAgent(upstreamUA, dockerUA)
	}

	return dockerUA
}

// escapeStr returns s with every rune in charsToEscape escaped by a backslash
func escapeStr(s string, charsToEscape string) string {
	var ret string
	for _, currRune := range s {
		appended := false
		for _, escapeableRune := range charsToEscape {
			if currRune == escapeableRune {
				ret += `\` + string(currRune)
				appended = true
				break
			}
		}
		if !appended {
			ret += string(currRune)
		}
	}
	return ret
}

// insertUpstreamUserAgent adds the upstream client useragent to create a user-agent
// string of the form:
//   $dockerUA UpstreamClient($upstreamUA)
func insertUpstreamUserAgent(upstreamUA string, dockerUA string) string {
	charsToEscape := `();\`
	upstreamUAEscaped := escapeStr(upstreamUA, charsToEscape)
	return fmt.Sprintf("%s UpstreamClient(%s)", dockerUA, upstreamUAEscaped)
}
