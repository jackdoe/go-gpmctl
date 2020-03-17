// # /dev/gpmctl control socket reader
//
// general purpose mouse daemon (https://linux.die.net/man/8/gpm)
// gives mouse support to the linux console
//
// it exposes /dev/gpmctl to which you can connect, send your current vc
// and pid and receive mouse events
//
// example:
//
//   package main
//
//   import (
//   	"log"
//
//   	gpmctl "github.com/jackdoe/go-gpmctl"
//   )
//
//   func main() {
//   	g, err := gpmctl.NewGPM(gpmctl.DefaultConf)
//   	if err != nil {
//   		panic(err)
//   	}
//   	for {
//   		event, err := g.Read()
//   		if err != nil {
//   			panic(err)
//   		}
//
//   		log.Printf("%s", event)
//   	}
//   }
//
//
//   ..
//   2020/03/16 23:18:57 type:move[buttons:, modifiers:0, vc:4] x:190[dx:0] y:28[dy:1], clicks:0 margin:, wdx:0, wdy:0
//   2020/03/16 23:18:57 type:move[buttons:, modifiers:0, vc:4] x:189[dx:-1] y:28[dy:0], clicks:0 margin:, wdx:0, wdy:0
//   2020/03/16 23:18:57 type:down,single[buttons:, modifiers:0, vc:4] x:189[dx:0] y:28[dy:0], clicks:0 margin:, wdx:0, wdy:0
//   2020/03/16 23:18:57 type:drag,single,mflag[buttons:, modifiers:0, vc:4] x:189[dx:0] y:29[dy:1], clicks:0 margin:, wdx:0, wdy:0
// ..
package gpmctl

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"unicode"
	"unsafe"
)

var nativeEndian binary.ByteOrder

func init() {
	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:
		nativeEndian = binary.LittleEndian
	case [2]byte{0xAB, 0xCD}:
		nativeEndian = binary.BigEndian
	default:
		nativeEndian = binary.LittleEndian
	}
}

const fd0 = "/proc/self/fd/0"

// from https://github.com/tudurom/ttyname
func getTTY() (int, error) {
	dest, err := os.Readlink(fd0)
	if err != nil {
		return 0, err
	}
	stty := strings.TrimFunc(path.Base(dest), func(r rune) bool {
		return !unicode.IsDigit(r)
	})
	tty, err := strconv.ParseInt(stty, 10, 64)
	if err != nil {
		return 0, err
	}

	return int(tty), nil
}

// /*....................................... Cfg buttons */
// /* Each button has one bit in the 16 bit buttons field.
//  * Mouse movement and wheel movement are not associated with a button
//  * i.e. buttons=GPM_B_NONE in these cases
//  * (except for ms3 mouse, for reasons unknown?)
//  * The middle button if pressed down (or clicked) is independent of
//  *  the wheel "device" which it happens to be associated with
//  * The use of GPM_B_UP/DOWN with ms3 is unclear. Maybe the wheel
//  * could be rolled forward then backward
//  * and this would generate a 'click' event on 'button 5' GPM_B_UP,
//  * but really the expected behaviour of wheel is movement, typically
//  * used for jump scrolling or for jumping between fields on a form. */
//
// #define GPM_B_DOWN      32
// #define GPM_B_UP        16
// #define GPM_B_FOURTH    8
// #define GPM_B_LEFT      4
// #define GPM_B_MIDDLE    2
// #define GPM_B_RIGHT     1
// #define GPM_B_NONE      0

type Buttons uint8

const (
	B_NONE   Buttons = 0
	B_RIGHT  Buttons = 1
	B_MIDDLE Buttons = 2
	B_LEFT   Buttons = 4
	B_FOURTH Buttons = 8
	B_UP     Buttons = 16
	B_DOWN   Buttons = 32
)

