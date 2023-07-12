package trace

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)
var (
	// set by the linker at build time
	buildVersion     = "unknown"
	buildGitRevision = "unknown"
	buildTime        = "unknown"

	componentName = ""
)

const requestId = "z-request-id"
const requestPath = "z-request-path"
const hopCount = "z-request-hops"

// We use -1 for uninitialized hop count as hop count 0 has a meaning of 1st service
const emptyHop = -1

func RequestIdFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	v := md.Get(requestId)
	if len(v) > 0 {
		return v[0]
	}
	return ""
}

func RequestPathFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	v := md.Get(requestPath)
	if len(v) > 0 {
		return v[0]
	}
	return ""
}

func HopCountFromContext(ctx context.Context) int {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return emptyHop
	}
	v := md.Get(hopCount)
	if len(v) > 0 {
		if i, err := strconv.Atoi(v[0]); err == nil {
			return i
		}
		return emptyHop
	}
	return emptyHop
}

func NewFromIncomingContext(ctx context.Context) context.Context {
	reqId := RequestIdFromContext(ctx)
	if len(reqId) == 0 {
		reqId = uuid.New().String()
	}
	reqPath := RequestPathFromContext(ctx)
	if len(reqPath) == 0 {
		reqPath = makeNodeInfo()
	} else {
		reqPath = fmt.Sprintf("%s -> %s", reqPath, makeNodeInfo())
	}
	hops := HopCountFromContext(ctx)
	if hops == emptyHop {
		hops = 0
	} else {
		hops++
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	md.Set(requestId, reqId)
	md.Set(requestPath, reqPath)
	md.Set(hopCount, strconv.Itoa(hops))
	ctx = metadata.NewIncomingContext(ctx, md)
	ctx = AppendToOutgoingContext(ctx)
	return ctx
}

func AppendToOutgoingContext(ctx context.Context) context.Context {
	reqId := RequestIdFromContext(ctx)
	if len(reqId) > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, requestId, reqId)
	}
	reqPath := RequestPathFromContext(ctx)
	if len(reqPath) > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, requestPath, reqPath)
	}
	hops := HopCountFromContext(ctx)
	if hops != emptyHop {
		ctx = metadata.AppendToOutgoingContext(ctx, hopCount, strconv.Itoa(hops))
	}
	return ctx
}
func makeNodeInfo() string {
	return fmt.Sprintf("(%s@%s)", ShortString(), getLocalIp())
}
func getLocalIp() (s string) {
	s = "unknown"
	iaddrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, i := range iaddrs {
		if ipnet, ok := i.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				s = ipnet.IP.String()
			}
		}
	}
	return
}

func SetComponentName(name string) {
	if len(componentName) > 0 {
		panic("component name is already set")
	}
	componentName = name
}

func ComponentName() string {
	if len(componentName) == 0 {
		return "unspecified"
	}
	return componentName
}
func ShortString() string {
	return fmt.Sprintf("%s~%s", ComponentName(), buildVersion)
}
