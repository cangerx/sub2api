import { describe, expect, it } from 'vitest'
import { escapeCsvCell } from '../csv'

describe('escapeCsvCell', () => {
  it('escapes formula-like values', () => {
    expect(escapeCsvCell('=SUM(1,1)')).toBe("\"'=SUM(1,1)\"")
    expect(escapeCsvCell('+1+2')).toBe("\"'+1+2\"")
    expect(escapeCsvCell('-1+2')).toBe("\"'-1+2\"")
    expect(escapeCsvCell('@cmd')).toBe("\"'@cmd\"")
  })

  it('escapes quotes and commas', () => {
    expect(escapeCsvCell('a,"b"')).toBe('"a,""b"""')
  })
})
