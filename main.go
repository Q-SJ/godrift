package main

import (
	"fmt"
	"strings"
	"strconv"
	"time"
	"bufio"
	"image/jpeg"
	"os"
	"flag"
	"path"

	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/layers"

	"./fbufio"
	. "./log"
	. "./media"
	"image"
	"image/png"
	"net"
)

var snaplen int
var filter string
var promisc bool
var dir string
var device string
var level string

func init() {
	flag.IntVar(&snaplen, "n", 1600, "the maximum size to read for each packet")
	flag.StringVar(&filter, "f", "tcp and src port 80", "the filter of the stream")
	flag.BoolVar(&promisc, "p", false, "set the NIC to the promiscuous mode")
	flag.StringVar(&dir, "d", "/tmp/godrift", "the directory to save the images you caught")
	flag.StringVar(&device, "i", "", "the NIC you want to monitor on")
	flag.StringVar(&level, "l", "INFO", "the log level")

	flag.Parse()

	SetLoggerLevel(level)
}

func get_devices() (string) {
	// Find all devices
	devices, err := pcap.FindAllDevs()
	if err != nil {
		Logger.Fatal(err)
		panic("No valid device detected!")
	}

	// Print device information
	Logger.Info("Devices found:")
	for i, device := range devices {
		fmt.Println(i, ":", device.Name)
		//fmt.Println("Description: ", device.Description)
		//fmt.Println("Devices addresses: ", device.Description)
		//for _, address := range device.Addresses {
		//	fmt.Println("- IP address: ", address.IP)
		//	fmt.Println("- Subnet mask: ", address.Netmask)
		//}
	}
	var i string
	for true {
		fmt.Scanln(&i)
		if strings.Trim(i, "") == "" {
			return devices[0].Name
		}
		index, err := strconv.ParseInt(i, 10, 0)
		if err != nil || int(index) >= len(devices) {
			fmt.Println("An invalid selection! Please select again:")
			continue
		}
		return devices[index].Name
	}
	return ""
}

type imageStreamFactory struct{}

type imageStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
}

func (stream *imageStream) run() {

	Logger.Debugf("Stream from %s:%s to %s:%s", stream.net.Src(), stream.transport.Src(), stream.net.Dst(), stream.transport.Dst())
	//the directory to save pictures belong to this stream
	dirname := fmt.Sprintf("%s<--%s", stream.net.Dst(), stream.net.Src())

	addrs, err := net.LookupAddr(stream.net.Src().String())
	if err != nil {
		Logger.Warningf("Failed to resolve the host name for ip %s: %s", stream.net.Src(), err)
	} else {
		dirname = fmt.Sprintf("%s<--%s", stream.net.Dst(), addrs[0])
		var addrfinal = ""
		for _, addr := range addrs {
			addrfinal = " " + addrfinal + addr
		}
		Logger.Infof("Host name for ip %s is %s", stream.net.Src(), addrfinal)
	}

	finalname := path.Join(dir, dirname)
	created := false
	var filename string

	buf := bufio.NewReader(&stream.r)
	reader := fbufio.NewReader(buf)

	for true {
		i, err := reader.ReadAfterAll(JPEG_PNG)
		if err != nil || i < 0 {
			Logger.Warningf("Stream from %s:%s to %s:%s encountered an error: %s.", stream.net.Src(), stream.transport.Src(), stream.net.Dst(), stream.transport.Dst(), err)
			return
		}

		if !created {
			//create a directory to save image
			err = os.MkdirAll(finalname, os.ModePerm)
			if err != nil {
				Logger.Warningf("Failed to create directory %s", finalname)
				return
			}
			Logger.Debugf("Created a directory: %s", finalname)
			created = true
		}

		var image image.Image
		var suffix string

		Logger.Info("Got a picture!")
		switch i {
		case JPEG:
			image, err = jpeg.Decode(reader)
			if err != nil {
				Logger.Warning("Cannot extract a jpeg picture.")
				continue
			}
			Logger.Debug("Got a jpeg!")
			suffix = ".jpg"
			filename = time.Now().Format("2006-01-02 15:04:05.000") + suffix
			Logger.Debugf("Create a file named: %s", path.Join(finalname, filename))
			f, err := os.Create(path.Join(finalname, filename))
			if err != nil {
				Logger.Warningf("Create file %s, error: %s.", filename, err)
				continue
			}
			jpeg.Encode(f, image, nil)

		case PNG:
			image, err = png.Decode(reader)
			if err != nil {
				Logger.Warning("Cannot extract a jpeg picture.")
				continue
			}
			Logger.Debug("Got a png!")
			suffix = ".png"
			filename = time.Now().Format("2006-01-02 15:04:05.000") + suffix
			Logger.Debugf("Create a file named: %s", path.Join(finalname, filename))
			f, err := os.Create(path.Join(finalname, filename))
			if err != nil {
				Logger.Warningf("Create file %s, error: %s.", filename, err)
				continue
			}
			png.Encode(f, image)
		}

		//err = reader.ReadAfter([]byte{0xFF, 0xD8})
		//if err != nil {
		//	Logger.Warningf("Stream from %s:%s to %s:%s encountered an error: %s.", stream.net.Src(), stream.transport.Src(), stream.net.Dst(), stream.transport.Dst(), err)
		//	return
		//}
		//
		//image, err := jpeg.Decode(reader)
		//if err != nil {
		//	Logger.Warning(err)
		//	continue
		//}
		//
		//Logger.Debug(image.Bounds())
		//filename = time.Now().Format("2006-01-02 15:04:05.000") +".jpg"
		//f, err := os.Create(path.Join(finalname, filename))
		//if err != nil {
		//	Logger.Warningf("Create file %s, error: %s.", filename, err)
		//	return
		//}
		//jpeg.Encode(f, image, nil)
	}
}

func (i *imageStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	imgstream := &imageStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
	}
	go imgstream.run()
	return &imgstream.r
}

func main() {

	//find the NIC you want to monitor on
	if device == "" {
		device = get_devices()
	}
	fmt.Println("You've selected ", device+".")

	var handle *pcap.Handle

	Logger.Debugf("Starting capture on interface %s", device)
	handle, err := pcap.OpenLive(device, int32(snaplen), promisc, pcap.BlockForever)
	if err != nil {
		Logger.Warning(err)
		return
	}
	if err := handle.SetBPFFilter(filter); err != nil {
		Logger.Fatal(err)
		return
	}

	streamFactory := &imageStreamFactory{}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	Logger.Debug("Begin reading in packets...")
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()
	ticker := time.Tick(time.Second * 20)

	for {
		select {
		case packet := <-packets:
			if packet == nil {
				return
			}

			//check to ensure the packet received has the right format
			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				Logger.Debug("Unusable packet")
				continue
			}
			tcp := packet.TransportLayer().(*layers.TCP)
			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)
		case <-ticker:
			Logger.Debug("Flush data...")
			assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
		}
	}
}
