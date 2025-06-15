package silog

import (
	"bytes"
	"cmp"
	"context"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
)

// Options defines options for the logger.
type Options struct {
	// Level is the minimum log level to log.
	// It must be one of the supported log levels.
	// The default is LevelInfo.
	Level slog.Leveler // optional
	// TODO: rename to Leveler

	// Style is the style to use for the logger.
	// If unset, the style will be picked based on whether
	// the output is a terminal or not.
	Style *Style // optional

	// TODO: kitchen time by default
}

// Handler is a slog.Handler that writes to an io.Writer
// with colored output.
//
// Output is in a logfmt-style format, with colored levels.
// Other features include:
//
//   - rendering of trace level
//   - multi-line fields are indented and aligned
type Handler struct {
	lvl   slog.Leveler // required
	style *Style       // required
	outMu *sync.Mutex  // required
	out   io.Writer    // required

	// attrs holds attributes that have already been serialized
	// with WithAttrs.
	//
	// This is set only at construction time (e.g. WithAttrs)
	// and not modified afterwards.
	attrs []byte

	// groups is the current group stack.
	groups []string

	// Number of levels to downgrade a log message
	// before writing it.
	lvlOffset int

	// prefix is the prefix to use for the logger.
	prefix string
}

var _ slog.Handler = (*Handler)(nil)

// NewHandler constructs a silog Handler for use with slog.
// Log output is written to the given io.Writer.
//
// The Handler synchronizes writes to the output writer,
// and is safe to use from multiple goroutines.
// Each log message is posted to the output writer
// in a single Writer.Write call.
func NewHandler(w io.Writer, opts *Options) *Handler {
	opts = cmp.Or(opts, &Options{})

	style := opts.Style
	if style == nil {
		if fder, ok := w.(interface {
			Fd() uintptr
		}); ok && isatty.IsTerminal(fder.Fd()) {
			style = DefaultStyle()
		} else {
			style = PlainStyle()
		}
	}

	lvl := opts.Level
	if lvl == nil {
		lvl = slog.LevelInfo // default level
	}

	return &Handler{
		lvl:   lvl,
		style: style,
		out:   w,
		outMu: new(sync.Mutex),
	}
}

func (h *Handler) Enabled(_ context.Context, lvl slog.Level) bool {
	lvl += slog.Level(h.lvlOffset)
	return h.lvl.Level() <= lvl
}

const (
	lvlDelim     = " "  // separator between level and message
	groupDelim   = "."  // separator between group names
	msgAttrDelim = "  " // separator between message and attributes
	attrDelim    = " "  // separator between attributes
	indent       = "  " // indentation for multi-line attributes
)

func (h *Handler) Handle(_ context.Context, rec slog.Record) error {
	bs := *takeBuf()
	defer releaseBuf(&bs)

	rec.Level += slog.Level(h.lvlOffset)
	lvlString := h.style.LevelLabels[rec.Level].String()
	// TODO: Use empty string with no space for unknown level

	// If the message is multi-line, we'll need to prepend the level
	// to each line.
	for line := range strings.Lines(rec.Message) {
		bs = append(bs, lvlString...)
		bs = append(bs, lvlDelim...)

		var msg bytes.Buffer
		if h.prefix != "" {
			msg.WriteString(h.prefix)
			msg.WriteString(h.style.PrefixDelimiter.Render())
		}

		// line may end with \n.
		// That should not be included in the rendering logic.
		var trailingNewline bool
		if line[len(line)-1] == '\n' {
			trailingNewline = true
			line = line[:len(line)-1]
		}
		msg.WriteString(line)

		line = h.style.Messages[rec.Level].Render(msg.String())
		bs = append(bs, line...)
		if trailingNewline {
			bs = append(bs, '\n')
		}
	}

	// First attribute after the message is separated by two spaces.
	bs = append(bs, msgAttrDelim...)

	// withAttrs attributes are serialized into the buffer
	if len(h.attrs) > 0 {
		bs = append(bs, h.attrs...)
	}

	// Write the attributes.
	formatter := attrFormatter{
		buf:    bs,
		style:  h.style,
		groups: slices.Clone(h.groups),
	}
	rec.Attrs(func(attr slog.Attr) bool {
		formatter.FormatAttr(attr)
		return true
	})
	bs = formatter.buf

	// Always a single trailing newline.
	bs = append(bytes.TrimRight(bs, " \n"), '\n')

	h.outMu.Lock()
	defer h.outMu.Unlock()
	_, err := h.out.Write(bs)
	return err
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	f := attrFormatter{
		buf:    slices.Clone(h.attrs),
		groups: slices.Clone(h.groups),
		style:  h.style,
	}
	for _, attr := range attrs {
		f.FormatAttr(attr)
	}
	bs := f.buf

	newL := *h
	newL.attrs = bs
	return &newL
}

func (h *Handler) WithGroup(name string) slog.Handler {
	newL := *h
	newL.groups = append(slices.Clone(h.groups), name)
	return &newL
}

