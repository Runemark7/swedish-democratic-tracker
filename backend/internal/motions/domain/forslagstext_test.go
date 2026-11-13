package domain

import "testing"

func TestPlainForslagstext(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"named entities", "Riksdagen a) godk&auml;nner &auml;ndrade ramar f&ouml;r utgiftsomr&aring;den", "Riksdagen a) godkänner ändrade ramar för utgiftsområden"},
		{"capitals", "&Aring;tg&auml;rder, &Auml;ndring, &Ouml;vrigt", "Åtgärder, Ändring, Övrigt"},
		{"numeric entities", "f&#246;rslag &#xE5;r &#229;tta", "förslag år åtta"},
		{"line breaks", "Rubrik<BR/>Riksdagen<br />a) ett,<br>b) två", "Rubrik\nRiksdagen\na) ett,\nb) två"},
		{"non-breaking space kept inside a number", "h&ouml;gst 19&#160;571&#160;000 kronor", "högst 19 571 000 kronor"},
		{"space before a break is dropped", "bilaga 4,&#160;<br />b) bemyndigar", "bilaga 4,\nb) bemyndigar"},
		{"soft hyphen removed", "Tran&#173;sport&#173;styrelsen", "Transportstyrelsen"},
		{"other tags stripped", "<p>lag om <i>&auml;ndring</i></p>", "lag om ändring"},
		{"escaped markup stays text", "villkoret x &lt; 5 &amp; y &gt; 2 &lt;script&gt;", "villkoret x < 5 & y > 2 <script>"},
		{"already plain", "Därmed bifaller riksdagen proposition 2025/26:275", "Därmed bifaller riksdagen proposition 2025/26:275"},
		{"source newlines kept", "motionerna \n\n2025/26:4199 av A (MP)", "motionerna\n\n2025/26:4199 av A (MP)"},
		{"empty", "", ""},
	}
	for _, c := range cases {
		if got := PlainForslagstext(c.in); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
