package dockerversion

import (
	"fmt"
	"runtime"

	"github.com/docker/docker/api/server/httputils"
	"github.com/docker/docker/pkg/parsers/kernel"
	"github.com/docker/docker/pkg/useragent"
	"golang.org/x/net/context"
)

// DockerUserAgent is the User-Agent the Docker client uses to identify itself.
// In accordance with RFC 7231 (5.5.3) is of the form:
//    [docker client's UA] UpstreamClient([upstream client's UA])
func DockerUserAgent(ctx context.Context) string {
	customUA := getCustomUserAgentFromContext(ctx)
	if len(customUA) > 0 {
		return customUA
	}

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
	upstreamUA := getUserAgentFromContext(ctx)
	if len(upstreamUA) > 0 {
		ret := insertUpstreamUserAgent(upstreamUA, dockerUA)
		return ret
	}
	return dockerUA
}

// getUserAgentFromContext returns the previously saved user-agent context stored in ctx, if one exists
func getUserAgentFromContext(ctx context.Context) string {
	var upstreamUA string
	if ctx != nil {
		var ki interface{} = ctx.Value(httputils.UAStringKey)
		if ki != nil {
			upstreamUA = ctx.Value(httputils.UAStringKey).(string)
		}
	}
	return upstreamUA
}

// getCustomUserAgentFromContext returns the custom user-agent context stored in ctx, if one exists
func getCustomUserAgentFromContext(ctx context.Context) string {
	var customUA string
	if ctx != nil {
		var ki interface{} = ctx.Value(httputils.CustomUAStringKey)
		if ki != nil {
			customUA = ctx.Value(httputils.CustomUAStringKey).(string)
		}
	}
	return customUA
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
