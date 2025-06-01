package utils

import (
	"fmt"
	"strings"

	"golang.org/x/sys/unix"
)

// formatClientAddr efficiently formats client address into the provided buffer
func FormatClientAddr(buf *strings.Builder, addr unix.Sockaddr) {
	if sa, ok := addr.(*unix.SockaddrInet4); ok {
		fmt.Fprintf(buf, "%d.%d.%d.%d:%d", sa.Addr[0], sa.Addr[1], sa.Addr[2], sa.Addr[3], sa.Port)
	} else if sa, ok := addr.(*unix.SockaddrInet6); ok {
		fmt.Fprintf(buf, "[%x:%x:%x:%x:%x:%x:%x:%x]:%d",
			sa.Addr[0:2], sa.Addr[2:4], sa.Addr[4:6], sa.Addr[6:8],
			sa.Addr[8:10], sa.Addr[10:12], sa.Addr[12:14], sa.Addr[14:16], sa.Port)
	} else {
		buf.WriteString("unknown")
	}
}

// validateString checks if the provided string is not empty and returns an error if it is.
func ValidateString(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("string cannot be empty")
	}
	return nil
}
