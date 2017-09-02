package listen

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"net"
)

func TestDummy(t *testing.T) {
	ips,err := GetListeningIp("tcp","")
	Convey("TestSS", t, func() {
		So(err, ShouldEqual, nil)
		// TODO actually run some service, relying on some other stuff running on localhost is a bit wonky
		So(ips, ShouldNotBeEmpty)
		So(net.ParseIP(ips[0].String()), ShouldNotBeNil)
	})
}