func (b Buttons) String() string {
	s := []string{}
	if b&B_NONE > 0 {
		s = append(s, "none")
	}
	if b&B_RIGHT > 0 {
		s = append(s, "right")
	}
	if b&B_MIDDLE > 0 {
		s = append(s, "middle")
	}
	if b&B_LEFT > 0 {
		s = append(s, "left")
	}
	if b&B_FOURTH > 0 {
		s = append(s, "fourth")
	}
	if b&B_UP > 0 {
		s = append(s, "up")
	}
	if b&B_DOWN > 0 {
		s = append(s, "down")
	}

	return strings.Join(s, ",")
}

// Gpm Event Type - as per gpm.h
//
//  enum Gpm_Etype {
//    GPM_MOVE=1,
//    GPM_DRAG=2,   /* exactly one of the bare ones is active at a time */
//    GPM_DOWN=4
//    GPM_UP=  8,
//
//  #define GPM_BARE_EVENTS(type) ((type)&(0x0f|GPM_ENTER|GPM_LEAVE))
//
//    GPM_SINGLE=16,            /* at most one in three is set */
//    GPM_DOUBLE=32,
//    GPM_TRIPLE=64,            /* WARNING: I depend on the values */
//
//    GPM_MFLAG=128,            /* motion during click? */
//    GPM_HARD=256,             /* if set in the defaultMask, force an already
//                     used event to pass over to another handler */
//
//    GPM_ENTER=512,            /* enter event, user in Roi's */
//    GPM_LEAVE=1024            /* leave event, used in Roi's */
//  };
type EventType uint16

const (
	MOVE EventType = 1 << iota
	DRAG
	DOWN
	UP
	SINGLE
	DOUBLE
	TRIPLE
	MFLAG
	HARD
	ENTER
	LEAVE
)

const ANY EventType = EventType(^uint16(0))

func (e EventType) String() string {
	s := []string{}
	if e&MOVE > 0 {
		s = append(s, "move")
	}
	if e&DRAG > 0 {
		s = append(s, "drag")
	}
	if e&DOWN > 0 {
		s = append(s, "down")
	}
	if e&UP > 0 {
		s = append(s, "up")
	}
	if e&SINGLE > 0 {
		s = append(s, "single")
	}

	if e&DOUBLE > 0 {
		s = append(s, "double")
	}

	if e&TRIPLE > 0 {
		s = append(s, "triple")
	}

	if e&MFLAG > 0 {
		s = append(s, "mflag")
	}

	if e&HARD > 0 {
		s = append(s, "hard")
	}
	if e&ENTER > 0 {
		s = append(s, "enter")
	}
	if e&LEAVE > 0 {
		s = append(s, "leave")
	}

	return strings.Join(s, ",")
}

// Gpm Margin Enum as per gpm.h
//
//   enum Gpm_Margin {GPM_TOP=1, GPM_BOT=2, GPM_LFT=4, GPM_RGT=8};
type Margin int

const (
	TOP = 1 << iota
	BOT
	LFT
	RGT
)

func (m Margin) String() string {
	s := []string{}
	if m&TOP > 0 {
		s = append(s, "top")
	}
	if m&BOT > 0 {
		s = append(s, "bot")
	}
	if m&LFT > 0 {
		s = append(s, "lft")
	}
	if m&RGT > 0 {
		s = append(s, "rgt")
	}
	return strings.Join(s, ",")
}

// Event defined as per gpm.h
//
//  typedef struct Gpm_Event {
//    unsigned char buttons, modifiers;  /* try to be a multiple of 4 */
//    unsigned short vc;
//    short dx, dy, x, y; /* displacement x,y for this event, and absolute x,y */
//    enum Gpm_Etype type;
//    /* clicks e.g. double click are determined by time-based processing */
//    int clicks;
//    enum Gpm_Margin margin;
//    /* wdx/y: displacement of wheels in this event. Absolute values are not
//     * required, because wheel movement is typically used for scrolling
//     * or selecting fields, not for cursor positioning. The application
//     * can determine when the end of file or form is reached, and not
//     * go any further.
//     * A single mouse will use wdy, "vertical scroll" wheel. */
//    short wdx, wdy;
//  } Gpm_Event;
type Event struct {
	Buttons   Buttons
	Modifiers uint8
	VC        uint16
	DX        int16
	DY        int16
	X         int16
	Y         int16
	Type      EventType
	Clicks    int32
	Margin    Margin
	WDX       int16
	WDY       int16
}

