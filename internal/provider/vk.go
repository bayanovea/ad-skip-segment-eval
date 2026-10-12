package provider

import (
	"fmt"
	"strings"

	"ad-skip-segment-eval/internal/transcript"
	"ad-skip-segment-eval/internal/ytdlp"
)

// VK: native ids look like "-<owner>_<video>"; the page is vk.com/video<id>.
type vk struct{}

func init() { Register(vk{}) }

func (vk) Name() string                    { return "vk" }
func (vk) VideoURL(nativeID string) string { return "https://vk.com/video" + nativeID }

// LinkAt is the vkvideo.ru player link with the start offset as ?t=1h2m3s.
func (vk) LinkAt(nativeID string, seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	h, m, s := seconds/3600, seconds%3600/60, seconds%60
	t := fmt.Sprintf("%dm%ds", m, s)
	if h > 0 {
		t = fmt.Sprintf("%dh%s", h, t)
	} else if m == 0 {
		t = fmt.Sprintf("%ds", s)
	}
	return fmt.Sprintf("https://vkvideo.ru/video%s?t=%s", nativeID, t)
}

// PickTrack prefers an author-uploaded track and falls back to auto-subtitles.
//
// The yt-dlp VK extractor lists two entries per language under `subtitles`: one
// derived from the stream manifest (manifest_url set, ext "vtt", no label) and
// one from the player's subtitle list, where `ext` carries the player's label.
// A label containing "auto" means VK auto-subtitles. A plain "vtt" label is NOT
// evidence of an uploaded track: a video labelled "Автоматические" in the player
// showed up as "vtt" too (checked 2026-10-10), so anything else stays unknown.
func (vk) PickTrack(info *ytdlp.Info, lang string) (t ytdlp.Track, typ transcript.TrackType, auto bool, ok bool) {
	subs := info.Subtitles[lang]
	if len(subs) > 0 {
		typ = transcript.TrackUnknown
		t = subs[0]
		for _, tr := range subs {
			if tr.ManifestURL != "" {
				continue // manifest-derived entry: format only, no label
			}
			t = tr
			if looksAuto(tr) {
				typ = transcript.TrackAuto
			}
			break
		}
		return t, typ, false, true
	}
	if trs := info.AutomaticCaptions[lang]; len(trs) > 0 {
		return trs[0], transcript.TrackAuto, true, true
	}
	return ytdlp.Track{}, transcript.TrackNone, false, false
}

func looksAuto(t ytdlp.Track) bool {
	s := strings.ToLower(t.Name + " " + t.Ext)
	return strings.Contains(s, "auto") || strings.Contains(s, "автомат")
}
