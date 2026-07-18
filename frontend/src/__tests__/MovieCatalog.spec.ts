import { describe, expect, it } from 'vitest'

import { movieCatalog } from '@/features/movies/catalog'

describe('movie catalog', () => {
  it('contains ten Thai movies with unique titles and local posters', () => {
    expect(movieCatalog).toHaveLength(10)
    expect(new Set(movieCatalog.map((movie) => movie.title)).size).toBe(10)

    expect(movieCatalog.map((movie) => movie.title)).toEqual(
      expect.arrayContaining(['หลานม่า', 'ธี่หยด 2', 'วิมานหนาม', 'ฉลาดเกมส์โกง']),
    )

    for (const movie of movieCatalog) {
      expect(movie.englishTitle).toBeTruthy()
      expect(movie.genres.length).toBeGreaterThan(0)
      expect(movie.poster).toMatch(/\.webp$/)
      expect(movie.description.length).toBeGreaterThan(30)
      expect(movie.sourceUrl).toMatch(/^https:\/\//)
    }
  })
})
