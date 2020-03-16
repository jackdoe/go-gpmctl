# gpmctl
--
    import "github.com/jackdoe/go-gpmctl"

# /dev/gpmctl control socket reader

general purpose mouse daemon (https://linux.die.net/man/8/gpm) gives mouse
support to the linux console

it exposes /dev/gpmctl to which you can connect, send your current vc and pid
and receive mouse events

example:

    package main

    import (
    	"log"

    	gpmctl "github.com/jackdoe/go-gpmctl"
    )

    func main() {
    	g, err := gpmctl.NewGPM(gpm.DefaultConf)
    	if err != nil {
    		panic(err)
    	}
    	for {
    		event, err := g.Read()
    		if err != nil {
    			panic(err)
    		}

    		log.Printf("%s", event)
    	}
    }

    ..
    2020/03/16 23:18:57 type:move[buttons:0, modifiers:0, vc:4] x:190[dx:0] y:28[dy:1], clicks:0 margin:, wdx:0, wdy:0
    2020/03/16 23:18:57 type:move[buttons:0, modifiers:0, vc:4] x:189[dx:-1] y:28[dy:0], clicks:0 margin:, wdx:0, wdy:0
    2020/03/16 23:18:57 type:down,single[buttons:4, modifiers:0, vc:4] x:189[dx:0] y:28[dy:0], clicks:0 margin:, wdx:0, wdy:0
    2020/03/16 23:18:57 type:drag,single,mflag[buttons:4, modifiers:0, vc:4] x:189[dx:0] y:29[dy:1], clicks:0 margin:, wdx:0, wdy:0

..

## Usage

```go
const (
	TOP = 1 << iota
	BOT
	LFT
	RGT
)
```

```go
var DefaultConf = GPMConnect{
	EventMask:   ^uint16(0),
	DefaultMask: ^uint16(0),
	MinMod:      0,
	MaxMod:      ^uint16(0),
}
```

#### func  TTY

```go
func TTY() (int, error)
```
from https://github.com/tudurom/ttyname

#### type Event

```go
type Event struct {
	Buttons   uint8
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
```

typedef struct Gpm_Event {

    unsigned char buttons, modifiers;  /* try to be a multiple of 4 */
    unsigned short vc;
    short dx, dy, x, y; /* displacement x,y for this event, and absolute x,y */
    enum Gpm_Etype type;
    /* clicks e.g. double click are determined by time-based processing */
    int clicks;
    enum Gpm_Margin margin;
    /* wdx/y: displacement of wheels in this event. Absolute values are not
     * required, because wheel movement is typically used for scrolling
     * or selecting fields, not for cursor positioning. The application
     * can determine when the end of file or form is reached, and not
     * go any further.
     * A single mouse will use wdy, "vertical scroll" wheel. */
    short wdx, wdy;

} Gpm_Event;

#### func (Event) String

```go
func (event Event) String() string
```

#### type EventType

```go
type EventType int
```

Gpm Event Type -

    enum Gpm_Etype {
      GPM_MOVE=1,
      GPM_DRAG=2,   /* exactly one of the bare ones is active at a time */
      GPM_DOWN=4
      GPM_UP=  8,

    #define GPM_BARE_EVENTS(type) ((type)&(0x0f|GPM_ENTER|GPM_LEAVE))

      GPM_SINGLE=16,            /* at most one in three is set */
      GPM_DOUBLE=32,
      GPM_TRIPLE=64,            /* WARNING: I depend on the values */

      GPM_MFLAG=128,            /* motion during click? */
      GPM_HARD=256,             /* if set in the defaultMask, force an already
                       used event to pass over to another handler */

      GPM_ENTER=512,            /* enter event, user in Roi's */
      GPM_LEAVE=1024            /* leave event, used in Roi's */
    };

```go
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
```

#### func (EventType) String

```go
func (e EventType) String() string
```

#### type GPM

```go
type GPM struct {
}
```

GPM connection

#### func  NewGPM

```go
func NewGPM(conf GPMConnect) (*GPM, error)
```
Create new gpm connection, it will detect current tty from "/proc/self/fd/0" and
it will use the current pid, then it will connect the /dev/gpmctl stream unix
socket nd send Gpm_Connect struct

#### func (*GPM) Close

```go
func (g *GPM) Close()
```
close the gpm connection

#### func (*GPM) Read

```go
func (g *GPM) Read() (Event, error)
```
Reads one event mouse, or blocks if there are no events NB: some gpm has #define
GPM_MAGIC 0x47706D4C in every message, at the moment that is not supported

#### type GPMConnect

```go
type GPMConnect struct {
	EventMask   uint16
	DefaultMask uint16
	MinMod      uint16
	MaxMod      uint16
}
```

Struct sent via the socket after connecting

    typedef struct Gpm_Connect {
      unsigned short eventMask, defaultMask; // 4
      unsigned short minMod, maxMod;         // 4
      int pid;                               // 4
      int vc;                                // 4
    }              Gpm_Connect;

#### type Margin

```go
type Margin int
```

Gpm Margin Enum

    enum Gpm_Margin {GPM_TOP=1, GPM_BOT=2, GPM_LFT=4, GPM_RGT=8};

#### func (Margin) String

```go
func (m Margin) String() string
```
