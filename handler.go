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
)

// HandlerOptions defines options for constructing a [Handler].
type HandlerOptions struct {
	// Level is the minimum log level to log.
	// It must be one of the supported log levels.
	// The default is LevelInfo.
	Level slog.Leveler // optional

	// Style is the style to use for the logger.
	// If unset, [DefaultStyle] is used.
	// You may use [PlainStyle] to get output with no colors.
	Style *Style // optional

	// TimeFormat is the format to use when rendering timestamps.
	// If unset, time.Kitchen will be used.
	TimeFormat string // optional

	// ReplaceAttr, if set, is called for each attribute
	// before it is rendered.
	//
	// For time and level, it is called with slog.TimeKey and slog.LevelKey
	// respectively.
	// It is not called if the associated time for the record is zero.
	ReplaceAttr func(groups []string, attr slog.Attr) slog.Attr // optional
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

	// timeFormat is the format to use when rendering timestamps.
	timeFormat string

	// replaceAttr is the attribute replacement function.
	replaceAttr func([]string, slog.Attr) slog.Attr
}

var _ slog.Handler = (*Handler)(nil)

// NewHandler constructs a silog Handler for use with slog.
// Log output is written to the given io.Writer.
//
// The Handler synchronizes writes to the output writer,
// and is safe to use from multiple goroutines.
// Each log message is posted to the output writer
// in a single Writer.Write call.
func NewHandler(w io.Writer, opts *HandlerOptions) *Handler {
	opts = cmp.Or(opts, &HandlerOptions{})
	style := cmp.Or(opts.Style, DefaultStyle())
	timeFormat := cmp.Or(opts.TimeFormat, time.Kitchen)

	lvl := opts.Level
	if lvl == nil {
		lvl = slog.LevelInfo // default level
	}

	return &Handler{
		lvl:         lvl,
		style:       style,
		out:         w,
		outMu:       new(sync.Mutex),
		timeFormat:  timeFormat,
		replaceAttr: opts.ReplaceAttr,
	}
}

// Enabled reports whether the handler is enabled for the given level.
//
// If Enabled returnsf alse, Handle should not be called for a record
// at that level.
func (h *Handler) Enabled(_ context.Context, lvl slog.Level) bool {
	lvl += slog.Level(h.lvlOffset)
	return h.lvl.Level() <= lvl
}

const (
	timeDelim    = " "  // separator between time and level
	lvlDelim     = " "  // separator between level and message
	groupDelim   = "."  // separator between group names
	msgAttrDelim = "  " // separator between message and attributes
	attrDelim    = " "  // separator between attributes
	indent       = "  " // indentation for multi-line attributes
)