// WithLevel returns a new handler with the given leveler,
// retaining all other attributes and groups.
//
// It will write to the same output writer as this handler.
func (h *Handler) WithLevel(lvl slog.Leveler) slog.Handler {
	newL := *h
	newL.lvl = lvl
	return &newL
}

// WithPrefix returns a new handler with the given prefix
func (l *Handler) WithPrefix(prefix string) slog.Handler {
	newL := *l
	newL.prefix = prefix
	return &newL
}

// TODO: rename to WithLevelOffset?
// TODO: or a general level remapper
// that includes a std level downgrader
func (l *Handler) WithDowngrade(n int) slog.Handler {
	newL := *l
	newL.lvlOffset -= n
	return &newL
}

type attrFormatter struct {
	buf    []byte
	style  *Style
	groups []string
}

func (f *attrFormatter) FormatAttr(attr slog.Attr) {
	if attr.Equal(slog.Attr{}) {
		return // skip empty attributes
	}

	value := attr.Value.Resolve()
	if value.Kind() == slog.KindGroup {
		// Groups just get splatted into their attributes
		// prefixed with the group name.
		f.groups = append(f.groups, attr.Key)
		for _, a := range value.Group() {
			f.FormatAttr(a)
		}
		f.groups = f.groups[:len(f.groups)-1]
		return
	}

	// We serialize the attribute into a byte slice,
	// and then decide how it goes into the output.
	// This is because we need to handle multi-line attributes
	// and indent them.
	valbs := *takeBuf()
	defer releaseBuf(&valbs)

	switch value.Kind() {
	case slog.KindBool:
		valbs = strconv.AppendBool(valbs, value.Bool())
	case slog.KindDuration:
		valbs = append(valbs, value.Duration().String()...)
	case slog.KindFloat64:
		valbs = strconv.AppendFloat(valbs, value.Float64(), 'g', -1, 64)
	case slog.KindInt64:
		valbs = strconv.AppendInt(valbs, value.Int64(), 10)
	case slog.KindString:
		valbs = append(valbs, value.String()...)
	case slog.KindTime:
		valbs = value.Time().AppendFormat(valbs, time.Kitchen)
	case slog.KindUint64:
		valbs = strconv.AppendUint(valbs, value.Uint64(), 10)
	default:
		// TODO: reflection to handle structs, maps, slices, etc.
		valbs = append(valbs, value.String()...)
	}

	// Add delimiter between attrs.
	if len(f.buf) > 0 {
		switch {
		case f.buf[len(f.buf)-1] == '\n':
			// If the last thing we wrote was multi-line,
			// then we need to indent the next attribute.
			f.buf = append(f.buf, indent...)
		case f.buf[len(f.buf)-1] != ' ':
			// All other attributes are separated by a space.
			f.buf = append(f.buf, attrDelim...)
		}
	}

	// Single-line attributes are rendered as:
	//
	//   key=value
	//
	// Multi-line attributes are rendered as:
	//
	//   key=
	//     | line 1
	//     | line 2
	isMultiline := bytes.ContainsAny(valbs, "\r\n")
	if isMultiline {
		f.buf = append(f.buf, '\n')
		f.buf = append(f.buf, indent...)
	}

	f.formatKey(attr.Key)
	f.buf = append(f.buf, f.style.KeyValueDelimiter.Render()...) // =

	valueStyle, hasStyle := f.style.Values[attr.Key]
	if isMultiline {
		prefixStyle := f.style.MultilinePrefix
		if hasStyle {
			prefixStyle = prefixStyle.Foreground(valueStyle.GetForeground())
		}
		prefix := indent + prefixStyle.Render()

		// TODO: \r handling
		f.buf = append(f.buf, '\n')
		for line := range bytes.Lines(valbs) {
			f.buf = append(f.buf, prefix...)
			line = bytes.TrimRight(line, "\r\n")
			if hasStyle {
				f.buf = append(f.buf, valueStyle.Render(string(line))...)
			} else {
				f.buf = append(f.buf, line...)
			}
			f.buf = append(f.buf, '\n')
		}

		// If multi-line attribute value does not end with a newline,
		// add one.
		if f.buf[len(f.buf)-1] != '\n' {
			f.buf = append(f.buf, '\n')
		}
	} else {
		if hasStyle {
			f.buf = append(f.buf, valueStyle.Render(string(valbs))...)
		} else {
			f.buf = append(f.buf, valbs...)
		}
	}
}

// formatKey writes a group-prefixed key to the buffer.
func (f *attrFormatter) formatKey(key string) {
	for _, group := range f.groups {
		if group != "" {
			f.buf = append(f.buf, f.style.Key.Render(group)...)
			f.buf = append(f.buf, groupDelim...)
		}
	}
	f.buf = append(f.buf, f.style.Key.Render(key)...)
}

var _bufPool = &sync.Pool{
	New: func() any {
		bs := make([]byte, 0, 1024)
		return &bs
	},
}

func takeBuf() *[]byte {
	bs := _bufPool.Get().(*[]byte)
	*bs = (*bs)[:0]
	return bs
}

func releaseBuf(bs *[]byte) {
	_bufPool.Put(bs)
}
