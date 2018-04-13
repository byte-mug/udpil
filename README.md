# udpil
UDP Internet Link (insipired by http://doc.cat-v.org/plan_9/4th_edition/papers/il/)

The API is unstable and yet subject to changes.

# Yet another Reliable UDP protocol.

There are already Protocols like 
[UDT](https://en.wikipedia.org/wiki/UDP-based_Data_Transfer_Protocol)
([go impl.](https://github.com/oxtoacart/go-udt)) or 
[KCP](https://github.com/xtaci/kcp-go). In fact, a lot of inspiration is taken from the KCP protocol.

The reason for yet another Reliable UDP protocol was, that KCP kind-of hogs CPU and Memory resources and the go-implementation of UDT is kind-of aging a bit, and the author of the package endorses KCP.

Unlike UDT and KCP, the implementation is pretty simplistic.

## Handshakes

There aren't such things as handshakes. The client simply starts to send data to the server. As the server sees a packet, for which no Connection exists, the Server will simply create one to receive the data stream.

Al the server does, is to validate the first incoming packets, to make sure, it is the start of a data stream.

