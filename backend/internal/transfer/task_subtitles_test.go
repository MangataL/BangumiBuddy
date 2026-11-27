package transfer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/magnet"
	"github.com/MangataL/BangumiBuddy/internal/notice"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle/ass"
)

type taskManagerStub struct {
	magnet.Interface
	task magnet.Task
}

func TestGenerateSubsetWhenCompanionExists(t *testing.T) {
	const original = "[V4+ Styles]\nFormat: Name, Fontname, Bold, Italic\nStyle: Default,多言語テスト 標準,0,0\n[Events]\nFormat: Style, Text\nDialogue: Default,A"
	const embedded = "\n[Fonts]\nfontname: existing.ttf\n!!!!\n"
	fonts := ass.NewFontMetaSet("../../pkg/subtitle/ass/testdata", ass.NewMemoryRepository())
	require.NoError(t, fonts.Init(t.Context(), false))
	for _, tc := range []struct {
		name, stem, suffix, subtitle, companion string
		cover                                   bool
	}{
		{"empty companion", "01", ".ass", original, "", false},
		{"regenerate embedded subtitle", "01", ".ass", original + embedded, strings.Replace(original, "Default,A", "Default,B", 1) + embedded, true},
		{"dotted video name", "show.01", ".ass", original, "old subset", false},
		{"language and track flags", "Series.Name.S01E01", ".zh-CN.forced.ass", original, "old subset", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			root, library := t.TempDir(), t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(root, tc.stem+tc.suffix), []byte(tc.subtitle), 0600))
			require.NoError(t, os.WriteFile(filepath.Join(root, tc.stem+".subset"+tc.suffix), []byte(tc.companion), 0600))
			subsetter := ass.NewSubsetter(fonts, ass.FontSubsetterConfig{GenerateNewFile: true, CoverExistSubFont: tc.cover})
			media := []taskMedia{{meta: Meta{FilePath: filepath.Join(root, tc.stem+".mkv")}}}
			require.Empty(t, prepareTaskSubtitles(ctx, root, media, subsetter))
			RegisterFileTransfer(t.Name(), taskTestLink{})
			tr := Transfer{config: Config{EnableSubtitleSubset: true, TransferType: t.Name()}, fontSubsetter: subsetter}
			require.NoError(t, tr.transferRelatedFiles(ctx, media[0].meta, filepath.Join(library, "S01E01")))
			data, err := os.ReadFile(filepath.Join(library, "S01E01.subset"+tc.suffix))
			require.NoError(t, err)
			require.Contains(t, string(data), "Dialogue: Default,A")
			require.Contains(t, string(data), "[Fonts]")
			require.NotEqual(t, tc.companion, string(data))
		})
	}
}

func (m taskManagerStub) GetTask(context.Context, string) (magnet.Task, error) { return m.task, nil }

type taskNoticeStub struct {
	notice.Notifier
	last notice.NoticeTaskTransferredReq
}

func (n *taskNoticeStub) NoticeTaskTransferred(_ context.Context, req notice.NoticeTaskTransferredReq) error {
	n.last = req
	return nil
}

type taskTestLink struct{}

func (taskTestLink) Transfer(_ context.Context, src, dst string) (string, error) {
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return "", err
	}
	return src, os.Link(src, dst)
}

