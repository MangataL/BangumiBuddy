package mikan

import (
	"strings"
	"testing"
	"time"

	"github.com/MangataL/BangumiBuddy/internal/discovery"
	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultTestBaseURL = normalizeBaseURL(discovery.DefaultMikanHost)

func TestParseSearchDocument(t *testing.T) {
	doc := strings.NewReader(`
		<html><body>
			<div class="js-search-results">
				<a href="/Home/Bangumi/681">
					<img src="/images/dora.jpg" alt="哆啦A梦">
					<span>哆啦A梦</span>
				</a>
				<table>
					<tr>
						<td><a href="/Home/Episode/abc">[字幕组] 剧场版 1080P</a></td>
						<td>1.4GB</td>
						<td>2026/07/03 12:30</td>
						<td><button data-clipboard-text="magnet:?xt=urn:btih:abc">复制磁连</button></td>
					</tr>
				</table>
			</div>
		</body></html>`)

	got, err := parseSearchDocument(doc, "https://mikanime.tv")

	require.NoError(t, err)
	require.Len(t, got.Bangumis, 1)
	assert.Equal(t, "681", got.Bangumis[0].MikanBangumiID)
	assert.Equal(t, "哆啦A梦", got.Bangumis[0].Name)
	assert.Equal(t, "https://mikanime.tv/images/dora.jpg", got.Bangumis[0].PosterURL)
	require.Len(t, got.Resources, 1)
	assert.Equal(t, "[字幕组] 剧场版 1080P", got.Resources[0].Title)
	assert.Equal(t, "magnet:?xt=urn:btih:abc", got.Resources[0].MagnetLink)
	assert.Equal(t, "1.4GB", got.Resources[0].Size)
	assert.Equal(t, downloader.DownloadTypeMovie, got.Resources[0].SuggestedDownloadType)
}

func TestParseWeekdayReportsMatch(t *testing.T) {
	weekday, ok := parseWeekday("放送日期：星期日")
	require.True(t, ok)
	assert.Equal(t, time.Sunday, weekday)

	weekday, ok = parseWeekday("没有放送日期")
	require.False(t, ok)
	assert.Equal(t, time.Sunday, weekday)
}

func TestParseSearchDocumentSkipsUntitledMagnetNodes(t *testing.T) {
	doc := strings.NewReader(`
		<html><body>
			<div data-clipboard-text="magnet:?xt=urn:btih:empty"></div>
			<table>
				<tr>
					<td>
						<a href="/Home/Episode/abc">[字幕组] 有标题的资源</a>
						<a data-clipboard-text="magnet:?xt=urn:btih:titled">复制磁连</a>
					</td>
					<td>500MB</td>
					<td>2026/07/03 12:30</td>
				</tr>
			</table>
		</body></html>`)

	got, err := parseSearchDocument(doc, defaultTestBaseURL)

	require.NoError(t, err)
	require.Len(t, got.Resources, 1)
	assert.Equal(t, "[字幕组] 有标题的资源", got.Resources[0].Title)
	assert.Equal(t, "magnet:?xt=urn:btih:titled", got.Resources[0].MagnetLink)
}

func TestParseSeasonDocumentParsesCurrentMikanScheduleMarkup(t *testing.T) {
	doc := strings.NewReader(`
		<html><body>
			<div class="sk-bangumi" data-dayofweek="5">
				<div class="row">星期五</div>
				<ul>
					<li>
						<span data-src="/images/Bangumi/200504/1df90634.jpg?width=400" class="js-expand_bangumi b-lazy" data-bangumiid="681"></span>
						<div class="num-node text-center">21</div>
						<div class="an-info">
							<div class="date-text">2026/07/04 更新</div>
							<a href="/Home/Bangumi/681" class="an-text" title="哆啦A梦">哆啦A梦</a>
						</div>
					</li>
					<li>
						<span data-src="/images/Bangumi/202607/empty.jpg" class="b-lazy greyout" data-bangumiid="4041"></span>
						<div class="an-info">
							<div class="date-text">此番组下暂无作品</div>
							<div class="date-text" title="占位番剧">占位番剧</div>
						</div>
					</li>
				</ul>
			</div>
		</body></html>`)

	got, err := parseSeasonDocument(doc, defaultTestBaseURL)

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "681", got[0].MikanBangumiID)
	assert.Equal(t, "哆啦A梦", got[0].Name)
	assert.Equal(t, time.Friday, got[0].Weekday)
	assert.Equal(t, "https://mikanani.me/images/Bangumi/200504/1df90634.jpg?width=400", got[0].PosterURL)
	assert.Equal(t, "4041", got[1].MikanBangumiID)
	assert.Equal(t, "占位番剧", got[1].Name)
	assert.Equal(t, time.Friday, got[1].Weekday)
	assert.Equal(t, "https://mikanani.me/images/Bangumi/202607/empty.jpg", got[1].PosterURL)
	assert.True(t, got[1].ReleaseGroupsKnownEmpty)
}

