package ifup

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"net"
)

func TestIfup(t *testing.T) {
	ip,err := GetLocalIp("lo","")
	Convey("N", t, func() {
		So(err,ShouldBeNil)
		// only checking IPv6 as there are systems with ipv4 only enabled
		So(ip,ShouldContain,net.ParseIP("127.0.0.1"))
	})
}
