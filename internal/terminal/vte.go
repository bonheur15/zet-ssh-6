package terminal

/*
#cgo pkg-config: gtk4 vte-2.91-gtk4
#include <vte/vte.h>
#include <stdlib.h>
#include <string.h>

extern void goOnChildExited(void *termPtr, int status);
extern void goOnWindowTitleChanged(void *termPtr, char *title);

static void c_child_exited_cb(VteTerminal *terminal, gint status, gpointer user_data) {
	goOnChildExited(terminal, (int)status);
}

static void c_window_title_changed_cb(VteTerminal *terminal, gpointer user_data) {
	const char *title = vte_terminal_get_window_title(terminal);
	goOnWindowTitleChanged(terminal, (char*)(title ? title : ""));
}

static void connect_vte_signals(VteTerminal *terminal) {
	g_signal_connect(terminal, "child-exited", G_CALLBACK(c_child_exited_cb), NULL);
	g_signal_connect(terminal, "window-title-changed", G_CALLBACK(c_window_title_changed_cb), NULL);
}

static void spawn_shell(VteTerminal *term, const char *shell_path, const char *working_dir) {
	char *argv[] = {(char*)shell_path, NULL};
	vte_terminal_spawn_async(
		term,
		VTE_PTY_DEFAULT,
		working_dir,
		argv,
		NULL, // inherit environment
		G_SPAWN_DEFAULT,
		NULL, NULL, NULL, // child setup
		-1, // timeout
		NULL, // cancellable
		NULL, // spawn callback (not strictly needed as we use child-exited signal)
		NULL  // user_data
	);
}

static void set_terminal_colors(VteTerminal *terminal, const char *fg_hex, const char *bg_hex, char **palette_hex, int palette_size) {
	GdkRGBA fg, bg;
	gdk_rgba_parse(&fg, fg_hex);
	gdk_rgba_parse(&bg, bg_hex);

	GdkRGBA *palette = g_new0(GdkRGBA, palette_size);
	for (int i = 0; i < palette_size; i++) {
		gdk_rgba_parse(&palette[i], palette_hex[i]);
	}

	vte_terminal_set_colors(terminal, &fg, &bg, palette, palette_size);
	g_free(palette);
}

static void set_terminal_font(VteTerminal *terminal, const char *font_name, int font_size) {
	char font_desc_str[256];
	snprintf(font_desc_str, sizeof(font_desc_str), "%s %d", font_name, font_size);
	PangoFontDescription *requested = pango_font_description_from_string(font_desc_str);
	PangoContext *context = gtk_widget_get_pango_context(GTK_WIDGET(terminal));

	PangoFontDescription *resolved = requested;
	char *resolved_family = NULL;
	PangoFont *loaded = pango_context_load_font(context, requested);
	if (loaded != NULL) {
		PangoFontDescription *loaded_desc = pango_font_describe(loaded);
		if (loaded_desc != NULL) {
			const char *family = pango_font_description_get_family(loaded_desc);
			if (family != NULL) {
				resolved_family = g_strdup(family);
			}
			pango_font_description_free(loaded_desc);
		}
		g_object_unref(loaded);
	}

	gboolean family_is_monospace = FALSE;
	if (resolved_family != NULL) {
		PangoFontFamily **families = NULL;
		int n_families = 0;
		pango_context_list_families(context, &families, &n_families);
		for (int i = 0; i < n_families; i++) {
			const char *family_name = pango_font_family_get_name(families[i]);
			if (family_name != NULL && strcmp(family_name, resolved_family) == 0) {
				family_is_monospace = pango_font_family_is_monospace(families[i]);
				break;
			}
		}
		g_free(families);
	}

	if (!family_is_monospace) {
		resolved = pango_font_description_from_string("monospace 11");
		pango_font_description_set_size(resolved, font_size * PANGO_SCALE);
	}

	vte_terminal_set_font(terminal, resolved);
	vte_terminal_set_cell_width_scale(terminal, 1.0);
	vte_terminal_set_cell_height_scale(terminal, 1.0);

	if (resolved != requested) {
		pango_font_description_free(resolved);
	}
	g_free(resolved_family);
	pango_font_description_free(requested);
}

static void terminal_copy(VteTerminal *terminal) {
	vte_terminal_copy_clipboard_format(terminal, VTE_FORMAT_TEXT);
}

static void terminal_paste(VteTerminal *terminal) {
	// Emit the paste-clipboard signal on the terminal
	g_signal_emit_by_name(terminal, "paste-clipboard", NULL);
}
*/
import "C"

