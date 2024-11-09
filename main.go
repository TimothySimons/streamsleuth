package main

import (
	"fmt"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/nareix/joy5/codec/h264"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
)

func main() {
	c := gortsplib.Client{}

	rtspURL := "rtsp://onvifuser:onvifpassword1@10.3.0.10/Streaming/Channels/101?channel=1&profile=Profile_1&subtype=0&transportmode=unicast"
	// rtspURL := "rtsp://10.0.0.104:554/cam/realmonitor?channel=1&subtype=0&unicast=true&proto=Onvif"
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
	s, err := h264.ParseSPS(sps)
	if err != nil {
		panic(err)
	}

	fmt.Println("SPS Information:")
	fmt.Println("====================")
	printByteArray(sps)

	fmt.Printf("ID:              %d\n", s.Id)
	fmt.Printf("Profile IDC:     %d\n", s.ProfileIdc)
	fmt.Printf("Level IDC:       %d\n", s.LevelIdc)
	fmt.Printf("Constraint Set:  %d\n", s.ConstraintSetFlag)
	fmt.Println()

	fmt.Println("Macroblock Dimensions:")
	fmt.Println("--------------------")
	fmt.Printf("Width:           %d\n", s.MbWidth)
	fmt.Printf("Height:          %d\n", s.MbHeight)
	fmt.Println()

	fmt.Println("Cropping:")
	fmt.Println("--------------------")
	fmt.Printf("Left:            %d\n", s.CropLeft)
	fmt.Printf("Right:           %d\n", s.CropRight)
	fmt.Printf("Top:             %d\n", s.CropTop)
	fmt.Printf("Bottom:          %d\n", s.CropBottom)
	fmt.Println()

	fmt.Println("Frame Dimensions:")
	fmt.Println("--------------------")
	fmt.Printf("Width:           %d\n", s.Width)
	fmt.Printf("Height:          %d\n", s.Height)
	fmt.Println()

	fmt.Printf("Frames per Second: %d\n", s.FPS)
	fmt.Println()
}

func printPPS(pps []byte) {
	p, err := h264.ParsePPS(pps)
	if err != nil {
		panic(err)
	}

	fmt.Println("PPS Information:")
	fmt.Println("====================")
	printByteArray(pps)

	fmt.Printf("ID:                %d\n", p.Id)
	fmt.Printf("Associated SPS ID: %d\n", p.SPSId)
	fmt.Println()
}

func printRTP(pkt *rtp.Packet) {
	fmt.Printf("\n%s\n", pkt.String())

	data := pkt.Payload
	startAddress := uintptr(0x1000)
	bytesPerLine := 16
	for i := 0; i < len(data); i += bytesPerLine {
		fmt.Printf("0x%04X | ", startAddress+uintptr(i))

		for j := 0; j < bytesPerLine; j++ {
			if i+j < len(data) {
				fmt.Printf("%02X ", data[i+j])
			}
		}

		fmt.Println()
	}
}

func printByteArray(data []byte) {
	for i, b := range data {
		fmt.Printf("%02X", b)
		if i < len(data)-1 {
			fmt.Print(" ")
		}
	}
	fmt.Println()
}