func TestParseSeasonDocumentIncludesGreyoutBangumisWithoutReleaseGroups(t *testing.T) {
	doc := strings.NewReader(`
		<html><body>
			<div class="sk-bangumi" data-dayofweek="3">
				<ul>
					<li>
						<span data-src="/images/Bangumi/202607/8752b276.jpg?width=400" class="b-lazy greyout" data-bangumiid="3988"></span>
						<div class="an-info">
							<div class="an-info-group">
								<div class="date-text">此番组下暂无作品</div>
								<div class="date-text" title="乡下大叔成为剑圣 第二季">乡下大叔成为剑圣 第二季</div>
							</div>
						</div>
					</li>
				</ul>
			</div>
		</body></html>`)

	got, err := parseSeasonDocument(doc, defaultTestBaseURL)

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "3988", got[0].MikanBangumiID)
	assert.Equal(t, "乡下大叔成为剑圣 第二季", got[0].Name)
	assert.Equal(t, time.Wednesday, got[0].Weekday)
	assert.Equal(t, "https://mikanani.me/images/Bangumi/202607/8752b276.jpg?width=400", got[0].PosterURL)
	assert.True(t, got[0].ReleaseGroupsKnownEmpty)
}

func TestParseDetailDocument(t *testing.T) {
	doc := strings.NewReader(`
		<html><body>
			<img src="/images/mikan-pic.png" alt="logo">
			<section class="bangumi-info">
				<div class="bangumi-poster" style="background-image: url('/images/Bangumi/200504/1df90634.jpg?width=400&height=560&format=webp');"></div>
				<h1>哆啦A梦</h1>
				<p>放送日期：星期五</p>
				<p>总集数：放送中</p>
				<p>放送开始：4/15/2005</p>
				<h2>概况介绍</h2>
				<p class="overview">一部长期放送的作品。</p>
			</section>
			<section class="subgroup" data-subgroupid="370">
				<a href="/Home/PublishGroup/370">夜莺工作室</a>
				<span>6/20/2026</span>
				<table>
					<tr>
						<td><a href="/Home/Episode/abc">[夜莺工作室] Doraemon 918</a></td>
						<td>455.5MB</td>
						<td>2026/06/20 13:28</td>
						<td><button data-clipboard-text="magnet:?xt=urn:btih:abc">复制磁连</button></td>
					</tr>
				</table>
			</section>
			<section class="subgroup">
				<a href="/Home/Bangumi/681?subgroupid=371">哆啦字幕组</a>
				<span>5/12/2026</span>
			</section>
		</body></html>`)

	got, err := parseDetailDocument(doc, "681", defaultTestBaseURL)

	require.NoError(t, err)
	assert.Equal(t, "681", got.MikanBangumiID)
	assert.Equal(t, "哆啦A梦", got.Name)
	assert.Equal(t, "https://mikanani.me/images/Bangumi/200504/1df90634.jpg?width=400&height=560&format=webp", got.PosterURL)
	assert.Equal(t, "4/15/2005", got.AirStartDate)
	assert.Equal(t, "放送中", got.EpisodeTotalText)
	assert.Equal(t, "一部长期放送的作品。", got.Overview)
	require.Len(t, got.ReleaseGroups, 2)
	assert.Equal(t, "370", got.ReleaseGroups[0].MikanSubgroupID)
	assert.Equal(t, "夜莺工作室", got.ReleaseGroups[0].Name)
	assert.Equal(t, "371", got.ReleaseGroups[1].MikanSubgroupID)
	require.Len(t, got.Resources, 1)
	assert.Equal(t, "magnet:?xt=urn:btih:abc", got.Resources[0].MagnetLink)
	assert.Equal(t, "370", got.Resources[0].MikanSubgroupID)
}

func TestParseDetailDocumentAssociatesSiblingEpisodeTableWithReleaseGroup(t *testing.T) {
	doc := strings.NewReader(`
		<html><body>
			<h1>欺诈游戏</h1>
			<ul>
				<li class="leftbar-item">
					<a class="subgroup-name subgroup-201" data-anchor="#201">Kirara Fantasia</a>
					<span class="date">7/4/2026</span>
				</li>
			</ul>
			<div class="subgroup-text" id="201">
				<div>
					Kirara Fantasia
					<a class="mikan-rss" href="/RSS/Bangumi?bangumiId=999&subgroupid=201">RSS</a>
				</div>
			</div>
			<div class="episode-table">
				<table>
					<tr>
						<td><a href="/Home/Episode/1">[Kirara Fantasia] LIAR GAME - 14</a></td>
						<td><button data-clipboard-text="magnet:?xt=urn:btih:abc">复制磁连</button></td>
					</tr>
				</table>
			</div>
		</body></html>`)

	got, err := parseDetailDocument(doc, "999", defaultTestBaseURL)

	require.NoError(t, err)
	require.Len(t, got.ReleaseGroups, 1)
	require.Len(t, got.Resources, 1)
	assert.Equal(t, "201", got.Resources[0].MikanSubgroupID)
}

