package discovery

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc/resolver"
)

type Server struct {
	Name    string `json:"name"`
	Addr    string `json:"addr"`
	Version string `json:"version"`
	Weight  int64  `json:"weight"`
}

func BuildPrefix(server Server) string {
	if server.Version == "" {
		return fmt.Sprintf("/%s/", server.Name)
	}
	return fmt.Sprintf("/%s/%s/", server.Name, server.Version)
}

func BuildRegisterPath(server Server) string {
	return fmt.Sprintf("%s%s", BuildPrefix(server), server.Addr)
}

// 将value反序列化到Server实例中
func ParseValue(value []byte) (Server, error) {
	server := Server{}
	if err := json.Unmarshal(value, &server); err != nil {
		return server, err
	}
	return server, nil
}

func SplitPath(path string) (Server, error) {
	server := Server{}
	strs := strings.Split(path, "/")
	if len(strs) == 0 {
		return server, errors.New("invalid path")
	}
	server.Addr = strs[len(strs)-1]
	return server, nil
}

func Exist(l []resolver.Address, addr resolver.Address) bool {
	for i := range l {
		if l[i].Addr == addr.Addr {
			return true
		}
	}
	return false
}
