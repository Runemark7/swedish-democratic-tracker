package http

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
	"time"

	"riksdagskollen/internal/loften/domain"
)

// csvHeader is the column order of the Löften download. The names are the
// ones the page uses, so the file reads the same as the list it came from.
var csvHeader = []string{
	"Id", "Parti", "Löfte", "Löftesstatus", "Status ändrad", "Mandatperiod", "Utskott", "Ämne",
	"Citat", "Plats i källan", "Källdokument", "Källa (URL)", "Arkiverat",
}

// listCSV serves the same selection as the JSON list as a file a spreadsheet
// opens directly: UTF-8 with a byte order mark (Excel otherwise misreads å, ä
// and ö), semicolon-separated (what Swedish-locale Excel expects; LibreOffice
// and Google Sheets detect it), CRLF line endings.
//
// The list is not paginated and neither is the file: public Löften are a
// curated set of a few hundred rows, and the file never holds more than the
// JSON list for the same filters already does.
func (h *Handler) listCSV(w http.ResponseWriter, r *http.Request) {
	out, ok := h.selection(w, r)
	if !ok {
		return
	}
	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF")
	cw := csv.NewWriter(&buf)
	cw.Comma = ';'
	cw.UseCRLF = true
	_ = cw.Write(csvHeader)
	for _, l := range out {
		_ = cw.Write(csvRow(l))
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="riksdagskollen-loften-`+time.Now().UTC().Format("2006-01-02")+`.csv"`)
	_, _ = w.Write(buf.Bytes())
}

func csvRow(l *domain.LoefteView) []string {
	// An archived Löfte has no Löftesstatus; the cell stays empty.
	status, changed := "", ""
	if l.Status != nil {
		status = string(*l.Status)
	}
	if l.StatusChangedAt != nil && len(*l.StatusChangedAt) >= 10 {
		changed = (*l.StatusChangedAt)[:10]
	}
	archived := "nej"
	if l.Archived {
		archived = "ja"
	}
	return []string{
		strconv.Itoa(l.ID), l.Party, csvText(l.Text), status, changed, l.MandatePeriod,
		csvText(strings.Join(l.Utskott, ", ")), csvText(l.Topic), csvText(l.Quote), csvText(l.Locator),
		csvText(l.SourceDoc), csvText(l.SourceURL), archived,
	}
}

// csvText keeps a spreadsheet from running free text as a formula: a cell that
// starts with =, +, -, @, tab or carriage return gets a leading apostrophe.
func csvText(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}