func TestMagnetTaskMissingFonts(t *testing.T) {
	const header = "[V4+ Styles]\nFormat: Name, Fontname, Bold, Italic\nStyle: Default,Missing A,0,0\n[Events]\nFormat: Style, Text\nDialogue: Default,"
	for _, mode := range []string{"strict", "ignore", "disabled"} {
		t.Run(mode, func(t *testing.T) {
			ctx := t.Context()
			ctrl := gomock.NewController(t)
			root, library := t.TempDir(), t.TempDir()
			write := func(name, content string) {
				require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(content), 0600))
			}
			for _, name := range []string{"01.mkv", "02.mkv", "03.mkv", "skip.mkv"} {
				write(name, "video")
			}
			write("01.sc.ass", header+`A{\fnMissing B\i1}B`)
			write("01.tc.ssa", header+`A{\b1}B{\i1}C{\b650}D{\i0\b600}E`)
			write("02.ass", header+"A")
			write("01.subset.sc.ass", "stale subset")
			const supplied = "[Fonts]\nfontname: supplied.ttf\n"
			write("02.subset.ass", supplied)
			write("skip.ass", strings.Replace(header, "Missing A", "Unselected", 1)+"A")
			write("03.ass", strings.Replace(header, "Missing A", "多言語テスト 標準", 1)+"A")
			fontDir := t.TempDir()
			font, err := os.ReadFile("../../pkg/subtitle/ass/testdata/multilingual.ttf")
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(fontDir, "font.ttf"), font, 0600))
			global := ass.NewFontMetaSet(fontDir, ass.NewMemoryRepository())
			require.NoError(t, global.Init(ctx, false))
			torrent := downloader.Torrent{Path: root, Hash: "task", FileNames: []string{"03.mkv", "01.mkv", "02.mkv", "skip.mkv"}}
			task := magnet.Task{DownloadType: downloader.DownloadTypeTV}
			for i, n := range torrent.FileNames {
				task.Torrent.Files = append(task.Torrent.Files, magnet.TorrentFile{FileName: n, Episode: i + 1, Media: true, Download: n != "skip.mkv"})
			}
			task.Torrent.Files[2].Meta = &magnet.TorrentFileMeta{MediaType: downloader.DownloadTypeMovie}
			torrent.FileNames = append(torrent.FileNames, "02.subset.ass")
			task.Torrent.Files = append(task.Torrent.Files, magnet.TorrentFile{FileName: "02.subset.ass", Download: true})
			registry := NewMockTransferFilesRepo(ctrl)
			wantStatus := downloader.TorrentStatusTransferredError
			if mode != "strict" {
				wantStatus = downloader.TorrentStatusTransferred
				registry.EXPECT().Get(gomock.Any(), gomock.Any()).Return(FileTransferred{}, ErrFileTransferredNotFound).Times(3)
				registry.EXPECT().Set(gomock.Any(), gomock.Any()).Return(nil).Times(3)
			}
			var detail string
			operator := downloader.NewMockTorrentOperator(ctrl)
			operator.EXPECT().SetTorrentStatus(ctx, torrent.Hash, wantStatus, gomock.Any(), gomock.Any()).Do(
				func(_ context.Context, _ string, _ downloader.TorrentStatus, s string, _ *downloader.SetTorrentStatusOptions) {
					detail = s
				},
			).Return(nil)
			captured := &taskNoticeStub{}
			RegisterFileTransfer(t.Name(), taskTestLink{})
			tr := &Transfer{
				config:        Config{EnableSubtitleSubset: mode != "disabled", IgnoreSubsetError: mode == "ignore", TransferType: t.Name(), TVPath: library, MoviePath: library, TVFormat: "{origin_name}", MovieFormat: "{origin_name}"},
				fontSubsetter: ass.NewSubsetter(global, ass.FontSubsetterConfig{GenerateNewFile: true}),
				magnetManager: taskManagerStub{task: task}, notifier: captured, transferFiles: registry, torrentOperator: operator,
			}
			err = tr.transferTorrent(ctx, torrent)
			contents, readErr := os.ReadDir(library)
			require.NoError(t, readErr)
			if mode == "strict" {
				require.ErrorIs(t, err, subtitle.ErrFontNotFound)
				require.EqualError(t, captured.last.Error, detail)
				require.Empty(t, captured.last.MediaFilePaths)
				require.Empty(t, contents)
				require.NoFileExists(t, filepath.Join(root, "03.subset.ass"))
			} else {
				require.NoError(t, err)
				require.Len(t, captured.last.MediaFilePaths, 3)
				require.NoError(t, captured.last.Error)
				require.Equal(t, detail, captured.last.Warning)
				data, err := os.ReadFile(filepath.Join(library, "01.sc.ass"))
				require.NoError(t, err)
				require.Equal(t, header+`A{\fnMissing B\i1}B`, string(data))
				subset, err := os.ReadFile(filepath.Join(library, "02.subset.ass"))
				require.NoError(t, err)
				require.Equal(t, supplied, string(subset))
				require.NoFileExists(t, filepath.Join(library, "skip.mkv"))
				if mode == "disabled" {
					require.Empty(t, detail)
					require.NoFileExists(t, filepath.Join(library, "03.subset.ass"))
					return
				}
				existing, err := os.ReadFile(filepath.Join(library, "01.subset.sc.ass"))
				require.NoError(t, err)
				require.Equal(t, "stale subset", string(existing))
				out, err := os.ReadFile(filepath.Join(library, "03.subset.ass"))
				require.NoError(t, err)
				require.Contains(t, string(out), "[Fonts]")
			}
			require.Contains(t, detail, "缺少 6 款字体")
			require.Contains(t, detail, "Missing A：01.sc.ass、01.tc.ssa、02.ass")
			require.Contains(t, detail, "Missing A（粗体）：01.tc.ssa")
			require.Contains(t, detail, "Missing A（粗斜体）：01.tc.ssa")
			require.Contains(t, detail, "Missing A（半粗体）：01.tc.ssa")
			require.Contains(t, detail, "Missing A（字重 650，斜体）：01.tc.ssa")
			require.Contains(t, detail, "Missing B（斜体）：01.sc.ass")
			require.NotContains(t, detail, "Unselected")
		})
	}
}