// Handle writes the given log record to the output writer.
//
// The write is synchronized with a mutex,
// so that multiple copies of the handler
// (e.g. those made with WithAttrs, WithPrefix, etc.)
// can be used concurrently without issues
// as long as they all are built from the same base handler.
func (h *Handler) Handle(_ context.Context, rec slog.Record) error {
	bs := *takeBuf()
	defer releaseBuf(&bs)

	// Level
	lvl := rec.Level + slog.Level(h.lvlOffset)
	var lvlString string
	if h.replaceAttr == nil {
		lvlString = h.style.LevelLabels[rec.Level].String()
	} else {
		attr := h.replaceAttr(nil, slog.Any(slog.LevelKey, lvl))
		if !attr.Equal(slog.Attr{}) {
			if lvl, ok := attr.Value.Any().(slog.Level); ok {
				// If the value is a known slog.Level,
				// we can use the level label from the style.
				lvlString = h.style.LevelLabels[lvl].String()
			} else {
				// Otherwise, just use the string representation.
				lvlString = attr.Value.String()
				// TODO: silog.Styled(lipgloss.Style, slog.Attr)
			}
		}
	}

	// Time
	var timeString string
	if !rec.Time.IsZero() {
		if h.replaceAttr == nil {
			timeString = rec.Time.Format(h.timeFormat)
		} else {
			timeAttr := h.replaceAttr(nil, slog.Time(slog.TimeKey, rec.Time))
			switch {
			case timeAttr.Equal(slog.Attr{}):
				// Skip the time.

			case timeAttr.Value.Kind() == slog.KindTime:
				// If the value is a time, format it with TimeFormat.
				timeString = timeAttr.Value.Time().Format(h.timeFormat)

			default:
				// Otherwise, just use the string representation of the value.
				timeString = timeAttr.Value.String()
			}
		}
	}
	if timeString != "" {
		timeString = h.style.Time.Render(timeString)
	}

	// If the message is multi-line,
	// we'll need to prepend the level and time to each line.
	for line := range strings.Lines(rec.Message) {
		if timeString != "" {
			bs = append(bs, timeString...)
			bs = append(bs, timeDelim...)
		}
		if lvlString != "" {
			bs = append(bs, lvlString...)
			bs = append(bs, lvlDelim...)
		}

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
	formatter := h.attrFormatter(bs)
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

// WithAttrs returns a copy of this handler
// that will always include the given slog attributes
// in its output.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	f := h.attrFormatter(slices.Clone(h.attrs))
	for _, attr := range attrs {
		f.FormatAttr(attr)
	}
	bs := f.buf

	newH := *h
	newH.attrs = bs
	return &newH
}

// WithGroup returns a copy of this handler
// that will always group the attributes that follow
// under the given group name.
func (h *Handler) WithGroup(name string) slog.Handler {
	newH := *h
	newH.groups = append(slices.Clone(h.groups), name)
	return &newH
}

// WithLevel returns a new handler with the given leveler,
// retaining all other attributes and groups.
//
// It will write to the same output writer as this handler.
func (h *Handler) WithLevel(lvl slog.Leveler) *Handler {
	newH := *h
	newH.lvl = lvl
	return &newH
}

// SetPrefix returns a copy of this handler
// that will use the given prefix for each log message.
//
// If the handler already has a prefix,
// this will replace it with the new prefix.
func (h *Handler) SetPrefix(prefix string) *Handler {
	newH := *h
	newH.prefix = prefix
	return &newH
}

// Prefix returns the current prefix for this handler, if any.
func (h *Handler) Prefix() string {
	return h.prefix
}

// WithLevelOffset returns a copy of this handler
// that will offset the log level by the given number of levels
// before writing it.
//
// Levels defined in log/slog are 4-levels apart,
// so you can use 4, or -4 to upgrade or downgrade log levels.
// For example:
//
//	handler = handler.WithLevelOffset(-4)
//
// This will result in the following remapping:
//
//	slog.LevelError -> slog.LevelWarn
//	slog.LevelWarn  -> slog.LevelInfo
//	slog.LevelInfo  -> slog.LevelDebug
//	slog.LevelDebug -> slog.LevelDebug - 4
//
// Any existing level offset is retained, so this operation is additive.
func (h *Handler) WithLevelOffset(n int) *Handler {
	newH := *h
	newH.lvlOffset += n
	return &newH
}

// LevelOffset returns the current level offset for this handler, if any.
func (h *Handler) LevelOffset() int {
	return h.lvlOffset
}

type attrFormatter struct {
	buf    []byte
	style  *Style
	groups []string

	replaceAttr func([]string, slog.Attr) slog.Attr
}

func (h *Handler) attrFormatter(buf []byte) *attrFormatter {
	return &attrFormatter{
		buf:         buf,
		style:       h.style,
		groups:      slices.Clone(h.groups),
		replaceAttr: h.replaceAttr,
	}
}

func (f *attrFormatter) FormatAttr(attr slog.Attr) {
	attr.Value = attr.Value.Resolve()
	if f.replaceAttr != nil {
		attr = f.replaceAttr(f.groups, attr)
	}

	if attr.Equal(slog.Attr{}) {
		return // skip empty attributes
	}

	value := attr.Value
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
		prefixStyle := f.style.MultilineValuePrefix
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