import (
	"unsafe"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// Active terminal mapping for callbacks
var activeTerminals = make(map[uintptr]*VteTerminalInstance)

type VteTerminalInstance struct {
	Widget               *gtk.Widget
	onChildExited        func(status int)
	onWindowTitleChanged func(title string)
}

//export goOnChildExited
func goOnChildExited(termPtr unsafe.Pointer, status C.int) {
	ptr := uintptr(termPtr)
	if inst, exists := activeTerminals[ptr]; exists && inst.onChildExited != nil {
		inst.onChildExited(int(status))
	}
}

//export goOnWindowTitleChanged
func goOnWindowTitleChanged(termPtr unsafe.Pointer, title *C.char) {
	ptr := uintptr(termPtr)
	if inst, exists := activeTerminals[ptr]; exists && inst.onWindowTitleChanged != nil {
		inst.onWindowTitleChanged(C.GoString(title))
	}
}

func NewVteTerminal() *VteTerminalInstance {
	termPtr := C.vte_terminal_new()
	gobj := glib.Take(unsafe.Pointer(termPtr))
	widget := gobj.Cast().(*gtk.Widget)
	
	C.connect_vte_signals((*C.VteTerminal)(unsafe.Pointer(termPtr)))

	inst := &VteTerminalInstance{
		Widget: widget,
	}
	
	activeTerminals[uintptr(unsafe.Pointer(termPtr))] = inst
	return inst
}

func (t *VteTerminalInstance) Destroy() {
	ptr := uintptr(unsafe.Pointer(t.Widget.Object.Native()))
	delete(activeTerminals, ptr)
}

func (t *VteTerminalInstance) SpawnShell(shellPath, workingDir string) {
	termPtr := (*C.VteTerminal)(unsafe.Pointer(t.Widget.Object.Native()))
	
	cShell := C.CString(shellPath)
	defer C.free(unsafe.Pointer(cShell))
	
	var cDir *C.char
	if workingDir != "" {
		cDir = C.CString(workingDir)
		defer C.free(unsafe.Pointer(cDir))
	}
	
	C.spawn_shell(termPtr, cShell, cDir)
}

func (t *VteTerminalInstance) SetFont(fontName string, fontSize int) {
	if fontName == "" {
		return
	}
	termPtr := (*C.VteTerminal)(unsafe.Pointer(t.Widget.Object.Native()))
	cFont := C.CString(fontName)
	defer C.free(unsafe.Pointer(cFont))
	C.set_terminal_font(termPtr, cFont, C.int(fontSize))
}

func (t *VteTerminalInstance) SetColors(fg, bg string, palette []string) {
	termPtr := (*C.VteTerminal)(unsafe.Pointer(t.Widget.Object.Native()))
	
	cFg := C.CString(fg)
	cBg := C.CString(bg)
	defer C.free(unsafe.Pointer(cFg))
	defer C.free(unsafe.Pointer(cBg))
	
	cPalette := make([]*C.char, len(palette))
	for i, color := range palette {
		cPalette[i] = C.CString(color)
	}
	defer func() {
		for _, ptr := range cPalette {
			C.free(unsafe.Pointer(ptr))
		}
	}()
	
	C.set_terminal_colors(termPtr, cFg, cBg, &cPalette[0], C.int(len(palette)))
}

func (t *VteTerminalInstance) SetScrollbackLines(lines int) {
	termPtr := (*C.VteTerminal)(unsafe.Pointer(t.Widget.Object.Native()))
	C.vte_terminal_set_scrollback_lines(termPtr, C.long(lines))
}

func (t *VteTerminalInstance) SetCursorBlinkMode(mode int) {
	termPtr := (*C.VteTerminal)(unsafe.Pointer(t.Widget.Object.Native()))
	C.vte_terminal_set_cursor_blink_mode(termPtr, C.VteCursorBlinkMode(mode))
}

func (t *VteTerminalInstance) SetCursorShape(shape int) {
	termPtr := (*C.VteTerminal)(unsafe.Pointer(t.Widget.Object.Native()))
	C.vte_terminal_set_cursor_shape(termPtr, C.VteCursorShape(shape))
}

func (t *VteTerminalInstance) Copy() {
	termPtr := (*C.VteTerminal)(unsafe.Pointer(t.Widget.Object.Native()))
	C.terminal_copy(termPtr)
}

func (t *VteTerminalInstance) Paste() {
	termPtr := (*C.VteTerminal)(unsafe.Pointer(t.Widget.Object.Native()))
	C.terminal_paste(termPtr)
}

func (t *VteTerminalInstance) OnChildExited(fn func(status int)) {
	t.onChildExited = fn
}

func (t *VteTerminalInstance) OnWindowTitleChanged(fn func(title string)) {
	t.onWindowTitleChanged = fn
}

func ActiveTerminals() map[uintptr]*VteTerminalInstance {
	return activeTerminals
}
