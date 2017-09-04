package listen

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"net"
)

func TestListening(t *testing.T) {
	ips,err := GetListeningIp("tcp","")
	Convey("TestSS", t, func() {
		So(err, ShouldEqual, nil)
		// TODO actually run some service, relying on some other stuff running on localhost is a bit wonky
		So(ips, ShouldNotBeEmpty)
		So(net.ParseIP(ips[0].String()), ShouldNotBeNil)
	})
}

// this is 100% dominated by ss timing;  included here just for easy testing.
func BenchmarkListening(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ , _  = GetListeningIp("tcp","")
	}
}
