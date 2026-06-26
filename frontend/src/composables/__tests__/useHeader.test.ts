import { describe, it, expect, beforeEach } from 'vitest'
import { useHeader } from '../useHeader'

describe('useHeader Composable', () => {
  let header: ReturnType<typeof useHeader>

  beforeEach(() => {
    header = useHeader()
    // Reset state for predictable test conditions
    header.isHeaderVisible.value = true
    header.lastScrollTop.value = 0
  })

  it('should initialize with defaults', () => {
    expect(header.isHeaderVisible.value).toBe(true)
    expect(header.lastScrollTop.value).toBe(0)
  })

  it('should allow modifying states', () => {
    header.isHeaderVisible.value = false
    header.lastScrollTop.value = 100

    expect(header.isHeaderVisible.value).toBe(false)
    expect(header.lastScrollTop.value).toBe(100)
  })

  it('should share state globally across separate calls', () => {
    const header2 = useHeader()
    header.isHeaderVisible.value = false
    header.lastScrollTop.value = 250

    expect(header2.isHeaderVisible.value).toBe(false)
    expect(header2.lastScrollTop.value).toBe(250)
  })
})