func (event Event) String() string {
	return fmt.Sprintf("type:%v[buttons:%s, modifiers:%v, vc:%v] x:%v[dx:%v] y:%v[dy:%v], clicks:%v margin:%v, wdx:%v, wdy:%v",
		event.Type,
		event.Buttons,
		event.Modifiers,
		event.VC,
		event.X, event.DX,
		event.Y, event.DY,
		event.Clicks,
		event.Margin,
		event.WDX,
		event.WDY)
}

// GPM connection
type GPM struct {
	c   net.Conn
	tty int
	pid int
}

// Struct sent via the socket after connecting
//   typedef struct Gpm_Connect {
//     unsigned short eventMask, defaultMask; // 4
//     unsigned short minMod, maxMod;         // 4
//     int pid;                               // 4
//     int vc;                                // 4
//   } Gpm_Connect;
type GPMConnect struct {
	EventMask   EventType
	DefaultMask EventType
	MinMod      uint16
	MaxMod      uint16
}

var DefaultConf = GPMConnect{
	EventMask:   EventType(^uint16(0)),
	DefaultMask: EventType(^uint16(0)),
	MinMod:      0,
	MaxMod:      ^uint16(0),
}

// Create new gpm connection, it will detect current tty from
// "/proc/self/fd/0" and it will use the current pid, then it will
// connect the /dev/gpmctl stream unix socket nd send Gpm_Connect struct
func NewGPM(conf GPMConnect) (*GPM, error) {
	tty, err := getTTY()
	if err != nil {
		return nil, err
	}
	c, err := net.Dial("unix", "/dev/gpmctl")
	if err != nil {
		return nil, err
	}

	pid := os.Getpid()
	gpmConnect := make([]byte, 16)

	nativeEndian.PutUint16(gpmConnect[0:], uint16(conf.EventMask))   // eventmask
	nativeEndian.PutUint16(gpmConnect[2:], uint16(conf.DefaultMask)) // defautmask
	nativeEndian.PutUint16(gpmConnect[4:], conf.MinMod)              // minmod
	nativeEndian.PutUint16(gpmConnect[6:], conf.MaxMod)              // maxmod

	nativeEndian.PutUint32(gpmConnect[8:], uint32(pid))  // pid
	nativeEndian.PutUint32(gpmConnect[12:], uint32(tty)) // vc

	_, err = c.Write(gpmConnect)
	if err != nil {
		c.Close()
		return nil, err
	}
	return &GPM{c: c, tty: tty, pid: pid}, nil
}

// Reads one event mouse, or blocks if there are no events
// NB: some gpm's could have `#define GPM_MAGIC 0x47706D4C` in every message, at the moment that is not supported
func (g *GPM) Read() (Event, error) {
	// sizeof Gpm_Event, this assumes sizeof Gpm_EventType to be 4
	// bytes and sizeof Margin to be 4 bytes, which is not guaranteed
	b := make([]byte, 28)
	_, err := g.c.Read(b)
	if err != nil {
		return Event{}, err
	}
	e := Event{
		Buttons:   Buttons(uint8(nativeEndian.Uint16(b[0:]))),
		Modifiers: uint8(nativeEndian.Uint16(b[1:])),
		VC:        nativeEndian.Uint16(b[2:]),
		DX:        int16(nativeEndian.Uint16(b[4:])),
		DY:        int16(nativeEndian.Uint16(b[6:])),
		X:         int16(nativeEndian.Uint16(b[8:])),
		Y:         int16(nativeEndian.Uint16(b[10:])),
		Type:      EventType(nativeEndian.Uint32(b[12:])),
		Clicks:    int32(nativeEndian.Uint32(b[16:])),
		Margin:    Margin(nativeEndian.Uint32(b[20:])),
		WDX:       int16(nativeEndian.Uint16(b[24:])),
		WDY:       int16(nativeEndian.Uint16(b[26:])),
	}
	return e, nil
}

// close the gpm connection
func (g *GPM) Close() {
	g.c.Close()
}
