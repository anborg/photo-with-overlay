package jpegmeta

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"math"
	"sort"
	"time"
)

type Metadata struct {
	CapturedAt          time.Time
	User                string
	Latitude, Longitude float64
	Accuracy            *float64
	Location, Source    string
}

type entry struct {
	tag, kind uint16
	count     uint32
	value     []byte
}

func Insert(jpeg []byte, m Metadata) ([]byte, error) {
	if len(jpeg) < 2 || jpeg[0] != 0xff || jpeg[1] != 0xd8 {
		return nil, fmt.Errorf("invalid JPEG")
	}
	exif := append([]byte("Exif\x00\x00"), buildTIFF(m)...)
	xmp := append([]byte("http://ns.adobe.com/xap/1.0/\x00"), buildXMP(m)...)
	if len(exif)+2 > math.MaxUint16 || len(xmp)+2 > math.MaxUint16 {
		return nil, fmt.Errorf("metadata exceeds JPEG APP1 limit")
	}
	out := make([]byte, 0, len(jpeg)+len(exif)+len(xmp)+8)
	out = append(out, jpeg[:2]...)
	out = appendSegment(out, exif)
	out = appendSegment(out, xmp)
	out = append(out, jpeg[2:]...)
	return out, nil
}

func appendSegment(dst, payload []byte) []byte {
	dst = append(dst, 0xff, 0xe1, byte((len(payload)+2)>>8), byte(len(payload)+2))
	return append(dst, payload...)
}

func buildTIFF(m Metadata) []byte {
	stamp := m.CapturedAt.Format("2006:01:02 15:04:05") + "\x00"
	description := fmt.Sprintf("User=%s; Coordinates=%.6f,%.6f; Location=%s; Source=%s", m.User, m.Latitude, m.Longitude, m.Location, m.Source) + "\x00"
	exifEntries := []entry{ascii(0x9003, stamp), ascii(0x9004, stamp)}
	gpsEntries := []entry{{0x0000, 1, 4, []byte{2, 3, 0, 0}}, ascii(0x0001, hemisphere(m.Latitude, "N", "S")+"\x00"),
		rationals(0x0002, dms(m.Latitude)), ascii(0x0003, hemisphere(m.Longitude, "E", "W")+"\x00"), rationals(0x0004, dms(m.Longitude)),
		rationals(0x0007, [][2]uint32{{uint32(m.CapturedAt.UTC().Hour()), 1}, {uint32(m.CapturedAt.UTC().Minute()), 1}, {uint32(m.CapturedAt.UTC().Second()), 1}}),
		ascii(0x001d, m.CapturedAt.UTC().Format("2006:01:02")+"\x00")}
	if m.Accuracy != nil {
		gpsEntries = append(gpsEntries, rationals(0x000b, [][2]uint32{{uint32(math.Round(*m.Accuracy * 100)), 100}}))
	}
	sort.Slice(gpsEntries, func(i, j int) bool { return gpsEntries[i].tag < gpsEntries[j].tag })

	ifd0Offset := uint32(8)
	ifd0Size := uint32(2 + 5*12 + 4)
	exifOffset := ifd0Offset + ifd0Size
	exifSize := uint32(2 + len(exifEntries)*12 + 4)
	gpsOffset := exifOffset + exifSize
	gpsSize := uint32(2 + len(gpsEntries)*12 + 4)
	dataStart := gpsOffset + gpsSize
	ifd0 := []entry{ascii(0x010e, description), ascii(0x0132, stamp), ascii(0x013b, m.User+"\x00"), long(0x8769, exifOffset), long(0x8825, gpsOffset)}
	tiff := make([]byte, dataStart)
	copy(tiff, []byte{'I', 'I', 42, 0})
	binary.LittleEndian.PutUint32(tiff[4:8], ifd0Offset)
	writeIFD := func(offset uint32, entries []entry) {
		binary.LittleEndian.PutUint16(tiff[offset:offset+2], uint16(len(entries)))
		for i, e := range entries {
			p := int(offset) + 2 + i*12
			binary.LittleEndian.PutUint16(tiff[p:p+2], e.tag)
			binary.LittleEndian.PutUint16(tiff[p+2:p+4], e.kind)
			binary.LittleEndian.PutUint32(tiff[p+4:p+8], e.count)
			if len(e.value) <= 4 {
				copy(tiff[p+8:p+12], e.value)
			} else {
				binary.LittleEndian.PutUint32(tiff[p+8:p+12], uint32(len(tiff)))
				tiff = append(tiff, e.value...)
				if len(tiff)%2 != 0 {
					tiff = append(tiff, 0)
				}
			}
		}
	}
	writeIFD(ifd0Offset, ifd0)
	writeIFD(exifOffset, exifEntries)
	writeIFD(gpsOffset, gpsEntries)
	return tiff
}

func ascii(tag uint16, s string) entry { return entry{tag, 2, uint32(len([]byte(s))), []byte(s)} }
func long(tag uint16, v uint32) entry {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return entry{tag, 4, 1, b}
}
func rationals(tag uint16, values [][2]uint32) entry {
	b := make([]byte, len(values)*8)
	for i, v := range values {
		binary.LittleEndian.PutUint32(b[i*8:], v[0])
		binary.LittleEndian.PutUint32(b[i*8+4:], v[1])
	}
	return entry{tag, 5, uint32(len(values)), b}
}
func hemisphere(v float64, positive, negative string) string {
	if v >= 0 {
		return positive
	}
	return negative
}
func dms(value float64) [][2]uint32 {
	v := math.Abs(value)
	d := math.Floor(v)
	mr := (v - d) * 60
	min := math.Floor(mr)
	sec := (mr - min) * 60
	return [][2]uint32{{uint32(d), 1}, {uint32(min), 1}, {uint32(math.Round(sec * 10000)), 10000}}
}

func buildXMP(m Metadata) []byte {
	escape := func(s string) string { var b bytes.Buffer; _ = xml.EscapeText(&b, []byte(s)); return b.String() }
	x := fmt.Sprintf(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?><x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:exif="http://ns.adobe.com/exif/1.0/" xmlns:field="https://markham.ca/ns/field-photo/1.0/" xmp:CreateDate="%s" dc:creator="%s" exif:GPSLatitude="%.6f" exif:GPSLongitude="%.6f" field:Operator="%s" field:Location="%s" field:LocationSource="%s"/></rdf:RDF></x:xmpmeta><?xpacket end="w"?>`, m.CapturedAt.Format(time.RFC3339), escape(m.User), m.Latitude, m.Longitude, escape(m.User), escape(m.Location), escape(m.Source))
	return []byte(x)
}