func TestParseDetailDocumentReadsCurrentMikanOverviewMarkup(t *testing.T) {
	doc := strings.NewReader(`
		<html><body>
			<h1>哆啦A梦</h1>
			<div class="central-container">
				<div class="header2">
					<span class="header2-text">概况介绍</span>
				</div>
				<p class="header2-desc">这里才是真正的概况内容。</p>
				<div class="header2">
					<span class="header2-text">字幕组列表</span>
				</div>
			</div>
		</body></html>`)

	got, err := parseDetailDocument(doc, "681", defaultTestBaseURL)

	require.NoError(t, err)
	assert.Equal(t, "这里才是真正的概况内容。", got.Overview)
}

func TestParseDetailDocumentKeepsOverviewSeparateFromResources(t *testing.T) {
	doc := strings.NewReader(`
		<html><body>
			<h1>欺诈游戏</h1>
			<section class="bangumi-info">
				<h2>概况介绍</h2>
				<div class="info-box">
					<p>真正的简介第一段。</p>
					<section class="subgroup" data-subgroupid="101">
						<a href="/Home/PublishGroup/101">Kirara Fantasia</a>
						<table>
							<tr>
								<td><a href="/Home/Episode/1">[Kirara Fantasia] LIAR GAME - 14</a></td>
								<td><button data-clipboard-text="magnet:?xt=urn:btih:abc">复制磁连</button></td>
							</tr>
						</table>
					</section>
				</div>
			</section>
		</body></html>`)

	got, err := parseDetailDocument(doc, "999", defaultTestBaseURL)

	require.NoError(t, err)
	assert.Equal(t, "真正的简介第一段。", got.Overview)
	assert.NotContains(t, got.Overview, "LIAR GAME")
	assert.NotContains(t, got.Overview, "Kirara Fantasia")
}

func TestParseDetailDocumentDeduplicatesReleaseGroupsByName(t *testing.T) {
	doc := strings.NewReader(`
		<html><body>
			<h1>欺诈游戏</h1>
			<section class="subgroup" data-subgroupid="101">
				<a href="/Home/PublishGroup/101">Kirara Fantasia</a>
				<a href="/Home/Bangumi/999?subgroupid=201">Kirara Fantasia</a>
			</section>
			<section class="subgroup" data-subgroupid="102">
				<a href="/Home/PublishGroup/102">LoliHouse</a>
			</section>
			<section class="subgroup" data-subgroupid="202">
				<a href="/Home/Bangumi/999?subgroupid=202">LoliHouse</a>
			</section>
			<section class="subgroup" data-subgroupid="103">
				<a href="/Home/PublishGroup/103">喵萌奶茶屋</a>
			</section>
			<section class="subgroup" data-subgroupid="203">
				<a href="/Home/Bangumi/999?subgroupid=203">喵萌奶茶屋</a>
			</section>
		</body></html>`)

	got, err := parseDetailDocument(doc, "999", defaultTestBaseURL)

	require.NoError(t, err)
	require.Len(t, got.ReleaseGroups, 3)
	assert.Equal(t, "Kirara Fantasia", got.ReleaseGroups[0].Name)
	assert.Equal(t, "LoliHouse", got.ReleaseGroups[1].Name)
	assert.Equal(t, "喵萌奶茶屋", got.ReleaseGroups[2].Name)
}

func TestParseDetailDocumentUsesRSSSubgroupIDForReleaseGroup(t *testing.T) {
	doc := strings.NewReader(`
		<html><body>
			<h1>欺诈游戏</h1>
			<div class="subgroup-text" id="201">
				<div>
					Kirara Fantasia
					<a href="/Home/PublishGroup/101">Kirara Fantasia</a>
					<a class="mikan-rss" href="/RSS/Bangumi?bangumiId=999&subgroupid=201">RSS</a>
				</div>
			</div>
			<div class="episode-table">
				<table><tbody></tbody></table>
			</div>
		</body></html>`)

	got, err := parseDetailDocument(doc, "999", defaultTestBaseURL)

	require.NoError(t, err)
	require.Len(t, got.ReleaseGroups, 1)
	assert.Equal(t, "201", got.ReleaseGroups[0].MikanSubgroupID)
	assert.Equal(t, "Kirara Fantasia", got.ReleaseGroups[0].Name)
}
