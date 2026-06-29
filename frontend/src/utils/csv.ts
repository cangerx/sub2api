export function escapeCsvCell(value: unknown): string {
  if (value === null || value === undefined) return ''
  const text = String(value)
  const escaped = text.replace(/"/g, '""')
  const prefixed = /^[=+\-@]/.test(escaped) ? `'${escaped}` : escaped
  if (/[",\n\r]/.test(prefixed) || prefixed.startsWith("'")) {
    return `"${prefixed}"`
  }
  return prefixed
}
