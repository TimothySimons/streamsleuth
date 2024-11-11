package main

import (
	"fmt"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	jh264 "github.com/nareix/joy5/codec/h264"

	mh264 "github.com/bluenviron/mediacommon/pkg/codecs/h264"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
)

type PacketFilter struct {
	// Filter to check if the packet is an IDR (Instantaneous Decoder Refresh) frame.
	FilterIDR bool

	// Filter to check for specific NALU unit types (e.g., non-IDR, SEI, etc.)
	FilterNALUType bool

	// Filter to check for specific payload types in the RTP packet.
	FilterPayloadType bool

	// Filter to check if the packet has a specific Marker bit set.
	FilterMarker bool
}

func main() {
	c := gortsplib.Client{}

	// rtspURL := "rtsp://onvifuser:onvifpassword1@10.3.0.10/Streaming/Channels/101?channel=1&profile=Profile_1&subtype=0&transportmode=unicast"
	rtspURL := "rtsp://10.0.0.104:554/cam/realmonitor?channel=1&subtype=0&unicast=true&proto=Onvif"

	u, err := base.ParseURL(rtspURL)
	if err != nil {
		panic(err)
	}

	err = c.Start(u.Scheme, u.Host)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	desc, _, err := c.Describe(u)
	if err != nil {
		panic(err)
	}

	var forma *format.H264
	medi := desc.FindFormat(&forma)
	if medi == nil {
		panic("media not found")
	}

	printSPS(forma.SPS)
	printPPS(forma.PPS)

	_, err = c.Setup(desc.BaseURL, medi, 0, 0)
	if err != nil {
		panic(err)
	}

	c.OnPacketRTP(medi, forma, func(pkt *rtp.Packet) {
		printRTP(pkt)
	})

	c.OnPacketRTCP(medi, func(_ rtcp.Packet) {})

	_, err = c.Play(nil)
	if err != nil {
		panic(err)
	}

	panic(c.Wait())
}

func printSPS(sps []byte) {
	s, err := jh264.ParseSPS(sps)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nSPS INFORMATION:")
	fmt.Println("\tID:", s.Id)
	fmt.Println("\tProfile IDC:", s.ProfileIdc)
	fmt.Println("\tLevel IDC:", s.LevelIdc)
	fmt.Println("\tConstraint Set:", s.ConstraintSetFlag)

	fmt.Println("\nMACROBLOCK DIMENSIONS:")
	fmt.Println("\tWidth:", s.MbWidth)
	fmt.Println("\tHeight:", s.MbHeight)

	fmt.Println("\nCROPPING:")
	fmt.Println("\tLeft:", s.CropLeft)
	fmt.Println("\tRight:", s.CropRight)
	fmt.Println("\tTop:", s.CropTop)
	fmt.Println("\tBottom:", s.CropBottom)

	fmt.Println("\nFRAME DIMENSIONS:")
	fmt.Println("\tWidth:", s.Width)
	fmt.Println("\tHeight:", s.Height)

	fmt.Println("\nFRAMES PER SECOND:", s.FPS)
}

func printPPS(pps []byte) {
	p, err := jh264.ParsePPS(pps)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nPPS INFORMATION:")
	fmt.Println("\tID:", p.Id)
	fmt.Println("\tAssociated SPS ID:", p.SPSId)
	fmt.Println()
}

func printRTP(pkt *rtp.Packet) {
	nalUnitType := mh264.NALUType(pkt.Payload[0] & 0x1F)
	// if nalUnitType != mh264.NALUTypeIDR && nalUnitType != mh264.NALUTypeSPS && nalUnitType != mh264.NALUTypePPS && nalUnitType != mh264.NALUTypeSEI {
	// 	return
	// }

	if pkt.Marker {
		return
	}

	fmt.Printf("\n%s\n", pkt.String())
	fmt.Println("\nRTP PACKET PAYLOAD HEADER")
	fmt.Printf("\tValue: %08b (%02X)\n", pkt.Payload[0], pkt.Payload[0])
	fmt.Printf("\tNALU Unit Type: %s\n", nalUnitType.String())
	fmt.Println()

	// data := pkt.Payload
	// startAddress := uintptr(0x1000)
	// bytesPerLine := 16
	// for i := 0; i < len(data); i += bytesPerLine {
	// 	fmt.Printf("0x%04X | ", startAddress+uintptr(i))

	// 	for j := 0; j < bytesPerLine; j++ {
	// 		if i+j < len(data) {
	// 			fmt.Printf("%02X ", data[i+j])
	// 		}
	// 	}

	// 	fmt.Println()
	// }
}

// RTP Payload Header
//
// +---------------+
// |0|1|2|3|4|5|6|7|
// +-+-+-+-+-+-+-+-+
// |F|NRI|  Type   |
// +---------------+
//
// F: 1 bit
// 		forbidden_zero_bit
//
// NRI: 2 bits
// 		nal_ref_idc
//
// TYPE: 5 bits
// 		nal_unit_type

// func printByteArray(data []byte) {
// 	for i, b := range data {
// 		fmt.Printf("%02X", b)
// 		if i < len(data)-1 {
// 			fmt.Print(" ")
// 		}
// 	}
// 	fmt.Println()
// }

// NOTES
//
// An active sequence parameter set remains
// unchanged throughout a coded video sequence, and an active picture
// parameter set remains unchanged within a coded picture.
//
//    This mechanism allows the decoupling of the transmission of parameter
//    sets from the packet stream and the transmission of them by external
//    means (e.g., as a side effect of the capability exchange) or through
//    a (reliable or unreliable) control protocol.  It may even be possible
//    that they are never transmitted but are fixed by an application
//    design specification.
