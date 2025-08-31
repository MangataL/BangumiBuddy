import { http } from "./index";

// TMDB 元数据接口
export interface Meta {
  chineseName: string;
  year: string;
  tmdbID: number;
  season: number;
  episodeTotalNum: number;
  airWeekday: number | null;
  posterURL: string;
  backdropURL: string;
  overview: string;
  genres: string;
}

const metaAPI = {
  // 搜索电视剧（番剧）
  searchTVs: async (name: string): Promise<Meta[]> => {
    return http.get(`/meta/tvs?name=${encodeURIComponent(name)}`);
  },

  // 搜索电影（剧场版）
  searchMovies: async (name: string): Promise<Meta[]> => {
    return http.get(`/meta/movies?name=${encodeURIComponent(name)}`);
  },

  // 根据TMDB ID获取电视剧元数据
  getTVMeta: async (tmdbID: number): Promise<Meta> => {
    return http.get(`/meta/tv/${tmdbID}`);
  },

  // 根据TMDB ID获取电影元数据
  getMovieMeta: async (tmdbID: number): Promise<Meta> => {
    return http.get(`/meta/movie/${tmdbID}`);
  },
};

export default metaAPI;
