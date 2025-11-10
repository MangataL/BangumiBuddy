package scrape

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)



func TestFileExists(t *testing.T) {
	// 测试存在的文件
	assert.True(t, fileExists("scraper.go"))

	// 测试不存在的文件
	assert.False(t, fileExists("nonexistent_file.txt"))
}

func Test_parseNFO(t *testing.T) {
	scraper := &Scraper{}
	plot := `真昼遇见了神秘的女性菊并与她逐渐打好关系，然而更却看穿了她的真面目。菊虽然很犹豫，还是跟真昼说明了真实情况……后来，更听荠说了菊的过去，便有了更多疑虑。`

	nfoData, err := scraper.parseNFO("./testdata/彻夜之歌 S02E02.nfo")

	require.NoError(t, err)
	assert.Equal(t, "好想见到你", nfoData.title)
	assert.Equal(t, 2, nfoData.season)
	assert.Equal(t, 2, nfoData.episode)
	assert.Equal(t, "/video/bangumi/彻夜之歌/Season 2/彻夜之歌 S02E02-thumb.jpg", nfoData.posterPath)
	assert.Equal(t, plot, nfoData.plot)
}