import { describe, expect, it } from 'vitest'
import { clampAppearance, DEFAULTS } from './appearance'

describe('clampAppearance', () => {
  it('falls back to the defaults for missing or corrupt storage', () => {
    expect(clampAppearance(null)).toEqual(DEFAULTS)
    expect(clampAppearance('nope')).toEqual(DEFAULTS)
    expect(clampAppearance({ opacity: 'x', blur: NaN, liquid: 1, calm: 'yes' })).toEqual(DEFAULTS)
  })

  it('keeps every value inside the range the UI offers', () => {
    expect(clampAppearance({ opacity: 5, blur: -3 })).toMatchObject({ opacity: 20, blur: 0 })
    expect(clampAppearance({ opacity: 900, blur: 99 })).toMatchObject({ opacity: 200, blur: 40 })
  })

  it('keeps valid saved values', () => {
    expect(clampAppearance({ opacity: 60, blur: 12, liquid: false, calm: true })).toEqual({ opacity: 60, blur: 12, liquid: false, calm: true })
  })
})
